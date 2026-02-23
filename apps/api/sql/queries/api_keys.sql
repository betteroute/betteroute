-- name: InsertAPIKey :one
INSERT INTO api_keys (
    id, workspace_id, created_by, name, key_hash, key_prefix, permission, scopes, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: FindAPIKeyByHash :one
-- Auth hot path: resolve key hash → workspace + scoping + creator.
-- Index-only scan via idx_api_keys_hash.
SELECT id, workspace_id, created_by, permission, scopes, expires_at
FROM api_keys
WHERE key_hash = $1 AND deleted_at IS NULL;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = NOW()
WHERE id = $1;

-- name: FindAPIKeyByID :one
SELECT * FROM api_keys
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: ListAPIKeysByWorkspace :many
SELECT * FROM api_keys
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CountAPIKeysByWorkspace :one
SELECT COUNT(*) FROM api_keys
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteAPIKey :execrows
UPDATE api_keys SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
