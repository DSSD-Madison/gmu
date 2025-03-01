import os
import json
import boto3
from dotenv import load_dotenv
import logging

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(filename="logs/processing.log", level=logging.INFO, format="%(asctime)s - %(message)s")

# AWS Credentials
AWS_ACCESS_KEY = os.getenv("AWS_ACCESS_KEY_ID")
AWS_SECRET_KEY = os.getenv("AWS_SECRET_ACCESS_KEY")
AWS_REGION = os.getenv("AWS_REGION", "us-east-1")

# Initialize S3 client
s3 = boto3.client(
    "s3",
    aws_access_key_id=AWS_ACCESS_KEY,
    aws_secret_access_key=AWS_SECRET_KEY,
    region_name=AWS_REGION
)

def list_s3_buckets():
    """Retrieve and return all S3 bucket names."""
    try:
        response = s3.list_buckets()
        return [bucket["Name"] for bucket in response.get("Buckets", [])]
    except Exception as e:
        print(f"Error listing buckets: {e}")
        return []


def list_s3_files(bucket_name):
    """Fetches all files from S3 bucket."""
    response = s3.list_objects_v2(Bucket=bucket_name)
    files = response.get("Contents", [])

    metadata_files = [f["Key"] for f in files if f["Key"].endswith(".json")]
    logging.info(f"Found {len(metadata_files)} JSON metadata files.")
    return metadata_files

def fetch_json_from_s3(file_key, bucket_name):
    """Reads a JSON file from S3."""
    response = s3.get_object(Bucket=bucket_name, Key=file_key)
    return json.loads(response["Body"].read().decode("utf-8"))