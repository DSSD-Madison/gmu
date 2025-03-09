import logging
import os
from logging.handlers import RotatingFileHandler

# Ensure logs directory exists
LOG_DIR = "logs"
os.makedirs(LOG_DIR, exist_ok=True)

# Define log file paths
EXECUTION_LOG_FILE = os.path.join(LOG_DIR, "execution.log")
ERROR_LOG_FILE = os.path.join(LOG_DIR, "errors.log")

# Create logger
logger = logging.getLogger(__name__)  # Standardized logger name
logger.setLevel(logging.DEBUG)  # Capture all log levels

# Prevent duplicate handlers if script is run multiple times
if logger.hasHandlers():
    logger.handlers.clear()

# Create handlers
execution_handler = RotatingFileHandler(
    EXECUTION_LOG_FILE, maxBytes=5 * 1024 * 1024, backupCount=3
)
error_handler = RotatingFileHandler(
    ERROR_LOG_FILE, maxBytes=5 * 1024 * 1024, backupCount=3
)

# Set log levels
execution_handler.setLevel(logging.INFO)  # Captures INFO, WARNING, ERROR, CRITICAL
error_handler.setLevel(logging.ERROR)  # Captures only ERROR and CRITICAL

# Define log format
log_format = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")

# Attach formatters to handlers
execution_handler.setFormatter(log_format)
error_handler.setFormatter(log_format)

# Add handlers to logger
logger.addHandler(execution_handler)  # Writes INFO+ logs to execution.log
logger.addHandler(error_handler)  # Writes ERROR+ logs to errors.log
