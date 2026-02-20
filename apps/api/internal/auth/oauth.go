package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"golang.org/x/oauth2/google"
)

// oauthProviders maps provider name to its oauth2 config.
// Populated in NewService if credentials are configured.
type oauthProviders struct {
	google *oauth2.Config
	github *oauth2.Config
}

// initOAuth builds the provider configs from auth.Config.
// Called once from NewService.
func initOAuth(cfg Config) oauthProviders {
	p := oauthProviders{}

	if cfg.GoogleClientID != "" {
		p.google = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.APIURL + "/api/v1/auth/oauth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	if cfg.GitHubClientID != "" {
		p.github = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.APIURL + "/api/v1/auth/oauth/github/callback",
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     endpoints.GitHub,
		}
	}

	return p
}

// providerConfig returns the oauth2 config for the given provider name.
func (s *Service) providerConfig(provider string) (*oauth2.Config, error) {
	switch provider {
	case "google":
		if s.oauth.google == nil {
			return nil, ErrOAuthUnavailable
		}
		return s.oauth.google, nil
	case "github":
		if s.oauth.github == nil {
			return nil, ErrOAuthUnavailable
		}
		return s.oauth.github, nil
	default:
		return nil, ErrOAuthUnavailable
	}
}

// OAuthURL returns the provider's authorization URL with the given state.
func (s *Service) OAuthURL(provider, state string) (string, error) {
	cfg, err := s.providerConfig(provider)
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// OAuthCallback exchanges the code, upserts the user+account, and opens a session.
func (s *Service) OAuthCallback(ctx context.Context, provider, code string, meta SessionMeta) (*User, *Session, error) {
	cfg, err := s.providerConfig(provider)
	if err != nil {
		return nil, nil, err
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("exchanging oauth code: %w", err)
	}

	user, err := s.upsertOAuthUser(ctx, provider, cfg, token)
	if err != nil {
		return nil, nil, err
	}

	sess, err := s.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, nil, err
	}

	return user, sess, nil
}

// upsertOAuthUser fetches the provider profile, then finds or creates the user.
func (s *Service) upsertOAuthUser(ctx context.Context, provider string, cfg *oauth2.Config, token *oauth2.Token) (*User, error) {
	profile, err := fetchProfile(ctx, provider, cfg, token)
	if err != nil {
		return nil, err
	}

	// Fast path: account already linked → return existing user.
	acc, err := s.store.FindAccountByProvider(ctx, provider, profile.id)
	if err == nil {
		return s.store.FindUserByID(ctx, acc.UserID)
	}

	// Slow path: new OAuth login — link to existing user or create one.
	email := strings.ToLower(strings.TrimSpace(profile.email))

	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		// No existing user → create one.
		user, err = s.store.InsertUser(ctx, &User{
			ID:        newUserID(),
			Name:      profile.name,
			Email:     email,
			AvatarURL: profile.avatarURL,
		})
		if err != nil {
			return nil, fmt.Errorf("creating oauth user: %w", err)
		}

		// Mark email verified — provider has already verified it.
		if err = s.store.UpdateUserEmailVerified(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("verifying oauth email: %w", err)
		}
	}

	// Link the OAuth account to the user (existing or newly created).
	if _, err = s.store.InsertAccount(ctx, &Account{
		ID:                newAccountID(),
		UserID:            user.ID,
		Provider:          provider,
		ProviderAccountID: profile.id,
	}); err != nil {
		return nil, fmt.Errorf("linking oauth account: %w", err)
	}

	return user, nil
}

// --- Provider profile fetching ---

// oauthProfile holds the normalized fields we care about from any provider.
type oauthProfile struct {
	id        string
	name      string
	email     string
	avatarURL string
}

func fetchProfile(ctx context.Context, provider string, cfg *oauth2.Config, token *oauth2.Token) (*oauthProfile, error) {
	switch provider {
	case "google":
		return fetchGoogleProfile(ctx, cfg, token)
	case "github":
		return fetchGitHubProfile(ctx, cfg, token)
	default:
		return nil, ErrOAuthUnavailable
	}
}

func fetchGoogleProfile(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*oauthProfile, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned %d", resp.StatusCode)
	}

	var raw struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding google userinfo: %w", err)
	}
	if !raw.EmailVerified {
		return nil, ErrEmailUnverified
	}

	return &oauthProfile{
		id:        raw.Sub,
		name:      raw.Name,
		email:     raw.Email,
		avatarURL: raw.Picture,
	}, nil
}

func fetchGitHubProfile(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*oauthProfile, error) {
	client := cfg.Client(ctx, token)

	// Fetch the user object.
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("fetching github user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading github user body: %w", err)
	}

	var raw struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err = json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decoding github user: %w", err)
	}

	name := raw.Name
	if name == "" {
		name = raw.Login // fall back to username if display name is unset
	}

	// GitHub may not include email if the user has set it private.
	email := raw.Email
	if email == "" {
		email, err = fetchGitHubPrimaryEmail(client)
		if err != nil {
			return nil, err
		}
	}

	return &oauthProfile{
		id:        fmt.Sprintf("%d", raw.ID),
		name:      name,
		email:     email,
		avatarURL: raw.AvatarURL,
	}, nil
}

func fetchGitHubPrimaryEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("fetching github emails: %w", err)
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decoding github emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified primary email on github account")
}
