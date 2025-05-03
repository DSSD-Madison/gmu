-- 1. Create a B-tree index on title
CREATE INDEX IF NOT EXISTS idx_documents_title
    ON documents (title);

-- 2. Create a B-tree index on file_name
CREATE INDEX IF NOT EXISTS idx_documents_file_name
    ON documents (file_name);

-- 3. Create a B-tree index on created_at
CREATE INDEX IF NOT EXISTS idx_documents_created_at
    ON documents (created_at);
