# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email us at: **security@quickbill.id**

### What to include

- A description of the vulnerability
- Steps to reproduce the issue
- The potential impact
- Any suggested fixes (optional)

### Response timeline

- **Acknowledgement**: Within 2 business days
- **Initial assessment**: Within 5 business days
- **Resolution target**: Within 30 days for critical issues

### Disclosure

We follow coordinated disclosure. We will work with you to understand and address the issue before any public disclosure. We appreciate your patience and responsible reporting.

## Security Best Practices

When using this library:

- Always use TLS in production (`WithInsecure(false)`)
- Keep dependencies up to date
- Review sampling rates to avoid leaking sensitive data in traces
- Use environment variables for configuration rather than hardcoding values
