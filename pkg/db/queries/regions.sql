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

-- name: ListAllRegions :many
SELECT id, name FROM regions ORDER BY name;

-- name: FindRegionByName :one
SELECT * FROM regions WHERE LOWER(name) = LOWER($1) LIMIT 1;

-- name: InsertRegion :exec
INSERT INTO regions (id, name) VALUES ($1, $2); 