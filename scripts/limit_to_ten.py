import os
import json
import psycopg2
import boto3
from pathlib import Path
from dotenv import load_dotenv
from extract_metadata import extract_text_first_n_pages_cleaned, call_claude, estimate_cost

load_dotenv()

# === Claude Prompt Helper ===
def build_pruning_prompt(tag_type: str, tags: list[str], text: str) -> str:
    return f'''
You are helping prune metadata tags for a document. You will be given:
1. The text content of a document.
2. A list of existing {tag_type} tags.

Your task is to select the **10 most relevant {tag_type} tags** from the provided list, based on the document content.

Return your answer as a **strict JSON object** in the following format:

{{ "{tag_type}": ["tag1", "tag2", ..., "tag10"] }}

⚠️ Guidelines:
- Return **exactly 10** tags.
- Do **not** invent new tags. Only use the ones provided.
- Return a valid JSON object with the key "{tag_type}".
- Do **not** include any explanation or non-JSON output.
- Escape all quotes properly (use `\"` for double quotes inside strings).

TAGS:
{json.dumps(tags, ensure_ascii=False, indent=2)}

DOCUMENT TEXT:
{text}
'''

# === DB + S3 Clients ===
conn = psycopg2.connect(
    host=os.getenv("PROD_HOST"),
    database=os.getenv("PROD_NAME"),
    user=os.getenv("PROD_USER"),
    password=os.getenv("PROD_PASSWORD")
)
cursor = conn.cursor()
s3 = boto3.client("s3")

# === Load docs with >10 tags ===
cursor.execute("""
    SELECT d.id, d.s3_file
    FROM documents d
    WHERE (
        (SELECT COUNT(*) FROM doc_keywords WHERE doc_id = d.id) > 10
        OR (SELECT COUNT(*) FROM doc_authors WHERE doc_id = d.id) > 10
        OR (SELECT COUNT(*) FROM doc_regions WHERE doc_id = d.id) > 10
        OR (SELECT COUNT(*) FROM doc_categories WHERE doc_id = d.id) > 10
    )
""")
documents = cursor.fetchall()
print("\n⚙️ Starting tag pruning...")
print(f"📄 Found {len(documents)} documents with >10 tags")

total_cost = 0.0
for doc_id, s3_path in documents:
    if not s3_path or not s3_path.startswith("s3://"):
        print(f"❌ Malformed s3_path: {s3_path}")
        continue

    bucket, key = s3_path.replace("s3://", "").split("/", 1)
    print(f"\n📄 Processing {key}...")
    print(f"🧾 Matched DB document UUID: {doc_id}")

    try:
        file_bytes = s3.get_object(Bucket=bucket, Key=key)["Body"].read()
        text = extract_text_first_n_pages_cleaned(file_bytes, max_pages=10, key=key)
        if not text:
            continue

        tag_tables = {
            "keyword_name": ("doc_keywords", "keywords", "keyword_id"),
            "author_name": ("doc_authors", "authors", "author_id"),
            "region_name": ("doc_regions", "regions", "region_id"),
            "category_name": ("doc_categories", "categories", "category_id")
        }

        for tag_type, (join_table, ref_table, id_field) in tag_tables.items():
            cursor.execute(f"""
                SELECT {ref_table}.id, {ref_table}.name
                FROM {join_table}
                JOIN {ref_table} ON {ref_table}.id = {join_table}.{id_field}
                WHERE {join_table}.doc_id = %s
            """, (doc_id,))
            rows = cursor.fetchall()
            if len(rows) <= 10:
                continue

            all_ids_names = [(r[0], r[1]) for r in rows]
            all_names = [r[1] for r in rows]

            prompt = build_pruning_prompt(tag_type, all_names, text)
            response, input_tokens, output_tokens = call_claude(prompt)
            total_cost += estimate_cost(input_tokens, output_tokens)

            try:
                json_start = response.find("{")
                if json_start == -1:
                    raise ValueError("No JSON object found")
                parsed = json.loads(response[json_start:])
                if tag_type not in parsed:
                    raise ValueError("Claude response must be a JSON object with tag type keys")

                selected = parsed[tag_type][:10]  # force trim to 10 just in case
                selected_lower = {s.lower() for s in selected}

                for tag_id, name in all_ids_names:
                    if name.lower() not in selected_lower:
                        cursor.execute(f"DELETE FROM {join_table} WHERE doc_id = %s AND {id_field} = %s", (doc_id, tag_id))
                conn.commit()
                print(f"✅ Pruned {tag_type} to: {selected}")
            except Exception as e:
                print(f"❌ Failed to parse pruning result for {tag_type}: {e}\nRaw response: {response}")

    except Exception as e:
        print(f"❗ Error reading document or pruning: {e}")

print(f"\n💰 Total Claude usage cost: ${total_cost:.6f}")
conn.close()
print("\n✅ Done.")
