import os
from dotenv import load_dotenv

# Get the parent directory
PARENT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

# Load the .env file from the parent directory
dotenv_path = os.path.join(PARENT_DIR, ".env")
load_dotenv(dotenv_path=dotenv_path)

# Database Configuration
LOCAL_DATABASE_URL = os.getenv("LOCAL_DATABASE_URL")
# Use a separate test database when running tests
if os.getenv("TEST_MODE"):
    DATABASE_URL = LOCAL_DATABASE_URL
else:
    DATABASE_URL = os.getenv("AWS_DATABASE_URL", LOCAL_DATABASE_URL)

# AWS Credentials
AWS_ACCESS_KEY = os.getenv("ACCESS_KEY")
AWS_SECRET_KEY = os.getenv("SECRET_KEY")
AWS_REGION = os.getenv("REGION", "us-east-1")

# Buckets
SKIP_BUCKETS = {"aws-cloudtrail-logs-676432721551-af2ce380"}
TEST_BUCKET = None
TEST_BUCKET = "bep-json-test-bucket"
