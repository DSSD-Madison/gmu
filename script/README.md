# S3 to PostgreSQL Processing Pipeline

This script extracts metadata files from Amazon S3, processes the data, and inserts it into a PostgreSQL database using SQLAlchemy.

### Project Structure

A brief explanation to the files

```
.
├── config.py           # Manages environment variables (database, AWS credentials)
├── db/
│   └── schema.sql      # PostgreSQL schema (table definitions & indexes)
├── logs/
│   ├── errors.log      # Logs for errors
│   ├── execution.log   # Logs for general execution
│   └── logger.py       # Centralized logging setup
├── models/
│   ├── base.py         # Database engine & session initialization
│   └── dynamic_models.py  # Auto-generated ORM models based on the schema
├── process_file.py     # Orchestrates the full pipeline (S3 → PostgreSQL)
├── requirements.txt    # Lists required Python dependencies
├── scripts/
│   ├── db_insert.py    # Inserts extracted data into PostgreSQL
│   ├── db_reset.py     # Resets and clears the local database
│   └── s3_manager.py   # Handles S3 operations (listing & fetching files)
```

### Setup instructions

1. Create a Virtual Environment

```bash
python -m venv venv
source venv/bin/activate  # On macOS/Linux
venv\Scripts\activate     # On Windows
```

2. Install Dependencies

```bash
pip install -r requirements.txt
```

3. Set Up Environment Variables

Create a `.env` file and define:

```
LOCAL_DATABASE_URL=postgresql://user:password@localhost/db_name
AWS_DATABASE_URL=postgresql://user:password@aws-host/db_name
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
BATCH_SIZE=100
```

4. Run the Pipeline

```bash
python process_file.py
```

### How It Works

1. Fetches JSON metadata files from S3 (s3_manager.py)
2. Parses and extracts metadata
3. Inserts data into PostgreSQL (db_insert.py)
4. Logs progress and errors (logger.py)
5. Resets database when needed (db_reset.py)
