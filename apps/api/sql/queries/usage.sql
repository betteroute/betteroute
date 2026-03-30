-- name: UpsertWorkspaceUsage :exec
INSERT INTO workspace_usage (workspace_id)
VALUES ($1)
ON CONFLICT (workspace_id) DO NOTHING;

-- name: IncrementUsage :execrows
UPDATE workspace_usage SET
    links_usage = CASE WHEN sqlc.arg(is_links)::BOOLEAN THEN links_usage + sqlc.arg(delta)::INT ELSE links_usage END,
    clicks_usage = CASE WHEN sqlc.arg(is_clicks)::BOOLEAN THEN clicks_usage + sqlc.arg(delta)::BIGINT ELSE clicks_usage END,
    updated_at = NOW()
WHERE workspace_id = $1
  AND usage_period_end > NOW()
  AND EXISTS (SELECT 1 FROM workspaces WHERE id = $1 AND deleted_at IS NULL);

-- name: RolloverUsageCycle :execrows
-- Resets consumable counters when the billing cycle has expired.
-- Catches up in one shot regardless of how many months have elapsed.
-- No-op if the cycle is still active (idempotent). Does NOT touch allocated counters.
-- The Polar webhook sets the authoritative boundaries via ForceResetWorkspaceUsage;
-- this is the safety net for when the webhook doesn't fire.
UPDATE workspace_usage SET
    links_usage = 0,
    clicks_usage = 0,
    usage_period_start = NOW(),
    usage_period_end = NOW() + interval '1 month',
    updated_at = NOW()
WHERE workspace_id = $1 AND usage_period_end <= NOW();

-- name: AdjustResource :exec
-- Adjusts a single allocated resource counter.
-- delta is +1 on create, -1 on delete. CHECK constraints prevent negatives.
UPDATE workspace_usage SET
    domains_active  = CASE WHEN sqlc.arg(is_domains)::BOOLEAN  THEN domains_active  + sqlc.arg(delta)::INT ELSE domains_active  END,
    webhooks_active = CASE WHEN sqlc.arg(is_webhooks)::BOOLEAN THEN webhooks_active + sqlc.arg(delta)::INT ELSE webhooks_active END,
    api_keys_active = CASE WHEN sqlc.arg(is_api_keys)::BOOLEAN THEN api_keys_active + sqlc.arg(delta)::INT ELSE api_keys_active END,
    members_active  = CASE WHEN sqlc.arg(is_members)::BOOLEAN  THEN members_active  + sqlc.arg(delta)::INT ELSE members_active  END,
    folders_active  = CASE WHEN sqlc.arg(is_folders)::BOOLEAN  THEN folders_active  + sqlc.arg(delta)::INT ELSE folders_active  END,
    tags_active     = CASE WHEN sqlc.arg(is_tags)::BOOLEAN     THEN tags_active     + sqlc.arg(delta)::INT ELSE tags_active     END,
    updated_at = NOW()
WHERE workspace_id = $1
  AND EXISTS (SELECT 1 FROM workspaces WHERE id = $1 AND deleted_at IS NULL);
