# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |

## Scope

picokit is a zero-dependency Go library providing small utility packages. Security concerns include:

- Path traversal when resolving file paths or arguments
- Unsafe file operations via untrusted input
- Dependency vulnerabilities in external packages

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do not** open a public issue
2. Use GitHub's [private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
3. Include steps to reproduce and the affected function/version
4. Allow reasonable time for a fix before public disclosure

## Security Measures

This project uses:
- [CodeQL](https://codeql.github.com/) for static analysis (Go)
- [Gitleaks](https://github.com/gitleaks/gitleaks) for secret scanning
- [OpenSSF Scorecard](https://securityscorecards.dev/) for supply chain security
- SHA-pinned GitHub Actions to prevent supply chain attacks
- Dependabot for automated dependency updates
