-- 1. Add back the old id as SERIAL
ALTER TABLE users ADD COLUMN old_id SERIAL;

-- 2. Drop the UUID id
ALTER TABLE users DROP COLUMN id;

-- 3. Rename old_id to id
ALTER TABLE users RENAME COLUMN old_id TO id;

-- 4. Set id as primary key
ALTER TABLE users ADD PRIMARY KEY (id);
