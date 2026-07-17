# Security

Aegis is intended to process local graph files. Treat externally supplied OSM/DIMACS files as untrusted.

- Run imports with ordinary user privileges.
- Keep the Web UI bound to `127.0.0.1` unless network exposure is explicitly required.
- The HTTP server limits request bodies and applies timeouts and basic security headers.
- Report malformed-file crashes or excessive memory use through a private GitHub security advisory.
