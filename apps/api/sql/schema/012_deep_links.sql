-- Deep Links
-- Per-link resolved deep link data. Separate table to keep links lean —
-- most links won't have deep link data. 1:1 relationship with links.
-- Auto-populated at link creation when dest_url matches a known app.

CREATE TABLE deep_links (
    link_id              TEXT PRIMARY KEY REFERENCES links(id) ON DELETE CASCADE,
    platform_app_id      TEXT REFERENCES platform_apps(id) ON DELETE SET NULL,
    workspace_app_id     TEXT REFERENCES workspace_apps(id) ON DELETE SET NULL,

    -- Resolved deep link URLs (stored at creation, no computation on redirect)
    ios_deep_link        TEXT,
    android_deep_link    TEXT,
    ios_fallback_url     TEXT,
    android_fallback_url TEXT,

    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT deep_links_ios_deep_link_length
        CHECK (ios_deep_link IS NULL OR char_length(ios_deep_link) <= 2048),

    CONSTRAINT deep_links_android_deep_link_length
        CHECK (android_deep_link IS NULL OR char_length(android_deep_link) <= 2048),

    CONSTRAINT deep_links_ios_fallback_url_length
        CHECK (ios_fallback_url IS NULL OR char_length(ios_fallback_url) <= 2048),

    CONSTRAINT deep_links_android_fallback_url_length
        CHECK (android_fallback_url IS NULL OR char_length(android_fallback_url) <= 2048)
);
