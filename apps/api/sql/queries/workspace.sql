-- name: InsertWorkspace :one
INSERT INTO workspaces (id, name, slug)
VALUES ($1, $2, $3)
RETURNING *;

-- name: FindWorkspaceByID :one
SELECT * FROM workspaces
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindWorkspaceBySlug :one
SELECT * FROM workspaces
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListWorkspacesByUser :many
SELECT w.*, wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1 AND w.deleted_at IS NULL
ORDER BY wm.created_at ASC;

-- name: SoftDeleteWorkspace :exec
UPDATE workspaces
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;


-- Workspace members

-- Middleware hot path: resolve workspace and member role in one round-trip.
-- name: FindWorkspaceBySlugAndMember :one
SELECT
    w.id, w.name, w.slug, w.deleted_at, w.created_at, w.updated_at,
    wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.user_id = sqlc.arg(user_id)
WHERE w.slug = sqlc.arg(slug) AND w.deleted_at IS NULL;

-- name: InsertWorkspaceMember :exec
INSERT INTO workspace_members (workspace_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4);

-- name: FindWorkspaceMember :one
SELECT * FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: ListWorkspaceMembers :many
SELECT
    wm.workspace_id,
    wm.user_id,
    wm.role,
    wm.invited_by,
    wm.created_at,
    wm.updated_at,
    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.created_at ASC;

-- name: UpdateWorkspaceMemberRole :exec
UPDATE workspace_members
SET role = $3, updated_at = NOW()
WHERE workspace_id = $1 AND user_id = $2;

-- name: DeleteWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: CountWorkspaceOwners :one
SELECT COUNT(*)::int FROM workspace_members
WHERE workspace_id = $1 AND role = 'owner';


-- Workspace invitations

-- name: InsertWorkspaceInvitation :one
INSERT INTO workspace_invitations (id, workspace_id, email, role, token_hash, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: FindWorkspaceInvitationByTokenHash :one
SELECT * FROM workspace_invitations
WHERE token_hash = $1
  AND accepted_at IS NULL
  AND expires_at > NOW();

-- name: ListWorkspaceInvitations :many
SELECT * FROM workspace_invitations
WHERE workspace_id = $1 AND accepted_at IS NULL
ORDER BY created_at DESC;

-- name: AcceptWorkspaceInvitation :exec
UPDATE workspace_invitations
SET accepted_at = NOW()
WHERE id = $1;

-- name: DeleteWorkspaceInvitation :exec
DELETE FROM workspace_invitations
WHERE id = $1 AND workspace_id = $2;
