-- name: ListAllAuthors :many
SELECT id, name FROM authors ORDER BY name;

-- name: ListAllKeywords :many
SELECT id, name FROM keywords ORDER BY name;

-- name: ListAllRegions :many
SELECT id, name FROM regions ORDER BY name;

-- name: ListAllCategories :many
SELECT id, name FROM categories ORDER BY name;
