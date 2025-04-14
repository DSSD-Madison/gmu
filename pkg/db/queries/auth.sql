-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :exec
INSERT INTO users (username, password_hash, is_master)
VALUES ($1, $2, $3);

-- name: ListUsers :many
SELECT username, is_master FROM users ORDER BY username;

-- name: DeleteUserByUsername :exec
DELETE FROM users WHERE username = $1;







