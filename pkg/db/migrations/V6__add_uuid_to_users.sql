-- Add a new UUID column with auto-generated UUIDs
ALTER TABLE users ADD COLUMN new_id UUID DEFAULT gen_random_uuid();

-- Make sure existing rows have UUIDs
UPDATE users SET new_id = gen_random_uuid() WHERE new_id IS NULL;

-- Drop the old SERIAL id column
ALTER TABLE users DROP COLUMN id;

-- Rename the new UUID column to 'id'
ALTER TABLE users RENAME COLUMN new_id TO id;

-- Set 'id' as the new primary key
ALTER TABLE users ADD PRIMARY KEY (id);
