import boto3
import os
import logging
from typing import List
from utils import get_db_connection, get_kendra_client
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv

# Logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# AWS Configuration
role_arn = os.getenv('ROLE_ARN')
role_session_name = 'delete-untitled'
index_id = os.getenv('INDEX_ID')

kendra = get_kendra_client(role_session_name)

def get_untitled_document_ids(conn) -> List[str]:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("""
            SELECT s3_file
            FROM documents
            WHERE LOWER(title) = 'untitled'
        """)
        rows = cur.fetchall()
        return [row['s3_file'] for row in rows if row['s3_file']]

def mark_as_unindexed(conn, s3_files: List[str]):
    with conn.cursor() as cur:
        cur.execute(
            """
            UPDATE documents
            SET to_index = false
            WHERE s3_file = ANY(%s)
            """,
            (s3_files,)
        )
    conn.commit()

def remove_untitled_documents():
    if not role_arn:
        raise ValueError("ROLE_ARN environment variable is not set")
    if not index_id:
        raise ValueError("INDEX_ID environment variable is not set")

    conn = get_db_connection()

    try:
        s3_file_ids = get_untitled_document_ids(conn)
        logger.info(f"Found {len(s3_file_ids)} 'Untitled' documents to delete from Kendra")

        batch_size = 10
        for i in range(0, len(s3_file_ids), batch_size):
            batch = s3_file_ids[i:i + batch_size]

            for doc_id in batch:
                logger.info(f"Attempting to delete document ID: {doc_id}")

            try:
                kendra.batch_delete_document(
                    IndexId=index_id,
                    DocumentIdList=batch
                )
                logger.info(f"Deleted batch of {len(batch)} documents from Kendra")
            except Exception as e:
                logger.error(f"Error deleting batch: {e}")
                raise

            mark_as_unindexed(conn, batch)
            logger.info(f"Marked batch of {len(batch)} documents as unindexed in DB")


    finally:
        conn.close()

if __name__ == "__main__":
    remove_untitled_documents()
