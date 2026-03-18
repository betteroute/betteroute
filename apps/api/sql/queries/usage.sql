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
-- Resets consumable counters natively when the exact billing anniversary period has expired.
-- No-op if the cycle is still active (idempotent).
UPDATE workspace_usage SET
    links_usage = 0,
    clicks_usage = 0,
    usage_period_start = usage_period_end,
    usage_period_end = usage_period_end + interval '1 month',
    updated_at = NOW()
WHERE workspace_id = $1 AND usage_period_end <= NOW();
