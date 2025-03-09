import logging
import os

# Ensure logs directory exists
LOG_DIR = "logs"
os.makedirs(LOG_DIR, exist_ok=True)

# Setup logging configuration
LOG_FILE = os.path.join(LOG_DIR, "app.log")

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(LOG_FILE),  # Log to file
        logging.StreamHandler(),  # Log to console
    ],
)

# Create logger instance
logger = logging.getLogger(__name__)
