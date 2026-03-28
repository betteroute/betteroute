// Package wellknown serves /.well-known endpoints for deep linking:
// apple-app-site-association (iOS Universal Links) and assetlinks.json (Android App Links).
//
// These files tell the OS which apps are allowed to open links on a given domain.
// Only workspace apps (custom domains) are served — platform apps handle their own AASA.
package wellknown

type aasaResponse struct {
	AppLinks aasaAppLinks `json:"applinks"`
}

type aasaAppLinks struct {
	Details []aasaDetail `json:"details"`
}

type aasaDetail struct {
	AppIDs     []string        `json:"appIDs"`
	Components []aasaComponent `json:"components"`
}

type aasaComponent struct {
	Path string `json:"/"` // JSON key is literally "/"
}

type assetLinkStatement struct {
	Relation []string        `json:"relation"`
	Target   assetLinkTarget `json:"target"`
}

type assetLinkTarget struct {
	Namespace              string   `json:"namespace"`
	PackageName            string   `json:"package_name"`
	SHA256CertFingerprints []string `json:"sha256_cert_fingerprints"`
}

// WorkspaceApp is a read-only projection of workspace_apps used for AASA/assetlinks.
type WorkspaceApp struct {
	TeamID             *string
	BundleID           *string
	PackageName        *string
	SHA256Fingerprints []string
}
