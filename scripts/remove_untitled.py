import boto3
import os
import psycopg2
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
import logging
from typing import List

# Logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# AWS Configuration
role_arn = os.getenv('ROLE_ARN')
role_session_name = 'kendra-delete-untitled-session'
index_id = os.getenv('INDEX_ID')

# Database Configuration
DB_HOST = os.getenv('DB_HOST')
DB_USER = os.getenv('DB_USER')
DB_NAME = os.getenv('DB_NAME')
DB_PASSWORD = os.getenv('DB_PASSWORD')

def get_db_connection():
    return psycopg2.connect(
        host=DB_HOST,
        database=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )

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
            SET indexed_by_kendra = FALSE
            WHERE s3_file = ANY(%s)
            """,
            (s3_files,)
        )
    conn.commit()

def main():
    if not role_arn:
        raise ValueError("ROLE_ARN environment variable is not set")
    if not index_id:
        raise ValueError("INDEX_ID environment variable is not set")

    sts = boto3.client('sts', region_name='us-east-1')
    creds = sts.assume_role(RoleArn=role_arn, RoleSessionName=role_session_name)['Credentials']

    kendra = boto3.client(
        'kendra',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )

    conn = get_db_connection()

    try:
        s3_file_ids = get_untitled_document_ids(conn)
        logger.info(f"Found {len(s3_file_ids)} 'Untitled' documents to delete from Kendra")

        batch_size = 10
        for i in range(0, len(s3_file_ids), batch_size):
            batch = s3_file_ids[i:i + batch_size]
            try:
                kendra.batch_delete_document(
                    IndexId=index_id,
                    DocumentIdList=batch,
                    RoleArn=role_arn
                )
                logger.info(f"Deleted batch of {len(batch)} documents from Kendra")
            except Exception as e:
                logger.error(f"Error deleting batch: {e}")
                raise

            # Mark documents as unindexed in the DB whether they existed in Kendra or not
            mark_as_unindexed(conn, batch)
            logger.info(f"Marked batch of {len(batch)} documents as unindexed in DB")

    finally:
        conn.close()

if __name__ == "__main__":
    main()
