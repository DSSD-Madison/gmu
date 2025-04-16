import boto3
import fitz  # PyMuPDF
import json
import re
from io import BytesIO

# === Config ===
BUCKET = "allianceforpeacebuilding-org"
PDF_KEYS = [
    "0010.pdf", "0011.pdf", "0022.pdf", "1860.pdf", "0027.pdf"
]

s3 = boto3.client("s3")
bedrock = boto3.client("bedrock-runtime", region_name="us-east-1")

# === Helpers ===
def estimate_tokens(text: str) -> int:
    return int(len(text.split()) * 0.75)

def estimate_cost(input_tokens: int, output_tokens: int = 200) -> float:
    return round((input_tokens / 1000 * 0.00025) + (output_tokens / 1000 * 0.001), 6)

def clean_text(text: str) -> str:
    lines = text.splitlines()
    cleaned = []
    for line in lines:
        if re.fullmatch(r"\s*\d+\s*", line): continue
        if len(line.strip()) < 6: continue
        cleaned.append(line.strip())
    return "\n".join(cleaned)

def extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10):
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    text = ""
    for i in range(min(max_pages, len(doc))):
        text += doc[i].get_text() + "\n"
    cleaned = clean_text(text)
    tokens = estimate_tokens(cleaned)
    print(f"📄 Estimated input tokens: {tokens}")
    print(f"💸 Estimated cost for this doc: ${estimate_cost(tokens):.4f}")
    return cleaned, tokens

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

Only return the JSON — no explanations or extra text.
TEXT:
{text}
"""

def call_claude(prompt: str) -> str:
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
    return result["content"][0]["text"].strip()

def clip_list(values, max_items=10):
    return list(set(values))[:max_items] if isinstance(values, list) else []

# === Main ===
results = []
total_cost = 0.0

for key in PDF_KEYS:
    print(f"\n📄 Processing {key}...")
    try:
        obj = s3.get_object(Bucket=BUCKET, Key=key)
        pdf_bytes = obj["Body"].read()

        text, tokens = extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10)
        prompt = build_prompt(text)
        response = call_claude(prompt)

        try:
            metadata = json.loads(response)
        except json.JSONDecodeError:
            print(f"❌ Failed to parse Claude response.")
            results.append({"s3_file": f"s3://{BUCKET}/{key}", "error": "Invalid JSON"})
            continue

        cost = estimate_cost(tokens)
        total_cost += cost

        results.append({
            "s3_file": f"s3://{BUCKET}/{key}",
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
        results.append({"s3_file": f"s3://{BUCKET}/{key}", "error": str(e)})

print(f"\n💰 Total estimated cost for {len(results)} documents: ${total_cost:.4f}")
