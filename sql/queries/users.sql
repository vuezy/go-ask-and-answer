-- name: CheckIfEmailExists :one
SELECT COUNT(*) FROM users
WHERE email = ?;

-- name: RegisterUser :exec
INSERT INTO users (`name`, email, `password`, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);

-- name: Login :one
SELECT * FROM users
WHERE email = ?;

-- name: GetUserPointsAndCredits :one
SELECT points, credits FROM users
WHERE id = ?;

-- name: AddUserCredit :exec
UPDATE users
SET credits = GREATEST(0, credits + ?), updated_at = ?
WHERE id = ?;

-- name: RemoveUserCredit :exec
UPDATE users
SET credits = GREATEST(0, credits - ?), updated_at = ?
WHERE id = ?;

-- name: UpdateUserPoints :exec
UPDATE users
SET points = points + ?, updated_at = ?
WHERE id = ?;