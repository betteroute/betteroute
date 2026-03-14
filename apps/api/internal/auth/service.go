package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/execrc/betteroute/internal/notify"
)

const (
	sessionDuration  = 30 * 24 * time.Hour // 30-day rolling session
	magicLinkTTL     = 15 * time.Minute    // Magic link expires quickly for security
	rateLimitPerHour = 3                   // max magic link emails per email per hour
)

// Service implements auth business logic.
type Service struct {
	store    *Store
	notifier notify.AuthNotifier
	cfg      Config
	oauth    oauthProviders
}

// NewService creates a new auth service.
func NewService(store *Store, notifier notify.AuthNotifier, cfg Config) *Service {
	return &Service{
		store:    store,
		notifier: notifier,
		cfg:      cfg,
		oauth:    initOAuth(cfg),
	}
}

// SendMagicLink issues a one-time token and sends it via email.
// If the user doesn't exist, they are auto-provisioned to streamline onboarding.
func (s *Service) SendMagicLink(ctx context.Context, input MagicLinkInput) error {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if err := s.checkRateLimit(ctx, input.Email, "magic_link"); err != nil {
		return err
	}

	user, err := s.store.FindUserByEmail(ctx, input.Email)
	if err != nil {
		// Auto-provision user. Name is optional — fall back to email prefix.
		if input.Name == "" {
			input.Name = strings.Split(input.Email, "@")[0]
		}

		user, err = s.store.InsertUser(ctx, &User{
			ID:    newUserID(),
			Name:  input.Name,
			Email: input.Email,
		})
		if err != nil {
			return fmt.Errorf("auto-provisioning user: %w", err)
		}

		if _, err = s.store.InsertAccount(ctx, &Account{
			ID:                newAccountID(),
			UserID:            user.ID,
			Provider:          "email",
			ProviderAccountID: user.Email,
		}); err != nil {
			return fmt.Errorf("creating email account: %w", err)
		}
	}

	// Send magic link in the background to keep the API response snappy.
	go s.sendMagicLinkEmail(context.Background(), user)
	return nil
}

// VerifyMagicLink validates the token, marks the email as verified, and opens a session.
func (s *Service) VerifyMagicLink(ctx context.Context, input VerifyMagicLinkInput, meta SessionMeta) (*User, *Session, error) {
	vt, err := s.store.FindVerificationTokenByToken(ctx, input.Token)
	if err != nil {
		return nil, nil, ErrTokenInvalid
	}
	if vt.Type != "magic_link" {
		return nil, nil, ErrTokenInvalid
	}

	// Check user status before consuming the token — a suspended user
	// shouldn't burn a one-time token only to get an error.
	user, err := s.store.FindUserByID(ctx, vt.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("finding user for magic link: %w", err)
	}
	if user.Status == "suspended" || user.Status == "banned" {
		return nil, nil, ErrAccountSuspended
	}

	if err = s.store.MarkVerificationTokenUsed(ctx, vt.ID); err != nil {
		return nil, nil, fmt.Errorf("consuming magic link token: %w", err)
	}

	// First login or never verified — mark email as verified.
	if user.EmailVerifiedAt == nil {
		if verifyErr := s.store.UpdateUserEmailVerified(ctx, user.ID); verifyErr != nil {
			slog.ErrorContext(ctx, "verifying email during magic link", "error", verifyErr, "user_id", user.ID)
		}
	}

	sess, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, nil, err
	}

	// Update last_login_at asynchronously.
	go func() {
		if err := s.store.UpdateUserLastLogin(context.Background(), user.ID); err != nil {
			slog.Error("updating last login", "error", err, "user_id", user.ID)
		}
	}()

	return user, sess, nil
}

// Logout deletes the active session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.store.DeleteSession(ctx, sessionID)
}

// UpdateProfile applies partial profile updates for the authenticated user.
func (s *Service) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*User, error) {
	return s.store.UpdateUserProfile(ctx, userID, input)
}

// ValidateSession checks a plain session token and returns the associated user and session.
// Called by the auth middleware on every protected request.
func (s *Service) ValidateSession(ctx context.Context, plainToken string) (*User, *Session, error) {
	user, sess, err := s.store.FindSessionByToken(ctx, plainToken)
	if err != nil {
		return nil, nil, ErrSessionNotFound
	}
	if user.Status == "suspended" || user.Status == "banned" {
		return nil, nil, ErrAccountSuspended
	}
	return user, sess, nil
}

// FindUserByID returns a user by ID. Used by the auth middleware to resolve
// the API key creator into an auth.User for context injection.
func (s *Service) FindUserByID(ctx context.Context, id string) (*User, error) {
	return s.store.FindUserByID(ctx, id)
}

// createSession generates a token, persists the session, and returns it with the plain token set.
func (s *Service) createSession(ctx context.Context, userID string, meta SessionMeta) (*Session, error) {
	plain, _, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	sess := &Session{
		ID:        newSessionID(),
		UserID:    userID,
		Token:     plain,
		ExpiresAt: time.Now().Add(sessionDuration),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
	}

	return s.store.InsertSession(ctx, sess)
}

// checkRateLimit returns ErrRateLimited if too many tokens were issued recently.
func (s *Service) checkRateLimit(ctx context.Context, email, tokenType string) error {
	count, err := s.store.CountRecentVerificationTokens(ctx, email, tokenType)
	if err != nil {
		return fmt.Errorf("checking rate limit: %w", err)
	}
	if count >= rateLimitPerHour {
		return ErrRateLimited
	}
	return nil
}

// sendMagicLinkEmail creates a token and sends the magic link email.
// Intended to run in a goroutine — errors are logged, not returned.
func (s *Service) sendMagicLinkEmail(ctx context.Context, user *User) {
	plain, _, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "generating magic link token", "error", err, "user_id", user.ID)
		return
	}

	if err = s.store.InsertVerificationToken(ctx, &VerificationToken{
		ID:         newTokenID(),
		UserID:     user.ID,
		Email:      user.Email,
		PlainToken: plain,
		Type:       "magic_link",
		ExpiresAt:  time.Now().Add(magicLinkTTL),
	}); err != nil {
		slog.ErrorContext(ctx, "inserting magic link token", "error", err, "user_id", user.ID)
		return
	}

	url := s.cfg.WebURL + "/verify?token=" + plain
	if err = s.notifier.SendMagicLinkEmail(ctx, user.Email, user.Name, url); err != nil {
		slog.ErrorContext(ctx, "sending magic link email", "error", err, "user_id", user.ID)
	}
}
