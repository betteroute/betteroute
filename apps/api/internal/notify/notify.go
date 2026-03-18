// Package notify defines notification interfaces for sending transactional
// messages. Implementations live in subpackages: notify/email, notify/slack, etc.
package notify

import (
	"context"
	"log/slog"
)

// AuthNotifier sends auth-related notifications (magic link).
type AuthNotifier interface {
	SendMagicLinkEmail(ctx context.Context, to, name, magicLinkURL string) error
}

// TeamNotifier sends team collaboration notifications (workspace invites).
type TeamNotifier interface {
	SendWorkspaceInviteEmail(ctx context.Context, to, inviterName, workspaceName, inviteURL string) error
}

// Notifier combines all notification interfaces.
// Used in main.go for wiring; services depend on specific sub-interfaces.
type Notifier interface {
	AuthNotifier
	TeamNotifier
}

// nop is a Notifier that logs warnings instead of delivering.
// Used when no provider is configured so operators can detect dropped notifications.
type nop struct{}

// Nop returns a Notifier that warns on each dropped notification.
func Nop() Notifier { return nop{} }

func (nop) SendMagicLinkEmail(ctx context.Context, to, _, _ string) error {
	slog.WarnContext(ctx, "notification dropped: magic link email", "to", to)
	return nil
}

func (nop) SendWorkspaceInviteEmail(ctx context.Context, to, _, _, _ string) error {
	slog.WarnContext(ctx, "notification dropped: workspace invite email", "to", to)
	return nil
}
