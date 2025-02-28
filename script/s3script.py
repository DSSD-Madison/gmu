import os
import json
import boto3
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Get AWS credentials
aws_access_key = os.getenv("AWS_ACCESS_KEY_ID")
aws_secret_key = os.getenv("AWS_SECRET_ACCESS_KEY")
aws_region = os.getenv("AWS_REGION", "us-east-1") 

# Initialize S3 client
s3 = boto3.client("s3",
    aws_access_key_id=aws_access_key,
    aws_secret_access_key=aws_secret_key,
    region_name=aws_region
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
# Helper functions / modules
################################################
def list_s3_buckets():
    buckets = s3.list_buckets()["Buckets"]
    bucket_names = [bucket["Name"] for bucket in buckets]
    return bucket_names

def list_s3_files(bucket_name):
    response = s3.list_objects_v2(Bucket=bucket_name)
    
    if "Contents" in response:
        for obj in response["Contents"]:
            print(f"File: {obj['Key']}")
    else:
        print("No files found in the bucket.")

################################################
# Script call sequence
################################################