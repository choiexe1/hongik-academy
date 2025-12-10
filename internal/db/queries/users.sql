-- name: GetUserByUsername :one
SELECT id, username, name, password_hash, role, created_at
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, name, password_hash, role, created_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (username, name, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING id, username, name, password_hash, role, created_at;

-- name: UpdateUser :one
UPDATE users
SET name = $2, role = $3
WHERE id = $1
RETURNING id, username, name, password_hash, role, created_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, name, role, created_at
FROM users
ORDER BY created_at DESC;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UpdateSessionToken :exec
UPDATE users
SET session_token = $2
WHERE id = $1;

-- name: GetSessionToken :one
SELECT session_token
FROM users
WHERE id = $1;

-- name: ClearSessionToken :exec
UPDATE users
SET session_token = NULL
WHERE id = $1;
