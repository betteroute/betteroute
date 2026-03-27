package deeplink

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/execrc/betteroute/internal/db"
	"github.com/execrc/betteroute/internal/ptr"
	"github.com/execrc/betteroute/internal/sqlc"
)

// Store handles database operations for the deeplink package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new deeplink store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// FindByID retrieves a platform app by ID.
func (s *Store) FindByID(ctx context.Context, id string) (*PlatformApp, error) {
	row, err := s.q.FindPlatformAppByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding platform app %s: %w", id, err)
	}
	return toPlatformApp(row), nil
}

// FindByHostname looks up a platform app whose url_patterns contain the hostname.
func (s *Store) FindByHostname(ctx context.Context, hostname string) (*PlatformApp, error) {
	row, err := s.q.FindPlatformAppByURLPattern(ctx, hostname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding platform app by hostname %s: %w", hostname, err)
	}
	return toPlatformApp(row), nil
}

// List returns all platform apps.
func (s *Store) List(ctx context.Context) ([]PlatformApp, error) {
	rows, err := s.q.ListPlatformApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing platform apps: %w", err)
	}
	apps := make([]PlatformApp, len(rows))
	for i, row := range rows {
		apps[i] = *toPlatformApp(row)
	}
	return apps, nil
}

func toPlatformApp(row sqlc.PlatformApp) *PlatformApp {
	return &PlatformApp{
		ID:             row.ID,
		Name:           row.Name,
		IconURL:        ptr.Val(row.IconUrl),
		URLPatterns:    row.UrlPatterns,
		IOSScheme:      ptr.Val(row.IosScheme),
		AndroidScheme:  ptr.Val(row.AndroidScheme),
		IOSAppID:       ptr.Val(row.IosAppID),
		IOSBundleID:    ptr.Val(row.IosBundleID),
		IOSTeamID:      ptr.Val(row.IosTeamID),
		AndroidPackage: ptr.Val(row.AndroidPackage),
		AndroidSHA256:  row.AndroidSha256,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// InsertWorkspaceApp creates a new workspace app record.
func (s *Store) InsertWorkspaceApp(ctx context.Context, wa *WorkspaceApp) (*WorkspaceApp, error) {
	row, err := s.q.InsertWorkspaceApp(ctx, sqlc.InsertWorkspaceAppParams{
		ID:                 wa.ID,
		WorkspaceID:        wa.WorkspaceID,
		CreatedBy:          ptr.ToNonZero(wa.CreatedBy),
		Name:               wa.Name,
		Platform:           wa.Platform,
		BundleID:           ptr.ToNonZero(wa.BundleID),
		TeamID:             ptr.ToNonZero(wa.TeamID),
		AppStoreUrl:        ptr.ToNonZero(wa.AppStoreURL),
		PackageName:        ptr.ToNonZero(wa.PackageName),
		Sha256Fingerprints: wa.SHA256Fingerprints,
		PlayStoreUrl:       ptr.ToNonZero(wa.PlayStoreURL),
		Scheme:             ptr.ToNonZero(wa.Scheme),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting workspace app: %w", err)
	}
	return toWorkspaceApp(row), nil
}

// FindWorkspaceAppByID retrieves a workspace app by ID.
func (s *Store) FindWorkspaceAppByID(ctx context.Context, id, workspaceID string) (*WorkspaceApp, error) {
	row, err := s.q.FindWorkspaceAppByID(ctx, sqlc.FindWorkspaceAppByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkspaceAppNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying workspace app %s: %w", id, err)
	}
	return toWorkspaceApp(row), nil
}

// ListWorkspaceApps retrieves all workspace apps for a workspace.
func (s *Store) ListWorkspaceApps(ctx context.Context, workspaceID string) ([]WorkspaceApp, error) {
	rows, err := s.q.ListWorkspaceAppsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing workspace apps: %w", err)
	}
	apps := make([]WorkspaceApp, len(rows))
	for i, row := range rows {
		apps[i] = *toWorkspaceApp(row)
	}
	return apps, nil
}

// UpdateWorkspaceApp partially updates a workspace app.
func (s *Store) UpdateWorkspaceApp(ctx context.Context, id, workspaceID string, input UpdateWorkspaceAppInput) (*WorkspaceApp, error) {
	var u db.Update

	if input.Name.Set {
		u.Set("name", input.Name.Value)
	}
	if input.BundleID.Set {
		u.Set("bundle_id", input.BundleID.Value)
	}
	if input.TeamID.Set {
		u.Set("team_id", input.TeamID.Value)
	}
	if input.AppStoreURL.Set {
		u.Set("app_store_url", input.AppStoreURL.Value)
	}
	if input.PackageName.Set {
		u.Set("package_name", input.PackageName.Value)
	}
	if input.SHA256Fingerprints.Set {
		u.Set("sha256_fingerprints", input.SHA256Fingerprints.Value)
	}
	if input.PlayStoreURL.Set {
		u.Set("play_store_url", input.PlayStoreURL.Value)
	}
	if input.Scheme.Set {
		u.Set("scheme", input.Scheme.Value)
	}

	if u.IsEmpty() {
		return s.FindWorkspaceAppByID(ctx, id, workspaceID)
	}

	sql, args := u.Build("workspace_apps", "id = ? AND workspace_id = ? AND deleted_at IS NULL", id, workspaceID)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.WorkspaceApp])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkspaceAppNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating workspace app %s: %w", id, err)
	}
	return toWorkspaceApp(row), nil
}

// SoftDeleteWorkspaceApp marks a workspace app as deleted.
func (s *Store) SoftDeleteWorkspaceApp(ctx context.Context, id, workspaceID string) error {
	rows, err := s.q.SoftDeleteWorkspaceApp(ctx, sqlc.SoftDeleteWorkspaceAppParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting workspace app %s: %w", id, err)
	}
	if rows == 0 {
		return ErrWorkspaceAppNotFound
	}
	return nil
}

func toWorkspaceApp(row sqlc.WorkspaceApp) *WorkspaceApp {
	return &WorkspaceApp{
		ID:                 row.ID,
		WorkspaceID:        row.WorkspaceID,
		CreatedBy:          ptr.Val(row.CreatedBy),
		Name:               row.Name,
		Platform:           row.Platform,
		BundleID:           ptr.Val(row.BundleID),
		TeamID:             ptr.Val(row.TeamID),
		AppStoreURL:        ptr.Val(row.AppStoreUrl),
		PackageName:        ptr.Val(row.PackageName),
		SHA256Fingerprints: row.Sha256Fingerprints,
		PlayStoreURL:       ptr.Val(row.PlayStoreUrl),
		Scheme:             ptr.Val(row.Scheme),
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}
