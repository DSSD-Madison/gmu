-- name: GetDocumentsByURIs :many
SELECT
    d.*,  -- Select all columns from the documents table
    -- Aggregate author names into a text array
    COALESCE(ARRAY_AGG(DISTINCT a.name) FILTER (WHERE a.id IS NOT NULL), '{}'::text[]) AS author_names,
    -- Aggregate region names into a text array
    COALESCE(ARRAY_AGG(DISTINCT r.name) FILTER (WHERE r.id IS NOT NULL), '{}'::text[]) AS region_names,
    -- Aggregate keyword names into a text array
    COALESCE(ARRAY_AGG(DISTINCT k.name) FILTER (WHERE k.id IS NOT NULL), '{}'::text[]) AS keyword_names,
    -- Aggregate category names into a text array
    COALESCE(ARRAY_AGG(DISTINCT c.name) FILTER (WHERE c.id IS NOT NULL), '{}'::text[]) AS category_names
FROM
    documents d
        LEFT JOIN
    doc_authors da ON d.id = da.doc_id
        LEFT JOIN
    authors a ON da.author_id = a.id
        LEFT JOIN
    doc_regions dr ON d.id = dr.doc_id
        LEFT JOIN
    regions r ON dr.region_id = r.id
        LEFT JOIN
    doc_keywords dk ON d.id = dk.doc_id
        LEFT JOIN
    keywords k ON dk.keyword_id = k.id
        LEFT JOIN
    doc_categories dc ON d.id = dc.doc_id
        LEFT JOIN
    categories c ON dc.category_id = c.id
WHERE
    d.s3_file = ANY($1::text[]) -- Filter documents by the provided list of s3_file paths
GROUP BY
    d.id -- Group by document ID to aggregate related names for each document
ORDER BY
    d.id; -- Optional: Add an order clause if needed

-- name: FindDocumentByID :one
SELECT
    d.*,
    ARRAY_REMOVE(ARRAY_AGG(DISTINCT a.name), NULL)::text[] AS author_names,
    ARRAY_REMOVE(ARRAY_AGG(DISTINCT r.name), NULL)::text[] AS region_names,
    ARRAY_REMOVE(ARRAY_AGG(DISTINCT k.name), NULL)::text[] AS keyword_names,
    ARRAY_REMOVE(ARRAY_AGG(DISTINCT c.name), NULL)::text[] AS category_names
FROM documents d
LEFT JOIN doc_authors da ON d.id = da.doc_id
LEFT JOIN authors a ON da.author_id = a.id
LEFT JOIN doc_regions dr ON d.id = dr.doc_id
LEFT JOIN regions r ON dr.region_id = r.id
LEFT JOIN doc_keywords dk ON d.id = dk.doc_id
LEFT JOIN keywords k ON dk.keyword_id = k.id
LEFT JOIN doc_categories dc ON d.id = dc.doc_id
LEFT JOIN categories c ON dc.category_id = c.id
WHERE d.id = $1
GROUP BY d.id;

-- name: FindDocumentByS3Path :one
SELECT *
FROM documents
WHERE s3_file = $1;

-- name: GetLatestDocuments :many
SELECT *
FROM documents
ORDER BY created_at DESC
    LIMIT 25;

