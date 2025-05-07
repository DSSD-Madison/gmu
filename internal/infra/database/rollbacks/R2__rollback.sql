-- Rollback: Rename name column back to keyword in keywords table
ALTER TABLE public.keywords RENAME COLUMN name TO keyword; 