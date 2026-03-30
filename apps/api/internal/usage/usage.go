// Package usage tracks workspace resource consumption against plan quotas.
//
// Three operations, three execution models:
//   - Consume: synchronous consumable increment inside a caller-provided transaction (links).
//   - Emit:    asynchronous buffered increment for high-throughput counters (clicks).
//   - Adjust:  synchronous allocated counter delta (+1 create, -1 delete).
package usage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/entitlement"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Quota aliases for call-site readability: usage.Links instead of entitlement.QuotaLinks.
const (
	Links  = entitlement.QuotaLinks
	Clicks = entitlement.QuotaClicks

	Domains  = entitlement.QuotaDomains
	Webhooks = entitlement.QuotaWebhooks
	APIKeys  = entitlement.QuotaAPIKeys
	Members  = entitlement.QuotaMembers
	Folders  = entitlement.QuotaFolders
	Tags     = entitlement.QuotaTags
)

// Quota is an alias for entitlement.Quota.
type Quota = entitlement.Quota

// consumableColumn maps consumable quotas to their workspace_usage column name.
var consumableColumn = map[Quota]string{
	Links:  "links_usage",
	Clicks: "clicks_usage",
}

// Meter tracks workspace resource consumption.
type Meter struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	buffer *buffer
}

// NewMeter creates a usage meter and starts the background flush goroutine.
func NewMeter(pool *pgxpool.Pool) *Meter {
	m := &Meter{
		pool: pool,
		q:    sqlc.New(pool),
	}
	m.buffer = newBuffer(m)
	return m
}

// Stop drains pending events and shuts down the background flusher.
func (m *Meter) Stop() {
	m.buffer.stop()
}

// Consume increments a consumable counter inside the caller's transaction.
// Used for link creation where the increment must be atomic with the insert.
// No-op if n <= 0. Returns an error if the usage row is missing or the cycle expired.
func (m *Meter) Consume(ctx context.Context, db sqlc.DBTX, workspaceID string, q Quota, n int) error {
	if n <= 0 {
		return nil
	}

	rows, err := sqlc.New(db).IncrementUsage(ctx, sqlc.IncrementUsageParams{
		WorkspaceID: workspaceID,
		IsLinks:     q == Links,
		IsClicks:    q == Clicks,
		Delta:       int32(n), //nolint:gosec // capped by plan quotas, never exceeds int32
	})
	if err != nil {
		return fmt.Errorf("consume %s: %w", q, err)
	}
	if rows == 0 {
		return fmt.Errorf("no usage row or expired cycle for workspace %s", workspaceID)
	}
	return nil
}

// Emit buffers a consumable increment for async batch flush (clicks).
// Non-blocking: drops the event if the buffer is full to protect the hot path.
func (m *Meter) Emit(workspaceID string, q Quota, n int64) {
	if n <= 0 {
		return
	}
	m.buffer.send(workspaceID, q, n)
}

// Adjust changes an allocated resource counter by delta (+1 on create, -1 on delete).
// Allocated counters are never reset by cycle rollover.
func (m *Meter) Adjust(ctx context.Context, workspaceID string, q Quota, delta int) error {
	err := m.q.AdjustResource(ctx, sqlc.AdjustResourceParams{
		WorkspaceID: workspaceID,
		Delta:       int32(delta), //nolint:gosec // always ±1
		IsDomains:   q == Domains,
		IsWebhooks:  q == Webhooks,
		IsApiKeys:   q == APIKeys,
		IsMembers:   q == Members,
		IsFolders:   q == Folders,
		IsTags:      q == Tags,
	})
	if err != nil {
		return fmt.Errorf("adjust %s: %w", q, err)
	}
	return nil
}

// flush batch-writes accumulated deltas to PostgreSQL in a single round-trip.
func (m *Meter) flush(ctx context.Context, deltas map[bufferKey]int64) error {
	// Group by workspace.
	type wsDeltas struct {
		quotas map[Quota]int64
	}
	byWorkspace := make(map[string]*wsDeltas, len(deltas))
	for k, n := range deltas {
		ws := byWorkspace[k.workspaceID]
		if ws == nil {
			ws = &wsDeltas{quotas: make(map[Quota]int64, 2)}
			byWorkspace[k.workspaceID] = ws
		}
		ws.quotas[k.quota] += n
	}

	batch := &pgx.Batch{}

	for wsID, ws := range byWorkspace {
		// Rollover expired cycles before incrementing.
		batch.Queue(`
			UPDATE workspace_usage SET
				links_usage = 0, clicks_usage = 0,
				usage_period_start = NOW(),
				usage_period_end = NOW() + interval '1 month',
				updated_at = NOW()
			WHERE workspace_id = $1 AND usage_period_end <= NOW()
			  AND EXISTS (SELECT 1 FROM workspaces WHERE id = $1 AND deleted_at IS NULL)
		`, wsID)

		// Build dynamic SET clause for consumable increments.
		var setClauses strings.Builder
		args := []any{wsID}

		for q, n := range ws.quotas {
			col, ok := consumableColumn[q]
			if !ok {
				continue
			}
			if setClauses.Len() > 0 {
				setClauses.WriteString(", ")
			}
			args = append(args, n)
			// Freeze counters when cycle has expired (rollover above handles the reset).
			fmt.Fprintf(&setClauses, "%s = CASE WHEN usage_period_end > NOW() THEN %s + $%d ELSE %s END",
				col, col, len(args), col)
		}

		if setClauses.Len() > 0 {
			query := fmt.Sprintf(`
				UPDATE workspace_usage SET %s, updated_at = NOW()
				WHERE workspace_id = $1
				  AND EXISTS (SELECT 1 FROM workspaces WHERE id = $1 AND deleted_at IS NULL)
			`, setClauses.String())
			batch.Queue(query, args...)
		}
	}

	br := m.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range batch.Len() {
		if _, err := br.Exec(); err != nil {
			slog.ErrorContext(ctx, "usage batch flush failed", "error", err)
			return err
		}
	}
	return nil
}
