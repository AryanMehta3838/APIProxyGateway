# TEST_PLAN.md

# API Gateway Proxy Test Plan

## 1. Testing principles
- Every slice must prove observable behavior, not just compile.
- Prefer a mix of unit, integration, and user-style validation.
- Treat logs, metrics, and error responses as product surface area.
- Keep test setup simple and reproducible.

## 2. Build / type / lint checks
Run on every slice:
```bash
go build ./...
go test ./...
```

Add when wiring is stable:
```bash
go vet ./...
golangci-lint run
```

Pass criteria:
- no build failures
- no failing tests
- no obvious lint regressions that hide real defects

## 3. Unit tests
Focus:
- config validation
- path rewrite logic
- request ID generation / preservation
- rate limit decision logic
- failure mapping helpers

Examples:
- invalid config missing upstream should fail validation
- `strip_prefix=/api/echo` rewrites `/api/echo/v1/ping` to `/v1/ping`
- request with no request ID gets generated ID
- request beyond quota returns deny decision
- timeout error maps to 504, connection failure to 502

Pass criteria:
- edge cases are covered
- tests do not require Redis or external processes unless specifically integration-level

## 4. Integration tests
Use `httptest` or fixture upstreams to validate:
- route assembly from config
- proxy forwarding
- response status/body passthrough
- timeout behavior
- request ID propagation
- middleware ordering
- metrics counter growth after requests

Pass criteria:
- gateway behavior is verified end to end in-process
- tests do not rely on sleeps unless unavoidable and bounded

## 5. Browser / user-flow tests
Even for a backend gateway, include reviewer-style flows that feel like product usage.

### Flow A — happy path proxy
1. Start gateway and echo upstream.
2. Open browser or API client to gateway route.
3. Send request to `/api/echo/hello?name=aryan`.
4. Confirm response comes from upstream.
5. Confirm route listing or logs make the path understandable.

### Flow B — request ID propagation
1. Send request without `X-Request-ID`.
2. Observe generated request ID in response header or debug output.
3. Confirm upstream received the same ID.

### Flow C — rate limiting
1. Hit limited route repeatedly from browser, shell loop, or API client.
2. Confirm first N requests pass.
3. Confirm next request returns 429.
4. Confirm metrics or logs reflect the throttle.

### Flow D — timeout / upstream failure
1. Route to slow or flaky upstream.
2. Trigger timeout.
3. Confirm 504 response and visible request ID.
4. Point to dead upstream and confirm 502.

### Flow E — metrics visibility
1. Generate traffic.
2. Open `/metrics`.
3. Confirm counters and latencies changed.

Suggested tools:
- browser
- curl / httpie
- Postman or Bruno
- optional tiny demo page under `examples/` for reviewers

Pass criteria:
- a human reviewer can reproduce the feature without reading tests
- visible behavior matches acceptance criteria

## 6. Visual checks
Even without a UI-heavy app, visually inspect:
- JSON log shape
- `/metrics` readability
- `/healthz` and `/readyz` response simplicity
- route listing/debug endpoint readability
- README command examples formatting

Pass criteria:
- outputs are easy to read
- no accidental secrets or noisy dumps
- field naming is consistent

## 7. Regression checks
Before closing each major stop point:
- rerun all unit and integration tests
- rerun core smoke flows
- verify health endpoints still work
- verify previously working routes still proxy
- verify rate limiting did not break non-limited routes
- verify metrics still increment after middleware changes

Regression checklist:
- routing still works
- rewrite still works
- request IDs still propagate
- failure mapping still returns correct codes
- logs still include request ID and route name
- admin endpoints remain accessible

## 8. Manual review points
### After T01
- Is the config model still simple?
- Are we adding abstractions too early?

### After T04
- Is proxy code still easy to follow?
- Did route wiring remain explicit?

### After T07
- Are logs useful and safe?

### After T10
- Did Redis support stay isolated?
- Is limiter behavior obvious when Redis is unavailable?

### After T11 / T12
- Are metrics and debug endpoints worth keeping?
- Is the project already impressive enough without extra scope?

## 9. Recommended validation cadence by task size
### Small slice
- build
- targeted tests
- one manual smoke flow

### Medium slice
- build
- full tests
- two or more user flows
- manual code review

### High-risk slice
- build
- full tests
- user flows
- regression sweep
- human review before merge

## 10. PR / review gate
A slice is mergeable only when:
- acceptance criteria are met
- validation commands are documented in the PR or task notes
- at least one user-style flow was executed
- reviewer can explain the diff in under a few minutes
