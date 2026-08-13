# Known Limitations (v0.1)

- Path, host, and method routing not implemented (M1 — path and header routing done)
- Gateway 4xx/5xx errors are JSON; upstream/proxy errors remain stdlib format
- No authentication or rate limiting
- Configuration is environment-only (no YAML)
- OpenTelemetry enabled but HTTP handlers not instrumented until M4
- Documentation website deferred to M12
