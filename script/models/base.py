from sqlalchemy import create_engine, MetaData
from sqlalchemy.orm import sessionmaker, declarative_base
from logs.logger import logger  # Centralized logging
from config import DATABASE_URL

# Initialize SQLAlchemy Engine
try:
    engine = create_engine(DATABASE_URL, echo=False)
    SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    metadata = MetaData()
    Base = declarative_base(metadata=metadata)

    logger.info("✅ Database connection established")
except Exception as e:
    logger.error(f"❌ Database connection failed: {e}")
    raise  # Raise exception to prevent silent failure
