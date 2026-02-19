-- Link queries

-- name: InsertLink :one
INSERT INTO links (
    id, workspace_id, folder_id, short_code, dest_url, title, description,
    starts_at, expires_at, expiration_url, max_clicks,
    utm_source, utm_medium, utm_campaign, utm_term, utm_content,
    og_title, og_description, og_image,
    notes, created_via
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, $15, $16,
    $17, $18, $19,
    $20, $21
) RETURNING *;

-- name: FindLinkByID :one
SELECT * FROM links
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: FindLinkByShortCode :one
-- Used by link CRUD lookups. Not on the redirect hot path (see ResolveLink).
SELECT * FROM links
WHERE short_code = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ResolveLink :one
-- Redirect hot path: atomic increment + return in one round-trip.
-- Returns no rows when the link is invalid (deleted, inactive, expired, not started, or click-limited).
UPDATE links SET
    click_count = click_count + 1,
    last_clicked_at = NOW()
WHERE short_code = $1
    AND deleted_at IS NULL
    AND is_active = TRUE
    AND (starts_at IS NULL OR starts_at <= NOW())
    AND (expires_at IS NULL OR expires_at > NOW())
    AND (max_clicks IS NULL OR click_count < max_clicks)
RETURNING id, dest_url,
    utm_source, utm_medium, utm_campaign, utm_term, utm_content,
    og_title, og_description, og_image;

-- name: FindRedirectFallback :one
-- Slim diagnostic query: only called when ResolveLink returns 0 rows.
SELECT is_active, starts_at, expires_at, expiration_url, max_clicks, click_count
FROM links
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

-- name: SoftDeleteLink :execrows
UPDATE links SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

