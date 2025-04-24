-- 1. Rename indexed_by_kendra → to_index
ALTER TABLE documents RENAME COLUMN indexed_by_kendra TO to_index;

-- 2. Flip all the values (TRUE → FALSE and vice versa)
UPDATE documents SET to_index = NOT to_index;

-- 3. Add new column: to_generate_preview (default false, or NULL if you prefer)
ALTER TABLE documents ADD COLUMN to_generate_preview BOOLEAN DEFAULT TRUE;

-- 4. Rename has_duplicate → to_delete
ALTER TABLE documents RENAME COLUMN has_duplicate TO to_delete;
