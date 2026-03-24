# PRD.md

# API Gateway Proxy PRD

## 1. Product summary
Build a spec-first API gateway proxy that is small enough to understand, but polished enough to be portfolio-grade. The project should feel like a real internal platform tool rather than a toy reverse proxy.

The gateway should route requests to upstream services, enforce rate limits, expose operational visibility, and be safe to iterate on in small slices.

Assumed implementation stack for this plan:
- Go
- Chi
- Redis
- Prometheus-style metrics endpoint
- YAML config for static route definitions

## 2. Primary goal
Create a clean, demoable gateway that shows:
- route-based proxying
- per-route policy enforcement
- structured logging
- request ID propagation
- health and readiness endpoints
- metrics
- rate limiting with Redis support
- safe failure behavior
- strong validation and test discipline

## 3. Success criteria
The project is successful when a reviewer can:
1. Run one command to start the gateway and one or two fake upstreams.
2. Send requests through the gateway and see them forwarded correctly.
3. Trigger rate limiting and receive correct 429 responses.
4. Inspect logs, metrics, and health endpoints.
5. Read the codebase and see clear separation of concerns.
6. Review small commits or PRs with acceptance criteria per slice.

## 4. Target users
### Primary user
A developer or platform engineer who needs a lightweight gateway in front of internal services.

### Secondary user
A technical reviewer or recruiter evaluating system design, backend engineering, and implementation discipline.

## 5. Core user stories
### Routing
- As a developer, I can define routes in config so requests are proxied to the correct upstream.
- As a developer, I can match routes by path prefix and method.
- As a developer, I can rewrite or strip prefixes before forwarding when needed.

### Reliability and safety
- As an operator, I can configure upstream timeout behavior.
- As an operator, I can get safe, consistent 5xx responses when upstreams fail.
- As an operator, I can shut the gateway down gracefully.

### Observability
- As an operator, I can see structured request logs.
- As an operator, I can inspect health, readiness, and metrics endpoints.
- As a developer, I can trace a request with a request ID.

### Traffic policy
- As an operator, I can apply per-route rate limits.
- As an operator, I can use Redis-backed rate limiting so limits work across instances.
- As an operator, I can exempt health endpoints from traffic controls.

### Demoability
- As a reviewer, I can run a browser-style or user-style flow that proves routing, failure handling, and rate limiting without reading source first.

## 6. V1 scope
### Must have
- static config file with schema validation
- startup validation with clear errors
- health and readiness endpoints
- config-driven reverse proxy for at least one route
- support for multiple routes
- per-route timeout
- structured JSON logging
- request ID generation / propagation
- in-memory rate limiting
- Redis-backed rate limiting
- Prometheus-style metrics endpoint
- graceful shutdown
- test fixtures for fake upstreams
- reproducible local demo flow

### Nice to have if slices stay small
- header injection to upstreams
- path prefix stripping / rewriting
- admin route to list loaded routes
- Docker Compose for local stack
- hot-reload explicitly deferred unless it becomes trivial

## 7. Out of scope for this version
- dynamic control plane
- distributed config sync
- OAuth/JWT auth system
- service discovery
- circuit breaker mesh
- retries with complex idempotency policy
- full API management product
- graphical dashboard beyond tiny debug/demo pages

## 8. Non-functional requirements
- minimal diffs per slice
- no unnecessary abstractions before the second concrete use case
- clean package boundaries
- deterministic local development
- clear failure messages
- tests should cover both code-level behavior and user-observable behavior
- every slice must be independently reviewable

## 9. Demo scenario
Local demo should include:
- gateway on one port
- one echo upstream
- one flaky or slow upstream
- a browser- or user-style test flow that shows:
  - successful proxying
  - path rewrite behavior
  - request ID propagation
  - rate limiting
  - timeout / upstream failure handling
  - metrics growth after traffic

## 10. Constraints and implementation principles
- Prefer configuration and composition over inheritance or framework magic.
- Prefer standard library capabilities where they are sufficient.
- Keep the request path through the system obvious.
- Avoid generic interfaces until a second real implementation exists.
- Build the smallest thing that can be validated end to end, then extend.

## 11. Release slices
Planned slices are defined in TASKS.md and should proceed in very small, validated increments with explicit stop points for human review.
