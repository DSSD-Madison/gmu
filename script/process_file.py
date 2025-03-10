from logs.logger import logger  # Import centralized logger
from scripts.s3_manager import list_s3_buckets, list_s3_files, fetch_json_from_s3
from scripts.db_insert import insert_document

# Configurable Options
# Set to None to process all buckets
TEST_BUCKET = "allianceforpeacebuilding-org"
TEST_BUCKET = None

# Buckets to exclude
SKIP_BUCKETS = {"aws-cloudtrail-logs-676432721551-af2ce380"}


def process_s3_files():
    """Orchestrates the S3-to-Postgres pipeline with logging and error handling."""

    if TEST_BUCKET:
        buckets = [TEST_BUCKET]  # Process a single test bucket
    else:
        buckets = list_s3_buckets()
        buckets = [
            b for b in buckets if b not in SKIP_BUCKETS
        ]  # Remove skipped buckets

    total_files_processed = 0

    for bucket in buckets:
        logger.info(f"📂 Processing bucket: {bucket}")

        processed_count = 0

        for file_batch in list_s3_files(bucket):
            batch_count = 0

            for file_key in file_batch:
                try:
                    data = fetch_json_from_s3(bucket, file_key)
                    if data:
                        insert_document(data)
                        processed_count += 1
                        batch_count += 1
                except Exception as e:
                    logger.error(
                        f"❌ Error processing file {file_key} in {bucket}: {e}"
                    )

            logger.info(f"🔄 Processed {batch_count} files in current batch")

        total_files_processed += processed_count
        logger.info(f"✅ Bucket {bucket} Completed - {processed_count} files processed")

    logger.info(
        f"🎉 File processing completed. Total files processed: {total_files_processed}"
    )


if __name__ == "__main__":
    logger.info("🚀 File processing started...")
    process_s3_files()
    logger.info("✅ File processing completed")
