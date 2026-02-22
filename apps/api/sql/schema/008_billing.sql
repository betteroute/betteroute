-- Plans
-- Minimal anchor table for FK references from subscriptions and plan_prices.
-- Features, quotas, and tier definitions live in application code (internal/entitlement).
-- Seeds: sql/seeds/plans.sql

CREATE TABLE plans (
    id         TEXT PRIMARY KEY,
    name       TEXT UNIQUE NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT plans_name_length
        CHECK (char_length(name) BETWEEN 1 AND 50)
);

-- Plan Prices
-- Maps plans to provider-specific price IDs for checkout flows.
-- Stored in DB so provider price IDs can be updated without a code deploy.

CREATE TABLE plan_prices (
    plan_id  TEXT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    interval TEXT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'usd',
    price_id TEXT NOT NULL,

    PRIMARY KEY (plan_id, provider, interval, currency),

    CONSTRAINT plan_prices_provider_check
        CHECK (provider IN ('stripe', 'polar')),

    CONSTRAINT plan_prices_interval_check
        CHECK (interval IN ('monthly', 'yearly')),

    CONSTRAINT plan_prices_currency_check
        CHECK (currency IN ('usd'))
);

-- Subscriptions
-- The billing contract between a workspace and a plan.

CREATE TABLE subscriptions (
    workspace_id             TEXT PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id                  TEXT NOT NULL REFERENCES plans(id),

    provider                 TEXT,
    provider_customer_id     TEXT,
    provider_subscription_id TEXT,

    currency                 TEXT,
    billing_interval         TEXT,
    status                   TEXT NOT NULL DEFAULT 'active',

    current_period_start     TIMESTAMPTZ,
    current_period_end       TIMESTAMPTZ,
    cancel_at_period_end     BOOLEAN DEFAULT FALSE,
    canceled_at              TIMESTAMPTZ,

    -- Enterprise overrides merged over plan defaults at request time
    custom_quotas            JSONB,
    custom_features          JSONB,

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscriptions_status_check
        CHECK (status IN ('active', 'trialing', 'past_due', 'canceled', 'paused')),

    CONSTRAINT subscriptions_provider_check
        CHECK (provider IN ('stripe', 'polar') OR provider IS NULL),

    CONSTRAINT subscriptions_interval_check
        CHECK (billing_interval IN ('monthly', 'yearly') OR billing_interval IS NULL),

    CONSTRAINT subscriptions_currency_check
        CHECK (currency IN ('usd') OR currency IS NULL)
);

-- Looking up subscriptions by plan.
CREATE INDEX idx_subscriptions_plan ON subscriptions(plan_id);

-- Filtering for inactive or delinquent subscriptions.
CREATE INDEX idx_subscriptions_status ON subscriptions(status) WHERE status != 'active';

-- Webhook resolution via provider subscription ID.
CREATE INDEX idx_subscriptions_provider_sub ON subscriptions(provider_subscription_id)
    WHERE provider_subscription_id IS NOT NULL;

-- Workspace Usage
-- Tracks current resource utilization against plan quotas.
-- Consumable counters reset each usage cycle; allocated counters are persistent.

CREATE TABLE workspace_usage (
    workspace_id   TEXT PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Consumable — reset each usage cycle (always calendar month)
    links_usage    INTEGER NOT NULL DEFAULT 0,
    clicks_usage   BIGINT  NOT NULL DEFAULT 0,

    -- Allocated — current active resource counts, never reset
    domains_active  INTEGER NOT NULL DEFAULT 0,
    webhooks_active INTEGER NOT NULL DEFAULT 0,
    api_keys_active INTEGER NOT NULL DEFAULT 0,
    members_active  INTEGER NOT NULL DEFAULT 0,
    folders_active  INTEGER NOT NULL DEFAULT 0,
    tags_active     INTEGER NOT NULL DEFAULT 0,

    -- Quotas that exceed the current plan cap (set on downgrade, cleared on delete)
    -- Keys are entitlement.Quota string values: "quota_links", "quota_folders", etc.
    over_quota     JSONB NOT NULL DEFAULT '{}',

    -- Usage cycle bounds (calendar month, decoupled from billing interval)
    usage_period_start TIMESTAMPTZ NOT NULL DEFAULT date_trunc('month', NOW()),
    usage_period_end   TIMESTAMPTZ NOT NULL DEFAULT (date_trunc('month', NOW()) + interval '1 month'),

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workspace_usage_links_positive    CHECK (links_usage >= 0),
    CONSTRAINT workspace_usage_clicks_positive   CHECK (clicks_usage >= 0),
    CONSTRAINT workspace_usage_domains_positive  CHECK (domains_active >= 0),
    CONSTRAINT workspace_usage_webhooks_positive CHECK (webhooks_active >= 0),
    CONSTRAINT workspace_usage_api_keys_positive CHECK (api_keys_active >= 0),
    CONSTRAINT workspace_usage_members_positive  CHECK (members_active >= 0),
    CONSTRAINT workspace_usage_folders_positive  CHECK (folders_active >= 0),
    CONSTRAINT workspace_usage_tags_positive     CHECK (tags_active >= 0)
);

-- Billing Webhook Events
-- Deduplication log for provider webhook delivery.
-- Prevents double-processing when providers retry events.

CREATE TABLE billing_webhook_events (
    id         TEXT        PRIMARY KEY,
    provider   TEXT        NOT NULL,
    event_type TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT billing_webhook_events_provider_check
        CHECK (provider IN ('stripe', 'polar'))
);

-- Periodic cleanup of events older than 30 days.
CREATE INDEX idx_billing_webhook_events_cleanup ON billing_webhook_events(created_at);
