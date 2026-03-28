-- Workspace Apps
-- User-managed app configurations for deep linking. Each workspace can register
-- their own iOS/Android apps with custom schemes, bundle IDs, and store URLs.
-- Used when platform_apps auto-detect doesn't cover the user's app.

CREATE TABLE workspace_apps (
    id                    TEXT PRIMARY KEY,
    workspace_id          TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by            TEXT REFERENCES users(id) ON DELETE SET NULL,

    name                  TEXT NOT NULL,
    platform              TEXT NOT NULL,        -- "ios" or "android"

    -- iOS fields
    bundle_id             TEXT,
    team_id               TEXT,
    app_store_url         TEXT,

    -- Android fields
    package_name          TEXT,
    sha256_fingerprints   TEXT[],
    play_store_url        TEXT,

    -- Deep link scheme (e.g. "myapp://{path}")
    scheme                TEXT,

    deleted_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workspace_apps_name_length
        CHECK (char_length(name) BETWEEN 1 AND 100),

    CONSTRAINT workspace_apps_platform_check
        CHECK (platform IN ('ios', 'android')),

    CONSTRAINT workspace_apps_scheme_length
        CHECK (scheme IS NULL OR char_length(scheme) <= 500),

    CONSTRAINT workspace_apps_bundle_id_length
        CHECK (bundle_id IS NULL OR char_length(bundle_id) <= 255),

    CONSTRAINT workspace_apps_package_name_length
        CHECK (package_name IS NULL OR char_length(package_name) <= 255),

    CONSTRAINT workspace_apps_app_store_url_length
        CHECK (app_store_url IS NULL OR char_length(app_store_url) <= 2048),

    CONSTRAINT workspace_apps_play_store_url_length
        CHECK (play_store_url IS NULL OR char_length(play_store_url) <= 2048)
);

-- List workspace apps for a workspace.
CREATE INDEX idx_workspace_apps_workspace
    ON workspace_apps(workspace_id)
    WHERE deleted_at IS NULL;

