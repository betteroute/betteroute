-- name: FindEntitlement :one
-- Middleware hot path: resolves the full entitlement payload for a workspace
-- in a single round-trip.
--
-- LEFT JOIN ensures workspaces with a subscription but no usage row yet still
-- return a row (usage columns will be NULL, coalesced to 0 / '{}').
-- If no subscription row exists at all, the query returns no rows — the
-- middleware defaults to Free plan + zero usage in that case.
SELECT
    s.plan_id,
    s.custom_quotas,
    s.custom_features,
    COALESCE(u.links_usage,    0)    AS links_usage,
    COALESCE(u.clicks_usage,   0)    AS clicks_usage,
    COALESCE(u.domains_active, 0)    AS domains_active,
    COALESCE(u.webhooks_active, 0)   AS webhooks_active,
    COALESCE(u.api_keys_active, 0)   AS api_keys_active,
    COALESCE(u.members_active,  0)   AS members_active,
    COALESCE(u.folders_active,  0)   AS folders_active,
    COALESCE(u.tags_active,    0)   AS tags_active,
    COALESCE(u.over_quota,     '{}') AS over_quota
FROM   subscriptions s
LEFT   JOIN workspace_usage u ON u.workspace_id = s.workspace_id
WHERE  s.workspace_id = $1;
