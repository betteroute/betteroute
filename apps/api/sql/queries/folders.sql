-- Folder queries

-- name: InsertFolder :one
INSERT INTO folders (
    id, workspace_id, name, color, position
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: FindFolderByID :one
SELECT * FROM folders
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;

-- name: ListFoldersByWorkspace :many
SELECT * FROM folders
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY position, created_at;

-- name: CountFoldersByWorkspace :one
SELECT COUNT(*) FROM folders
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- name: UpdateFolder :one
UPDATE folders SET
    name       = COALESCE(NULLIF(sqlc.narg('name')::TEXT, ''), name),
    color      = COALESCE(NULLIF(sqlc.narg('color')::TEXT, ''), color),
    position   = CASE WHEN sqlc.arg('set_position')::BOOLEAN THEN sqlc.narg('position') ELSE position END,
    updated_at = NOW()
WHERE id = @id AND workspace_id = @workspace_id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteFolder :execrows
UPDATE folders SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
