-- name: SearchCategoriesByNamePrefix :many
SELECT
    id,
    name
FROM
    categories
WHERE
    name ILIKE $1 || '%'
ORDER BY
    name
    LIMIT 10;