-- name: CheckIfTokenIsActive :one
SELECT COUNT(*) FROM active_tokens
WHERE sub = ? AND jti = ?;

-- name: SetActiveToken :exec
INSERT INTO active_tokens (sub, jti)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
sub = VALUES(sub), jti = VALUES(jti);