# Usage Package — Wiring Guide

Reference for integrating `usage.Meter` into feature packages.

## Core Rule
**Usage tracking lives in the service layer, not handlers, not stores.**
- **Allocations** (Domains, Folders): Tracked universally entirely by PostgreSQL `COUNT(*)` subqueries via the `entitlement` middleware. Zero manual tracking is required.
- **Consumables** (Links, Clicks): Handled explicitly by the `usage.Meter`.

## Wiring in main.go
```go
meter := usage.NewMeter(pool)
defer meter.Stop() // flush pending async events on shutdown

// Inject ONLY into services that track consumables (Links/Clicks) or initialize workspaces.
linkSvc := link.NewService(link.NewStore(pool), deeplinkSvc, meter)
workspaceSvc := workspace.NewService(workspace.NewStore(pool), meter)

// Redirect handler uses meter.Emit() for async click tracking.
redirectSvc := redirect.NewService(pool, cfg.PlatformDomains, meter)
```

## Integration Patterns

### Pattern 1: Allocated Metrics (Domains, Folders, Tags, Members)
Allocated limits define "Peak Concurrency". There is **no need** to manually track these.
The database inherently enforces limits exclusively at the authorization gate before your handler runs:
```go
func (h *Handler) Create(c fiber.Ctx) error {
    ctx := c.Context()
    
    // Natively runs SELECT COUNT(*) behind the scenes and blocks if over limit.
    // There is no meter tracking code needed in your creation service!
    if err := guard.Quota(ctx, entitlement.QuotaDomains, 1); err != nil {
        return err 
    }
}
```

### Pattern 2: Consumable Metrics (Links)
Consumables asynchronously bypass pure row counts and burn a monthly capacity natively. 
```go
func (s *Service) Create(ctx context.Context, ...) (*Link, error) {
    // Notice we use Burn(), guaranteeing it cannot be arbitrarily refunded.
    if err := s.meter.Burn(ctx, s.store.Pool(), created.WorkspaceID, usage.Links, 1); err != nil {
        slog.Warn("burning link quota", "error", err)
    }
}
```

### Pattern 3: Async Telemetry (Clicks)
Clicks are extremely high-frequency — use `meter.Emit()` which continuously buffers and autonomously flushes via `pgx.Batch` in the background perfectly insulating your database from DDOS.
```go
func (s *Service) Resolve(ctx context.Context, code string) (*Resolution, error) {
    // Fire-and-forget — perfectly batched, DDOS-protected, and non-blocking.
    s.meter.Emit(res.WorkspaceID, usage.Clicks, 1)
}
```

### Pattern 4: Component Lifecycles
When a workspace is created, initialize its autonomous ledger row.
```go
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Workspace, error) {
    // Initialize usage tracking row for the new workspace's consumable counters.
    // Note: For Free plans, this initialization implicitly anchors their limits to perpetually 
    // roll over precisely 30 days from this exact timestamp (unless a Polar upgrade specifically overrides it).
    if err := s.meter.Init(ctx, ws.ID); err != nil {
        slog.Warn("initializing workspace usage", "error", err)
    }
}
```

## Quota → Capability Mapping

| Quota | Tracked On | Tracker Mechanism |
|-------|-----------|-----------|
| `usage.Links` | link.Service.Create | `meter.Burn()` explicitly |
| `usage.Clicks` | redirect.Service.Resolve | `meter.Emit()` explicitly |
| `usage.Domains` | handler | `guard.Quota` (Native DB Count) |
| `usage.Folders` | handler | `guard.Quota` (Native DB Count) |
| `usage.Tags` | handler | `guard.Quota` (Native DB Count) |
| `usage.Members` | handler | `guard.Quota` (Native DB Count) |
| `usage.APIKeys` | handler | `guard.Quota` (Native DB Count) |
| `usage.Webhooks` | handler | `guard.Quota` (Native DB Count) |

## Guard vs Meter
These are distinct:
- **`guard.Quota(ctx, quota, n)`** — Runs in the handler **BEFORE** the operation. If it's an Allocation, it validates the live `COUNT(*)` DB numbers. If it's a Consumable, it validates the current month's usage integer. This is the **gate**.
- **`meter.*`** — Runs in the service **AFTER** the creation succeeds to permanently increment Consumable counters (`usage.Links` and `usage.Clicks`). This is the **ledger**.
