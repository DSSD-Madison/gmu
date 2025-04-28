-- name: UpdateDocumentMetadata :exec
UPDATE documents
SET
  title = $2,
  abstract = $3,
  publish_date = $4,
  source = $5,
  to_index = $6
WHERE id = $1;

-- name: UpdateDocumentDeletionStatus :exec
UPDATE documents
SET
    to_delete = $2
WHERE id = $1;
