# API Gateway Proxy

## Overview

Spec-first API gateway in Go: YAML-defined routes, reverse proxy to upstreams, request IDs, structured access logs, optional rate limiting (in-memory or Redis), Prometheus-style metrics, health/readiness, and debug endpoints. Intended for demos, learning, and portfolio review.

## Tech stack

- **Go** (see `go.mod` for version)
- **chi** HTTP router
- **YAML** configuration (`gopkg.in/yaml.v3`)
- **Redis** (optional) for distributed rate limits via `github.com/redis/go-redis/v9`

## Quick start

Prerequisites: Go toolchain matching `go.mod`, `git`.

```bash
git clone <repository-url>
cd Apiproxy
go build ./...
go test ./...
```

### Run gateway + echo upstream (no Docker)

Two terminals:

```bash
# Terminal 1 — echo upstream (default :9091)
go run ./examples/echo-upstream
```

```bash
# Terminal 2 — gateway (uses in-memory rate limiting in configs/gateway.dev.yaml)
go run ./cmd/gateway -config configs/gateway.dev.yaml
```

Smoke checks:

```bash
curl -sS "http://127.0.0.1:8080/healthz"
curl -sS "http://127.0.0.1:8080/readyz"
curl -sS "http://127.0.0.1:8080/api/echo/hello?name=aryan"
```

The proxied JSON should show `"path":"/hello"` (prefix `/api/echo` is stripped per config).

### One-command regression (build / vet / test)

```bash
./scripts/smoke.sh
```

### Full stack with Docker (Redis-backed rate limits)

Uses `configs/gateway.docker.yaml` (Redis enabled; upstream `echo-upstream` hostname).

```bash
docker compose -f deployments/docker-compose.yml up --build
```

Then hit `http://localhost:8080` the same way as above.

## Configuration

| File | Purpose |
|------|---------|
| `configs/gateway.dev.yaml` | Local development: **Redis off** by default; in-memory rate limiting so you do not need Redis running. |
| `configs/gateway.docker.yaml` | Docker Compose: Redis on, upstream points at `echo-upstream` service. |

Important keys:

- **`server.port`** — gateway listen port.
- **`redis.enabled`** — if `true`, rate-limited routes need a reachable Redis at **`redis.addr`** or they return **503** on those routes.
- **`routes[]`** — `name`, `path_prefix`, `strip_prefix`, `upstream`, `timeout_ms`, `methods`, `rate_limit`.

See `PRD.md` / `ARCHITECTURE.md` for the full model.

### Environment variables

None are required. See `.env.example` for a short note.

## Project structure

```text
cmd/gateway/          # main entrypoint
configs/              # YAML configs (dev vs docker)
deployments/          # docker-compose
examples/echo-upstream/
internal/
  admin/              # health, ready, metrics, debug routes
  config/             # load + validate YAML
  middleware/         # request ID, access log, metrics, rate limit
  proxy/              # reverse proxy + timeouts + errors
  ratelimit/          # in-memory + Redis
  router/             # chi wiring
  telemetry/          # Prometheus metrics
  testkit/            # echo handler for tests
scripts/              # smoke.sh
```

## Development

| Command | Description |
|---------|-------------|
| `go build ./...` | Compile all packages |
| `go vet ./...` | Static checks |
| `go test ./...` | Unit and integration tests |
| `./scripts/smoke.sh` | `build` + `vet` + `test` |

Suggested commit style: [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, etc.).

## Demo walkthrough (TEST_PLAN flows)

Run gateway + echo as in **Quick start**. Optional: `curl -i` to see headers.

| Flow | What to do |
|------|------------|
| **A — Happy path** | `curl -sS "http://127.0.0.1:8080/api/echo/hello?name=aryan"` — JSON from upstream. `curl -sS "http://127.0.0.1:8080/debug/routes"` — loaded routes (no secrets). |
| **B — Request ID** | `curl -i "http://127.0.0.1:8080/api/echo/hello"` — response includes `X-Request-ID`. |
| **C — Rate limit** | With dev config (`requests: 3`), run four requests: `for i in 1 2 3 4; do curl -s -o /dev/null -w "%{http_code}\n" "http://127.0.0.1:8080/api/echo/hello"; done` — expect `200` then `429`. With **Redis enabled** and Redis down, limited routes return **503**. |
| **D — Timeout / 502** | Covered by automated tests: `go test ./internal/proxy -run 'TestNew_TimeoutReturns504WithRequestID|TestNew_UnreachableUpstreamReturns502WithRequestID' -v` |
| **E — Metrics** | `curl -sS "http://127.0.0.1:8080/metrics"` — Prometheus text after traffic. |

## Data & SQL

No database; route state is static YAML at startup.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|----------------|-----|
| `listen tcp :8080: address already in use` | Another process on 8080 | Stop the other process or change `server.port` in your YAML. |
| `rate limit backend unavailable` (503) on `/api/echo/...` | `redis.enabled: true` but Redis not reachable | Start Redis (e.g. `docker run --rm -p 6379:6379 redis:7-alpine`) or set **`redis.enabled: false`** in your config for in-memory limits. |
| Upstream errors / empty body | Echo not running | Start `go run ./examples/echo-upstream` (or match `upstream` URL in config). |
| Docker Compose gateway cannot reach echo | Wrong `upstream` in config | Use `configs/gateway.docker.yaml` (upstream `http://echo-upstream:9091`). |

## Validate

```bash
./scripts/smoke.sh
```

Invalid config example (should exit with a clear error):

```bash
cp configs/gateway.dev.yaml /tmp/gateway.invalid.yaml
perl -0pi -e 's/port: 8080/port: 0/' /tmp/gateway.invalid.yaml
go run ./cmd/gateway -config /tmp/gateway.invalid.yaml
```

### Redis-backed rate limiting (optional)

To exercise Redis on the laptop:

1. Start Redis: `docker run --rm -p 6379:6379 redis:7-alpine`
2. Set **`redis.enabled: true`** in a copy of `configs/gateway.dev.yaml` (or temporarily edit), **`addr: localhost:6379`**
3. Run gateway + echo; use the rate-limit loop from the table above (expect **429** on the 4th request with `requests: 3`).
