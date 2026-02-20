package notify

import (
	"context"
	"log/slog"
)

// logNotifier logs emails to stdout instead of sending them.
// Used in development when no email API key is configured.
type logNotifier struct{}

// Log returns a Notifier that logs emails instead of sending them.
func Log() Notifier { return logNotifier{} }

func (logNotifier) SendVerificationEmail(_ context.Context, to, _, url string) error {
	slog.Info("verification email", "to", to, "url", url)
	return nil
}

func (logNotifier) SendPasswordResetEmail(_ context.Context, to, _, url string) error {
	slog.Info("password reset email", "to", to, "url", url)
	return nil
}
