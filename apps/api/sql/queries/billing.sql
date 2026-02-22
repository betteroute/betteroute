-- name: FindSubscriptionByWorkspace :one
SELECT * FROM subscriptions
WHERE workspace_id = $1;

-- name: FindSubscriptionByProviderSub :one
-- Webhook path: resolve provider subscription ID → workspace subscription.
SELECT * FROM subscriptions
WHERE provider_subscription_id = $1;

-- name: UpsertSubscription :one
-- Webhook path: triggered by subscription lifecycle events.
INSERT INTO subscriptions (
    workspace_id, plan_id, provider, provider_customer_id, provider_subscription_id,
    currency, billing_interval, status, current_period_start, current_period_end,
    cancel_at_period_end, canceled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
ON CONFLICT (workspace_id) DO UPDATE SET
    plan_id = EXCLUDED.plan_id,
    status = EXCLUDED.status,
    current_period_start = EXCLUDED.current_period_start,
    current_period_end = EXCLUDED.current_period_end,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    canceled_at = EXCLUDED.canceled_at,
    updated_at = NOW()
RETURNING *;

-- name: UpdateSubscriptionStatus :execrows
UPDATE subscriptions SET
    status = $2,
    canceled_at = $3,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: SyncUsageCycle :execrows
-- Webhook path: triggered on new billing cycle.
-- Resets consumable counts and bumps the usage period forward.
UPDATE workspace_usage SET
    links_usage = 0,
    clicks_usage = 0,
    usage_period_start = $2,
    usage_period_end = $3,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: UpsertWorkspaceUsage :exec
-- Ensures a usage row exists for a workspace (created on first subscription).
INSERT INTO workspace_usage (workspace_id)
VALUES ($1)
ON CONFLICT (workspace_id) DO NOTHING;

-- name: RecordBillingWebhookEvent :exec
-- Idempotency gatekeeper. PK constraint rejects duplicate event IDs.
INSERT INTO billing_webhook_events (id, provider, event_type)
VALUES ($1, $2, $3);

-- name: FindBillingWebhookEvent :one
SELECT * FROM billing_webhook_events
WHERE id = $1;

-- name: FindPlanPrice :one
SELECT * FROM plan_prices
WHERE plan_id = $1 AND provider = $2 AND interval = $3 AND currency = $4;

-- name: ListPlanPrices :many
SELECT * FROM plan_prices
WHERE provider = $1 AND currency = $2
ORDER BY plan_id, interval;
