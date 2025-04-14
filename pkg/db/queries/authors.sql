-- name: SearchAuthorsByNamePrefix :many
SELECT
    id,
    name
FROM
    authors
WHERE
    name ILIKE $1 || '%'  -- Case-insensitive prefix search
ORDER BY
    name -- Optional: order results alphabetically
    LIMIT 10; -- Optional: limit the number of results (good for autocomplete)

-- name: ListAllAuthors :many
SELECT id, name FROM authors ORDER BY name;