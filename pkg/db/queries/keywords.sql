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

-- name: ListAllKeywords :many
SELECT id, name FROM keywords ORDER BY name;

-- name: FindKeywordByName :one
SELECT * FROM keywords WHERE LOWER(name) = LOWER($1) LIMIT 1;

-- name: InsertKeyword :exec
INSERT INTO keywords (id, name) VALUES ($1, $2);    