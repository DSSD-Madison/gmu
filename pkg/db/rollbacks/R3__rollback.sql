-- Rollback: Remove 'has_duplicate' column from 'documents' table
ALTER TABLE public.documents
DROP COLUMN has_duplicate;
