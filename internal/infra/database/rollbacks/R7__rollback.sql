-- 1. Drop the B-tree index on title
DROP INDEX IF EXISTS idx_documents_title;

-- 2. Drop the B-tree index on file_name
DROP INDEX IF EXISTS idx_documents_file_name;

-- 3. Drop the B-tree index on created_at
DROP INDEX IF EXISTS idx_documents_created_at;
