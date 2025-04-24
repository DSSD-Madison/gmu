-- 1. Rename to_index back to indexed_by_kendra
ALTER TABLE documents RENAME COLUMN to_index TO indexed_by_kendra;

-- 2. Flip values back to original
UPDATE documents SET indexed_by_kendra = NOT indexed_by_kendra;

-- 3. Drop the to_generate_preview column
ALTER TABLE documents DROP COLUMN to_generate_preview;

-- 4. Rename to_delete back to has_duplicate
ALTER TABLE documents RENAME COLUMN to_delete TO has_duplicate;
