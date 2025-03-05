# import psycopg2

# DATABASE_URL = "postgresql://postgres:R5umUQOEhSp69OrjbnAm@18.205.40.248:5432/postgres"

# try:
#     conn = psycopg2.connect(DATABASE_URL, connect_timeout=5)
#     print("✅ PostgreSQL connection successful!")
#     conn.close()
# except Exception as e:
#     print(f"❌ Connection failed: {e}")

from sqlalchemy import create_engine, text

DATABASE_URL = "postgresql://postgres:R5umUQOEhSp69OrjbnAm@18.205.40.248:5432/postgres"

engine = create_engine(DATABASE_URL)

with engine.connect() as connection:
    result = connection.execute(
        text(
            "SELECT table_name FROM information_schema.tables WHERE table_schema='public';"
        )
    )
    tables = result.fetchall()

print("📂 Tables in the database:")
for table in tables:
    print(f"- {table[0]}")
