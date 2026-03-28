-- name: InsertWorkspaceApp :one
INSERT INTO workspace_apps (
    id, workspace_id, created_by, name, platform,
    bundle_id, team_id, app_store_url,
    package_name, sha256_fingerprints, play_store_url,
    scheme
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12
) RETURNING *;

-- name: FindWorkspaceAppByID :one
SELECT * FROM workspace_apps
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: ListWorkspaceAppsByWorkspace :many
SELECT * FROM workspace_apps
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: ListWorkspaceAppsByPlatform :many
-- "platform" here means the target OS ("ios" or "android"), not platform apps.
-- Used by AASA/assetlinks to fetch only the relevant OS's apps for a domain.
SELECT * FROM workspace_apps
WHERE workspace_id = $1 AND platform = $2 AND deleted_at IS NULL
ORDER BY name;

-- name: CountWorkspaceAppsByWorkspace :one
SELECT COUNT(*) FROM workspace_apps
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteWorkspaceApp :execrows
UPDATE workspace_apps SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
