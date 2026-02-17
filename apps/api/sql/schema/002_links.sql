-- Links
-- Core of the product. Maps a short_code to a destination URL.
-- Every link belongs to exactly one workspace.

CREATE TABLE links (
    id              TEXT PRIMARY KEY,
    workspace_id    TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Core
    short_code      TEXT NOT NULL,
    dest_url        TEXT NOT NULL,
    title           TEXT,
    description     TEXT,

    -- Status & Scheduling
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at       TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    expiration_url  TEXT,                         -- redirect here after expiry

    -- Click limits
    max_clicks      INTEGER,                      -- link expires after N clicks (1 = single use)

    -- UTM parameters (flat — queryable, indexable)
    utm_source      TEXT,
    utm_medium      TEXT,
    utm_campaign    TEXT,
    utm_term        TEXT,
    utm_content     TEXT,

    -- OG metadata overrides
    og_title        TEXT,
    og_description  TEXT,
    og_image        TEXT,

    -- Denormalized counters updated async to avoid COUNT on every read.
    click_count         BIGINT NOT NULL DEFAULT 0,
    unique_click_count  BIGINT NOT NULL DEFAULT 0,
    last_clicked_at     TIMESTAMPTZ,

    -- Internal
    notes           TEXT,
    created_via     TEXT NOT NULL DEFAULT 'web',   -- web, api, import

    -- Timestamps
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
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

    CONSTRAINT links_expiration_url_length
        CHECK (expiration_url IS NULL OR char_length(expiration_url) <= 2048),

    CONSTRAINT links_notes_length
        CHECK (notes IS NULL OR char_length(notes) <= 5000),

    CONSTRAINT links_max_clicks_positive
        CHECK (max_clicks IS NULL OR max_clicks > 0),

    CONSTRAINT links_click_count_non_negative
        CHECK (click_count >= 0),

    CONSTRAINT links_unique_click_count_non_negative
        CHECK (unique_click_count >= 0),

    CONSTRAINT links_created_via_check
        CHECK (created_via IN ('web', 'api', 'import'))
);

-- Unique short_code for active links. Allows reuse after soft delete.
CREATE UNIQUE INDEX idx_links_short_code_active
    ON links(short_code)
    WHERE deleted_at IS NULL;

-- Redirect hot path. Covering index — no heap fetch needed.
-- Only includes columns checked during redirect resolution.
CREATE INDEX idx_links_redirect
    ON links(short_code)
    INCLUDE (dest_url, is_active, starts_at, expires_at, expiration_url, max_clicks, click_count)
    WHERE deleted_at IS NULL;

-- Dashboard listing by workspace, newest first.
CREATE INDEX idx_links_workspace
    ON links(workspace_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- Background job: find expired links for deactivation.
CREATE INDEX idx_links_expiring
    ON links(expires_at)
    WHERE expires_at IS NOT NULL AND is_active = TRUE AND deleted_at IS NULL;
