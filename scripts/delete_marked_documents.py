import os
import logging
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

def delete_document_from_s3(s3, s3_uri):
    bucket, key = parse_s3_uri(s3_uri)
    try:
        s3.delete_object(Bucket=bucket, Key=key)
        logger.info(f"Deleted from S3: {s3_uri}")
    except Exception as e:
        logger.warning(f"Failed to delete {s3_uri}: {e}")

def delete_documents_from_kendra():
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

def delete_documents_from_s3():
    logger.info("Starting S3 cleanup of marked documents...")
    
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("SELECT s3_file, s3_file_preview FROM documents WHERE to_delete = TRUE")
            rows = cur.fetchall()

            for row in rows:
                if row["s3_file"]:
                    delete_document_from_s3(s3, row["s3_file"])
                if row["s3_file_preview"]:
                    delete_document_from_s3(s3, row["s3_file_preview"])

        logger.info("S3 cleanup complete.")
    finally:
        conn.close()
        
def delete_documents_from_db():
    logger.info("Starting database cleanup of marked documents...")

    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            # Delete from join tables
            cur.execute("""
                DELETE FROM doc_keywords WHERE doc_id IN (SELECT id FROM documents WHERE to_delete = TRUE);
                DELETE FROM doc_authors WHERE doc_id IN (SELECT id FROM documents WHERE to_delete = TRUE);
                DELETE FROM doc_regions WHERE doc_id IN (SELECT id FROM documents WHERE to_delete = TRUE);
                DELETE FROM doc_categories WHERE doc_id IN (SELECT id FROM documents WHERE to_delete = TRUE);
            """)
            # Delete from documents table
            cur.execute("DELETE FROM documents WHERE to_delete = TRUE;")

            # Clean up orphaned metadata
            cur.execute("""
                DELETE FROM keywords WHERE id NOT IN (SELECT DISTINCT keyword_id FROM doc_keywords);
                DELETE FROM authors WHERE id NOT IN (SELECT DISTINCT author_id FROM doc_authors);
                DELETE FROM regions WHERE id NOT IN (SELECT DISTINCT region_id FROM doc_regions);
                DELETE FROM categories WHERE id NOT IN (SELECT DISTINCT category_id FROM doc_categories);
            """)

            conn.commit()
            logger.info("Database cleanup complete.")
    except Exception as e:
        conn.rollback()
        logger.error(f"Database deletion failed: {e}")
    finally:
        conn.close()



if __name__ == "__main__":
    delete_documents_from_kendra()
    delete_documents_from_s3()
    delete_documents_from_db()
