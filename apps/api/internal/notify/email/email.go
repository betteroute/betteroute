// Package email implements notify.Notifier using the Resend API.
package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

// Mailer sends transactional emails via the Resend API.
type Mailer struct {
	client *resend.Client
	from   string
	tmpl   *tmpl
}

// New creates a Mailer backed by the Resend API.
func New(apiKey, from string) *Mailer {
	return &Mailer{
		client: resend.NewClient(apiKey),
		from:   from,
		tmpl:   loadTemplates(),
	}
}

func (m *Mailer) SendVerificationEmail(_ context.Context, to, name, url string) error {
	return m.send(to, "Verify your email — Betteroute", "verification", map[string]string{
		"Name": name,
		"URL":  url,
	})
}

func (m *Mailer) SendPasswordResetEmail(_ context.Context, to, name, url string) error {
	return m.send(to, "Reset your password — Betteroute", "password_reset", map[string]string{
		"Name": name,
		"URL":  url,
	})
}

// send renders a template and delivers the email via Resend.
func (m *Mailer) send(to, subject, tmplName string, data any) error {
	body, err := m.tmpl.render(tmplName, data)
	if err != nil {
		return fmt.Errorf("rendering %s: %w", tmplName, err)
	}

	_, err = m.client.Emails.Send(&resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	})
	if err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}
	return nil
}
