-- Link queries

-- name: InsertLink :one
INSERT INTO links (
    id, workspace_id, short_code, dest_url, title, description,
    starts_at, expires_at, expiration_url, max_clicks,
    utm_source, utm_medium, utm_campaign, utm_term, utm_content,
    og_title, og_description, og_image,
    notes, created_via
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $13, $14, $15,
    $16, $17, $18,
    $19, $20
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
    starts_at = CASE WHEN sqlc.arg('set_starts_at')::BOOLEAN THEN sqlc.narg('starts_at') ELSE starts_at END,
    expires_at = CASE WHEN sqlc.arg('set_expires_at')::BOOLEAN THEN sqlc.narg('expires_at') ELSE expires_at END,
    expiration_url = COALESCE(sqlc.narg('expiration_url'), expiration_url),
    max_clicks = CASE WHEN sqlc.arg('set_max_clicks')::BOOLEAN THEN sqlc.narg('max_clicks') ELSE max_clicks END,
    utm_source = COALESCE(sqlc.narg('utm_source'), utm_source),
    utm_medium = COALESCE(sqlc.narg('utm_medium'), utm_medium),
    utm_campaign = COALESCE(sqlc.narg('utm_campaign'), utm_campaign),
    utm_term = COALESCE(sqlc.narg('utm_term'), utm_term),
    utm_content = COALESCE(sqlc.narg('utm_content'), utm_content),
    og_title = COALESCE(sqlc.narg('og_title'), og_title),
    og_description = COALESCE(sqlc.narg('og_description'), og_description),
    og_image = COALESCE(sqlc.narg('og_image'), og_image),
    notes = COALESCE(sqlc.narg('notes'), notes),
    updated_at = NOW()
WHERE id = @id AND workspace_id = @workspace_id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteLink :execrows
UPDATE links SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: IncrementClickCount :exec
-- Called async after redirect. Simple increment until batching is added.
UPDATE links SET
    click_count = click_count + 1,
    last_clicked_at = NOW()
WHERE id = $1;
