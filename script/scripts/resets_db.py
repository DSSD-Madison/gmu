import sys
import os

from sqlalchemy.orm import sessionmaker
from sqlalchemy import create_engine, text
# Add the parent directory to sys.path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

from models.base import engine  # Now it should work

Session = sessionmaker(bind=engine)
session = Session()

def reset_database():
    sql = text("TRUNCATE TABLE documents, regions, authors, doc_authors, keywords, doc_keywords RESTART IDENTITY CASCADE;")
    session.execute(sql)
    session.commit()
    print("✅ Database reset successfully!")

if __name__ == "__main__":
    reset_database()