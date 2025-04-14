-- name: SearchKeywordsByNamePrefix :many
SELECT
    id,
    name
FROM
    keywords
WHERE
    name ILIKE $1 || '%'
ORDER BY
    name
    LIMIT 10;