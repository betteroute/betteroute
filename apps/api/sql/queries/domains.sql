-- name: InsertDomain :one
INSERT INTO domains (
    id, workspace_id, created_by, hostname, verification_token, fallback_url
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: FindDomainByID :one
SELECT * FROM domains
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: FindDomainByHostname :one
SELECT * FROM domains
WHERE hostname = $1 AND deleted_at IS NULL;

-- name: ResolveDomain :one
-- Redirect hot path: resolve hostname → domain_id + workspace_id + fallback_url.
-- Uses the covering index on hostname.
SELECT id, workspace_id, fallback_url FROM domains
WHERE hostname = $1 AND deleted_at IS NULL;

-- name: ListDomainsByWorkspace :many
SELECT * FROM domains
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CountDomainsByWorkspace :one
SELECT COUNT(*) FROM domains
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: UpdateDomainStatus :one
UPDATE domains SET
    status = $3,
    verified_at = $4,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateDomainLastChecked :exec
UPDATE domains SET
    last_checked_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteDomain :execrows
UPDATE domains SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
