# Security Policy

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, use [GitHub Security Advisories](https://github.com/pscheid92/uuid/security/advisories/new) to privately report a vulnerability.

You should receive an acknowledgment within 48 hours. We aim to provide an initial assessment within 1 week.

## Security Design

This library delegates all randomness to Go's `crypto/rand`, which uses the operating system's CSPRNG. There is no custom random number generation, no global mutable random source, and no option to substitute a weaker source.
