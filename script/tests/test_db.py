import os
import pytest
import copy
from sqlalchemy import create_engine, text
from sqlalchemy.exc import IntegrityError
from models.base import SessionLocal, Base
from models.dynamic_models import (
    DynamicDocument,
    DynamicRegion,
    DynamicAuthor,
    DynamicDocAuthor,
    DynamicKeyword,
    DynamicDocKeyword,
)
from config import DATABASE_URL

# ✅ Ensure tests only run on a local database
if "localhost" not in DATABASE_URL and "127.0.0.1" not in DATABASE_URL:
    pytest.exit(f"❌ Tests must run on a local database, but using: {DATABASE_URL}")

# ✅ Verify database connection before running tests
engine = create_engine(DATABASE_URL)
with engine.connect() as conn:
    result = conn.execute(text("SELECT current_database();"))
    db_name = result.scalar()
    print(f"✅ Connected to local test database: {db_name}")


@pytest.fixture(scope="function")
def db_session():
    """Provide a clean database session for each test."""
    session = SessionLocal()

    session.execute(
        text(
            "TRUNCATE TABLE doc_authors, doc_keywords, documents, authors, keywords, regions RESTART IDENTITY CASCADE;"
        )
    )
    session.commit()

    yield session  # Provide session for the test

    # Rollback any uncommitted transactions
    session.rollback()
    session.close()


# Sample test data
DOCUMENT_ID = "test_file.pdf"
DOCUMENT_DATA = {
    "file_name": DOCUMENT_ID,
    "title": "Test Document",
    "abstract": "A simple test case",
    "category": "Test",
    "publish_date": "2025-01-01",
    "source": "Unit Test",
    "region_id": None,
    "s3_file": "s3://test-bucket/test_file.pdf",
    "pdf_link": "https://example.com/test_file.pdf",
}


def test_insert_document_success(db_session):
    """Ensure a document can be inserted successfully."""
    doc = DynamicDocument(**copy.deepcopy(DOCUMENT_DATA))  # Prevent mutation
    db_session.add(doc)
    db_session.commit()

    # Fetch from DB and check
    inserted_doc = (
        db_session.query(DynamicDocument).filter_by(file_name=DOCUMENT_ID).first()
    )
    assert inserted_doc is not None, "Document was not inserted correctly!"
    assert inserted_doc.title == "Test Document"


def test_insert_duplicate_document(db_session):
    """Ensure inserting a duplicate file_name fails."""
    doc1 = DynamicDocument(**copy.deepcopy(DOCUMENT_DATA))
    db_session.add(doc1)
    db_session.commit()

    # Try inserting the same document again
    doc2 = DynamicDocument(**copy.deepcopy(DOCUMENT_DATA))
    db_session.add(doc2)

    with pytest.raises(IntegrityError):
        db_session.commit()
    db_session.rollback()  # Prevent transaction blocking


def test_insert_document_with_invalid_region(db_session):
    """Ensure inserting a document with an invalid region_id fails."""
    invalid_document = copy.deepcopy(DOCUMENT_DATA)
    invalid_document["region_id"] = (
        "00000000-0000-0000-0000-000000000000"  # Invalid UUID
    )

    doc = DynamicDocument(**invalid_document)
    db_session.add(doc)

    with pytest.raises(IntegrityError):
        db_session.commit()
    db_session.rollback()


def test_insert_document_with_authors(db_session):
    """Ensure authors are correctly linked to documents."""
    # Insert document
    doc = DynamicDocument(**copy.deepcopy(DOCUMENT_DATA))
    db_session.add(doc)
    db_session.commit()

    # Insert author and link to document
    author = DynamicAuthor(name="Test Author")
    db_session.add(author)
    db_session.commit()

    doc_author = DynamicDocAuthor(doc_id=doc.id, author_id=author.id)
    db_session.add(doc_author)
    db_session.commit()

    # Check if linked correctly
    linked = db_session.query(DynamicDocAuthor).filter_by(doc_id=doc.id).first()
    assert linked is not None, "Author was not linked to document!"


def test_soft_delete_document(db_session):
    """Ensure a soft-deleted document does not appear in active queries."""
    doc = DynamicDocument(**copy.deepcopy(DOCUMENT_DATA))
    db_session.add(doc)
    db_session.commit()

    # Soft delete
    doc.deleted_at = "2025-03-02 12:00:00"
    db_session.commit()

    # Ensure it's excluded from active queries
    active_doc = (
        db_session.query(DynamicDocument)
        .filter_by(file_name=DOCUMENT_ID, deleted_at=None)
        .first()
    )
    assert active_doc is None, "Soft-deleted document is still appearing!"
