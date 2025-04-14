import boto3
import os
import psycopg2
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
import logging
from typing import List, Tuple
from urllib.parse import urlparse
import argparse

# Logging setup
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# AWS Configuration
role_arn = os.getenv('ROLE_ARN')
role_session_name = 's3-clean-database-session'
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

def get_duplicate_docs(conn) -> List[Tuple[str, str]]:
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("""
            SELECT id, s3_file
            FROM documents
            WHERE has_duplicate = TRUE
        """)
        rows = cur.fetchall()
        return [(row['id'], row['s3_file']) for row in rows if row['s3_file']]

def parse_s3_uri(uri: str) -> Tuple[str, str]:
    parsed = urlparse(uri)
    return parsed.netloc, parsed.path.lstrip("/")

def delete_from_s3(bucket_to_keys: dict, s3, dry_run: bool):
    for bucket, keys in bucket_to_keys.items():
        for i in range(0, len(keys), 1000):
            batch = keys[i:i + 1000]
            if dry_run:
                logger.info(f"DRY RUN: would delete {len(batch)} files from S3 bucket '{bucket}': {batch}")
            else:
                try:
                    s3.delete_objects(
                        Bucket=bucket,
                        Delete={'Objects': [{'Key': key} for key in batch]}
                    )
                    logger.info(f"Deleted {len(batch)} files from bucket: {bucket}")
                except Exception as e:
                    logger.error(f"Error deleting from S3 bucket {bucket}: {e}")
                    raise

def delete_from_kendra(kendra, doc_ids: List[str], dry_run: bool):
    for i in range(0, len(doc_ids), 10):
        batch = doc_ids[i:i + 10]
        if dry_run:
            logger.info(f"DRY RUN: would delete {len(batch)} documents from Kendra index: {batch}")
        else:
            try:
                kendra.batch_delete_document(
                    IndexId=index_id,
                    DocumentIdList=batch
                )
                logger.info(f"Deleted {len(batch)} documents from Kendra index")
            except Exception as e:
                logger.error(f"Error deleting from Kendra: {e}")
                raise

def clean_related_db_entries(conn, doc_ids: List[str], dry_run: bool):
    with conn.cursor() as cur:
        if dry_run:
            logger.info(f"DRY RUN: would delete {len(doc_ids)} documents and associated join entries")
        else:
            # Delete from join tables
            cur.execute("DELETE FROM doc_regions WHERE document_id = ANY(%s)", (doc_ids,))
            cur.execute("DELETE FROM doc_authors WHERE document_id = ANY(%s)", (doc_ids,))
            cur.execute("DELETE FROM doc_categories WHERE document_id = ANY(%s)", (doc_ids,))
            cur.execute("DELETE FROM doc_keywords WHERE document_id = ANY(%s)", (doc_ids,))

            # Delete from documents
            cur.execute("DELETE FROM documents WHERE id = ANY(%s)", (doc_ids,))

            # Clean up orphaned metadata
            cur.execute("""
                DELETE FROM regions
                WHERE id NOT IN (SELECT DISTINCT region_id FROM doc_regions)
            """)
            cur.execute("""
                DELETE FROM authors
                WHERE id NOT IN (SELECT DISTINCT author_id FROM doc_authors)
            """)
            cur.execute("""
                DELETE FROM categories
                WHERE id NOT IN (SELECT DISTINCT category_id FROM doc_categories)
            """)
            cur.execute("""
                DELETE FROM keywords
                WHERE id NOT IN (SELECT DISTINCT keyword_id FROM doc_keywords)
            """)

            conn.commit()
            logger.info("Deleted document rows and cleaned up orphaned metadata.")

def main():
    parser = argparse.ArgumentParser(description="Clean up duplicate documents from S3, Kendra, and DB.")
    parser.add_argument("--dry-run", action="store_true", help="Preview what will be deleted without making any changes.")
    args = parser.parse_args()
    dry_run = args.dry_run

    if not role_arn or not index_id:
        raise ValueError("ROLE_ARN or INDEX_ID environment variables not set")

    sts = boto3.client('sts', region_name='us-east-1')
    creds = sts.assume_role(RoleArn=role_arn, RoleSessionName=role_session_name)['Credentials']

    s3 = boto3.client(
        's3',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )

    kendra = boto3.client(
        'kendra',
        region_name='us-east-1',
        aws_access_key_id=creds['AccessKeyId'],
        aws_secret_access_key=creds['SecretAccessKey'],
        aws_session_token=creds['SessionToken']
    )

    conn = get_db_connection()

    try:
        duplicate_docs = get_duplicate_docs(conn)
        if not duplicate_docs:
            logger.info("No duplicate documents found.")
            return

        doc_ids = [doc_id for doc_id, _ in duplicate_docs]
        s3_uris = [s3_uri for _, s3_uri in duplicate_docs]
        kendra_ids = s3_uris  # s3_file is used as DocumentId in Kendra

        # Organize S3 keys by bucket
        bucket_to_keys = {}
        for _, uri in duplicate_docs:
            bucket, key = parse_s3_uri(uri)
            bucket_to_keys.setdefault(bucket, []).append(key)

        # Dry run aware deletion steps
        delete_from_kendra(kendra, kendra_ids, dry_run)
        delete_from_s3(bucket_to_keys, s3, dry_run)
        clean_related_db_entries(conn, doc_ids, dry_run)

        logger.info(f"{'Would remove' if dry_run else 'Removed'} {len(doc_ids)} duplicate documents and associated data.")

    finally:
        conn.close()

if __name__ == "__main__":
    main()
