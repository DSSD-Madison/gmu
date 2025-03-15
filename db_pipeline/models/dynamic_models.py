from sqlalchemy import MetaData
from sqlalchemy.orm import declarative_base
from models.base import engine

# Automatically Reflect Schema
metadata = MetaData()
metadata.reflect(bind=engine)  # Fetch schema from DB

# Generate ORM Models Dynamically
Base = declarative_base(metadata=metadata)


class DynamicDocument(Base):
    __tablename__ = "documents"
    __table__ = metadata.tables["documents"]


class DynamicRegion(Base):
    __tablename__ = "regions"
    __table__ = metadata.tables["regions"]


class DynamicAuthor(Base):
    __tablename__ = "authors"
    __table__ = metadata.tables["authors"]


class DynamicDocAuthor(Base):
    __tablename__ = "doc_authors"
    __table__ = metadata.tables["doc_authors"]


class DynamicKeyword(Base):
    __tablename__ = "keywords"
    __table__ = metadata.tables["keywords"]


class DynamicDocKeyword(Base):
    __tablename__ = "doc_keywords"
    __table__ = metadata.tables["doc_keywords"]
