package domain

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/xid"
)

// Service implements domain business logic.
type Service struct {
	store     *Store
	txtPrefix string
}

// NewService creates a new domain service.
func NewService(store *Store, txtPrefix string) *Service {
	return &Service{store: store, txtPrefix: txtPrefix}
}

// Create generates a verification token and persists a new domain.
func (s *Service) Create(ctx context.Context, workspaceID, userID string, input CreateInput) (*Domain, error) {
	hostname := strings.ToLower(input.Hostname)

	token, err := generateVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("generating verification token: %w", err)
	}

	d := &Domain{
		ID:                "dom_" + xid.New().String(),
		WorkspaceID:       workspaceID,
		CreatedBy:         userID,
		Hostname:          hostname,
		VerificationToken: token,
		FallbackURL:       input.FallbackURL,
	}

	return s.store.Insert(ctx, d)
}

// Get retrieves a domain by ID within a workspace.
func (s *Service) Get(ctx context.Context, id, workspaceID string) (*Domain, error) {
	return s.store.FindByID(ctx, id, workspaceID)
}

// List returns all domains for a workspace.
func (s *Service) List(ctx context.Context, workspaceID string) ([]Domain, error) {
	return s.store.List(ctx, workspaceID)
}

// Update partially updates a domain.
func (s *Service) Update(ctx context.Context, id, workspaceID string, input UpdateInput) (*Domain, error) {
	return s.store.Update(ctx, id, workspaceID, input)
}

// Delete soft-deletes a domain.
func (s *Service) Delete(ctx context.Context, id, workspaceID string) error {
	return s.store.SoftDelete(ctx, id, workspaceID)
}

// FindByHostname retrieves a domain by its hostname (used by internal endpoints).
func (s *Service) FindByHostname(ctx context.Context, hostname string) (*Domain, error) {
	return s.store.FindByHostname(ctx, hostname)
}

// Verify performs a DNS TXT lookup to confirm domain ownership and activates the domain.
func (s *Service) Verify(ctx context.Context, id, workspaceID string) (*Domain, error) {
	d, err := s.store.FindByID(ctx, id, workspaceID)
	if err != nil {
		return nil, err
	}

	if d.Status == "active" {
		return nil, ErrAlreadyVerified
	}

	// Look up TXT records with a bounded timeout to avoid blocking on slow DNS.
	txtHost := s.txtPrefix + d.Hostname
	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resolver := &net.Resolver{}
	records, err := resolver.LookupTXT(lookupCtx, txtHost)
	if err != nil || len(records) == 0 {
		return nil, ErrDNSNotFound
	}

	found := false
	for _, r := range records {
		if strings.TrimSpace(r) == d.VerificationToken {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrDNSMismatch
	}

	now := time.Now()
	return s.store.UpdateStatus(ctx, id, workspaceID, "active", &now)
}
