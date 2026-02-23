package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"

	"github.com/execrc/betteroute/internal/notify"
)

const (
	sessionDuration  = 30 * 24 * time.Hour // 30-day rolling session
	verificationTTL  = 24 * time.Hour      // email verification link
	passwordResetTTL = 1 * time.Hour       // password reset link (shorter for security)
	rateLimitPerHour = 3                   // max verification/reset emails per email per hour
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

// Register creates a new user with a credential account and opens a session.
// A verification email is sent asynchronously — registration succeeds regardless.
func (s *Service) Register(ctx context.Context, input RegisterInput, meta SessionMeta) (*User, *Session, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	hash, err := argon2id.CreateHash(input.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.store.InsertUser(ctx, &User{
		ID:    newUserID(),
		Name:  input.Name,
		Email: input.Email,
	})
	if err != nil {
		return nil, nil, err // ErrEmailTaken or wrapped DB error
	}

	if _, err = s.store.InsertAccount(ctx, &Account{
		ID:                newAccountID(),
		UserID:            user.ID,
		Provider:          "credential",
		ProviderAccountID: user.Email, // email is the stable credential identifier
		PasswordHash:      hash,
	}); err != nil {
		return nil, nil, fmt.Errorf("creating credential account: %w", err)
	}

	sess, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, nil, err
	}

	// Send verification email in the background — never block registration on email.
	go s.sendVerificationEmail(context.Background(), user)

	return user, sess, nil
}

// Login authenticates via email/password and opens a session.
func (s *Service) Login(ctx context.Context, input LoginInput, meta SessionMeta) (*User, *Session, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.store.FindUserByEmail(ctx, input.Email)
	if err != nil {
		// Return a generic error — never reveal whether the email exists.
		return nil, nil, ErrInvalidCredential
	}

	if user.Status == "suspended" || user.Status == "banned" {
		return nil, nil, ErrAccountSuspended
	}

	acc, err := s.store.FindAccountByProvider(ctx, "credential", user.Email)
	if err != nil {
		return nil, nil, ErrInvalidCredential
	}

	ok, err := argon2id.ComparePasswordAndHash(input.Password, acc.PasswordHash)
	if err != nil || !ok {
		return nil, nil, ErrInvalidCredential
	}

	sess, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, nil, err
	}

	// Update last_login_at asynchronously — not critical path.
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

// VerifyEmail validates the token and marks the user's email as verified.
func (s *Service) VerifyEmail(ctx context.Context, input VerifyEmailInput) error {
	vt, err := s.store.FindVerificationTokenByToken(ctx, input.Token)
	if err != nil {
		return ErrTokenInvalid
	}
	if vt.Type != "email_verification" {
		return ErrTokenInvalid
	}

	if err = s.store.MarkVerificationTokenUsed(ctx, vt.ID); err != nil {
		return fmt.Errorf("consuming verification token: %w", err)
	}

	return s.store.UpdateUserEmailVerified(ctx, vt.UserID)
}

// ResendVerification sends a new verification email, subject to rate limiting.
func (s *Service) ResendVerification(ctx context.Context, input ResendVerificationInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if err := s.checkRateLimit(ctx, email, "email_verification"); err != nil {
		return err
	}

	user, err := s.store.FindUserByEmail(ctx, email)
	if err == nil {
		go s.sendVerificationEmail(context.Background(), user)
	}
	return nil
}

// ForgotPassword sends a password reset email, subject to rate limiting.
// Always returns nil — never reveals whether the email is registered.
func (s *Service) ForgotPassword(ctx context.Context, input ForgotPasswordInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if err := s.checkRateLimit(ctx, email, "password_reset"); err != nil {
		return err
	}

	user, err := s.store.FindUserByEmail(ctx, email)
	if err == nil {
		go s.sendPasswordResetEmail(context.Background(), user)
	}
	return nil
}

// ResetPassword validates the token, updates the password, and invalidates all sessions.
func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	vt, err := s.store.FindVerificationTokenByToken(ctx, input.Token)
	if err != nil {
		return ErrTokenInvalid
	}
	if vt.Type != "password_reset" {
		return ErrTokenInvalid
	}

	hash, err := argon2id.CreateHash(input.Password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	acc, err := s.store.FindAccountByProvider(ctx, "credential", vt.Email)
	if err != nil {
		return ErrNotFound
	}

	if err = s.store.UpdateAccountPassword(ctx, acc.ID, hash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err = s.store.MarkVerificationTokenUsed(ctx, vt.ID); err != nil {
		return fmt.Errorf("consuming reset token: %w", err)
	}

	// Sign out everywhere — compromised password means compromised sessions.
	return s.store.DeleteUserSessions(ctx, vt.UserID)
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

// sendVerificationEmail creates a token and sends the verification email.
// Intended to run in a goroutine — errors are logged, not returned.
func (s *Service) sendVerificationEmail(ctx context.Context, user *User) {
	plain, _, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "generating verification token", "error", err, "user_id", user.ID)
		return
	}

	if err = s.store.InsertVerificationToken(ctx, &VerificationToken{
		ID:         newTokenID(),
		UserID:     user.ID,
		Email:      user.Email,
		PlainToken: plain,
		Type:       "email_verification",
		ExpiresAt:  time.Now().Add(verificationTTL),
	}); err != nil {
		slog.ErrorContext(ctx, "inserting verification token", "error", err, "user_id", user.ID)
		return
	}

	url := s.cfg.WebURL + "/verify-email?token=" + plain
	if err = s.notifier.SendVerificationEmail(ctx, user.Email, user.Name, url); err != nil {
		slog.ErrorContext(ctx, "sending verification email", "error", err, "user_id", user.ID)
	}
}

// sendPasswordResetEmail creates a token and sends the reset email.
// Intended to run in a goroutine — errors are logged, not returned.
func (s *Service) sendPasswordResetEmail(ctx context.Context, user *User) {
	plain, _, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "generating reset token", "error", err, "user_id", user.ID)
		return
	}

	if err = s.store.InsertVerificationToken(ctx, &VerificationToken{
		ID:         newTokenID(),
		UserID:     user.ID,
		Email:      user.Email,
		PlainToken: plain,
		Type:       "password_reset",
		ExpiresAt:  time.Now().Add(passwordResetTTL),
	}); err != nil {
		slog.ErrorContext(ctx, "inserting reset token", "error", err, "user_id", user.ID)
		return
	}

	url := s.cfg.WebURL + "/reset-password?token=" + plain
	if err = s.notifier.SendPasswordResetEmail(ctx, user.Email, user.Name, url); err != nil {
		slog.ErrorContext(ctx, "sending password reset email", "error", err, "user_id", user.ID)
	}
}
