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
