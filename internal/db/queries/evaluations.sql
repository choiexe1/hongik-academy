-- name: CreateEvaluation :one
INSERT INTO evaluations (student_id, author_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEvaluationByID :one
SELECT e.*, u.name as author_name, s.name as student_name
FROM evaluations e
JOIN users u ON e.author_id = u.id
JOIN students s ON e.student_id = s.id
WHERE e.id = $1;

-- name: ListEvaluationsByStudent :many
SELECT e.*, u.name as author_name
FROM evaluations e
JOIN users u ON e.author_id = u.id
WHERE e.student_id = $1
    AND (sqlc.narg('search')::text IS NULL OR e.content ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('start_date')::date IS NULL OR e.created_at::date >= sqlc.narg('start_date')::date)
    AND (sqlc.narg('end_date')::date IS NULL OR e.created_at::date <= sqlc.narg('end_date')::date)
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEvaluationsByStudent :one
SELECT COUNT(*) FROM evaluations e
WHERE e.student_id = $1
    AND (sqlc.narg('search')::text IS NULL OR e.content ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('start_date')::date IS NULL OR e.created_at::date >= sqlc.narg('start_date')::date)
    AND (sqlc.narg('end_date')::date IS NULL OR e.created_at::date <= sqlc.narg('end_date')::date);

-- name: UpdateEvaluation :one
UPDATE evaluations
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEvaluation :exec
DELETE FROM evaluations WHERE id = $1;
