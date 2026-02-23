package auth

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

// Store handles database operations for the auth package.
type Store struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

// NewStore creates a new auth store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{q: sqlc.New(pool), pool: pool}
}

// InsertUser creates a new user record.
func (s *Store) InsertUser(ctx context.Context, u *User) (*User, error) {
	row, err := s.q.InsertUser(ctx, sqlc.InsertUserParams{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email, // service normalizes to lowercase before insert
		AvatarUrl: ptr.ToNonZero(u.AvatarURL),
	})
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserting user: %w", err)
	}
	return toUser(row), nil
}

// FindUserByID retrieves a user by their ID.
func (s *Store) FindUserByID(ctx context.Context, id string) (*User, error) {
	row, err := s.q.FindUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user %s: %w", id, err)
	}
	return toUser(row), nil
}

// FindUserByEmail retrieves a user by their email address.
func (s *Store) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	row, err := s.q.FindUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	return toUser(row), nil
}

// UpdateUserEmailVerified marks a user's email as verified.
func (s *Store) UpdateUserEmailVerified(ctx context.Context, userID string) error {
	if err := s.q.UpdateUserEmailVerified(ctx, userID); err != nil {
		return fmt.Errorf("verifying email for user %s: %w", userID, err)
	}
	return nil
}

// UpdateUserLastLogin updates the user's last login timestamp.
func (s *Store) UpdateUserLastLogin(ctx context.Context, userID string) error {
	if err := s.q.UpdateUserLastLogin(ctx, userID); err != nil {
		return fmt.Errorf("updating last login for user %s: %w", userID, err)
	}
	return nil
}

// UpdateUserProfile partially updates a user's profile information.
func (s *Store) UpdateUserProfile(ctx context.Context, userID string, input UpdateProfileInput) (*User, error) {
	var u db.Update

	if input.Name.Set {
		u.Set("name", input.Name.Value)
	}
	if input.AvatarURL.Set {
		u.Set("avatar_url", input.AvatarURL.Value)
	}
	if input.Timezone.Set {
		u.Set("timezone", input.Timezone.Value)
	}

	if u.IsEmpty() {
		return s.FindUserByID(ctx, userID)
	}

	sql, args := u.Build("users", "id = ? AND deleted_at IS NULL", userID)
	rows, _ := s.pool.Query(ctx, sql, args...)
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[sqlc.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating profile for user %s: %w", userID, err)
	}
	return toUser(row), nil
}

// InsertAccount creates a new authentication account (password or OAuth).
func (s *Store) InsertAccount(ctx context.Context, a *Account) (*Account, error) {
	row, err := s.q.InsertAccount(ctx, sqlc.InsertAccountParams{
		ID:                a.ID,
		UserID:            a.UserID,
		Provider:          a.Provider,
		ProviderAccountID: a.ProviderAccountID,
		PasswordHash:      ptr.ToNonZero(a.PasswordHash),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting account: %w", err)
	}
	return toAccount(row), nil
}

// FindAccountByProvider retrieves an account by its OAuth provider and provider ID.
func (s *Store) FindAccountByProvider(ctx context.Context, provider, providerAccountID string) (*Account, error) {
	row, err := s.q.FindAccountByProvider(ctx, sqlc.FindAccountByProviderParams{
		Provider:          provider,
		ProviderAccountID: providerAccountID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying account by provider: %w", err)
	}
	return toAccount(row), nil
}

// UpdateAccountPassword updates the hashed password for an account.
func (s *Store) UpdateAccountPassword(ctx context.Context, accountID, passwordHash string) error {
	if err := s.q.UpdateAccountPassword(ctx, sqlc.UpdateAccountPasswordParams{
		ID:           accountID,
		PasswordHash: ptr.ToNonZero(passwordHash),
	}); err != nil {
		return fmt.Errorf("updating password for account %s: %w", accountID, err)
	}
	return nil
}

// InsertSession creates a new active session.
func (s *Store) InsertSession(ctx context.Context, sess *Session) (*Session, error) {
	row, err := s.q.InsertSession(ctx, sqlc.InsertSessionParams{
		ID:        sess.ID,
		UserID:    sess.UserID,
		TokenHash: hashToken(sess.Token), // hash plain token before storing
		ExpiresAt: sess.ExpiresAt,
		IpAddress: ptr.ToNonZero(sess.IPAddress),
		UserAgent: ptr.ToNonZero(sess.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting session: %w", err)
	}
	// Restore plain token — store only holds the hash, caller needs the plain value.
	created := toSession(row)
	created.Token = sess.Token
	return created, nil
}

// FindSessionByToken retrieves a session and its associated user by the plain token.
func (s *Store) FindSessionByToken(ctx context.Context, plainToken string) (*User, *Session, error) {
	row, err := s.q.FindSessionByTokenHash(ctx, hashToken(plainToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("querying session: %w", err)
	}

	user := &User{
		ID:              row.ID,
		Name:            row.Name,
		Email:           row.Email,
		EmailVerifiedAt: row.EmailVerifiedAt,
		AvatarURL:       ptr.From(row.AvatarUrl),
		Status:          row.Status,
		OnboardedAt:     row.OnboardedAt,
		Timezone:        row.Timezone,
		LastLoginAt:     row.LastLoginAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	sess := &Session{
		ID:        row.SessionID,
		ExpiresAt: row.SessionExpiresAt,
		CreatedAt: row.SessionCreatedAt,
	}
	return user, sess, nil
}

// DeleteSession invalidates a specific session by its ID.
func (s *Store) DeleteSession(ctx context.Context, id string) error {
	if err := s.q.DeleteSession(ctx, id); err != nil {
		return fmt.Errorf("deleting session %s: %w", id, err)
	}
	return nil
}

// DeleteUserSessions invalidates all active sessions for a user.
func (s *Store) DeleteUserSessions(ctx context.Context, userID string) error {
	if err := s.q.DeleteUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("deleting sessions for user %s: %w", userID, err)
	}
	return nil
}

// InsertVerificationToken stores a new email verification or password reset token.
func (s *Store) InsertVerificationToken(ctx context.Context, vt *VerificationToken) error {
	if err := s.q.InsertVerificationToken(ctx, sqlc.InsertVerificationTokenParams{
		ID:        vt.ID,
		UserID:    vt.UserID,
		Email:     vt.Email,
		TokenHash: hashToken(vt.PlainToken), // hash plain token before storing
		Type:      vt.Type,
		ExpiresAt: vt.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("inserting verification token: %w", err)
	}
	return nil
}

// FindVerificationTokenByToken retrieves a token by its plain value.
func (s *Store) FindVerificationTokenByToken(ctx context.Context, plainToken string) (*VerificationToken, error) {
	row, err := s.q.FindVerificationTokenByHash(ctx, hashToken(plainToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("querying verification token: %w", err)
	}
	return toVerificationToken(row), nil
}

// MarkVerificationTokenUsed marks a token as consumed so it cannot be reused.
func (s *Store) MarkVerificationTokenUsed(ctx context.Context, id string) error {
	if err := s.q.MarkVerificationTokenUsed(ctx, id); err != nil {
		return fmt.Errorf("marking verification token %s used: %w", id, err)
	}
	return nil
}

// CountRecentVerificationTokens returns the number of tokens issued recently for rate limiting.
func (s *Store) CountRecentVerificationTokens(ctx context.Context, email, tokenType string) (int, error) {
	count, err := s.q.CountRecentVerificationTokens(ctx, sqlc.CountRecentVerificationTokensParams{
		Email: email,
		Type:  tokenType,
	})
	if err != nil {
		return 0, fmt.Errorf("counting recent verification tokens: %w", err)
	}
	return int(count), nil
}

func toUser(row sqlc.User) *User {
	return &User{
		ID:              row.ID,
		Name:            row.Name,
		Email:           row.Email,
		EmailVerifiedAt: row.EmailVerifiedAt,
		AvatarURL:       ptr.From(row.AvatarUrl),
		Status:          row.Status,
		OnboardedAt:     row.OnboardedAt,
		Timezone:        row.Timezone,
		LastLoginAt:     row.LastLoginAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toSession(row sqlc.Session) *Session {
	return &Session{
		ID:        row.ID,
		UserID:    row.UserID,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
}

func toAccount(row sqlc.Account) *Account {
	return &Account{
		ID:                row.ID,
		UserID:            row.UserID,
		Provider:          row.Provider,
		ProviderAccountID: row.ProviderAccountID,
		PasswordHash:      ptr.From(row.PasswordHash),
	}
}

func toVerificationToken(row sqlc.VerificationToken) *VerificationToken {
	return &VerificationToken{
		ID:        row.ID,
		UserID:    row.UserID,
		Email:     row.Email,
		Type:      row.Type,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    row.UsedAt,
		CreatedAt: row.CreatedAt,
	}
}
