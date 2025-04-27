import os
import logging
import boto3
from dotenv import load_dotenv
from psycopg2.extras import RealDictCursor
from utils import get_db_connection, parse_s3_uri, get_s3_client, get_kendra_client


load_dotenv()
index_id = os.getenv("INDEX_ID")

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

role_session_name = "document-deletion"
s3 = get_s3_client(role_session_name)
kendra_client = get_kendra_client(role_session_name)

def delete_from_s3(s3, s3_uri):
    bucket, key = parse_s3_uri(s3_uri)
    try:
        s3.delete_object(Bucket=bucket, Key=key)
        logger.info(f"Deleted from S3: {s3_uri}")
    except Exception as e:
        logger.warning(f"Failed to delete {s3_uri}: {e}")

def delete_duplicates_from_kendra():
    logger.info("Starting Kendra cleanup...")
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("SELECT s3_file FROM documents WHERE to_delete = TRUE")
            doc_ids = [row["s3_file"] for row in cur.fetchall()]
            
        if not doc_ids:
            logger.info("No documents found for Kendra deletion.")
            return

        for i in range(0, len(doc_ids), 10):
            batch = doc_ids[i:i + 10]
            logger.info(f"Deleting batch from Kendra: {batch}")
            kendra_client.batch_delete_document(
                IndexId=index_id,
                DocumentIdList=batch
            )

        logger.info("Finished deleting from Kendra.")
    finally:
        conn.close()

def delete_duplicates_from_s3():
    logger.info("Starting S3 cleanup of marked documents...")
    
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("SELECT s3_file, s3_file_preview FROM documents WHERE to_delete = TRUE")
            rows = cur.fetchall()

            for row in rows:
                if row["s3_file"]:
                    delete_from_s3(s3, row["s3_file"])
                if row["s3_file_preview"]:
                    delete_from_s3(s3, row["s3_file_preview"])

        logger.info("S3 cleanup complete.")
    finally:
        conn.close()

if __name__ == "__main__":
    delete_duplicates_from_kendra()
    delete_duplicates_from_s3()
