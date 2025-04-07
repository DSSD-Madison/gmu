import boto3
import os
import urllib.parse
import psycopg2
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
from typing import List, Dict, Any
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# AWS Configuration
role_arn = os.getenv('ROLE_ARN')
role_session_name = 'kendra-indexing-session'
index_id = os.getenv('INDEX_ID')

# Database Configuration
DB_HOST = os.getenv('DB_HOST')
DB_USER = os.getenv('DB_USER')
DB_NAME = os.getenv('DB_NAME')
DB_PASSWORD = os.getenv('DB_PASSWORD')

def get_db_connection():
    """Create and return a database connection"""
    return psycopg2.connect(
        host=DB_HOST,
        database=DB_NAME,
        user=DB_USER,
        password=DB_PASSWORD
    )

def get_unindexed_documents(conn) -> List[Dict[str, Any]]:
    """Fetch documents that haven't been indexed by Kendra"""
    with conn.cursor(cursor_factory=RealDictCursor) as cur:
        cur.execute("""
            SELECT id, file_name, title, abstract, publish_date, source, s3_file
            FROM documents
            WHERE indexed_by_kendra = false
            AND deleted_at IS NULL
        """)
        return cur.fetchall()

def convert_s3_uri_to_url(s3_uri: str) -> str:
    """Convert S3 URI to HTTPS URL"""
    if not s3_uri.startswith("s3://"):
        raise ValueError(f"Invalid S3 URI: {s3_uri}")

    uri_parts = s3_uri[len("s3://"):]
    parts = uri_parts.split("/", 1)

    if len(parts) != 2:
        raise ValueError(f"S3 URI format is incorrect: {s3_uri}")

    bucket, file_path = parts
    encoded_path = urllib.parse.quote(file_path)

    return f"https://{bucket}.s3.amazonaws.com/{encoded_path}"

def create_kendra_document(doc: Dict[str, Any]) -> Dict[str, Any]:
    """Create a Kendra document from database record"""
    s3_uri = doc['s3_file']
    bucket, key = s3_uri.replace('s3://', '').split('/', 1)
    
    attributes = [
        {'Key': '_file_type', 'Value': {'StringValue': 'PDF'}},
        {'Key': 'Title', 'Value': {'StringValue': doc['title']}},
        {'Key': '_source_uri', 'Value': {'StringValue': convert_s3_uri_to_url(s3_uri)}}
    ]

    # Add optional attributes if they exist
    if doc['abstract']:
        attributes.append({'Key': 'Abstract', 'Value': {'StringValue': doc['abstract']}})
    if doc['source']:
        attributes.append({'Key': 'source', 'Value': {'StringListValue': [doc['source']]}})
    if doc['publish_date']:
        attributes.append({'Key': 'publish_date', 'Value': {'StringValue': doc['publish_date'].isoformat()}})

    return {
        'Id': s3_uri,  # Using S3 URI as document ID
        'S3Path': {
            'Bucket': bucket,
            'Key': key
        },
        'ContentType': 'PDF',
        'Attributes': attributes,
        'Title': doc['title']
    }

def update_document_indexed_status(conn, doc_id: str):
    """Update the document's indexed_by_kendra status in the database"""
    with conn.cursor() as cur:
        cur.execute("""
            UPDATE documents
            SET indexed_by_kendra = true
            WHERE id = %s
        """, (doc_id,))
    conn.commit()

def main():
    # Validate required environment variables
    if not role_arn:
        raise ValueError("ROLE_ARN environment variable is not set")
    if not index_id:
        raise ValueError("INDEX_ID environment variable is not set")

    # Initialize AWS clients
    sts_client = boto3.client('sts', region_name='us-east-1')
    
    # Assume the IAM role
    response = sts_client.assume_role(
        RoleArn=role_arn,
        RoleSessionName=role_session_name
    )

    # Extract temporary credentials
    credentials = response['Credentials']

    # Initialize Kendra client with temporary credentials
    kendra = boto3.client(
        'kendra',
        region_name='us-east-1',
        aws_access_key_id=credentials['AccessKeyId'],
        aws_secret_access_key=credentials['SecretAccessKey'],
        aws_session_token=credentials['SessionToken']
    )

    # Connect to database
    conn = get_db_connection()
    
    try:
        # Get unindexed documents
        documents = get_unindexed_documents(conn)
        logger.info(f"Found {len(documents)} documents to index")

        # Process documents in batches of 10 (Kendra's batch limit)
        batch_size = 10
        for i in range(0, len(documents), batch_size):
            batch = documents[i:i + batch_size]
            kendra_docs = [create_kendra_document(doc) for doc in batch]
            
            # Submit batch to Kendra
            response = kendra.batch_put_document(
                IndexId=index_id,
                Documents=kendra_docs,
                RoleArn=role_arn
            )

            # Handle failed documents
            if response.get('FailedDocuments'):
                for failed_doc in response['FailedDocuments']:
                    logger.error(f"Failed to index document {failed_doc['Id']}: {failed_doc['ErrorMessage']}")
            else:
                # Update database for successfully indexed documents
                for doc in batch:
                    update_document_indexed_status(conn, doc['id'])
                logger.info(f"Successfully indexed batch of {len(batch)} documents")

    except Exception as e:
        logger.error(f"Error during indexing: {str(e)}")
        raise
    finally:
        conn.close()

if __name__ == "__main__":
    main()
