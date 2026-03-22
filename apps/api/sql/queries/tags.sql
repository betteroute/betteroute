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

-- name: SoftDeleteTag :execrows
UPDATE tags SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- Link-tag association

-- name: AddTagToLink :exec
INSERT INTO link_tags (link_id, tag_id)
SELECT l.id, t.id
FROM links l, tags t
WHERE l.id = sqlc.arg(link_id) AND t.id = sqlc.arg(tag_id)
  AND l.workspace_id = sqlc.arg(workspace_id) AND t.workspace_id = sqlc.arg(workspace_id)
  AND l.deleted_at IS NULL AND t.deleted_at IS NULL
ON CONFLICT (link_id, tag_id) DO NOTHING;

-- name: RemoveTagFromLink :execrows
DELETE FROM link_tags lt
USING links l, tags t
WHERE lt.link_id = l.id AND lt.tag_id = t.id
  AND lt.link_id = $1 AND lt.tag_id = $2
  AND l.workspace_id = $3 AND t.workspace_id = $3;

-- name: ListTagsByLink :many
SELECT t.* FROM tags t
JOIN link_tags lt ON lt.tag_id = t.id
JOIN links l ON l.id = lt.link_id
WHERE lt.link_id = $1 AND l.workspace_id = $2
  AND t.deleted_at IS NULL AND l.deleted_at IS NULL
ORDER BY t.name;


