// Package domain handles custom domain management for branded short links.
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/execrc/betteroute/internal/opt"
)

// Domain represents a custom domain belonging to a workspace.
type Domain struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspace_id"`
	CreatedBy         string     `json:"created_by,omitempty"`
	Hostname          string     `json:"hostname"`
	VerificationToken string     `json:"verification_token"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	FallbackURL       string     `json:"fallback_url,omitempty"`
	Status            string     `json:"status"`
	LastCheckedAt     *time.Time `json:"last_checked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	DNS *DNSSetup `json:"dns,omitempty"` // Attached dynamically by the handler for presentation
}

// DNSInstructions returns the DNS records the user must configure.
func (d *Domain) DNSInstructions(txtPrefix, proxyCNAME, proxyIP string) *DNSSetup {
	setup := &DNSSetup{
		TXTHost:  txtPrefix + d.Hostname,
		TXTValue: d.VerificationToken,
	}

	// Calculate if the domain is Apex (e.g. "example.com") or Subdomain (e.g. "link.example.com")
	eTLD1, err := publicsuffix.EffectiveTLDPlusOne(d.Hostname)
	isApex := err == nil && eTLD1 == d.Hostname

	if isApex {
		// Apex domains strictly require A Records
		setup.ARecordHost = "@"
		setup.ARecordValue = proxyIP
	} else {
		// Subdomains strictly require CNAME Records
		setup.CNAMEHost = d.Hostname
		setup.CNAMEValue = proxyCNAME
	}

	return setup
}

// DNSSetup describes the DNS records a user needs to add.
type DNSSetup struct {
	TXTHost      string `json:"txt_host"`
	TXTValue     string `json:"txt_value"`
	CNAMEHost    string `json:"cname_host,omitempty"`
	CNAMEValue   string `json:"cname_value,omitempty"`
	ARecordHost  string `json:"a_record_host,omitempty"`
	ARecordValue string `json:"a_record_value,omitempty"`
}

// CreateInput is the input for adding a custom domain.
type CreateInput struct {
	Hostname    string `json:"hostname"     validate:"required,hostname,min=4,max=253"`
	FallbackURL string `json:"fallback_url" validate:"omitempty,url,max=2048"`
}

// UpdateInput is the input for partially updating a domain.
type UpdateInput struct {
	FallbackURL opt.Field[*string] `json:"fallback_url" validate:"omitempty,url,max=2048" swaggertype:"string"`
}

var (
	ErrNotFound        = errors.New("domain not found")
	ErrHostnameTaken   = errors.New("hostname already in use")
	ErrAlreadyVerified = errors.New("domain is already verified")
	ErrDNSNotFound     = errors.New("no TXT record found")
	ErrDNSMismatch     = errors.New("TXT record value does not match")
)

// generateVerificationToken creates a 32-byte hex-encoded random token.
func generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
