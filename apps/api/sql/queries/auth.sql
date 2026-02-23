-- Users

-- name: InsertUser :one
INSERT INTO users (id, name, email, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE lower(email) = sqlc.arg(email) AND deleted_at IS NULL;

-- name: UpdateUserEmailVerified :exec
UPDATE users
SET email_verified_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;


-- Accounts

-- name: InsertAccount :one
INSERT INTO accounts (id, user_id, provider, provider_account_id, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING *;

-- name: FindAccountByProvider :one
SELECT * FROM accounts
WHERE provider = $1 AND provider_account_id = $2;

-- name: FindAccountsByUser :many
SELECT * FROM accounts
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: UpdateAccountPassword :exec
UPDATE accounts
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- Sessions

-- name: InsertSession :one
INSERT INTO sessions (id, user_id, token_hash, expires_at, ip_address, user_agent, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- Auth middleware hot path: validates session and loads user in one round-trip.
-- The covering index on token_hash INCLUDE (user_id, expires_at) satisfies the
-- session lookup without a heap fetch; the JOIN then loads the user row.
-- name: FindSessionByTokenHash :one
SELECT
    s.id         AS session_id,
    s.expires_at AS session_expires_at,
    s.created_at AS session_created_at,
    u.id,
    u.name,
    u.email,
    u.email_verified_at,
    u.avatar_url,
    u.status,
    u.onboarded_at,
    u.timezone,
    u.last_login_at,
    u.created_at,
    u.updated_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1
  AND s.expires_at > NOW()
  AND u.deleted_at IS NULL;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- Verification tokens

-- name: InsertVerificationToken :exec
INSERT INTO verification_tokens (id, user_id, email, token_hash, type, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: FindVerificationTokenByHash :one
SELECT * FROM verification_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: MarkVerificationTokenUsed :exec
UPDATE verification_tokens
SET used_at = NOW()
WHERE id = $1;

-- Rate limit: count tokens issued for this email+type in the last hour.
-- Used to prevent abuse of password reset and verification email endpoints.
-- name: CountRecentVerificationTokens :one
SELECT COUNT(*)::int FROM verification_tokens
WHERE lower(email) = sqlc.arg(email)
  AND type = sqlc.arg(type)
  AND created_at > NOW() - INTERVAL '1 hour';
