-- name: FindWorkspaceUsage :one
SELECT * FROM workspace_usage
WHERE workspace_id = $1;

-- name: IncrementLinkUsage :execrows
-- Consumable: fails if the usage cycle has ended, preventing billing drift.
UPDATE workspace_usage SET
    links_usage = links_usage + $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND usage_period_end > NOW();

-- name: IncrementClickUsage :execrows
-- Consumable: async increment triggered by the link resolver.
UPDATE workspace_usage SET
    clicks_usage = clicks_usage + $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND usage_period_end > NOW();

-- name: IncrementDomainActive :execrows
UPDATE workspace_usage SET
    domains_active = domains_active + $2,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: DecrementDomainActive :execrows
UPDATE workspace_usage SET
    domains_active = domains_active - $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND domains_active >= $2;

-- name: IncrementWebhookActive :execrows
UPDATE workspace_usage SET
    webhooks_active = webhooks_active + $2,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: DecrementWebhookActive :execrows
UPDATE workspace_usage SET
    webhooks_active = webhooks_active - $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND webhooks_active >= $2;

-- name: IncrementAPIKeyActive :execrows
UPDATE workspace_usage SET
    api_keys_active = api_keys_active + $2,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: DecrementAPIKeyActive :execrows
UPDATE workspace_usage SET
    api_keys_active = api_keys_active - $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND api_keys_active >= $2;

-- name: IncrementMemberActive :execrows
UPDATE workspace_usage SET
    members_active = members_active + $2,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: DecrementMemberActive :execrows
UPDATE workspace_usage SET
    members_active = members_active - $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND members_active >= $2;

-- name: IncrementFolderActive :execrows
UPDATE workspace_usage SET
    folders_active = folders_active + $2,
    updated_at = NOW()
WHERE workspace_id = $1;

-- name: DecrementFolderActive :execrows
UPDATE workspace_usage SET
    folders_active = folders_active - $2,
    updated_at = NOW()
WHERE workspace_id = $1 AND folders_active >= $2;
