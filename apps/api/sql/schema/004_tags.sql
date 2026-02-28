-- Tags
-- Cross-cutting labels for links. A link can have many tags.
-- Every tag belongs to exactly one workspace.

CREATE TABLE tags (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      TEXT,                         -- FK in 005_auth.sql

    name            TEXT NOT NULL,
    color           TEXT NOT NULL DEFAULT '#6366f1',

    -- Timestamps
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT tags_name_length
        CHECK (char_length(name) BETWEEN 1 AND 50),

    CONSTRAINT tags_color_hex
        CHECK (color ~ '^#[0-9a-fA-F]{6}$')
);

-- Case-insensitive unique name per workspace. Prevents "Social" + "social" dupes.
CREATE UNIQUE INDEX idx_tags_workspace_name
    ON tags(workspace_id, LOWER(name))
    WHERE deleted_at IS NULL;

-- Listing by workspace.
CREATE INDEX idx_tags_workspace
    ON tags(workspace_id)
    WHERE deleted_at IS NULL;


-- Link ↔ Tag (many-to-many).
-- CASCADE both sides: deleting a link or tag removes the association.

CREATE TABLE link_tags (
    link_id         TEXT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    tag_id          TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (link_id, tag_id)
);

-- Reverse lookup: find links by tag.
CREATE INDEX idx_link_tags_tag ON link_tags(tag_id);
