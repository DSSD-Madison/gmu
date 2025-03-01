from sqlalchemy import Column, String, Date, ForeignKey, Text, TIMESTAMP, UniqueConstraint
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.sql import func
import uuid
from sqlalchemy.orm import relationship
from .base import Base  # Import Base from base.py

# Document Table
class Document(Base):
    __tablename__ = "documents"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    file_name = Column(String(255), unique=True, nullable=False)
    title = Column(Text, nullable=False)
    abstract = Column(Text)
    category = Column(String(100))
    publish_date = Column(Date)
    source = Column(String(255))
    region_id = Column(UUID(as_uuid=True), ForeignKey("regions.id", ondelete="SET NULL"))

    s3_bucket = Column(String(255), nullable=False)
    s3_key = Column(String(1024), unique=True, nullable=False)
    pdf_link = Column(String(1024))

    created_at = Column(TIMESTAMP, server_default=func.now())
    last_modified = Column(TIMESTAMP, server_default=func.now(), onupdate=func.now())
    deleted_at = Column(TIMESTAMP, nullable=True)

# Region Table
class Region(Base):
    __tablename__ = "regions"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(255), unique=True, nullable=False)

# Author Table
class Author(Base):
    __tablename__ = "authors"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(255), unique=True, nullable=False)

# Many-to-Many Relationship: Documents ↔ Authors
class DocAuthor(Base):
    __tablename__ = "doc_authors"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    doc_id = Column(UUID(as_uuid=True), ForeignKey("documents.id", ondelete="CASCADE"), nullable=False)
    author_id = Column(UUID(as_uuid=True), ForeignKey("authors.id", ondelete="CASCADE"), nullable=False)

    __table_args__ = (UniqueConstraint("doc_id", "author_id"),)

# Keyword Table
class Keyword(Base):
    __tablename__ = "keywords"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    keyword = Column(String(255), unique=True, nullable=False)

# Many-to-Many Relationship: Documents ↔ Keywords
class DocKeyword(Base):
    __tablename__ = "doc_keywords"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    doc_id = Column(UUID(as_uuid=True), ForeignKey("documents.id", ondelete="CASCADE"), nullable=False)
    keyword_id = Column(UUID(as_uuid=True), ForeignKey("keywords.id", ondelete="CASCADE"), nullable=False)

    __table_args__ = (UniqueConstraint("doc_id", "keyword_id"),)