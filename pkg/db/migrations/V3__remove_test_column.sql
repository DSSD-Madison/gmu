ALTER TABLE public.documents DROP COLUMN test;

COMMENT ON TABLE public.documents IS 'Removed test column as it was only for demonstration purposes'; 