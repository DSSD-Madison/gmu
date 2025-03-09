import difflib
from sqlalchemy import MetaData
from models.base import engine

# Load database schema
metadata = MetaData()
metadata.reflect(bind=engine)

# Read schema.sql
with open("db/schema.sql", "r") as f:
    expected_schema = f.readlines()

# Fetch the current schema from the database
actual_schema = []
for table_name, table in metadata.tables.items():
    actual_schema.append(str(table.compile(engine)).strip())

# Compare expected and actual schema
diff = difflib.unified_diff(
    expected_schema, actual_schema, fromfile="schema.sql", tofile="Live DB"
)

if list(diff):
    print("⚠️ Schema Mismatch Detected! Here’s the difference:")
    print("\n".join(diff))
else:
    print("✅ Schema matches database!")
