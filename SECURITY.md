# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Betteroute, please report it responsibly. **Do not open a public issue.**

Report via [GitHub Security Advisories](https://github.com/betteroute/betteroute/security/advisories/new) with:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (optional but appreciated)

We will acknowledge reports within **48 hours** and provide a fix or mitigation plan within **7 days**, depending on severity.

## In Scope

- Authentication or session bypasses
- Unauthorized access to workspaces, links, or user data
- Redirect engine manipulation (open redirects, destination tampering)
- SQL injection, XSS, CSRF, or other OWASP Top 10 vulnerabilities
- API key or token leakage
- Privilege escalation across roles or workspaces

## Out of Scope

- Denial of service (DoS) attacks
- Social engineering
- Attacks requiring MITM or physical device access
- Dependency vulnerabilities with no demonstrated exploit in Betteroute
- Self-XSS or issues requiring physical device access
- Missing security headers on non-sensitive routes
- Rate limiting or brute force on non-production instances

## Guidelines for Researchers

- Do not run automated scanners against production infrastructure — contact us for a sandbox
- Do not access, modify, or delete other users' data
- Do not exploit a vulnerability beyond what's necessary to demonstrate it
- Do not disclose publicly until we've released a fix

## Our Commitments

- We will respond within 48 hours with an evaluation and expected timeline
- If you follow the guidelines above, we will not take legal action against you
- We will handle your report with strict confidentiality
- We will keep you informed of progress toward a resolution
- We will credit you in the release notes when the fix ships (unless you prefer to remain anonymous)

We don't currently run a bug bounty program, but we appreciate every report.

## Supported Versions

During the pre-beta phase, security fixes are applied to the `main` branch only. Once we reach stable releases, we will maintain a supported versions table here.
