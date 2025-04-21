import boto3
import fitz  # PyMuPDF
import json
import re
import psycopg2
from psycopg2.extras import RealDictCursor
import os
from io import BytesIO
from dotenv import load_dotenv
from pathlib import Path
from docx import Document
import contextlib
import sys

load_dotenv()

import sys

log_path = "scripts/metadata_run.log"
log_file = open(log_path, "a")

class Tee:
    def __init__(self, *streams):
        self.streams = streams
    def write(self, data):
        for s in self.streams:
            s.write(data)
            s.flush()
    def flush(self):
        for s in self.streams:
            s.flush()

# Send all `print()` output to both terminal and file
sys.stdout = Tee(sys.stdout, log_file)
sys.stderr = Tee(sys.stderr, log_file)
    


# === AWS Clients ===
s3 = boto3.client("s3")
bedrock = boto3.client("bedrock-runtime", region_name="us-east-1")
include_buckets = {"ipinst-org", "allianceforpeacebuilding-org", "better-evidence-repository", "video-kendra-testing"}

# === DB Connection ===
conn = psycopg2.connect(
    host=os.getenv("PROD_HOST"),
    database=os.getenv("PROD_NAME"),
    user=os.getenv("PROD_USER"),
    password=os.getenv("PROD_PASSWORD")
)
cursor = conn.cursor()

# === Token + Cost Estimation ===
# only used if information is not in claude
def estimate_tokens(text: str) -> int:
    return int(len(text.split()) * 0.75)

def estimate_cost(input_tokens: int, output_tokens: int = 200) -> float:
    return round((input_tokens / 1000 * 0.00025) + (output_tokens / 1000 * 0.00125), 6)

# === Text Cleaning ===
def clean_text(text: str) -> str:
    lines = text.splitlines()
    cleaned = []
    for line in lines:
        if re.fullmatch(r"\s*\d+\s*", line): continue
        if len(line.strip()) < 6: continue
        cleaned.append(line.strip())
    return "\n".join(cleaned)

# === PDF Extraction ===
def extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10, key=None):
    try:
        doc = fitz.open(stream=pdf_bytes, filetype="pdf")
        text = ""
        for i in range(min(max_pages, len(doc))):
            text += doc[i].get_text() + "\n"
        cleaned = clean_text(text)
        return cleaned
    except Exception as e:
        print(f"❌ Failed to open PDF {key or '[unknown key]'}: {e}")
        return None

# === DOCX Extraction ===
def extract_text_from_docx(docx_bytes, max_chars=5000, key=None):
    try:
        with BytesIO(docx_bytes) as f:
            with open(os.devnull, "w") as devnull:
                doc = Document(f)
            text = "\n".join(p.text for p in doc.paragraphs)
            trimmed = text[:max_chars]  
            cleaned = clean_text(trimmed)
            return cleaned
    except Exception as e:
        print(f"❌ Failed to open DOCX {key or '[unknown key]'}: {e}")
        return None


# === Prompt Builder ===
def build_prompt(text: str) -> str:
    categories = [
        "Article", "Background Paper", "Blog Post", "Book", "Brief", "Case Study", "Dataset", "Educational Guide",
        "Evaluation", "Fact Sheet", "Government Report", "Organizational Study", "Paper", "Policy Brief", "Policy Paper",
        "Project Evaluation", "Project Evaluations", "Report", "Working Paper"
    ]

    regions = [
        "Afghanistan", "Africa", "Albania", "Angola", "Asia", "Bangladesh", "Benin", "Bosnia And Herzegovina",
        "Burkina Faso", "Burundi", "Cambodia", "Caribean", "Central African Republic Car", "Central America",
        "Democratic Republic Of Congo Drc", "Democratic Republic Of Congo Drc / Central African Republic Car",
        "Ecuador", "Egypt", "El Salvador", "Ethiopia", "Europe", "Georgia", "Ghana", "Global", "Guatemala",
        "Guinea", "Indonesia", "Indo Pacific", "Iraq", "Israel", "Jamaica", "Jerusalem", "Jordan", "Kenya",
        "Kosovo", "Kyrgyzstan", "Latin America", "Lebanon", "Liberia", "Macedonia", "Madagascar", "Mali",
        "Middle East", "Morocco", "Myanmar", "Nepal", "Nigeria", "North America", "Oceana", "Oceania",
        "Pakistan", "Papua New Guinea", "Peru", "Philippines", "Russia", "Rwanda", "Senegal", "Somalia",
        "South Africa", "South America", "South Sudan", "Sri Lanka", "Sudan", "Tajikistan", "Tanzania",
        "Timor Leste", "Uganda", "Ukraine", "West Bank", "Yemen", "Zambia", "Zimbabwe"
    ]

    with open("scripts/full_keyword_list.txt") as f:
        keywords = [line.strip() for line in f if line.strip()]

    return f"""
You are an assistant extracting structured metadata from an academic policy document.

Prefer to select from the following known lists if relevant:

CATEGORIES:
{', '.join(categories)}

REGIONS:
{', '.join(regions)}

KEYWORDS:
{', '.join(keywords)}

Normalize all values by removing dashes and replacing them with spaces. For example, "conflict-resolution" becomes "conflict resolution".
All string fields must be enclosed in double quotes.
Double quotes inside any string **must** be escaped using a backslash: use `\"`, never `”` or `“`. Do not escape any characters that are not quotes.
Do not include any preamble, explanation, commentary, or non-JSON output — just return the JSON.
Only generate regions that are widely known and well-represented in global datasets and literature.
Focus on fully recognized countries or broad, commonly referenced geographic areas (e.g., Central America, Southeast Asia).
Avoid small, obscure, or low-data regions (e.g., Kurdistan, Upper Nile, Northern Ireland), as these are less likely to be relevant or supported by sufficient context.

Return only a valid JSON object with the following fields:
- "title" (string, required)
- "abstract" (string, max 1800 characters)
- "publish_date" (date)
- "source" (string, max 255 characters)
- "region_name" (array of unique strings, required, max 10)
- "keyword_name" (array of unique strings, required, max 10)
- "author_name" (array of unique strings, required, max 10)
- "category_name" (array of unique strings, required, max 10)

Only return a JSON object, and nothing else. Do not include any explanation or header — even in French. Do not say "Here is the JSON" or anything similar.
Return a JSON object that passes strict validation (e.g., `json.loads(...)` in Python). It cannot be an incomplete JSON object.
TEXT:
{text}
"""

# === Claude API Call ===
def call_claude(prompt: str) -> tuple[str, int, int]:
    body = {
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": 3500,
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
    raw_result = response["body"].read()
    result = json.loads(raw_result)

    output_text = result["content"][0]["text"].strip()

    usage = result.get("usage", {})
    input_tokens = usage.get("input_tokens", estimate_tokens(prompt))
    output_tokens = usage.get("output_tokens", estimate_tokens(output_text))

    print(f"📊 Input tokens: {input_tokens}")
    print(f"📝 Output tokens: {output_tokens}")
    print(f"💸 Cost: ${estimate_cost(input_tokens, output_tokens):.6f}")

    return output_text, input_tokens, output_tokens


def extract_first_json(text: str, key: str = "[unknown]"):
    try:
        # Try parsing the JSON (might be a dict or a string)
        parsed = json.loads(text)

        # If Claude returned a JSON *string* of a JSON object, decode again
        if isinstance(parsed, str):
            parsed = json.loads(parsed)

        # If we got a dict, it's good to go
        if isinstance(parsed, dict):
            return parsed

        raise ValueError("Parsed JSON is not a dict")

    except Exception as e:
        print(f"❌ JSON parse error for {key}: {e}")
        print(f"🔍 Raw JSON snippet:\n{text[:500]}")
        return None



def clip_list(values, max_items=10):
    return list(dict.fromkeys(values))[:max_items] if isinstance(values, list) else []

def get_or_create_name(table, name):
    cursor.execute(f"SELECT id FROM {table} WHERE LOWER(name) = LOWER(%s)", (name,))
    result = cursor.fetchone()
    if result:
        return result[0]
    cursor.execute(f"INSERT INTO {table} (name) VALUES (%s) RETURNING id", (name,))
    return cursor.fetchone()[0]

def normalize_date(date_str):
    """Normalize partial date strings to YYYY-MM-DD, or return None if invalid."""
    if not date_str or not isinstance(date_str, str):
        return None
    if re.fullmatch(r"\d{4}", date_str):
        return f"{date_str}-01-01"
    if re.fullmatch(r"\d{4}-\d{2}", date_str):
        return f"{date_str}-01"
    if re.fullmatch(r"\d{4}-\d{2}-\d{2}", date_str):
        return date_str
    return None  # Reject anything else

def insert_doc_metadata(doc_id, metadata):
    cursor.execute("SELECT title, abstract, publish_date, source FROM documents WHERE id = %s", (doc_id,))
    current = cursor.fetchone()
    title, abstract, publish_date, source = current

    new_title = metadata.get("title") if title == "Untitled" else title
    new_abstract = metadata.get("abstract") if not abstract else abstract
    proposed_date = normalize_date(metadata.get("publish_date"))
    new_publish_date = proposed_date if not publish_date and proposed_date else publish_date
    new_source = metadata.get("source") if source in (None, "", "Unknown") else source

    cursor.execute("""
        UPDATE documents 
        SET title = %s, abstract = %s, publish_date = %s, source = %s 
        WHERE id = %s
    """, (new_title, new_abstract, new_publish_date, new_source, doc_id))

    tag_mappings = {
        "author_name": ("doc_authors", "authors", "author_id"),
        "keyword_name": ("doc_keywords", "keywords", "keyword_id"),
        "region_name": ("doc_regions", "regions", "region_id"),
        "category_name": ("doc_categories", "categories", "category_id")
    }

    for tag_type, (join_table, ref_table, id_field) in tag_mappings.items():
        if tag_type not in metadata:
            continue

        names = clip_list(metadata[tag_type])
        cursor.execute(f"""
            SELECT {ref_table}.name FROM {join_table}
            JOIN {ref_table} ON {ref_table}.id = {join_table}.{id_field}
            WHERE {join_table}.doc_id = %s
        """, (doc_id,))
        existing_names = {r[0].lower() for r in cursor.fetchall()}
        count = len(existing_names)

        for name in names:
            if count >= 10:
                break
            if name.lower() in existing_names:
                continue
            ref_id = get_or_create_name(ref_table, name)
            cursor.execute(
                f"INSERT INTO {join_table} (doc_id, {id_field}) VALUES (%s, %s)",
                (doc_id, ref_id)
            )
            count += 1

    conn.commit()


# === MAIN ===
print("⚙️ Starting metadata enrichment...")
output_log_path = Path("scripts/successful_metadata_updates.txt")
output_log_path.touch(exist_ok=True)
with output_log_path.open() as f:
    processed_uuids = {line.strip() for line in f if line.strip()}


buckets = s3.list_buckets()["Buckets"]
total_cost = 0.0
for bucket in buckets:
    bucket_name = bucket["Name"]
    if bucket_name not in include_buckets:
        print(f"⏭️ Skipping bucket: {bucket_name}")
        continue
    print(f"🪣 Checking bucket: {bucket_name}")
    try:
        paginator = s3.get_paginator('list_objects_v2')
        page_iterator = paginator.paginate(Bucket=bucket_name)

        file_keys = []
        for page in page_iterator:
            if "Contents" in page:
                for obj in page["Contents"]:
                    key = obj["Key"]
                    if key.lower().endswith(".pdf") or key.lower().endswith(".docx"):
                        file_keys.append(key)


        if not file_keys:
            print("📭 No supported files found in this bucket.")
            continue

        for key in file_keys:
            try:
                cursor.execute("SELECT id FROM documents WHERE s3_file = %s", (f"s3://{bucket_name}/{key}",))
                doc_row = cursor.fetchone()
                if not doc_row:
                    print(f"❌ No matching document in DB for this file: {key}")
                    continue
                doc_id = doc_row[0]
                if str(doc_id) in processed_uuids:
                    # print(f"⏭️ Skipping already-processed document: {doc_id}")
                    continue
                print(f"📄 Processing {key}...")
                print(f"🧾 Matched DB document UUID: {doc_id}")


                obj = s3.get_object(Bucket=bucket_name, Key=key)
                file_bytes = obj["Body"].read()

                if key.lower().endswith(".pdf"):
                    text = extract_text_first_n_pages_cleaned(file_bytes, max_pages=10, key=key)
                else:
                    text = extract_text_from_docx(file_bytes, key=key)
                if text is None:
                    continue

                prompt = build_prompt(text)
                response, input_tokens, output_tokens = call_claude(prompt)
                doc_cost = estimate_cost(input_tokens, output_tokens)
                total_cost += doc_cost
                metadata = extract_first_json(response, key)

                if metadata:
                    insert_doc_metadata(doc_id, metadata)
                    print("✅ Metadata updated.")
                    with output_log_path.open("a") as f:
                        f.write(f"{doc_id}\n")
                else:
                    print("❌ Failed to parse Claude response.")
                    # debug_path = Path("scripts/debug_claude_responses.txt")
                    # with debug_path.open("a") as debug_f:
                    #     debug_f.write(f"\n--- FAILED {doc_id} ({key}) ---\n{response}\n")
            except Exception as doc_err:
                print(f"❗ Error processing document {key}: {doc_err}")

    except Exception as e:
        print(f"⚠️ Error accessing bucket {bucket_name}: {e}")
print(f"\n💰 Total Claude usage cost: ${total_cost:.6f}")
conn.close()

print("\n✅ Done.")
