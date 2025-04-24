-- name: UpdateDocumentMetadata :exec
UPDATE documents
SET
  title = $2,
  abstract = $3,
  publish_date = $4,
  source = $5
WHERE id = $1;
