// Package notify defines the Notifier interface for sending transactional
// and marketing notifications across channels (email, SMS, WhatsApp, etc.).
// Implementations live in subpackages: notify/email, notify/sms, etc.
package notify

import "context"

// Notifier is the interface for sending user-facing notifications.
type Notifier interface {
	SendVerificationEmail(ctx context.Context, to, name, verificationURL string) error
	SendPasswordResetEmail(ctx context.Context, to, name, resetURL string) error
}
