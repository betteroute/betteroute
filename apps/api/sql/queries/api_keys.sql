-- name: InsertAPIKey :one
INSERT INTO api_keys (
    id, workspace_id, created_by, name, key_hash, key_prefix, permission, scopes, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: FindAPIKeyByHash :one
-- Auth hot path: resolve key hash -> API key.
SELECT * FROM api_keys
WHERE key_hash = $1 AND deleted_at IS NULL;

-- name: FindAPIKeyWithCreator :one
-- Auth hot path: resolve key hash -> API key + creator in a single round-trip.
SELECT
    k.id, k.workspace_id, k.created_by, k.name, k.key_prefix, k.permission, k.scopes,
    k.expires_at, k.last_used_at, k.created_at, k.updated_at,
    u.id AS user_id, u.name AS user_name, u.email AS user_email,
    u.email_verified_at, u.avatar_url AS user_avatar_url, u.status AS user_status,
    u.onboarded_at, u.last_login_at, u.timezone, u.created_at AS user_created_at, u.updated_at AS user_updated_at
FROM api_keys k
JOIN users u ON u.id = k.created_by AND u.deleted_at IS NULL
WHERE k.key_hash = $1 AND k.deleted_at IS NULL;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindAPIKeyByID :one
SELECT * FROM api_keys
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: ListAPIKeysByWorkspace :many
SELECT * FROM api_keys
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SoftDeleteAPIKey :execrows
UPDATE api_keys SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
