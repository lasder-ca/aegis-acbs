# Security policy

Aegis ACBS processes local graph files and can optionally expose a local HTTP interface. Treat externally supplied OSM, DIMACS, and Aegis graph files as untrusted input.

## Supported versions

Only the latest research-preview tag receives security fixes. The project is pre-1.0 and does not promise long-term support for older previews.

## Reporting a vulnerability

Use a private GitHub security advisory in `lasder-ca/aegis-acbs`. Do not open a public issue for a vulnerability before a fix or coordinated disclosure plan exists.

Include:

- affected version and platform,
- a minimal reproducer or malformed input,
- observed impact,
- expected behavior,
- and whether the issue is reachable through the local HTTP server.

## Deployment guidance

- Run imports with ordinary user privileges.
- Keep the Web UI bound to `127.0.0.1` unless network exposure is explicitly required.
- Apply operating-system resource limits when importing untrusted large files.
- Verify release assets against `SHA256SUMS`.
- Do not treat this research preview as a hardened routing service without independent review.
