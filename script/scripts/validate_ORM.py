from sqlalchemy import inspect
from models.base import engine
from models import Document, Author

inspector = inspect(engine)

# Check for missing tables
db_tables = inspector.get_table_names()
model_tables = [Document.__tablename__, Author.__tablename__]

missing_in_db = set(model_tables) - set(db_tables)
missing_in_models = set(db_tables) - set(model_tables)

if missing_in_db:
    print(
        f"⚠️ Warning: These tables exist in `models.py` but NOT in the DB: {missing_in_db}"
    )
if missing_in_models:
    print(
        f"⚠️ Warning: These tables exist in the DB but NOT in `models.py`: {missing_in_models}"
    )
