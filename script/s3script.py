import os
import json
import boto3
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Get AWS credentials from env variables
AWS_ACCESS_KEY = os.getenv("AWS_ACCESS_KEY_ID")
AWS_SECRET_KEY = os.getenv("AWS_SECRET_ACCESS_KEY")
AWS_REGION = os.getenv("AWS_REGION", "us-east-1")  

# Initialize S3 client
s3 = boto3.client("s3",
    aws_access_key_id=AWS_ACCESS_KEY,
    aws_secret_access_key=AWS_SECRET_KEY,
    region_name=AWS_REGION
)

################################################
# Big idea
################################################
# Get all buckets
# For every bucket go through file
# For every file parse the json and add data into database
#   If there are any conflicts, document down in a diff log file or something

################################################
# TODO
################################################
# Figure out a local database to mess around
# Double check and document the schema

################################################
# Helper functions
################################################
def list_s3_buckets():
    """Retrieve and return all S3 bucket names."""
    try:
        response = s3.list_buckets()
        return [bucket["Name"] for bucket in response.get("Buckets", [])]
    except Exception as e:
        print(f"Error listing buckets: {e}")
        return []

def list_s3_files(bucket_name):
    """Retrieve and return all file objects in a given S3 bucket."""
    try:
        response = s3.list_objects_v2(Bucket=bucket_name)
        return response.get("Contents", [])  # List of files
    except Exception as e:
        print(f"Error listing files in {bucket_name}: {e}")
        return []

def process_json_file(bucket_name, file_key):
    """Reads and parses a JSON file from S3."""
    try:
        response = s3.get_object(Bucket=bucket_name, Key=file_key)
        json_content = json.loads(response["Body"].read().decode("utf-8"))

        print(f"Processed JSON: {file_key}")
        return json_content  # Returns parsed JSON data
    except Exception as e:
        print(f"Error reading JSON file {file_key}: {e}")
        return None

def process_pdf_file(bucket_name, file_key):
    """Handles PDFs (currently logs them)."""
    print(f"Found PDF: {file_key} (Storing metadata)")
    return {"file": file_key, "type": "PDF"}

def process_unknown_file(bucket_name, file_key):
    """Handles unknown file types."""
    print(f"Unknown File Type: {file_key}")
    return {"file": file_key, "type": "Unknown"}

def categorize_files(bucket_name):
    """Categorizes and processes files in an S3 bucket."""
    print(f"Processing files in bucket: {bucket_name}")
    files = list_s3_files(bucket_name)

    processed_data = []
    
    for file in files:
        file_key = file["Key"]
        _, ext = os.path.splitext(file_key)  # Extract file extension

        if ext == ".json":
            processed_data.append(process_json_file(bucket_name, file_key))
        elif ext == ".pdf":
            processed_data.append(process_pdf_file(bucket_name, file_key))
        else:
            processed_data.append(process_unknown_file(bucket_name, file_key))

    return processed_data  # Return categorized file data

################################################
# Script Execution
################################################
# if __name__ == "__main__":
    # Main idea on the script
    # print("Starting S3 File Processing...")

    # # Get all S3 buckets
    # buckets = list_s3_buckets()
    
    # if not buckets:
    #     print("No S3 buckets found. Exiting.")
    #     exit()

    # # Process files in each bucket
    # for bucket in buckets:
    #     categorize_files(bucket)

    # print("Finished Processing All Files.")

    # BUCKET_NAME = "allianceforpeacebuilding-org"
    # print(process_json_file(BUCKET_NAME, list_s3_files(BUCKET_NAME)[1]["Key"]))
    # print(list_s3_files(BUCKET_NAME)[1]["Key"])