import os
import logging
from sqlalchemy.orm import sessionmaker
from sqlalchemy import create_engine
from sqlalchemy.exc import IntegrityError
from models.models import Document, Region, Author, DocAuthor, Keyword, DocKeyword

# Ensure logs directory exists
os.makedirs("logs", exist_ok=True)

# Configure logging properly
logging.basicConfig(
    filename="logs/errors.log",
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    filemode="a"
)

# Test Logging
logging.info("Test Error - Logging system is working.")

# Setup DB connection
DATABASE_URL = "postgresql://postgres:password@localhost/gmu_test_dev_db"
engine = create_engine(DATABASE_URL)
Session = sessionmaker(bind=engine)
session = Session()

def insert_document(data):
    """Inserts a document into the database."""
    try:
        # Ensure region exists or create
        region_name = data.get("Attributes", {}).get("Region", ["Unknown"])[0]
        region = session.query(Region).filter_by(name=region_name).first()
        if not region:
            region = Region(name=region_name)
            session.add(region)
            session.commit()

        # Insert document
        document = Document(
            file_name=data["DocumentId"][0],
            title=data["Title"],
            abstract="",
            category="",
            publish_date=data["Attributes"].get("Date_Published"),
            source=data["Attributes"].get("source", ["Unknown"])[0],
            region_id=region.id,
            s3_bucket="my-bucket",
            s3_key=data["Attributes"].get("Link"),
            pdf_link=data["Attributes"].get("Link")
        )
        session.add(document)
        session.commit()

        # Insert authors
        for author_name in data["Attributes"].get("_authors", []):
            author = session.query(Author).filter_by(name=author_name).first()
            if not author:
                author = Author(name=author_name)
                session.add(author)
                session.commit()
            session.add(DocAuthor(doc_id=document.id, author_id=author.id))
            session.commit()

        # Insert keywords
        for keyword_text in data["Attributes"].get("Subject_Keywords", []):
            keyword = session.query(Keyword).filter_by(keyword=keyword_text).first()
            if not keyword:
                keyword = Keyword(keyword=keyword_text)
                session.add(keyword)
                session.commit()
            session.add(DocKeyword(doc_id=document.id, keyword_id=keyword.id))
            session.commit()

        print(f"Successfully inserted: {data['DocumentId'][0]}")
    
    except IntegrityError as ie:
        session.rollback()
        logging.info(f"Integrity Error inserting {data['DocumentId'][0]}: {ie}")
        print(f"Integrity error inserting {data['DocumentId'][0]}. Check logs/errors.log")
    
    except Exception as e:
        session.rollback()
        logging.info(f"Error inserting {data['Title']}: {e}")
        print(f"Error inserting {data['DocumentId'][0]}. Check logs/errors.log")