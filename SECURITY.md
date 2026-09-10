# Security Policy

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, use [GitHub Security Advisories](https://github.com/pscheid92/uuid/security/advisories/new) to privately report a vulnerability.

You should receive an acknowledgment within 48 hours. We aim to provide an initial assessment within 1 week.

## Security Design

This library delegates all randomness to Go's `crypto/rand`, which uses the operating system's CSPRNG. There is no custom random number generation, no global mutable random source, and no option to substitute a weaker source.

One caveat: the `Pool` type buffers pre-generated randomness in process memory to amortize `crypto/rand` overhead. Unlike direct `crypto/rand` reads, this buffer is not fork-safe — a process fork or VM snapshot/clone duplicates the buffer, and both copies can then emit identical UUIDs. Use the package-level functions (which read `crypto/rand` directly) where fork or VM-clone safety matters. The batch APIs are unaffected: they read fresh randomness on every call.
