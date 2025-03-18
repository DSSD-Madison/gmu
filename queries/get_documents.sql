-- name: GetDocumentsByURIs :many
SELECT *
FROM documents 
WHERE s3_file = ANY($1::text[])
;