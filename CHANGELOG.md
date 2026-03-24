# CHANGELOG.md

# Changelog

All notable changes to this project will be tracked here.

## [Unreleased]
### Added
- `internal/testkit` echo HTTP handler and unit tests; runnable `examples/echo-upstream` (default `:9091`) for local demos and future proxy validation.
- Initial spec-first project planning docs:
  - `PRD.md`
  - `ARCHITECTURE.md`
  - `TASKS.md`
  - `TEST_PLAN.md`
  - `ROUTING.md`
- Bootstrap gateway skeleton with typed YAML config loading and startup validation.
- Admin router with browser-friendly `GET /healthz`.
- Dev config, README run instructions, and initial config/admin tests.
- Browser-friendly `GET /readyz` endpoint and graceful shutdown on interrupt.
- First static proxy route wired from config to the echo upstream with router/proxy tests.
- Configurable `strip_prefix` path rewriting with proxy and integration test coverage.
- Request ID middleware that generates missing IDs, preserves provided IDs, and forwards them upstream.
- JSON access logging (`msg=access`) on stdout with `request_id`, `route_name`, `method`, `path`, `status`, `duration_ms`, and `ts`; headers are not logged. Global middleware on the chi router plus `NamedRoute` for config route names and admin endpoints.
- Per-route upstream timeouts with `504` timeout mapping, `502` upstream failure mapping, and error bodies that include the request ID.
- In-memory per-route rate limiting with `429` responses, request ID propagation on throttled responses, and health endpoint exemption.
- Prometheus-style `/metrics` endpoint with request count, latency histogram, status visibility, and explicit rate-limit-denied metrics.
- Redis-backed per-route rate limiting selected by config (`redis.enabled`), including fixed-window enforcement and startup validation for `redis.addr`.
- Integration and unit coverage for Redis limiter behavior, Redis-unavailable fallback (`503`), and shared-limit behavior across router instances.
- `deployments/docker-compose.yml` for local gateway + echo upstream + Redis startup.
- `GET /debug/routes` JSON debug endpoint listing route names, paths, methods, timeouts, rate-limit policy, and sanitized upstream targets (`upstream_scheme` / `upstream_host` only, no userinfo or path); Redis `addr` only when `redis.enabled` is true.
- **T13 — Demo / docs:** Expanded README (overview, stack, quick start, config table, layout, dev commands, TEST_PLAN demo table, troubleshooting). `configs/gateway.dev.yaml` defaults **`redis.enabled: false`** so local runs work without Redis; **`configs/gateway.docker.yaml`** + compose use Redis and service hostnames. **`scripts/smoke.sh`** runs `go build`, `go vet`, `go test`. **`.env.example`** notes no env vars; **`examples/README.md`** points to echo upstream.

### Changed
- `deployments/docker-compose.yml` gateway now uses `configs/gateway.docker.yaml`.

### Fixed
- None yet.
