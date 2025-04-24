-- name: DeleteDocAuthorsByDocID :exec
DELETE FROM doc_authors WHERE doc_id = $1;

-- name: DeleteDocKeywordsByDocID :exec
DELETE FROM doc_keywords WHERE doc_id = $1;

-- name: DeleteDocCategoriesByDocID :exec
DELETE FROM doc_categories WHERE doc_id = $1;

-- name: DeleteDocRegionsByDocID :exec
DELETE FROM doc_regions WHERE doc_id = $1;
