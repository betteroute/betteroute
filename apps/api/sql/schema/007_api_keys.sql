-- API keys
-- Programmatic access to the API. Each key is scoped to one workspace.
-- The raw key is shown once on creation; only the hash is stored.
-- Format: br_live_<random> (live) or br_test_<random> (test).

CREATE TABLE api_keys (
    id            TEXT PRIMARY KEY,
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by    TEXT REFERENCES users(id) ON DELETE SET NULL,

    name          TEXT NOT NULL,                -- human label ("CI/CD", "Marketing Tool")
    key_hash      TEXT NOT NULL,                -- SHA-256 of the raw key
    key_prefix    TEXT NOT NULL,                -- first 8 chars for identification ("br_live_")
    permission    TEXT NOT NULL DEFAULT 'all',  -- "all" | "read_only" | "restricted"
    scopes        TEXT[] NOT NULL DEFAULT '{}', -- e.g. {"links:read","links:write"}
    expires_at    TIMESTAMPTZ,                  -- NULL = never expires
    last_used_at  TIMESTAMPTZ,

    deleted_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT api_keys_name_length CHECK (char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT api_keys_permission_valid CHECK (permission IN ('all', 'read_only', 'restricted')),
    CONSTRAINT api_keys_scopes_required CHECK (permission != 'restricted' OR array_length(scopes, 1) > 0)
);

-- Auth hot path: lookup by key hash, return workspace + scoping without heap fetch.
CREATE UNIQUE INDEX idx_api_keys_hash
    ON api_keys(key_hash)
    INCLUDE (workspace_id, created_by, expires_at, permission, scopes)
    WHERE deleted_at IS NULL;

-- List keys for a workspace ("API Keys" settings page).
CREATE INDEX idx_api_keys_workspace
    ON api_keys(workspace_id)
    WHERE deleted_at IS NULL;
