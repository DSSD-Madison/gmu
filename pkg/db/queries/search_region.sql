-- name: SearchRegionsByNamePrefix :many
SELECT
    id,
    name
FROM
    regions
WHERE
    name ILIKE $1 || '%'
ORDER BY
    name
    LIMIT 10;