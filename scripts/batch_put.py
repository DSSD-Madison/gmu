import os
import urllib.parse
import logging
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
from typing import List, Dict, Any
from utils import get_db_connection, get_kendra_client

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# AWS Configuration
role_arn = os.getenv('ROLE_ARN')
index_id = os.getenv('INDEX_ID')

def get_unindexed_documents(conn) -> List[Dict[str, Any]]:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("""
            SELECT 
                d.id,
                d.title,
                d.s3_file,
                d.source,
                (
                    SELECT array_agg(r.name)
                    FROM doc_regions dr
                    JOIN regions r ON dr.region_id = r.id
                    WHERE dr.doc_id = d.id
                ) AS regions,
                (
                    SELECT array_agg(a.name)
                    FROM doc_authors da
                    JOIN authors a ON da.author_id = a.id
                    WHERE da.doc_id = d.id
                ) AS authors,
                (
                    SELECT array_agg(k.name)
                    FROM doc_keywords dk
                    JOIN keywords k ON dk.keyword_id = k.id
                    WHERE dk.doc_id = d.id
                ) AS keywords,
                (
                    SELECT array_agg(c.name)
                    FROM doc_categories dc
                    JOIN categories c ON dc.category_id = c.id
                    WHERE dc.doc_id = d.id
                ) AS categories
  
            FROM documents d
            WHERE d.to_index = true
              AND d.to_delete = false
        """)
        return cur.fetchall()

def convert_s3_uri_to_url(s3_uri: str) -> str:
    if not s3_uri.startswith("s3://"):
        raise ValueError(f"Invalid S3 URI: {s3_uri}")
    bucket, file_path = s3_uri[len("s3://"):].split("/", 1)
    encoded_path = urllib.parse.quote(file_path)
    return f"https://{bucket}.s3.amazonaws.com/{encoded_path}"

def truncate(value: str, max_length: int = 2048) -> str:
    return value[:max_length] if value and len(value) > max_length else value

def create_kendra_document(doc: Dict[str, Any]) -> Dict[str, Any]:
    s3_uri = doc['s3_file']

    if s3_uri.endswith('.temp') or os.path.basename(s3_uri).startswith('.'):
        raise ValueError(f"Skipping temp/system file: {s3_uri}")

    bucket, key = s3_uri.replace('s3://', '').split('/', 1)

    file_ext = os.path.splitext(key)[1].lower()
    if file_ext == '.pdf':
        content_type = 'PDF'
    elif file_ext == '.docx':
        content_type = 'MS_WORD'
    else:
        raise ValueError(f"Unsupported file type: {file_ext}")

    def norm_list(values):
        return sorted({v.strip().title() for v in values if v and v.strip()})

    attributes = [
        {'Key': '_file_type', 'Value': {'StringValue': file_ext[1:].upper()}},
        {'Key': '_source_uri', 'Value': {'StringValue': convert_s3_uri_to_url(s3_uri)}}
    ]

    if doc.get('regions'):
        regions = norm_list(doc['regions'])
        if regions:
            attributes.append({'Key': 'Region', 'Value': {'StringListValue': regions}})

    if doc.get('keywords'):
        keywords = norm_list(doc['keywords'])
        if keywords:
            attributes.append({'Key': 'Keyword', 'Value': {'StringListValue': keywords}})

    if doc.get('authors'):
        authors = norm_list(doc['authors'])
        if authors:
            attributes.append({'Key': 'Author', 'Value': {'StringListValue': authors}})

    if doc.get('categories'):
        categories = norm_list(doc['categories'])
        if categories:
            attributes.append({'Key': 'Category', 'Value': {'StringListValue': categories}})

    if doc.get('source') and doc['source'].strip():
        attributes.append({'Key': 'Source', 'Value': {'StringValue': truncate(doc['source'].strip())}})

    return {
        'Id': s3_uri,
        'S3Path': {
            'Bucket': bucket,
            'Key': key
        },
        'ContentType': content_type,
        'Attributes': attributes,
        'Title': truncate(doc['title'].strip())
    }



def update_document_indexed_status(conn, doc_id: str):
    with conn.cursor() as cur:
        cur.execute("""
            UPDATE documents
            SET to_index = false
            WHERE id = %s
        """, (doc_id,))
    conn.commit()

def index_documents():
    if not role_arn:
        raise ValueError("ROLE_ARN environment variable is not set")
    if not index_id:
        raise ValueError("INDEX_ID environment variable is not set")

    role_session_name = 'kendra-indexing-session'
    kendra = get_kendra_client(role_session_name)

    conn = get_db_connection()

    try:
        documents = get_unindexed_documents(conn)
        logger.info(f"Found {len(documents)} documents to index")

        batch_size = 10
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            kendra_docs = []
            valid_docs = []

            for doc in batch:
                try:
                    if doc['title'].strip().lower() == 'untitled':
                        continue

                    k_doc = create_kendra_document(doc)
                    kendra_docs.append(k_doc)
                    valid_docs.append(doc)

                except Exception as skip_reason:
                    logger.warning(f"Skipping document {doc['s3_file']}: {skip_reason}")

            try:
                if not kendra_docs:
                    logger.info("No valid documents to index in this batch. Skipping batch_put_document.")
                    continue

                response = kendra.batch_put_document(
                    IndexId=index_id,
                    Documents=kendra_docs,
                    RoleArn=role_arn
                )

                failed = response.get('FailedDocuments', [])
                if failed:
                    for fail in failed:
                        logger.error(f"Failed to index document {fail['Id']}: {fail['ErrorMessage']}")
                else:
                    for doc in valid_docs:
                        update_document_indexed_status(conn, doc['id'])
                        logger.info(f"Indexed document: {doc['s3_file']}")
                    logger.info(f"Successfully indexed batch of {len(valid_docs)} documents")


            except kendra.exceptions.ValidationException as ve:
                logger.error(f"ValidationException: {ve}")
            except Exception as e:
                logger.error(f"Unexpected error: {e}")
                raise

    finally:
        conn.close()

if __name__ == "__main__":
    index_documents()
