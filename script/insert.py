import json
from sqlalchemy.orm import sessionmaker
from sqlalchemy import create_engine
from models.base import Base
from models.models import Document, Region, Author, Keyword, DocAuthor, DocKeyword

import logging

# Configure logging
logging.basicConfig(
    filename="errors.log",   # Log file name
    filemode="a",            # Append mode
    format="%(asctime)s - %(levelname)s - %(message)s",
    level=logging.ERROR      # Log only errors and above
)

# Database connection
DATABASE_URL = "postgresql://postgres:password@localhost/gmu_test_dev_db"
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(bind=engine)

def get_or_create(session, model, **kwargs):
    """Helper function to get an existing record or create a new one."""
    instance = session.query(model).filter_by(**kwargs).first()
    if not instance:
        instance = model(**kwargs)
        session.add(instance)
        session.commit()
    return instance

def insert_document_from_json(json_data):
    """Parses and inserts a document entry from JSON metadata."""
    session = SessionLocal()
    try:
        # Extract document details
        file_name = json_data["DocumentId"][0]  
        title = json_data["Attributes"].get("Title", "")
        pdf_link = json_data["Attributes"].get("Link", "")
        region_name = json_data["Attributes"].get("Region", [None])[0]  # Assume first region
        source = json_data["Attributes"].get("source", [None])[0]
        publish_date = json_data["Attributes"].get("Date_Published", None)
        authors = json_data["Attributes"].get("_authors", [])
        keywords = json_data["Attributes"].get("Subject_Keywords", [])

        # Ensure region exists
        region = get_or_create(session, Region, name=region_name) if region_name else None

        # Insert document
        document = get_or_create(
            session, Document,
            file_name=file_name,
            title=title,
            abstract=None,  # Modify if your JSON has this
            category=None,  # Modify if needed
            publish_date=publish_date,
            source=source,
            region_id=region.id if region else None,
            s3_bucket="my-bucket",  # Modify based on where files are stored
            s3_key=file_name,   # To be double checked
            pdf_link=pdf_link
        )

        # Handle authors (Many-to-Many)
        for author_name in authors:
            author = get_or_create(session, Author, name=author_name)
            get_or_create(session, DocAuthor, doc_id=document.id, author_id=author.id)

        # Handle keywords (Many-to-Many)
        for keyword_text in keywords:
            keyword = get_or_create(session, Keyword, keyword=keyword_text)
            get_or_create(session, DocKeyword, doc_id=document.id, keyword_id=keyword.id)

        print(f"Document '{file_name}' inserted successfully!")

    except Exception as e:
        session.rollback()
        error_message = f"Error inserting document '{file_name}' : {e}"
        logging.error(error_message)

    finally:
        session.close()

# Load JSON Example
with open("example.json", "r") as file:
    json_data = json.load(file)

insert_document_from_json(json_data)