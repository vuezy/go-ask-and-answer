-- name: CheckIfVoteExists :one
SELECT id, val FROM votes
WHERE answer_id = ? AND user_id = ?;

-- name: CreateVote :exec
INSERT INTO votes (val, answer_id, user_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);

-- name: Upvote :exec
UPDATE votes
SET val = 1, updated_at = ?
WHERE id = ?;

-- name: Downvote :exec
UPDATE votes
SET val = -1, updated_at = ?
WHERE id = ?;