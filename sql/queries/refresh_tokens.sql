-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES ($1, $2, $2, $3, NULL, $4)
RETURNING *;
