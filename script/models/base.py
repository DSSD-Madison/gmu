from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base

# Setup Database Connection
DATABASE_URL = "postgresql://postgres:password@localhost/gmu_test_dev_db"
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(bind=engine)

# Base Model
Base = declarative_base()