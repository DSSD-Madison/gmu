# S3 to PostgreSQL Processing Pipeline

This project extracts metadata files from Amazon S3, processes the data, and inserts it into a PostgreSQL database using SQLAlchemy. The system ensures efficient data management, preventing duplicate records while maintaining relationships between documents, authors, keywords, and regions.

### Project Structure

A brief explanation to the files

```
.
├── config.py           # Manages environment variables (database, AWS credentials)
├── logs/
│   ├── errors.log      # Logs for errors
│   ├── execution.log   # Logs for general execution
│   └── logger.py       # Centralized logging setup
├── models/
│   ├── base.py         # Database engine & session initialization
│   └── dynamic_models.py  # Auto-generated ORM models based on the schema
├── requirements.txt    # Lists required Python dependencies
├── utils/
│   ├── db_manager.py   # Handles database interactions (add/update documents)
│   └── s3_manager.py   # Handles S3 operations (listing & fetching files)
├── main.py             # Orchestrates the full pipeline (S3 → PostgreSQL)
```

### Workflow Breakdown:

1. Fetching Files from S3:

   - The script scans the S3 bucket for files using s3_manager.py.
   - It processes only metadata files (JSON) and associated documents (PDF, DOCX, etc.) while ignoring irrelevant files.

2. Processing Metadata & Files:

   - Metadata files are extracted and parsed to collect document attributes such as title, authors, keywords, region, and file references.
   - If a corresponding document (.pdf, .docx, etc.) exists, its file path is stored in the database.

3. Inserting or Updating Data in PostgreSQL:

   - The db_manager.py script uses the add_or_update_document function to insert new records or update existing ones based on file_name.
   - It ensures:
     - No duplicate authors or keywords are created.
     - File uniqueness is enforced (s3_file and s3_file_preview cannot be duplicated across documents).
     - Regions, authors, and keywords are auto-managed, avoiding manual intervention.

4. Logging & Error Handling:

   - Execution logs track the processing flow in execution.log.
   - Errors and database integrity issues (such as duplicate file paths) are logged in errors.log for debugging.

### Setup instructions

1. Move into `db_population/` directory

```bash
cd db_population/
```

2. Create a Virtual Environment

```bash
python -m venv venv
source venv/bin/activate  # On macOS/Linux
venv\Scripts\activate     # On Windows
```

3. Install Dependencies

```bash
pip install -r requirements.txt
```

4. Set Up Environment Variables

Create a `.env` file and define:

```
LOCAL_DATABASE_URL=postgresql://user:password@localhost/db_name
AWS_DATABASE_URL=postgresql://user:password@aws-host/db_name
ACCESS_KEY=your_access_key
SECRET_ACCESS=your_secret_key
REGION=us-east-1
```

5. Run the Pipeline

```bash
python main.py
```

### How to Add or Update a Document

For manual data entry, you can use the add_or_update_document function from db_manager.py.

**Function Overview**

```python
from utils.db_manager import add_or_update_document

data = {
    "file_name": "test_file.pdf",
    "title": "Test Document",
    "abstract": "A simple test case",
    "category": "Test",
    "publish_date": "2025-01-01",
    "source": "Unit Test",
    "region": "Global",
    "s3_file": "s3://test-bucket/test_file.pdf",
    "s3_file_preview": "s3://test-bucket/test_file_preview.webm",
    "pdf_link": "https://example.com/test_file.pdf",
    "authors": ["John Doe"],
    "keywords": ["Machine Learning", "AI"],
}

add_or_update_document(data)
```

**Function Behavior**

- If a document with the same file_name already exists, it will be updated. Otherwise, a new entry will be created. New authors and keywords are added without duplication to prevent redundancy.
- File name uniqueness is required to avoid conflicts. Regions, authors, and keywords are automatically managed, so manual creation is unnecessary.
- If s3_file or s3_file_preview already exists in another document, an IntegrityError will be raised. For debugging, check errors.log for any processing issues.
