-- Platform Apps
-- Catalog of known apps with deep link schemes and store metadata.
-- Maintained by the platform; users can request additions.
-- Used to auto-detect apps from dest_url and resolve deep link URLs.

CREATE TABLE platform_apps (
    id                TEXT PRIMARY KEY,     -- e.g. "app_youtube"

    name              TEXT NOT NULL,
    icon_url          TEXT,

    -- URL patterns to match against dest_url (e.g. "youtube.com", "youtu.be")
    url_patterns      TEXT[] NOT NULL,

    -- Deep link URI schemes
    ios_scheme        TEXT,                 -- e.g. "youtube://{path}"
    android_scheme    TEXT,                 -- e.g. "vnd.youtube://{path}"

    -- iOS metadata (for AASA / Universal Links)
    ios_app_id        TEXT,                 -- App Store ID
    ios_bundle_id     TEXT,
    ios_team_id       TEXT,

    -- Android metadata (for assetlinks / App Links)
    android_package   TEXT,                 -- e.g. "com.google.android.youtube"
    android_sha256    TEXT[],               -- assetlinks fingerprints

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT platform_apps_name_length
        CHECK (char_length(name) BETWEEN 1 AND 100),

    CONSTRAINT platform_apps_url_patterns_not_empty
        CHECK (array_length(url_patterns, 1) > 0)
);

-- Lookup by URL pattern (GIN index for array contains).
CREATE INDEX idx_platform_apps_url_patterns
    ON platform_apps USING GIN (url_patterns);
