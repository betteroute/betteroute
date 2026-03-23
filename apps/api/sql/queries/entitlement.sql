-- name: FindEntitlement :one
-- Workspace existence is already verified by the workspace middleware.
SELECT
    s.plan_id,
    COALESCE(u.links_usage,     0) AS links_usage,
    COALESCE(u.clicks_usage,    0) AS clicks_usage,
    COALESCE(u.domains_active,  0) AS domains_active,
    COALESCE(u.webhooks_active, 0) AS webhooks_active,
    COALESCE(u.api_keys_active, 0) AS api_keys_active,
    COALESCE(u.members_active,  0) AS members_active,
    COALESCE(u.folders_active,  0) AS folders_active,
    COALESCE(u.tags_active,     0) AS tags_active,
    u.usage_period_end
FROM   subscriptions s
LEFT   JOIN workspace_usage u ON u.workspace_id = s.workspace_id
WHERE  s.workspace_id = $1;
