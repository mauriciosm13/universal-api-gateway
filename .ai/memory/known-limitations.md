# Known Limitations (v0.1)

- Regex routing and URL rewrite not implemented (M1)
- Auth and rate limiting wired as no-op passthrough until M2/M3
- Gateway 4xx/5xx errors are JSON; upstream/proxy errors remain stdlib format
- No real JWT/API key enforcement yet (M2)
- Configuration is environment-only (no YAML)
- OpenTelemetry enabled but HTTP handlers not instrumented until M4
- Documentation website deferred to M12
