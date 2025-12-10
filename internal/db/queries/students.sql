-- name: CountStudents :one
SELECT COUNT(*) FROM students
WHERE
    (sqlc.narg('search')::text IS NULL OR
        name ILIKE '%' || sqlc.narg('search')::text || '%' OR
        phone ILIKE '%' || sqlc.narg('search')::text || '%' OR
        parent_phone ILIKE '%' || sqlc.narg('search')::text || '%' OR
        remarks ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('gender')::text IS NULL OR gender = sqlc.narg('gender')::text);

-- name: ListStudents :many
SELECT id, name, gender, phone, parent_phone, remarks, created_at, updated_at
FROM students
WHERE
    (sqlc.narg('search')::text IS NULL OR
        name ILIKE '%' || sqlc.narg('search')::text || '%' OR
        phone ILIKE '%' || sqlc.narg('search')::text || '%' OR
        parent_phone ILIKE '%' || sqlc.narg('search')::text || '%' OR
        remarks ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('gender')::text IS NULL OR gender = sqlc.narg('gender')::text)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetStudentByID :one
SELECT id, name, gender, phone, parent_phone, remarks, created_at, updated_at
FROM students
WHERE id = $1;

-- name: CreateStudent :one
INSERT INTO students (name, gender, phone, parent_phone, remarks)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, gender, phone, parent_phone, remarks, created_at, updated_at;

-- name: UpdateStudent :one
UPDATE students
SET name = $2, gender = $3, phone = $4, parent_phone = $5, remarks = $6, updated_at = NOW()
WHERE id = $1
RETURNING id, name, gender, phone, parent_phone, remarks, created_at, updated_at;

-- name: DeleteStudent :exec
DELETE FROM students
WHERE id = $1;
