import os
from dotenv import load_dotenv

# Load environment variables from .env
load_dotenv()

# Database Configuration
LOCAL_DATABASE_URL = os.getenv("LOCAL_DATABASE_URL")
# Use a separate test database when running tests
if os.getenv("TEST_MODE"):
    DATABASE_URL = LOCAL_DATABASE_URL
else:
    DATABASE_URL = os.getenv("AWS_DATABASE_URL", LOCAL_DATABASE_URL)

# Logging Configuration
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")  # Can be DEBUG, INFO, WARNING, ERROR

# AWS Credentials
AWS_ACCESS_KEY = os.getenv("AWS_ACCESS_KEY_ID")
AWS_SECRET_KEY = os.getenv("AWS_SECRET_ACCESS_KEY")
AWS_REGION = os.getenv("AWS_REGION", "us-east-1")

# Buckets
SKIP_BUCKETS = {"aws-cloudtrail-logs-676432721551-af2ce380"}
TEST_BUCKET = "bep-json-test-bucket"
TEST_BUCKET = None
