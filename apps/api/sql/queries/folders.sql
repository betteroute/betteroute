-- name: InsertFolder :one
INSERT INTO folders (
    id, workspace_id, created_by, name, color, position
) VALUES (
    $1, $2, $3, $4, $5, $6
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


-- name: SoftDeleteFolder :execrows
UPDATE folders SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL;
