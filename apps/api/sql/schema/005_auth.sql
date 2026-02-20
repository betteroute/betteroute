-- Auth tables: users, accounts, sessions, verification tokens.
-- Identity only — workspace membership lives in 006_workspace_members.sql.

-- Users
-- Soft delete: deleted_at IS NOT NULL means user requested account deletion.
-- Status: active/suspended/banned controls access for existing users.

CREATE TABLE users (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    email             TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ,
    avatar_url        TEXT,

    status            TEXT NOT NULL DEFAULT 'active',
    onboarded_at      TIMESTAMPTZ,
    last_login_at     TIMESTAMPTZ,
    timezone          TEXT NOT NULL DEFAULT 'UTC',

    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'banned')),
    CONSTRAINT users_email_length CHECK (char_length(email) BETWEEN 3 AND 254),
    CONSTRAINT users_name_length CHECK (char_length(name) BETWEEN 1 AND 100)
);

-- Email unique (case-insensitive) for active users. Allows re-registration after deletion.
CREATE UNIQUE INDEX idx_users_email
    ON users(lower(email))
    WHERE deleted_at IS NULL;


-- Accounts (multi-provider: credential, google, github)
-- One user can have multiple accounts (e.g. email+password AND Google).
-- Password lives here, not on users — OAuth-only users have no password row.

CREATE TABLE accounts (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider            TEXT NOT NULL,    -- 'credential', 'google', 'github'
    provider_account_id TEXT NOT NULL,    -- provider's user ID (email for credential)
    password_hash       TEXT,             -- only for provider = 'credential'
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One account per provider per user.
CREATE UNIQUE INDEX idx_accounts_provider
    ON accounts(provider, provider_account_id);

-- Find all accounts for a user ("Connected accounts" settings page).
CREATE INDEX idx_accounts_user
    ON accounts(user_id);


-- Sessions
-- Opaque token in httpOnly cookie; hashed in DB for security.
-- If the DB is breached, hashed tokens can't be used to hijack sessions.

CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    ip_address  TEXT,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auth hot path: lookup by token hash, return user_id and expires_at without heap fetch.
CREATE UNIQUE INDEX idx_sessions_token_hash
    ON sessions(token_hash)
    INCLUDE (user_id, expires_at);

-- Find all sessions for a user ("Active Sessions" UI, "Sign out everywhere").
CREATE INDEX idx_sessions_user
    ON sessions(user_id);

-- Cleanup expired sessions (background job).
CREATE INDEX idx_sessions_expires
    ON sessions(expires_at);


-- Verification tokens (email verification, password reset)
-- Hashed tokens with explicit types. Consumed tokens marked used_at (not deleted)
-- for rate limiting. Background job cleans up tokens older than 30 days.

CREATE TABLE verification_tokens (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email       TEXT NOT NULL,
    token_hash  TEXT NOT NULL,
    type        TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT verification_tokens_type_check CHECK (type IN ('email_verification', 'password_reset', 'magic_link'))
);

-- Lookup by token hash (auth hot path). Only unused tokens.
CREATE INDEX idx_verification_tokens_hash
    ON verification_tokens(token_hash)
    WHERE used_at IS NULL;

-- Cleanup expired tokens (background job).
CREATE INDEX idx_verification_tokens_expires
    ON verification_tokens(expires_at)
    WHERE used_at IS NULL;

-- Rate limit: prevent spamming password reset / verification emails.
CREATE INDEX idx_verification_tokens_rate
    ON verification_tokens(email, type, created_at DESC)
    WHERE used_at IS NULL;
