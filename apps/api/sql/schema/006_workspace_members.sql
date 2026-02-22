-- Workspace membership and invitations.
-- Bridges auth (identity) with workspaces (multi-tenancy).

-- Roles:
--   owner  — full control, billing, delete/transfer workspace
--   admin  — manage members (not owner), manage settings (domains, API keys, webhooks)
--   member — CRUD on links, folders, tags
--   viewer — read-only access, view analytics (agency clients, executives)

-- Workspace members (composite PK — join table, no synthetic ID needed)

CREATE TABLE workspace_members (
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL DEFAULT 'member',
    invited_by    TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (workspace_id, user_id),
    CONSTRAINT workspace_members_role_check CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

-- Find all workspaces for a user.
CREATE INDEX idx_workspace_members_user
    ON workspace_members(user_id);


-- Workspace invitations
-- Supports inviting by email before the user has an account.
-- On signup, pending invitations for that email are auto-accepted.

CREATE TABLE workspace_invitations (
    id            TEXT PRIMARY KEY,
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email         TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'member',
    token_hash    TEXT NOT NULL,
    invited_by    TEXT REFERENCES users(id) ON DELETE SET NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    accepted_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workspace_invitations_role_check CHECK (role IN ('admin', 'member', 'viewer'))
);

-- Accept invite via link (hot path).
CREATE INDEX idx_workspace_invitations_token
    ON workspace_invitations(token_hash)
    WHERE accepted_at IS NULL;

-- One pending invitation per email per workspace.
CREATE UNIQUE INDEX idx_workspace_invitations_pending
    ON workspace_invitations(workspace_id, email)
    WHERE accepted_at IS NULL;

-- Find pending invitations on signup (auto-join workspaces).
CREATE INDEX idx_workspace_invitations_email
    ON workspace_invitations(email)
    WHERE accepted_at IS NULL;
