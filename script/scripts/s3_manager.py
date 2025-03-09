import json
import boto3
from botocore.exceptions import BotoCoreError, ClientError
from logs.logger import logger
from config import AWS_ACCESS_KEY, AWS_SECRET_KEY, AWS_REGION, S3_BATCH_SIZE

# Initialize S3 client using credentials from config.py
s3 = boto3.client(
    "s3",
    aws_access_key_id=AWS_ACCESS_KEY,
    aws_secret_access_key=AWS_SECRET_KEY,
    region_name=AWS_REGION,
)


def list_s3_buckets():
    """Retrieve and return all S3 bucket names."""
    try:
        response = s3.list_buckets()
        bucket_names = [bucket["Name"] for bucket in response.get("Buckets", [])]
        logger.info(f"✅ Found {len(bucket_names)} S3 buckets.")
        return bucket_names
    except (BotoCoreError, ClientError) as e:
        logger.error(f"❌ Error listing buckets: {e}")
        return []


def list_s3_files(bucket_name, prefix=""):
    """Fetches all files from an S3 bucket, optionally filtering by prefix."""
    try:
        paginator = s3.get_paginator("list_objects_v2")
        files = []

        for page in paginator.paginate(
            Bucket=bucket_name,
            Prefix=prefix,
            PaginationConfig={"MaxItems": S3_BATCH_SIZE},
        ):
            files.extend(page.get("Contents", []))

        file_keys = [f["Key"] for f in files]
        logger.info(f"✅ Found {len(file_keys)} files in bucket '{bucket_name}'")
        return file_keys
    except (BotoCoreError, ClientError) as e:
        logger.error(f"❌ Error listing files in {bucket_name}: {e}")
        return []


def fetch_json_from_s3(bucket_name, file_key):
    """Reads and returns a JSON file from S3."""
    try:
        response = s3.get_object(Bucket=bucket_name, Key=file_key)
        json_content = json.loads(response["Body"].read().decode("utf-8"))
        logger.info(f"✅ Successfully fetched JSON: {file_key}")
        return json_content
    except json.JSONDecodeError as je:
        logger.error(f"❌ JSON decoding error for {file_key}: {je}")
        return None
    except (BotoCoreError, ClientError) as e:
        logger.error(f"❌ Error fetching {file_key} from {bucket_name}: {e}")
        return None
