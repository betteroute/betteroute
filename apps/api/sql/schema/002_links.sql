-- Links
-- Core of the product. Maps a short_code to a destination URL.
-- Every link belongs to exactly one workspace.

CREATE TABLE links (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    short_code      TEXT NOT NULL,
    dest_url        TEXT NOT NULL,
    title           TEXT,
    description     TEXT,

    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,

    -- Denormalized counters updated async to avoid COUNT on every read.
    click_count     BIGINT NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMPTZ,

    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT links_short_code_format
        CHECK (short_code ~ '^[a-zA-Z0-9_-]+$'),

    CONSTRAINT links_short_code_length
        CHECK (char_length(short_code) BETWEEN 1 AND 50),

    CONSTRAINT links_dest_url_length
        CHECK (char_length(dest_url) <= 2048),

    CONSTRAINT links_title_length
        CHECK (title IS NULL OR char_length(title) <= 200),

    CONSTRAINT links_description_length
        CHECK (description IS NULL OR char_length(description) <= 500),

    CONSTRAINT links_click_count_non_negative
        CHECK (click_count >= 0)
);

-- Unique short_code for active links. Allows reuse after soft delete.
CREATE UNIQUE INDEX idx_links_short_code_active
    ON links(short_code)
    WHERE deleted_at IS NULL;

-- Redirect hot path. Covering index — no heap fetch needed.
CREATE INDEX idx_links_redirect
    ON links(short_code)
    INCLUDE (dest_url, is_active, expires_at)
    WHERE deleted_at IS NULL;

-- Dashboard listing by workspace, newest first.
CREATE INDEX idx_links_workspace
    ON links(workspace_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- Background job: find expired links for deactivation.
CREATE INDEX idx_links_expiring
    ON links(expires_at)
    WHERE expires_at IS NOT NULL AND is_active = TRUE AND deleted_at IS NULL;
