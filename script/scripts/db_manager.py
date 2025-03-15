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


def add_or_update_document(data):
    """Adds or updates a document in the database while managing relationships."""
    session = Session()

    if not isinstance(data.get("file_name"), str) or not data["file_name"]:
        raise ValueError("file_name must be a valid string")

    try:
        # Ensure region exists
        region_name = data.get("region", "Unknown")
        region = get_or_create(session, DynamicRegion, name=region_name)

        # Check if the document already exists
        existing_doc = (
            session.query(DynamicDocument)
            .filter_by(file_name=data["file_name"])
            .first()
        )

        if existing_doc:
            # Update existing document
            existing_doc.title = data.get("title", existing_doc.title)
            existing_doc.abstract = data.get("abstract", existing_doc.abstract)
            existing_doc.category = data.get("category", existing_doc.category)
            existing_doc.publish_date = data.get(
                "publish_date", existing_doc.publish_date
            )
            existing_doc.source = data.get("source", existing_doc.source)
            existing_doc.region_id = region.id
            existing_doc.s3_file = data.get("s3_file", existing_doc.s3_file)
            existing_doc.s3_file_preview = data.get(
                "s3_file_preview", existing_doc.s3_file_preview
            )
            existing_doc.pdf_link = data.get("pdf_link", existing_doc.pdf_link)
            logger.info(f"🔄 Updated document: {data['file_name']}")
        else:
            # Insert new document
            new_doc = DynamicDocument(
                file_name=data["file_name"],
                title=data.get("title", "Untitled"),
                abstract=data.get("abstract", ""),
                category=data.get("category", ""),
                publish_date=data.get("publish_date"),
                source=data.get("source", "Unknown"),
                region_id=region.id,
                s3_file=data.get("s3_file", ""),
                s3_file_preview=data.get("s3_file_preview", None),
                pdf_link=data.get("pdf_link", ""),
            )
            session.add(new_doc)
            session.flush()  # Ensure we get the document ID
            existing_doc = new_doc
            logger.info(f"✅ Created new document: {data['file_name']}")

        # Manage authors (Avoid duplicates)
        session.query(DynamicDocAuthor).filter_by(doc_id=existing_doc.id).delete()
        unique_authors = set(data.get("authors", []))  # Remove duplicates
        for author_name in unique_authors:
            author = get_or_create(session, DynamicAuthor, name=author_name)
            session.add(DynamicDocAuthor(doc_id=existing_doc.id, author_id=author.id))

        # Manage keywords (Avoid duplicates)
        session.query(DynamicDocKeyword).filter_by(doc_id=existing_doc.id).delete()
        unique_keywords = set(data.get("keywords", []))  # Remove duplicates
        for keyword_text in unique_keywords:
            keyword = get_or_create(session, DynamicKeyword, keyword=keyword_text)
            session.add(
                DynamicDocKeyword(doc_id=existing_doc.id, keyword_id=keyword.id)
            )

        session.commit()

    except IntegrityError as ie:
        session.rollback()
        logger.warning(f"⚠️ Integrity Error processing {data['file_name']}: {ie}")

    except Exception as e:
        session.rollback()
        logger.exception(f"❌ Error processing {data['file_name']}: {e}")

    finally:
        session.close()
