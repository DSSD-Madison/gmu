import json
import boto3
from botocore.exceptions import BotoCoreError, ClientError
from logs.logger import logger
from config import (
    AWS_ACCESS_KEY,
    AWS_SECRET_KEY,
    AWS_REGION,
    TEST_BUCKET,
    SKIP_BUCKETS,
)

# Initialize S3 client using credentials from config.py
s3 = boto3.client(
    "s3",
    aws_access_key_id=AWS_ACCESS_KEY,
    aws_secret_access_key=AWS_SECRET_KEY,
    region_name=AWS_REGION,
)


def get_s3_buckets():
    """Retrieves and returns all S3 bucket names."""
    try:
        response = s3.list_buckets()
        bucket_names = [bucket["Name"] for bucket in response.get("Buckets", [])]
        logger.info(f"Found {len(bucket_names)} S3 buckets.")
        return bucket_names
    except (BotoCoreError, ClientError) as e:
        logger.error(f"Error listing S3 buckets: {e}")
        return []


def get_s3_files(bucket_name, prefix=""):
    """Fetches files from an S3 bucket in batches and yields them."""
    try:
        paginator = s3.get_paginator("list_objects_v2")
        batch_count = 0

        for page in paginator.paginate(
            Bucket=bucket_name,
            Prefix=prefix,
        ):
            batch_files = [f["Key"] for f in page.get("Contents", [])]

            if batch_files:
                batch_count += 1
                yield batch_files  # Process each batch immediately

        logger.info(f"Processed {batch_count} batches in bucket '{bucket_name}'")

    except (BotoCoreError, ClientError) as e:
        logger.error(f"Error listing files in {bucket_name}: {e}")
        yield []


def get_s3_preview_file(bucket_name, base_name):
    """Checks if a corresponding .webp preview file exists and returns its S3 URI."""
    preview_file_key = f"{base_name}.webp"
    try:
        s3.head_object(Bucket=bucket_name, Key=preview_file_key)  # Check if exists
        return f"s3://{bucket_name}/{preview_file_key}"  # Return full S3 URI
    except s3.exceptions.ClientError:
        return None


def get_s3_metadata(bucket_name, base_name):
    """Fetches JSON metadata if a corresponding .json file exists."""
    metadata_file_key = f"{base_name}.pdf.metadata.json"
    return read_json_from_s3(bucket_name, metadata_file_key)


def read_json_from_s3(bucket_name, file_key):
    """Reads and returns a JSON file from S3."""
    try:
        response = s3.get_object(Bucket=bucket_name, Key=file_key)
        json_content = json.loads(response["Body"].read().decode("utf-8"))
        logger.info(f"Successfully read JSON file: {file_key}")
        return json_content
    except json.JSONDecodeError as je:
        logger.error(f"JSON decoding error in {file_key}")
        return None
    except (BotoCoreError, ClientError) as e:
        logger.error(f"Error retrieving {file_key} from {bucket_name}:")
        return None


def convert_metadata_to_document(metadata):
    """Converts S3 metadata JSON into a simplified document format."""
    if not metadata or "Attributes" not in metadata:
        return {}  # Invalid format, return {}

    attributes = metadata["Attributes"]

    # Construct a simplified document format
    document_data = {
        "title": attributes.get(
            "Title", metadata.get("Title", "")
        ),  # Prioritize Attributes.Title
        "pdf_link": attributes.get("Link"),  # S3 file link
        "region": attributes.get("Region", ["Unknown"])[0],  # Default to 'Unknown'
        "authors": attributes.get("_authors", []),  # List of authors
        "publish_date": attributes.get("Date_Published"),  # Publication date
        "source": attributes.get("source", ["Unknown"])[0],  # Default to 'Unknown'
        "keywords": attributes.get("Subject_Keywords", []),  # List of keywords
    }

    return document_data


def process_s3_files(process_function):
    """Processes files from S3 and applies a given function."""

    # Determine which buckets to process
    buckets = [TEST_BUCKET] if TEST_BUCKET else get_s3_buckets()
    buckets = [b for b in buckets if b not in SKIP_BUCKETS]

    total_files_processed = 0

    for bucket in buckets:
        logger.info(f"Processing bucket: {bucket}")
        processed_count = 0

        for file_batch in get_s3_files(bucket):
            logger.info(f"Processing batch: {len(file_batch)}")
            for file_key in file_batch:
                logger.info(f"Processing file: {file_key}")

                # Apply the function to process each file
                process_function(bucket, file_key)

                processed_count += 1

        total_files_processed += processed_count
        logger.info(f"Completed {processed_count} files in bucket: {bucket}")

    logger.info(f"🎉 Total files processed across all buckets: {total_files_processed}")
