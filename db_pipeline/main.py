import os
from scripts.s3_manager import (
    get_s3_preview_file,
    get_s3_metadata,
    convert_metadata_to_document,
    process_s3_files,
)
from scripts.db_manager import add_or_update_document
from logs.logger import logger
import time


def migrate(bucket_name, file_key):
    """Migrates all S3 file into the PostgreSQL database."""
    base_name, ext = os.path.splitext(file_key)

    # Skip entries that end with /
    if file_key.endswith("/"):
        logger.info(f"Skipping directory: {file_key}")
        return

    # Ignore metadata and preview files at this stage
    if ext in [".json", ".webp"]:
        return

    # Construct S3 URI
    s3_file = f"s3://{bucket_name}/{file_key}"

    # Check for associated files
    s3_file_preview = get_s3_preview_file(bucket_name, base_name)
    json_data = get_s3_metadata(bucket_name, base_name)

    # Convert JSON metadata if available
    document_data = {"file_name": file_key, "s3_file": s3_file}
    if s3_file_preview:
        logger.info(f"Preview found for {file_key}")
        document_data["s3_file_preview"] = s3_file_preview
    else:
        logger.error(f"Preview not found for {file_key}")

    if json_data:
        logger.info(f"json_data found for {file_key}")
        new_json = convert_metadata_to_document(json_data)
        document_data.update(new_json)
    else:
        logger.error(f"json_data not found for {file_key}")

    # Insert or update the document
    add_or_update_document(document_data)


start_time = time.time()
process_s3_files(migrate)
end_time = time.time()

print(f"Execution time: {end_time - start_time:.4f} seconds")
