# Known Limitations (v0.1)

- Regex routing and URL rewrite not implemented (M1)
- Rate limiting wired as no-op passthrough until M3
- Gateway 4xx/5xx errors are JSON; upstream/proxy errors remain stdlib format
- JWT auth applies to all proxied traffic when enabled; no per-path public bypass (M2 Week 4+)
- JWKS keys loaded at startup; no background key rotation refresh in MVP
- Identity not propagated to upstream headers yet (M2 Week 4)
- API keys not implemented yet (M2 Week 4)
- Configuration is environment-only (no YAML)
- OpenTelemetry enabled but HTTP handlers not instrumented until M4
- Documentation website deferred to M12
