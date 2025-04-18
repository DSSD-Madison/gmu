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


load_dotenv()

# === AWS Clients ===
s3 = boto3.client("s3")
bedrock = boto3.client("bedrock-runtime", region_name="us-east-1")
include_buckets = {"ipinst-org", "allianceforpeacebuilding-org", "better-evidence-repository"}

# === DB Connection ===
conn = psycopg2.connect(
    host=os.getenv("DB_HOST"),
    database=os.getenv("DB_NAME"),
    user=os.getenv("DB_USER"),
    password=os.getenv("DB_PASSWORD")
)
cursor = conn.cursor()

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
    input_tokens = estimate_tokens(cleaned)
    print(f"📄 Estimated input tokens: {input_tokens}")
    print(f"💸 Estimated cost for this doc (input only): ${estimate_cost(input_tokens, 0):.4f}")
    return cleaned, input_tokens

def build_prompt(text: str) -> str:
    categories = [
        "article", "background paper", "blog post", "book", "brief", "case study", "dataset", "educational guide",
        "evaluation", "fact sheet", "government report", "organizational study", "paper", "policy brief", "policy paper",
        "project evaluation", "project evaluations", "report", "working paper"
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

Return only a valid JSON object with the following fields:
- "title" (string, required)
- "abstract" (string)
- "category" (string, max 100 characters): e.g., article, research paper, etc.
- "publish_date" (date)
- "source" (string, max 255 characters): use "bucket" as a placeholder
- "region_name" (array of unique strings, required, max 10)
- "keyword_name" (array of unique strings, required, max 10)
- "author_name" (array of unique strings, required, max 10)
- "category_name" (array of unique strings, required, max 10)

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

def extract_first_json(text: str):
    match = re.search(r"\{.*\}", text, re.DOTALL)
    if match:
        try:
            return json.loads(match.group(0))
        except json.JSONDecodeError:
            return None
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

def insert_doc_metadata(doc_id, metadata):
    cursor.execute("SELECT title, abstract, publish_date, source FROM documents WHERE id = %s", (doc_id,))
    current = cursor.fetchone()
    title, abstract, publish_date, source = current

    new_title = metadata.get("title") if title == "Untitled" else title
    new_abstract = metadata.get("abstract") if not abstract else abstract
    new_publish_date = metadata.get("publish_date") if not publish_date else publish_date
    new_source = metadata.get("source") if source in (None, "", "Unknown") else source

    cursor.execute("UPDATE documents SET title=%s, abstract=%s, publish_date=%s, source=%s WHERE id=%s",
                   (new_title, new_abstract, new_publish_date, new_source, doc_id))

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
output_log_path = Path("successful_metadata_updates.txt")
output_log_path.touch(exist_ok=True)  # Creates the file if it doesn't exist


buckets = s3.list_buckets()["Buckets"]

for bucket in buckets:
    bucket_name = bucket["Name"]
    if bucket_name not in include_buckets:
        print(f"⏭️ Skipping bucket: {bucket_name}")
        continue
    bucket_name = bucket["Name"]
    print(f"🪣 Checking bucket: {bucket_name}")
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
            print(f"📄 Processing {key}...")
            try:
                cursor.execute("SELECT id FROM documents WHERE s3_file = %s", (f"s3://{bucket_name}/{key}",))
                doc_row = cursor.fetchone()
                if not doc_row:
                    print("❌ No matching document in DB for this file (or it has has_duplicate = true).")
                    continue
                doc_id = doc_row[0]

                print(f"🧾 Matched DB document UUID: {doc_id}")

                obj = s3.get_object(Bucket=bucket_name, Key=key)
                pdf_bytes = obj["Body"].read()

                text, input_tokens = extract_text_first_n_pages_cleaned(pdf_bytes, max_pages=10)
                prompt = build_prompt(text)
                response, output_tokens = call_claude(prompt)
                metadata = extract_first_json(response)

                if metadata:
                    insert_doc_metadata(doc_id, metadata)
                    print("✅ Metadata updated.")
                    with output_log_path.open("a") as f:
                        f.write(f"{doc_id}\n")
                else:
                    print("❌ Failed to parse Claude response.")
            except Exception as doc_err:
                print(f"❗ Error processing document {key}: {doc_err}")

    except Exception as e:
        print(f"⚠️ Error accessing bucket {bucket_name}: {e}")

conn.close()
print("\n✅ Done.")
