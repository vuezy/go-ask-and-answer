-- name: SearchQuestions :many
SELECT id, title, body, `priority_level`, closed, updated_at FROM questions
WHERE title LIKE ?
ORDER BY closed ASC, `priority_level` DESC, responded_at ASC, updated_at DESC;

-- name: GetQuestionsByUserId :many
SELECT id, title, body, `priority_level`, closed, updated_at FROM questions
WHERE user_id = ?
ORDER BY closed ASC, `priority_level` DESC, responded_at ASC, updated_at DESC;

-- name: GetQuestionById :one
SELECT questions.*, `name`, email FROM questions
INNER JOIN users ON users.id = questions.user_id
WHERE questions.id = ?;

-- name: CreateQuestion :exec
INSERT INTO questions (title, body, `priority_level`, user_id, responded_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateQuestion :exec
UPDATE questions
SET title = ?, body = ?, `priority_level` = `priority_level` + ?, updated_at = ?
WHERE id = ? AND user_id = ? AND closed != 1;

-- name: DeleteQuestion :exec
DELETE FROM questions
WHERE id = ? AND user_id = ? AND closed != 1;

-- name: CloseQuestion :exec
UPDATE questions
SET closed = 1, updated_at = ?
WHERE id = ? AND user_id = ?;

-- name: RespondToQuestion :exec
UPDATE questions
SET responded_at = ?, updated_at = ?
WHERE id = ?;