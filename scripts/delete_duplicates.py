import psycopg2
from psycopg2.extras import RealDictCursor
import os
from dotenv import load_dotenv
from collections import defaultdict
import boto3
from urllib.parse import urlparse, quote
import logging

load_dotenv()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

role_arn = os.getenv("ROLE_ARN")
role_session_name = "deduplication-session"
index_id = os.getenv("INDEX_ID")

DB_HOST = os.getenv('DB_HOST')
DB_NAME = os.getenv('DB_NAME')
DB_USER = os.getenv('DB_USER')
DB_PASSWORD = os.getenv('DB_PASSWORD')

aws_access_key_id = os.getenv('AWS_ACCESS_KEY_ID')
aws_secret_access_key = os.getenv('AWS_SECRET_ACCESS_KEY')
aws_session_token = os.getenv('AWS_SESSION_TOKEN')

def get_db_connection():
    return psycopg2.connect(
        host=DB_HOST,
        dbname=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )

def parse_s3_uri(uri):
    parsed = urlparse(uri)
    return parsed.netloc, parsed.path.lstrip("/")

def get_s3_metadata(s3, s3_uri):
    bucket, key = parse_s3_uri(s3_uri)
    try:
        head = s3.head_object(Bucket=bucket, Key=key)
        return head['ContentLength'], head['ContentType']
    except Exception as e:
        logger.warning(f"Could not get metadata for {s3_uri}: {e}")
        return None, None

def prefer_foreign(docs):
    for doc in docs:
        if any(lang in doc['s3_file'].lower() for lang in ['french', 'spanish']):
            return doc
    return docs[0]

def group_by(docs, key_func):
    groups = defaultdict(list)
    for doc in docs:
        groups[key_func(doc)].append(doc)
    return [group for group in groups.values() if len(group) > 1]

def approx_equal(a, b, tolerance=3000):
    if a is None or b is None:
        return False
    return abs(a - b) <= tolerance


def process_duplicates():
    s3 = boto3.client(
        's3',
        region_name='us-east-1',
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
        aws_session_token=aws_session_token
    )

    conn = get_db_connection()
    cur = conn.cursor(cursor_factory=RealDictCursor)

    cur.execute("SELECT id, s3_file, title FROM documents WHERE to_delete = false")
    docs = cur.fetchall()

    basename_groups = group_by(docs, lambda d: os.path.basename(d['s3_file']))
    title_groups = group_by(docs, lambda d: d['title'].strip().lower())
    all_groups = {frozenset(doc['id'] for doc in group) for group in basename_groups + title_groups}

    try:
        cur.execute("BEGIN")
        for group_ids in all_groups:
            group = [doc for doc in docs if doc['id'] in group_ids]
            metas = [(doc, *get_s3_metadata(s3, doc['s3_file'])) for doc in group]

            valid = all(m[1] is not None and m[2] is not None for m in metas)
            base_size, base_type = metas[0][1], metas[0][2]
            if not valid or not all(approx_equal(m[1], base_size) and m[2] == base_type for m in metas):
                logger.info("SKIPPING GROUP due to mismatch or missing metadata:")
                for doc, size, ctype in metas:
                    size_note = "MATCH" if size == base_size else f"DIFF (expected {base_size})"
                    type_note = "MATCH" if ctype == base_type else f"DIFF (expected {base_type})"
                    logger.info(f"  - {doc['s3_file']}\n    → size: {size} [{size_note}], type: {ctype} [{type_note}]")
                continue

            keep = prefer_foreign(group)
            logger.info(f"KEEPING: {keep['s3_file']}")
            for doc in group:
                if doc['id'] != keep['id']:
                    logger.info(f"MARKING AS DUPLICATE: {doc['s3_file']}")
                    cur.execute("UPDATE documents SET to_delete = true WHERE id = %s", (doc['id'],))
        cur.execute("COMMIT")
        logger.info("Done marking to_index.")
    except Exception as e:
        logger.error(f"Error: {e}")
        conn.rollback()
    finally:
        cur.close()
        conn.close()

def delete_duplicates_from_kendra():
    print("Starting Kendra cleanup...")
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute("SELECT s3_file FROM documents WHERE to_delete = TRUE AND s3_file IS NOT NULL")
            doc_ids = [row["s3_file"] for row in cur.fetchall()]
            
        if not doc_ids:
            print("No duplicates found for Kendra deletion.")
            return

        sts = boto3.client("sts", region_name="us-east-1")
        creds = sts.assume_role(RoleArn=role_arn, RoleSessionName=role_session_name)["Credentials"]

        kendra_client = boto3.client(
            "kendra",
            region_name="us-east-1",
            aws_access_key_id=creds["AccessKeyId"],
            aws_secret_access_key=creds["SecretAccessKey"],
            aws_session_token=creds["SessionToken"]
        )

        for i in range(0, len(doc_ids), 10):
            batch = doc_ids[i:i + 10]
            print(f"Deleting batch from Kendra: {batch}")
            kendra_client.batch_delete_document(
                IndexId=index_id,
                DocumentIdList=batch
            )

        print("Finished deleting marked duplicates from Kendra.")
    finally:
        conn.close()

def delete_from_s3(s3, s3_uri):
    bucket, key = parse_s3_uri(s3_uri)
    try:
        s3.delete_object(Bucket=bucket, Key=key)
        logger.info(f"Deleted from S3: {s3_uri}")
    except Exception as e:
        logger.warning(f"Failed to delete {s3_uri} from S3: {e}")

def delete_duplicates_from_s3():
    logger.info("Starting S3 cleanup of duplicate documents and previews...")
    s3 = boto3.client(
        's3',
        region_name='us-east-1',
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
        aws_session_token=aws_session_token
    )

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
    process_duplicates()
    delete_duplicates_from_kendra()
    delete_duplicates_from_s3()