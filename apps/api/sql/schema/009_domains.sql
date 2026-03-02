-- Domains
-- Custom domains for branded short links. Each domain belongs to one workspace.
-- Verification via DNS TXT record on _betteroute.<hostname> proves ownership.
-- Short codes are unique per domain (see 002_links.sql).

CREATE TABLE domains (
    id                  TEXT PRIMARY KEY,
    workspace_id        TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by          TEXT REFERENCES users(id) ON DELETE SET NULL,

    hostname            TEXT NOT NULL,
    verification_token  TEXT NOT NULL,
    verified_at         TIMESTAMPTZ,

    -- Where to redirect when a short code doesn't exist on this domain.
    fallback_url        TEXT,

    status              TEXT NOT NULL DEFAULT 'pending',
    last_checked_at     TIMESTAMPTZ,

    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT domains_hostname_length
        CHECK (char_length(hostname) BETWEEN 4 AND 253),

    CONSTRAINT domains_hostname_format
        CHECK (hostname ~ '^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$'),

    CONSTRAINT domains_status_check
        CHECK (status IN ('pending', 'active', 'suspended')),

    CONSTRAINT domains_fallback_url_length
        CHECK (fallback_url IS NULL OR char_length(fallback_url) <= 2048)
);

-- Redirect hot path: Host header → domain lookup.
-- Covering index returns id, workspace_id, fallback_url without heap fetch.
CREATE UNIQUE INDEX idx_domains_hostname_active
    ON domains(hostname)
    INCLUDE (id, workspace_id, fallback_url)
    WHERE deleted_at IS NULL;

-- List domains for a workspace ("Domains" settings page).
CREATE INDEX idx_domains_workspace
    ON domains(workspace_id)
    WHERE deleted_at IS NULL;

-- FK from links.domain_id → domains.id (defined here because domains must exist first).
ALTER TABLE links
    ADD CONSTRAINT fk_links_domain
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE SET NULL;
