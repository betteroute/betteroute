-- name: InsertTag :one
INSERT INTO tags (
    id, workspace_id, created_by, name, color
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: FindTagByID :one
SELECT * FROM tags
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: ListTagsByWorkspace :many
SELECT * FROM tags
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: CountTagsByWorkspace :one
SELECT COUNT(*) FROM tags
WHERE workspace_id = $1 AND deleted_at IS NULL;


-- name: SoftDeleteTag :execrows
UPDATE tags SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- Link-tag association

-- name: AddTagToLink :exec
INSERT INTO link_tags (link_id, tag_id)
VALUES ($1, $2)
ON CONFLICT (link_id, tag_id) DO NOTHING;

-- name: RemoveTagFromLink :execrows
DELETE FROM link_tags
WHERE link_id = $1 AND tag_id = $2;

-- name: ListTagsByLink :many
SELECT t.* FROM tags t
JOIN link_tags lt ON lt.tag_id = t.id
WHERE lt.link_id = $1 AND t.deleted_at IS NULL
ORDER BY t.name;

-- name: ListLinkIDsByTag :many
SELECT lt.link_id FROM link_tags lt
JOIN tags t ON t.id = lt.tag_id
WHERE lt.tag_id = $1 AND t.deleted_at IS NULL;

-- name: SetLinkTags :exec
-- Replace all tags on a link. Called within a transaction:
-- 1. DELETE FROM link_tags WHERE link_id = $1
-- 2. INSERT each new tag
-- This query handles step 1.
DELETE FROM link_tags WHERE link_id = $1;
