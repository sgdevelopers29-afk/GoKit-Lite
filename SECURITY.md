# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 1.0.x   | ✅ Active support  |
| < 1.0   | ❌ Not supported   |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in GoKit-Lite, please report it responsibly.

### How to Report

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report vulnerabilities by emailing:

📧 **sgdevelopers29@gmail.com**

Include the following in your report:

- **Description** of the vulnerability.
- **Steps to reproduce** the issue.
- **Impact assessment** — what can an attacker do?
- **Affected package(s)** and version(s).
- **Suggested fix** (if you have one).

### What to Expect

| Timeframe | Action |
|-----------|--------|
| **48 hours** | We will acknowledge receipt of your report. |
| **7 days** | We will provide an initial assessment and severity classification. |
| **30 days** | We aim to release a fix for confirmed vulnerabilities. |

### After Reporting

- We will work with you to understand and validate the vulnerability.
- We will credit you in the release notes (unless you prefer to remain anonymous).
- We will coordinate disclosure timing with you.
- We will issue a patch release and update the CHANGELOG.

### Scope

The following are in scope for security reports:

- Authentication bypass in the `auth` package
- Token forgery or signature validation issues
- Input validation bypass in the `validator` package
- Race conditions leading to data corruption
- Denial of service through resource exhaustion

### Out of Scope

- Vulnerabilities in dependencies (report those upstream)
- Issues that require physical access to the server
- Social engineering attacks

## Security Best Practices

When using GoKit-Lite in production:

1. **Never hard-code secrets** — use environment variables or a secrets manager for `auth.SetSecret()`.
2. **Set appropriate token durations** — shorter durations reduce the window of compromise.
3. **Validate all inputs** — use the `validator` package on every user-facing endpoint.
4. **Keep dependencies updated** — run `go get -u` periodically.
5. **Enable the race detector in CI** — run `go test -race ./...` in your pipeline.

## Acknowledgments

We appreciate the security research community's efforts in helping us keep GoKit-Lite secure.
