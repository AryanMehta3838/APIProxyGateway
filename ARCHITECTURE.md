# ARCHITECTURE.md

# API Gateway Proxy Architecture

## 1. Architecture goals
- clean request path
- simple local development
- easy-to-review diffs
- policies added as middleware, not baked into proxy core
- testable components with minimal mocking

## 2. Proposed project structure
```text
cmd/gateway/main.go
internal/app/
internal/config/
internal/router/
internal/proxy/
internal/middleware/
internal/ratelimit/
internal/telemetry/
internal/admin/
internal/testkit/
examples/
deployments/docker-compose.yml
configs/gateway.dev.yaml
```

## 3. High-level components
### `internal/config`
Responsible for:
- loading YAML config
- validating required fields
- returning typed configuration structs

Keep validation here, not scattered across runtime code.

### `internal/router`
Responsible for:
- translating config into chi routes
- attaching middleware stacks per route
- mounting operational endpoints separately

### `internal/proxy`
Responsible for:
- constructing reverse proxy handlers
- path rewriting / prefix stripping
- upstream timeout and error mapping
- forwarding request ID and selected headers

This is the core traffic path and should stay small.

### `internal/middleware`
Responsible for:
- request ID handling
- structured logging
- panic recovery
- rate limiting adapter wiring
- metrics recording hooks

### `internal/ratelimit`
Responsible for:
- rate limit policy model
- in-memory limiter
- Redis-backed limiter
- shared limiter interface only if both implementations need the same call shape

Start with a narrow interface:
```go
type Limiter interface {
    Allow(ctx context.Context, key string, policy Policy) (Decision, error)
}
```

### `internal/telemetry`
Responsible for:
- Prometheus-style metrics registration
- request duration / count / status metrics
- helper functions only, not business logic

### `internal/admin`
Responsible for:
- health endpoint
- readiness endpoint
- optional route listing endpoint
- optional debug/demo helpers

### `internal/testkit`
Responsible for:
- fixture upstream servers
- helper config builders
- end-to-end test helpers

## 4. Runtime request flow
1. Server starts and loads config.
2. Config is validated before serving.
3. Chi router mounts admin endpoints.
4. For each configured route, router builds middleware chain.
5. Request enters:
   - panic recovery
   - request ID
   - logging
   - metrics
   - rate limiting
   - proxy handler
6. Proxy handler rewrites the request if configured.
7. Request is forwarded to upstream.
8. Response code, duration, and metadata are recorded.

## 5. Config model
Suggested route config shape:
```yaml
server:
  port: 8080

redis:
  addr: localhost:6379
  enabled: true

routes:
  - name: echo
    path_prefix: /api/echo
    methods: [GET, POST]
    upstream: http://localhost:9091
    strip_prefix: /api/echo
    timeout_ms: 3000
    rate_limit:
      enabled: true
      requests: 5
      window_seconds: 60
      key_strategy: ip
```

## 6. Key design choices
### Static config first
Start with static YAML. It is enough for a strong demo and keeps the system explainable.

### Middleware-driven policy
Rate limiting, logging, request IDs, and metrics should wrap routes, not live inside proxy code.

### Separate operational endpoints
Health and readiness must not share business route middleware. They should always remain available and lightweight.

### Two-stage rate limiting strategy
- first implementation: in-memory limiter
- second implementation: Redis-backed limiter
This preserves forward progress while keeping early slices easy to validate.

### Explicit stop points
Human review should happen before:
- proxy core is generalized
- Redis is introduced
- metrics shape is finalized
- any feature touches more than one vertical slice at once

## 7. Error handling policy
- invalid config: fail fast on startup
- upstream timeout: return 504
- upstream connection failure: return 502
- limiter backend unavailable:
  - if Redis-only route and fail-open is false, return 503
  - otherwise degrade according to explicit policy
- panic in middleware or handler: recover and return 500 with request ID

## 8. Security and safety defaults
- do not proxy arbitrary destinations from request params
- only proxy to configured upstreams
- bound request and header sizes using server defaults or explicit settings later
- redact sensitive headers in logs
- do not log full bodies in v1
- keep debug endpoints non-sensitive

## 9. Testing architecture
Testing pyramid:
- unit tests for config validation, path rewrite logic, limiter logic
- integration tests for route wiring and proxy behavior with httptest upstreams
- user-style tests with real local processes or docker-compose
- browser-style checks through a tiny demo page or plain browser-openable endpoints

## 10. Clean architecture rules
- config types may be used by router assembly, not by lower-level runtime components unless necessary
- proxy package should not know about YAML parsing
- limiter package should not know about chi
- telemetry package should not know route config internals
- main.go should wire dependencies, not implement logic

## 11. What not to do
- no plugin system
- no generic middleware factories unless duplication is proven
- no repository pattern
- no abstract service layer with one implementation
- no dynamic config loader in v1
