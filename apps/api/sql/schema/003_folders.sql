-- Folders
-- Flat grouping for links. A link belongs to one folder (or none).
-- Every folder belongs to exactly one workspace.

CREATE TABLE folders (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      TEXT REFERENCES users(id) ON DELETE SET NULL,

    name            TEXT NOT NULL,
    color           TEXT NOT NULL DEFAULT '#6366f1',
    position        INTEGER NOT NULL DEFAULT 0,           -- sidebar ordering

    -- Timestamps
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT folders_name_length
        CHECK (char_length(name) BETWEEN 1 AND 100),

    CONSTRAINT folders_color_hex
        CHECK (color ~ '^#[0-9a-fA-F]{6}$'),

    CONSTRAINT folders_position_non_negative
        CHECK (position >= 0)
);

-- Unique name per workspace. Allows reuse after soft delete.
CREATE UNIQUE INDEX idx_folders_workspace_name
    ON folders(workspace_id, name)
    WHERE deleted_at IS NULL;

-- Sidebar listing by workspace, ordered by position.
CREATE INDEX idx_folders_workspace
    ON folders(workspace_id, position)
    WHERE deleted_at IS NULL;

-- FK from links.folder_id → folders.id (defined here because folders must exist first).
ALTER TABLE links
    ADD CONSTRAINT fk_links_folder
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL;
