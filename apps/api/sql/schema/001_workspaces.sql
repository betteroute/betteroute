-- Workspaces
-- Multi-tenancy boundary. Every resource belongs to exactly one workspace.
-- A workspace is the unit of billing, permissions, and team collaboration.

CREATE TABLE workspaces (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL,

    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workspaces_name_length
        CHECK (char_length(name) BETWEEN 1 AND 100),

    CONSTRAINT workspaces_slug_length
        CHECK (char_length(slug) BETWEEN 1 AND 50),

    CONSTRAINT workspaces_slug_format
        CHECK (slug ~ '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$')
);

-- Unique slug for active workspaces. Partial index allows reuse after delete.
CREATE UNIQUE INDEX idx_workspaces_slug_active
    ON workspaces(slug)
    WHERE deleted_at IS NULL;
