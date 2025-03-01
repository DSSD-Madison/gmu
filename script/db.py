from models.base import Base, engine

# Create all tables
Base.metadata.create_all(bind=engine)
print("✅ Database tables created successfully!")