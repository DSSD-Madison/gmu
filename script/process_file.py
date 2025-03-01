from scripts.s3_fetcher import list_s3_buckets, list_s3_files, fetch_json_from_s3
from scripts.db_insert import insert_document

def process_s3_files():
    """Orchestrates the S3-to-Postgres pipeline."""
    buckets = list_s3_buckets()

    for bucket in buckets:
        metadata_files = list_s3_files(bucket)
        
        # for file_key in metadata_files:
        #     data = fetch_json_from_s3(file_key, bucket)
        #     insert_document(data)
        
        print("################################################")
        print(f"Bucket {bucket} Completed")
        print(f"{len(metadata_files)} files parsed")
        print("################################################")


if __name__ == "__main__":
    print(f"File process started...")
    process_s3_files()
    print(f"File process completed")