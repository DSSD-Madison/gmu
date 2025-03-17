import os
from dotenv import load_dotenv

# Get the parent directory
PARENT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

# Load the .env file from the parent directory
dotenv_path = os.path.join(PARENT_DIR, ".env")
load_dotenv(dotenv_path=dotenv_path)

# Local Database Configuration
LOCAL_DB_HOST = os.getenv("LOCAL_DB_HOST")
LOCAL_DB_PORT = os.getenv("LOCAL_DB_PORT")
LOCAL_DB_USER = os.getenv("LOCAL_DB_USER")
LOCAL_DB_PASSWORD = os.getenv("LOCAL_DB_PASSWORD")
LOCAL_DB_NAME = os.getenv("LOCAL_DB_NAME")

# Production Database Configuration
PROD_HOST = os.getenv("PROD_HOST")
PROD_USER = os.getenv("PROD_USER")
PROD_DB = os.getenv("PROD_DB")
PROD_PASSWORD = os.getenv("PROD_PASSWORD")

# Construct Local Database URL
LOCAL_DATABASE_URL = f"postgresql://{LOCAL_DB_USER}:{LOCAL_DB_PASSWORD}@{LOCAL_DB_HOST}:{LOCAL_DB_PORT}/{LOCAL_DB_NAME}"

# Construct Production Database URL
PROD_DATABASE_URL = (
    f"postgresql://{PROD_USER}:{PROD_PASSWORD}@{PROD_HOST}/{PROD_DB}"
    if PROD_HOST and PROD_USER and PROD_DB and PROD_PASSWORD
    else None
)

# Determine which database to use
if os.getenv("TEST_MODE"):
    DATABASE_URL = LOCAL_DATABASE_URL  # Use local DB for testing
else:
    DATABASE_URL = PROD_DATABASE_URL if PROD_DATABASE_URL else LOCAL_DATABASE_URL

# AWS Credentials
AWS_S3_ACCESS_KEY = os.getenv("S3_ACCESS_KEY")
AWS_S3_SECRET_KEY = os.getenv("S3_SECRET_KEY")
AWS_REGION = os.getenv("REGION")

# Buckets
SKIP_BUCKETS = {"aws-cloudtrail-logs-676432721551-af2ce380"}
TEST_BUCKET = None
