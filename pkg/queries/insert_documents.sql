-- name: InsertUploadedDocument :exec
INSERT INTO documents (
  id,
  s3_file,
  file_name,
  title,
  created_at,
  has_duplicate
) VALUES ($1, $2, $3, $4, NOW(), false);
