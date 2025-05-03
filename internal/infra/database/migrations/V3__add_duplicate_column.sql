-- Migration: Add 'has_duplicate' column to 'documents' table
ALTER TABLE public.documents
ADD COLUMN has_duplicate boolean NOT NULL DEFAULT false;
