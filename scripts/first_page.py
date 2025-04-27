import os
import io
import subprocess
import pymupdf
from urllib.parse import urlparse
from psycopg2.extras import RealDictCursor
from utils import get_db_connection, get_s3_client, get_s3_resource
from io import BytesIO
from PIL import Image

role_session_name = 'preview-generation'
s3_client = get_s3_client(role_session_name)
s3_resource = get_s3_resource(role_session_name)

TEMP_DIR = "/tmp/doc_preview"
os.makedirs(TEMP_DIR, exist_ok=True)

def docx_to_pdf(docx_path, output_dir):
    subprocess.run([
        "libreoffice",
        "--headless",
        "--convert-to", "pdf",
        "--outdir", output_dir,
        docx_path
    ], check=True)

def extract_bucket_key(s3_uri):
    parsed = urlparse(s3_uri)
    bucket = parsed.netloc
    key = parsed.path.lstrip('/')
    return bucket, key

def process_document(doc, conn):
    s3_uri = doc['s3_file']
    doc_id = doc['id']
    bucket, key = extract_bucket_key(s3_uri)
    file_name = os.path.basename(key)

    try:
        obj = s3_client.get_object(Bucket=bucket, Key=key)
        content = obj['Body'].read()
        file_stream = io.BytesIO(content)

        if key.lower().endswith('.pdf'):
            pdf_document = pymupdf.open(stream=file_stream, filetype='pdf')
        elif key.lower().endswith('.docx'):
            local_docx = os.path.join(TEMP_DIR, file_name)
            with open(local_docx, "wb") as f:
                f.write(content)
            docx_to_pdf(local_docx, TEMP_DIR)
            pdf_path = local_docx.replace('.docx', '.pdf')
            pdf_document = pymupdf.open(pdf_path)
        else:
            print(f"::warning::Skipping unsupported file type: {key}")
            return

        page = pdf_document[0]
        pix = page.get_pixmap()
        img_data = pix.tobytes("png")
        image = Image.open(io.BytesIO(img_data))
        image.thumbnail((10000, 120))

        output_buffer = BytesIO()
        image.save(output_buffer, format="webp")

        webp_key = key.rsplit('.', 1)[0].strip() + '.webp'
        webp_uri = f"s3://{bucket}/{webp_key}"

        s3_resource.Object(bucket, webp_key).put(
            Body=output_buffer.getvalue(),
            ACL='public-read',
            ContentType='image/webp'
        )

        with conn.cursor() as cur:
            cur.execute("""
                UPDATE documents
                SET s3_file_preview = %s, to_generate_preview = FALSE
                WHERE id = %s
            """, (webp_uri, doc_id))

        conn.commit()
        print(f"✅ Generated preview for {file_name} → {webp_uri}")

    except Exception as e:
        print(f"::error::Failed to process {file_name}: {e}")

def generate_preview_images():
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("""
                SELECT id, s3_file
                FROM documents
                WHERE to_generate_preview = true
            """)
            docs = cur.fetchall()
            print(f"Found {len(docs)} unprocessed documents")

            for doc in docs:
                process_document(doc, conn)
    finally:
        conn.close()

if __name__ == "__main__":
    generate_preview_images()
