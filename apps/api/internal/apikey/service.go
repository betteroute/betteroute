package apikey

import (
	"context"
	"time"

	"github.com/rs/xid"

	"github.com/execrc/betteroute/internal/ptr"
)

// Service implements API key business logic.
type Service struct {
	store *Store
}

// NewService creates a new API key service.
func NewService(store *Store) *Service {
	return &Service{store: store}
}

// Created is returned once on key creation — includes the plain key
// that cannot be retrieved again.
type Created struct {
	APIKey
	PlainKey string `json:"key"`
}

// Create generates a new API key. The plain key is returned once and never stored.
func (s *Service) Create(ctx context.Context, workspaceID, userID string, input CreateInput) (*Created, error) {
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	plain, hash, err := generateKey()
	if err != nil {
		return nil, err
	}

	scopes := input.Scopes
	if input.Permission != PermissionRestricted {
		scopes = nil
	}

	id := "key_" + xid.New().String()

	key, err := s.store.Insert(ctx, InsertParams{
		ID:          id,
		WorkspaceID: workspaceID,
		CreatedBy:   ptr.ToNonZero(userID),
		Name:        input.Name,
		KeyHash:     hash,
		KeyPrefix:   plain[:len(Prefix)+8], // "btr_a1b2c3d4"
		Permission:  input.Permission,
		Scopes:      scopesToStrings(scopes),
		ExpiresAt:   input.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &Created{APIKey: *key, PlainKey: plain}, nil
}

// ValidateKey resolves a plain API key to its domain object.
// Used by the auth middleware for Bearer token authentication.
func (s *Service) ValidateKey(ctx context.Context, plain string) (*APIKey, error) {
	hash := hashKey(plain)

	key, err := s.store.FindByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpired
	}

	// Update last_used_at asynchronously — don't slow the auth path.
	go func() {
		_ = s.store.UpdateLastUsed(context.Background(), key.ID)
	}()

	return key, nil
}

// ValidateKeyWithCreator resolves a plain API key to its domain object and creator
// in a single DB round-trip. Used by the auth middleware.
func (s *Service) ValidateKeyWithCreator(ctx context.Context, plain string) (*APIKey, *Creator, error) {
	hash := hashKey(plain)

	key, creator, err := s.store.FindByHashWithCreator(ctx, hash)
	if err != nil {
		return nil, nil, err
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, nil, ErrExpired
	}

	go func() {
		_ = s.store.UpdateLastUsed(context.Background(), key.ID)
	}()

	return key, creator, nil
}

// Get retrieves a single API key by ID.
func (s *Service) Get(ctx context.Context, id, workspaceID string) (*APIKey, error) {
	return s.store.FindByID(ctx, id, workspaceID)
}

// List retrieves all active API keys for a workspace.
func (s *Service) List(ctx context.Context, workspaceID string) ([]APIKey, error) {
	return s.store.List(ctx, workspaceID)
}

// Delete soft-deletes an API key.
func (s *Service) Delete(ctx context.Context, id, workspaceID string) error {
	return s.store.SoftDelete(ctx, id, workspaceID)
}

func validateCreateInput(input CreateInput) error {
	if !input.Permission.Valid() {
		return ErrInvalidScope
	}
	if input.Permission == PermissionRestricted {
		if len(input.Scopes) == 0 {
			return ErrScopesRequired
		}
		for _, s := range input.Scopes {
			if !s.Valid() {
				return ErrInvalidScope
			}
		}
	}
	return nil
}
