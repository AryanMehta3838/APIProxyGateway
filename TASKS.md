# TASKS.md

# API Gateway Proxy Task Breakdown

## Workflow rules
- Complete one task at a time.
- Do not start the next task until validation passes.
- Keep each diff small and easy to review.
- If a task grows beyond roughly 4 to 8 touched files or needs a new concept, split it.
- After each stop point, do a human review before continuing.

## Status legend
- `[ ]` not started
- `[-]` in progress
- `[x]` done
- `[R]` human review required before continuing

---

## Next pending task
- **Task ID:** T10
- **Title:** Redis-backed rate limiting
- **Goal:** Add Redis-backed limiter behind the same policy shape.
- **Why next:** The in-memory limiter proves the policy path locally; the next step is the distributed-backed implementation.
- **Recommended tool:** Cursor Cloud Agent
- **Recommended effort:** medium
- **Human review required after completion:** yes (Stop point D review should happen before continuing)

---

## [x] T01 — Bootstrap skeleton + typed config + health endpoint
- **Goal:** Create the initial Go project structure, a typed config loader, startup validation, and `/healthz`.
- **Files likely involved:**
  - `go.mod`
  - `cmd/gateway/main.go`
  - `internal/config/*`
  - `internal/admin/*`
  - `configs/gateway.dev.yaml`
  - `README.md`
- **Validation:**
  - `go build ./...`
  - start app with valid config
  - open `/healthz` in browser and get 200
  - start app with broken config and confirm startup fails with clear error
- **Done condition:** Project starts cleanly, reads config, and exposes health endpoint.
- **Recommended tool:** Codex local
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [R] Stop point A
Human review:
- folder/package structure still feels simple
- config shape is understandable
- health endpoint response shape is acceptable

## [x] T02 — Add readiness endpoint + graceful shutdown
- **Goal:** Add `/readyz` and graceful server shutdown behavior.
- **Files likely involved:**
  - `cmd/gateway/main.go`
  - `internal/admin/*`
- **Validation:**
  - `go test ./...`
  - start app, hit `/readyz`
  - send interrupt signal and confirm clean shutdown without panic
- **Done condition:** Operational endpoints are stable and shutdown path is in place.
- **Recommended tool:** Codex local
- **Recommended effort:** low
- **Browser/user-style testing required:** yes

## [x] T03 — Add fixture echo upstream for end-to-end testing
- **Goal:** Add a tiny fake upstream service used by tests and local demos.
- **Files likely involved:**
  - `internal/testkit/*`
  - `examples/echo-upstream/*` or test helper package
  - `README.md`
- **Validation:**
  - start fixture upstream locally
  - call it directly and confirm echo behavior
- **Done condition:** A reviewer can run a stable upstream fixture without hand-building one.
- **Recommended tool:** Cursor local agent
- **Recommended effort:** low
- **Browser/user-style testing required:** yes

## [x] T04 — Single static proxy route happy path
- **Goal:** Proxy one configured route to the echo upstream.
- **Files likely involved:**
  - `internal/router/*`
  - `internal/proxy/*`
  - `cmd/gateway/main.go`
  - `configs/gateway.dev.yaml`
- **Validation:**
  - `go test ./...`
  - gateway forwards request to upstream
  - response status/body preserved
  - browser or API client can hit gateway path and receive upstream response
- **Done condition:** One route proxies end to end from config.
- **Recommended tool:** Codex local
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [R] Stop point B
Human review:
- request path through code is still obvious
- no premature abstractions appeared in router/proxy split
- config-to-route mapping reads cleanly

## [x] T05 — Path prefix stripping / rewriting
- **Goal:** Support `strip_prefix` so public route shape can differ from upstream shape.
- **Files likely involved:**
  - `internal/proxy/*`
  - `internal/config/*`
  - tests
- **Validation:**
  - unit tests for rewrite logic
  - end-to-end request proves stripped path reaches upstream
- **Done condition:** Rewrites are deterministic and tested.
- **Recommended tool:** Codex local
- **Recommended effort:** low
- **Browser/user-style testing required:** yes

## [x] T06 — Request ID middleware + propagation
- **Goal:** Generate request IDs when absent and forward them upstream.
- **Files likely involved:**
  - `internal/middleware/*`
  - `internal/proxy/*`
  - tests
- **Validation:**
  - request without ID gets one in response/logs/upstream-visible headers
  - request with existing ID preserves it
- **Done condition:** Traceability works end to end.
- **Recommended tool:** Codex local
- **Recommended effort:** low
- **Browser/user-style testing required:** yes

## [x] T07 — Structured logging middleware
- **Goal:** Add JSON request logs with status, latency, route name, and request ID.
- **Files likely involved:**
  - `internal/middleware/*`
  - `internal/telemetry/*` maybe minimal helpers
- **Validation:**
  - log output is machine-readable JSON
  - includes request ID and route name
  - sensitive headers are not dumped
- **Done condition:** Logs are useful for a demo and operational debugging.
- **Recommended tool:** Cursor local agent
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [R] Stop point C
Human review:
- log schema is acceptable
- request ID field names are consistent
- no noisy or unsafe logging

## [x] T08 — Upstream timeout handling + failure mapping
- **Goal:** Add per-route timeouts and map failures to clean 502/504 responses.
- **Files likely involved:**
  - `internal/proxy/*`
  - `internal/config/*`
  - integration tests
- **Validation:**
  - slow upstream returns 504
  - unreachable upstream returns 502
  - error response includes request ID
- **Done condition:** Failure behavior is intentional and test-covered.
- **Recommended tool:** Codex local
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [x] T09 — In-memory rate limiting
- **Goal:** Add simple per-route in-memory rate limiting as the first policy implementation.
- **Files likely involved:**
  - `internal/ratelimit/*`
  - `internal/middleware/*`
  - `internal/config/*`
- **Validation:**
  - burst within policy passes
  - next request exceeds policy and returns 429
  - health endpoints stay exempt
- **Done condition:** Basic policy enforcement works without Redis.
- **Recommended tool:** Codex local
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [R] Stop point D
Human review:
- limiter API is still narrow
- rate-limit config is readable
- no distributed concerns leaked into base design

## [ ] T10 — Redis-backed rate limiting
- **Goal:** Add Redis-backed limiter behind the same policy shape.
- **Files likely involved:**
  - `internal/ratelimit/*`
  - `deployments/docker-compose.yml`
  - `configs/gateway.dev.yaml`
  - tests
- **Validation:**
  - `docker compose up` starts Redis and gateway
  - rate limits hold across repeated requests
  - clear behavior when Redis is down
- **Done condition:** Gateway supports realistic multi-instance-ready rate limiting.
- **Recommended tool:** Cursor Cloud Agent
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [ ] T11 — Metrics endpoint
- **Goal:** Expose request count, latency, status, and rate-limit metrics.
- **Files likely involved:**
  - `internal/telemetry/*`
  - `internal/middleware/*`
  - `internal/admin/*`
- **Validation:**
  - `/metrics` renders metrics text
  - traffic increments counters
  - 429 and 5xx paths are visible
- **Done condition:** Operational visibility is demoable.
- **Recommended tool:** Codex local
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [ ] T12 — Route listing debug endpoint
- **Goal:** Add a small debug endpoint that lists loaded routes and policies.
- **Files likely involved:**
  - `internal/admin/*`
  - `internal/config/*`
  - `internal/router/*`
- **Validation:**
  - browser-openable endpoint shows route names and key settings
  - no secrets exposed
- **Done condition:** A reviewer can inspect config effects without reading code.
- **Recommended tool:** Cursor local agent
- **Recommended effort:** low
- **Browser/user-style testing required:** yes

## [R] Stop point E
Human review:
- metrics names look stable
- debug output is useful but not risky
- project is now impressive enough without adding complexity for its own sake

## [ ] T13 — Demo flow + regression hardening
- **Goal:** Add a polished local demo flow, regression checks, and docs for reviewers.
- **Files likely involved:**
  - `README.md`
  - `examples/*`
  - test files
  - helper scripts
- **Validation:**
  - fresh clone can be run with documented commands
  - full smoke flow passes
  - reviewer can reproduce rate limit and timeout behavior in minutes
- **Done condition:** Project is review-ready and can be presented confidently.
- **Recommended tool:** Cursor local agent
- **Recommended effort:** medium
- **Browser/user-style testing required:** yes

## [ ] T14 — PR review pass and bug sweep
- **Goal:** Run a final bug-focused pass using agent review plus manual review.
- **Files likely involved:**
  - touched files only
- **Validation:**
  - lint/build/tests clean
  - no obvious API inconsistencies
  - no dead code from earlier slices
- **Done condition:** Codebase is ready for portfolio/demo use.
- **Recommended tool:** Cursor Cloud Agent + Bugbot or Codex cloud review
- **Recommended effort:** high
- **Browser/user-style testing required:** yes
