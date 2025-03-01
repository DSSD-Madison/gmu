Local docker image command

```
docker run --name local-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=s3_metadata -p 5432:5432 -d postgres:16
```

psql -h localhost -U postgres

### Potential issues

Author name format

    1. Single column (name TEXT)
        - Pros: Simple, easy to implement
        - Cons: Hard to standardize (e.g., “John Doe” vs. “Doe, John”)
    2. Multiple columns (first_name, middle_name, last_name)
        - Pros: Easier to sort/filter/search
        - Cons: Not all names fit neatly (e.g., mononyms like “Aristotle”)
    3. Flexible JSON Format (name JSONB)
        - Pros: Handles complex name structures (e.g., initials, suffixes)
        - Cons: Requires extra parsing for queries

S3 Bucket source

    1. Redundant Data → If a document comes from the same bucket, the bucket name gets repeated multiple times.
    2. Changing Buckets is Hard → If an S3 bucket name changes, all rows must be updated manually.
    3. Query Performance Issues → Searching by bucket name is slower when stored in documents as plain text.

    Suggested change

    ```
        -- Table: s3_buckets (Stores unique S3 buckets)
        CREATE TABLE s3_buckets (
            id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
            name VARCHAR(255) UNIQUE NOT NULL  -- Unique bucket name
        );

        -- Table: documents (Main Storage Table)
        CREATE TABLE documents (
            id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
            file_name VARCHAR(255) UNIQUE NOT NULL,
            title TEXT NOT NULL,
            abstract TEXT,
            category VARCHAR(100),
            publish_date DATE,
            source VARCHAR(255),
            region_id UUID REFERENCES regions(id) ON DELETE SET NULL,

            s3_bucket_id UUID REFERENCES s3_buckets(id) ON DELETE CASCADE, -- Link to s3_buckets table
            s3_key VARCHAR(1024) UNIQUE NOT NULL,
            pdf_link VARCHAR(1024),

            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP NULL -- Soft delete
        );

        -- Index for fast lookups
        CREATE INDEX idx_documents_s3_bucket ON documents(s3_bucket_id);
    ```
