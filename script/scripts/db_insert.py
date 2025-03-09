import os
from sqlalchemy.orm import sessionmaker
from sqlalchemy.exc import IntegrityError
from models.dynamic_models import (
    DynamicDocument,
    DynamicRegion,
    DynamicAuthor,
    DynamicDocAuthor,
    DynamicKeyword,
    DynamicDocKeyword,
)
from models.base import engine
from logs.logger import logger  # Centralized logging

# Ensure logs directory exists
os.makedirs("logs", exist_ok=True)

# Initialize Session
Session = sessionmaker(bind=engine)


def get_or_create(session, model, **kwargs):
    """Fetch an existing record or create a new one."""
    instance = session.query(model).filter_by(**kwargs).first()
    if not instance:
        instance = model(**kwargs)
        session.add(instance)
        session.commit()
    return instance


def insert_document(data):
    """Inserts a document and its related data into the database."""
    session = Session()
    try:
        # Ensure region exists
        region_name = data.get("Attributes", {}).get("Region", ["Unknown"])[0]
        region = get_or_create(session, DynamicRegion, name=region_name)

        # Insert document
        document = DynamicDocument(
            file_name=data["DocumentId"][0],
            title=data["Title"],
            abstract="",
            category="",
            publish_date=data["Attributes"].get("Date_Published"),
            source=data["Attributes"].get("source", ["Unknown"])[0],
            region_id=region.id,
            s3_file=data["Attributes"].get("Link"),
            pdf_link=data["Attributes"].get("Link"),
        )
        session.add(document)
        session.commit()

        # Insert authors
        for author_name in data["Attributes"].get("_authors", []):
            author = get_or_create(session, DynamicAuthor, name=author_name)
            session.add(DynamicDocAuthor(doc_id=document.id, author_id=author.id))

        # Insert keywords
        for keyword_text in data["Attributes"].get("Subject_Keywords", []):
            keyword = get_or_create(session, DynamicKeyword, keyword=keyword_text)
            session.add(DynamicDocKeyword(doc_id=document.id, keyword_id=keyword.id))

        session.commit()
        logger.info(f"✅ Successfully inserted: {data['DocumentId'][0]}")

    except IntegrityError as ie:
        session.rollback()
        logger.warning(f"⚠️ Integrity Error inserting {data['DocumentId'][0]}: {ie}")

    except Exception as e:
        session.rollback()
        logger.exception(f"❌ Error inserting {data['DocumentId'][0]}: {e}")

    finally:
        session.close()
