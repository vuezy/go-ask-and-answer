// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: active_tokens.sql

package database

import (
	"context"
)

const checkIfTokenIsActive = `-- name: CheckIfTokenIsActive :one
SELECT COUNT(*) FROM active_tokens
WHERE sub = ? AND jti = ?
`

type CheckIfTokenIsActiveParams struct {
	Sub int32
	Jti string
}

func (q *Queries) CheckIfTokenIsActive(ctx context.Context, arg CheckIfTokenIsActiveParams) (int64, error) {
	row := q.queryRow(ctx, q.checkIfTokenIsActiveStmt, checkIfTokenIsActive, arg.Sub, arg.Jti)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const setActiveToken = `-- name: SetActiveToken :exec
INSERT INTO active_tokens (sub, jti)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
sub = VALUES(sub), jti = VALUES(jti)
`

type SetActiveTokenParams struct {
	Sub int32
	Jti string
}

func (q *Queries) SetActiveToken(ctx context.Context, arg SetActiveTokenParams) error {
	_, err := q.exec(ctx, q.setActiveTokenStmt, setActiveToken, arg.Sub, arg.Jti)
	return err
}
