-- Link queries

-- name: InsertLink :one
INSERT INTO links (
    id, workspace_id, short_code, dest_url, title, description, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: FindLinkByID :one
SELECT * FROM links
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: FindLinkByShortCode :one
-- Redirect hot path. Uses covering index idx_links_redirect.
SELECT * FROM links
WHERE short_code = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListLinksByWorkspace :many
SELECT * FROM links
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountLinksByWorkspace :one
SELECT COUNT(*) FROM links
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: UpdateLink :one
UPDATE links SET
    dest_url = COALESCE(sqlc.narg('dest_url'), dest_url),
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    expires_at = CASE WHEN sqlc.arg('set_expires_at')::BOOLEAN THEN sqlc.narg('expires_at') ELSE expires_at END,
    updated_at = NOW()
WHERE id = @id AND workspace_id = @workspace_id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteLink :execrows
UPDATE links SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
