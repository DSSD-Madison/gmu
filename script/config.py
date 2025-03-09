import os
from dotenv import load_dotenv

# Load environment variables from .env
load_dotenv()

# Database Configuration
DATABASE_URL = os.getenv(
    "AWS_DATABASE_URL", "postgresql://postgres:password@localhost/gmu_test_dev_db"
)

# Logging Configuration
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")  # Can be DEBUG, INFO, WARNING, ERROR

# Batch Processing
BATCH_SIZE = int(os.getenv("BATCH_SIZE", "100"))  # Number of records per batch
