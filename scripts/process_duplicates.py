import os
import logging
from dotenv import load_dotenv
from collections import defaultdict
from psycopg2.extras import RealDictCursor
from utils import get_db_connection, parse_s3_uri, get_s3_client

load_dotenv()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

role_session_name = "process-duplicates"
s3 = get_s3_client(role_session_name)

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
                    cur.execute("UPDATE documents SET to_delete = TRUE WHERE id = %s", (doc['id'],))
        cur.execute("COMMIT")
        logger.info("Done marking documents to_delete = TRUE.")
    except Exception as e:
        logger.error(f"Error during processing; rolling back transaction: {e}")
        conn.rollback()
    finally:
        cur.close()
        conn.close()
