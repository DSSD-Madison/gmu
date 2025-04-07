ALTER TABLE public.documents ADD COLUMN test VARCHAR(255);

COMMENT ON COLUMN public.documents.test IS 'Test column added back during rollback'; 