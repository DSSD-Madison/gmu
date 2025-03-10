import sys
import os
from sqlalchemy.orm import sessionmaker
from sqlalchemy import text

# Add the project root directory to sys.path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from models.base import engine
from logs.logger import logger  # Centralized logging

# Initialize database session
Session = sessionmaker(bind=engine)
session = Session()


def reset_database():
    """Truncates all tables and resets IDs."""
    try:
        sql = text(
            """
            TRUNCATE TABLE 
                documents, 
                regions, 
                authors, 
                doc_authors, 
                keywords, 
                doc_keywords 
            RESTART IDENTITY CASCADE;
        """
        )
        session.execute(sql)
        session.commit()
        logger.info("✅ Database reset successfully!")
        print("✅ Database reset successfully!")

    except Exception as e:
        session.rollback()
        logger.error(f"❌ Database reset failed: {e}")
        print(f"❌ Database reset failed. Check logs/errors.log")

    finally:
        session.close()


if __name__ == "__main__":
    reset_database()
