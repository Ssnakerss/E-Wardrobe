-- name: CreateUser :one
INSERT INTO users (
    id, email, password_hash, status
) VALUES (
    $1, $2, $3, 'active'
)
ON CONFLICT (email) DO NOTHING
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateInvitation :one
INSERT INTO invitations (
    id, token_hash, email, expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetInvitationByTokenHash :one
SELECT * FROM invitations
WHERE token_hash = $1
LIMIT 1;

-- name: MarkInvitationUsed :one
UPDATE invitations
SET used_at = now(), used_by_user_id = $2
WHERE id = $1 AND used_at IS NULL AND revoked_at IS NULL
RETURNING *;

-- name: RevokeInvitation :one
UPDATE invitations
SET revoked_at = now()
WHERE id = $1 AND used_at IS NULL AND revoked_at IS NULL
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (
    id, user_id, token_hash, csrf_token_hash, client, device_name, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > now()
LIMIT 1;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = now()
WHERE id = $1
RETURNING *;

-- name: RevokeUserSessions :many
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL
RETURNING id;

-- name: UpdateSessionLastUsed :one
UPDATE sessions
SET last_used_at = now()
WHERE id = $1 AND revoked_at IS NULL
RETURNING *;

-- name: CreateRecoveryToken :one
INSERT INTO account_recovery_tokens (
    id, user_id, token_hash, expires_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetRecoveryTokenByHash :one
SELECT * FROM account_recovery_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > now()
LIMIT 1;

-- name: MarkRecoveryTokenUsed :one
UPDATE account_recovery_tokens
SET used_at = now()
WHERE id = $1 AND used_at IS NULL
RETURNING *;
