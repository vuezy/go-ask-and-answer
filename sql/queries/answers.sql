-- name: GetAnswersByQuestionId :many
SELECT answers.*, `name`, email FROM answers
INNER JOIN users ON users.id = answers.user_id
WHERE question_id = ?
ORDER BY votes DESC, answers.updated_at ASC;

-- name: GetAnswersByUserId :many
SELECT * FROM answers
WHERE user_id = ?
ORDER BY updated_at DESC;

-- name: GetAnswerById :one
SELECT * FROM answers
WHERE id = ?;

-- name: CreateAnswer :exec
INSERT INTO answers (body, question_id, user_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateAnswer :exec
UPDATE answers
SET body = ?, updated_at = ?
WHERE id = ? AND user_id = ?;

-- name: DeleteAnswer :exec
DELETE FROM answers
WHERE id = ? AND user_id = ?;

-- name: UpdateAnswerVotes :exec
UPDATE answers
SET votes = votes + ?, updated_at = ?
WHERE id = ?;