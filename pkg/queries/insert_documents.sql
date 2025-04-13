-- name: InsertUploadedDocument :exec
INSERT INTO documents (
  id,
  s3_file,
  file_name,
  title,
  created_at,
  has_duplicate
) VALUES ($1, $2, $3, $4, NOW(), false);

-- name: InsertDocAuthor :exec
INSERT INTO doc_authors (id, doc_id, author_id)
VALUES ($1, $2, $3);

-- name: InsertDocKeyword :exec
INSERT INTO doc_keywords (id, doc_id, keyword_id)
VALUES ($1, $2, $3);

-- name: InsertDocCategory :exec
INSERT INTO doc_categories (id, doc_id, category_id)
VALUES ($1, $2, $3);

-- name: InsertDocRegion :exec
INSERT INTO doc_regions (id, doc_id, region_id)
VALUES ($1, $2, $3);
