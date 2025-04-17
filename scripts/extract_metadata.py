import boto3
import fitz  # PyMuPDF
import json
import re
from io import BytesIO

# === AWS Clients ===
s3 = boto3.client("s3")
bedrock = boto3.client("bedrock-runtime", region_name="us-east-1")
flag = False

# === Token + Cost Estimation ===
def estimate_tokens(text: str) -> int:
    return int(len(text.split()) * 0.75)

def estimate_cost(input_tokens: int, output_tokens: int = 200) -> float:
    return round((input_tokens / 1000 * 0.00025) + (output_tokens / 1000 * 0.00125), 6)

# === PDF Text Cleaning ===
def clean_text(text: str) -> str:
    lines = text.splitlines()
    cleaned = []
    for line in lines:
        if re.fullmatch(r"\s*\d+\s*", line): continue  # Remove standalone page numbers
        if len(line.strip()) < 6: continue  # Likely headers/footers
        cleaned.append(line.strip())
    return "\n".join(cleaned)

def extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10):
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    text = ""
    for i in range(min(max_pages, len(doc))):
        text += doc[i].get_text() + "\n"
    cleaned = clean_text(text)
    input_tokens = estimate_tokens(cleaned)
    print(f"📄 Estimated input tokens: {input_tokens}")
    print(f"💸 Estimated cost for this doc (input only): ${estimate_cost(input_tokens, 0):.4f}")
    return cleaned, input_tokens

# === Prompt + Claude Call ===
def build_prompt(text: str) -> str:
    return f"""
You are an assistant extracting structured metadata from an academic policy document.

Return only a valid JSON object with the following fields:
- "title" (string, required)
- "abstract" (string)
- "category" (string, max 100 characters): e.g., article, research paper, etc.
- "publish_date" (date)
- "source" (string, max 255 characters): use "bucket" as a placeholder
- "region_name" (array of unique strings, required, max 10)
- "keyword_name" (array of unique strings, required, max 10)
- "author_name" (array of unique strings, required, max 10)
- "category_name" (array of unique strings, required, max 10 - like "article", "research paper", etc.)

Do not explain. Do not say "Here is the JSON". Do not use Markdown. Just return the JSON object.
TEXT:
{text}
"""

def call_claude(prompt: str) -> tuple[str, int]:
    body = {
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": 512,
        "temperature": 0.3,
        "top_p": 1.0,
        "anthropic_version": "bedrock-2023-05-31"
    }

    response = bedrock.invoke_model(
        modelId="anthropic.claude-3-haiku-20240307-v1:0",
        body=json.dumps(body),
        contentType="application/json",
        accept="application/json"
    )

    result = json.loads(response["body"].read())
    output_text = result["content"][0]["text"].strip()
    output_tokens = estimate_tokens(output_text)
    print(f"📝 Estimated output tokens: {output_tokens}")
    return output_text, output_tokens

# === JSON Cleaner ===
def extract_first_json(text: str):
    match = re.search(r"\{.*\}", text, re.DOTALL)
    if match:
        try:
            return json.loads(match.group(0))
        except json.JSONDecodeError:
            return None
    return None

def clip_list(values, max_items=10):
    return list(set(values))[:max_items] if isinstance(values, list) else []

# === MAIN ===
results = []
total_cost = 0.0

buckets = s3.list_buckets()["Buckets"]

for bucket in buckets:
    bucket_name = bucket["Name"]
    print(f"\n🪣 Checking bucket: {bucket_name}")
    try:
        paginator = s3.get_paginator('list_objects_v2')
        page_iterator = paginator.paginate(Bucket=bucket_name)

        pdf_keys = []
        for page in page_iterator:
            if "Contents" in page:
                for obj in page["Contents"]:
                    key = obj["Key"]
                    if key.lower().endswith(".pdf"):
                        pdf_keys.append(key)

        if not pdf_keys:
            print("📭 No PDFs found in this bucket.")
            continue

        for key in pdf_keys[:5]:  # Max 5 PDFs per bucket
            print(f"\n📄 Processing {key}...")
            try:
                obj = s3.get_object(Bucket=bucket_name, Key=key)
                pdf_bytes = obj["Body"].read()

                text, input_tokens = extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10)
                if not flag:
                    print("🔎 Text:")
                    print(text)
                    flag = True
                    
                prompt = build_prompt(text)
                response, output_tokens = call_claude(prompt)

                print("🔎 Claude raw response:")
                print(response)

                metadata = extract_first_json(response)
                if not metadata:
                    raise ValueError("❌ Failed to parse Claude response.")

                cost = estimate_cost(input_tokens, output_tokens)
                total_cost += cost

                results.append({
                    "s3_file": f"s3://{bucket_name}/{key}",
                    "title": metadata.get("title"),
                    "abstract": metadata.get("abstract"),
                    "category": (metadata.get("category") or "")[:100],
                    "publish_date": metadata.get("publish_date"),
                    "source": "bucket",
                    "region_name": clip_list(metadata.get("region_name", [])),
                    "keyword_name": clip_list(metadata.get("keyword_name", [])),
                    "author_name": clip_list(metadata.get("author_name", [])),
                    "estimated_cost": f"${cost:.4f}"
                })

            except Exception as e:
                results.append({
                    "s3_file": f"s3://{bucket_name}/{key}",
                    "error": str(e),
                    "claude_output": response if 'response' in locals() else "No response"
                })

    except Exception as bucket_error:
        print(f"⚠️ Error accessing bucket {bucket_name}: {bucket_error}")

print(f"\n💰 Total estimated cost for all processed documents: ${total_cost:.4f}")

# Save results
with open("metadata_results.json", "w") as f:
    json.dump(results, f, indent=2)

print("\n✅ Done processing all buckets.")
