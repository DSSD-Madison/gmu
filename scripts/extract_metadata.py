import boto3
import fitz  # PyMuPDF
import json
from io import BytesIO

# === Configuration ===
BUCKET = "allianceforpeacebuilding-org"

PDF_KEYS = [
    "0010.pdf",
    "0011.pdf",
    "0022.pdf",
    "1860.pdf",
    "0027.pdf"
]


# === AWS Clients ===
s3 = boto3.client("s3")
bedrock = boto3.client("bedrock-runtime", region_name="us-east-1")

# === Helpers ===
def extract_text_until_metadata(pdf_bytes, max_pages=15):
    doc = fitz.open(stream=pdf_bytes, filetype="pdf")
    text = ""
    for i in range(min(max_pages, len(doc))):
        page = doc[i].get_text()
        text += page + "\n"
        if any(word in page.lower() for word in ["abstract", "introduction", "region", "keyword", "author", "category"]):
            break
    return text.strip()

def build_prompt(text):
    return f"""
You are an assistant extracting structured metadata from an academic policy document.

Return only a valid JSON object with these fields:
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

def call_claude(prompt):
    body = {
        "prompt": f"\n\nHuman: {prompt}\n\nAssistant:",
        "max_tokens_to_sample": 512,
        "temperature": 0.3,
        "top_k": 250,
        "top_p": 1.0,
        "stop_sequences": ["\n\nHuman:"]
    }
    response = bedrock.invoke_model(
        modelId = "anthropic.claude-3-5-haiku-20241022-v1:0",
        body=json.dumps(body),
        contentType="application/json",
        accept="application/json"
    )
    result = json.loads(response["body"].read())
    return result["completion"].strip()

def clip_list(values, max_items=10):
    return list(set(values))[:max_items] if isinstance(values, list) else []

# === Main ===
def main():
    for key in PDF_KEYS:
        print(f"\n📄 Processing {key}...")

        try:
            obj = s3.get_object(Bucket=BUCKET, Key=key)
            pdf_bytes = obj["Body"].read()
            text = extract_text_until_metadata(pdf_bytes)

            prompt = build_prompt(text)
            raw_result = call_claude(prompt)

            try:
                metadata = json.loads(raw_result)
            except json.JSONDecodeError:
                print(f"❌ Invalid JSON from Claude for {key}")
                continue

            final = {
                "s3_file": f"s3://{BUCKET}/{key}",
                "title": metadata.get("title"),
                "abstract": metadata.get("abstract"),
                "category": (metadata.get("category") or "")[:100],
                "publish_date": metadata.get("publish_date"),
                "source": "bucket",
                "region_name": clip_list(metadata.get("region_name", [])),
                "keyword_name": clip_list(metadata.get("keyword_name", [])),
                "author_name": clip_list(metadata.get("author_name", []))
            }

            print("✅ Metadata:\n", json.dumps(final, indent=2))

        except Exception as e:
            print(f"❌ Error processing {key}: {e}")

if __name__ == "__main__":
    main()
