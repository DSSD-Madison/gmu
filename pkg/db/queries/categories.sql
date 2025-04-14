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

-- name: ListAllCategories :many
SELECT id, name FROM categories ORDER BY name;

-- name: FindCategoryByName :one
SELECT * FROM categories WHERE LOWER(name) = LOWER($1) LIMIT 1;

-- name: InsertCategory :exec
INSERT INTO categories (id, name) VALUES ($1, $2);