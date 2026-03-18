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
    COALESCE(u.links_usage,    0)    AS links_usage,
    COALESCE(u.clicks_usage,   0)    AS clicks_usage,
    (SELECT COUNT(*) FROM domains WHERE domains.workspace_id = $1)::integer AS domains_active,
    (SELECT COUNT(*) FROM workspace_apps WHERE workspace_apps.workspace_id = $1)::integer AS webhooks_active,
    (SELECT COUNT(*) FROM api_keys WHERE api_keys.workspace_id = $1)::integer AS api_keys_active,
    (SELECT COUNT(*) FROM workspace_members WHERE workspace_members.workspace_id = $1)::integer AS members_active,
    (SELECT COUNT(*) FROM folders WHERE folders.workspace_id = $1)::integer AS folders_active,
    (SELECT COUNT(*) FROM tags WHERE tags.workspace_id = $1)::integer AS tags_active
FROM   subscriptions s
LEFT   JOIN workspace_usage u ON u.workspace_id = s.workspace_id
WHERE  s.workspace_id = $1;
