# Security Policy

## Reporting a Vulnerability

**Do not open a public issue.** Use [GitHub Security Advisories](https://github.com/betteroute/betteroute/security/advisories/new).

Include what the vulnerability is, steps to reproduce, potential impact, and a suggested fix if you have one.

We acknowledge reports within **48 hours** and target a fix within **7 days** depending on severity.

## In Scope

- Authentication or session bypasses
- Unauthorized access to workspaces, links, or user data
- Redirect engine manipulation (open redirects, destination tampering)
- SQL injection, XSS, CSRF, OWASP Top 10
- API key or token leakage
- Privilege escalation across roles or workspaces

## Out of Scope

- Denial of service
- Social engineering
- MITM or physical device access
- Dependency CVEs without a demonstrated exploit in Betteroute
- Self-XSS
- Missing headers on non-sensitive routes

## For Researchers

- Don't run automated scanners against production — reach out for a sandbox
- Don't access, modify, or delete other users' data
- Don't exploit beyond what's needed to demonstrate the issue
- Don't disclose publicly before a fix ships

## What We Guarantee

- 48-hour response
- No legal action if you follow the above
- Confidential handling
- Credit in release notes when the fix ships (unless you prefer anonymity)

No bug bounty program yet.

## Supported Versions

During pre-beta, security fixes land on `main` only. Supported versions table will be added at stable release.
