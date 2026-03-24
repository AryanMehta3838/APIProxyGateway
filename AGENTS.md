# AGENTS.md

## Mission
Build this project incrementally, safely, and cleanly.
Do not make large sweeping changes without first grounding yourself in the existing architecture, constraints, and tests.

Your job is to:
1. understand the current state
2. make the smallest correct change that moves the project forward
3. verify the change
4. summarize exactly what changed and what remains

## Core workflow
For every non-trivial task, follow this order:

1. Read relevant files first
2. Identify the smallest viable implementation slice
3. Implement only that slice
4. Run validation
5. Report:
   - what changed
   - what files changed
   - what tests were run
   - what passed
   - any risks or follow-ups

Do not skip validation unless impossible.

## Project priorities
Prioritize in this order:
1. correctness
2. preserving existing working behavior
3. maintainability
4. clarity
5. speed

## Scope control
Make focused, PR-sized changes.
Do not refactor unrelated parts of the codebase unless required for the task.
Do not introduce new frameworks, dependencies, or patterns unless clearly justified.

When the task is large:
- break it into steps
- complete one step at a time
- verify each step before moving on

## Before coding
Before making changes, quickly determine:
- relevant files
- relevant state/data flow
- likely risks
- how success will be validated

If something is ambiguous, make the most reasonable assumption that preserves stability and note it in your summary.

## Editing rules
Prefer:
- minimal diffs
- existing patterns
- existing naming conventions
- simple solutions over clever ones

Avoid:
- broad rewrites
- hidden side effects
- dead code
- placeholder implementations unless explicitly requested
- fake completions that are not actually wired up

## Architecture discipline
Respect the existing architecture unless the task explicitly includes changing it.

Do not:
- mix concerns across layers
- put business logic in UI code unless the project already does that
- duplicate state unnecessarily
- create parallel systems that solve the same problem

Prefer:
- single source of truth
- predictable data flow
- explicit boundaries between UI, state, domain logic, and persistence

## Testing and validation
After every meaningful change, validate using the strongest available method.

Validation preference order:
1. existing automated tests
2. targeted new automated tests if appropriate
3. local build/typecheck/lint
4. manual verification steps
5. screenshots/log output if helpful

Never claim something works unless it was actually validated.

If you cannot run validation, explicitly say:
- what you could not run
- why
- what should be tested manually

## UI work
For UI changes:
- preserve responsive behavior
- avoid visual regressions
- do not break navigation
- keep styling consistent with the existing app
- prefer accessible patterns
- do not add clutter

When implementing UI:
- wire real state, not fake display-only UI
- ensure empty/loading/error states exist where appropriate
- ensure core interaction flow can actually be completed

## State management
For stateful apps:
- avoid duplicated or conflicting state
- keep derived state derived
- use stable identifiers
- handle resume/reload/re-entry cases if relevant
- think through cancellation, back navigation, and partial progress

## Data and persistence
When persistence is involved:
- preserve existing data compatibility when possible
- avoid destructive migrations unless explicitly requested
- handle missing/old data safely
- validate inputs before saving
- keep models simple and extensible

## Debugging workflow
When fixing a bug:
1. identify the reproduction path
2. find the true root cause
3. implement the smallest reliable fix
4. check likely adjacent regressions
5. summarize root cause clearly

Do not patch symptoms if the underlying cause is identifiable.

## Dependencies
Before adding a dependency, ask:
- is this already solvable with the current stack?
- is the dependency worth the maintenance cost?
- is it aligned with project scope?

Prefer not to add dependencies for small problems.

## File hygiene
Keep files organized.
Do not create new files unless they improve clarity or are needed.
Do not leave commented-out code or temporary debug code behind.

## Output format for each completed task
At the end of each task, provide:
1. Summary
2. Files changed
3. Validation performed
4. Result
5. Risks / next step

## If asked to implement a large feature
Do not try to do everything at once.
Instead:
1. propose the smallest correct first slice
2. implement that slice
3. verify it
4. then continue incrementally

## If working from a product description
Convert the request into:
- requirements
- constraints
- architecture implications
- implementation plan
- test plan

Then begin with the first validated increment.

## Non-goals
Do not:
- overengineer
- optimize prematurely
- invent features not requested
- silently change product behavior
- skip reporting what changed

## Cursor Cloud specific instructions

### Services overview

This is a pure Go project (Go 1.24.0) with two runnable services:

| Service | Command | Default Port |
|---|---|---|
| API Gateway | `go run ./cmd/gateway -config configs/gateway.dev.yaml` | 8080 |
| Echo Upstream (fixture) | `go run ./examples/echo-upstream` | 9091 |

Start the echo upstream **before** the gateway so proxy routes have something to forward to.

### Standard commands

See `README.md` for full details. Quick reference:

- **Build:** `go build ./...`
- **Test:** `go test ./...`
- **Lint:** `go vet ./...`
- **Live demo:** `curl http://127.0.0.1:8080/healthz` (gateway), `curl http://127.0.0.1:8080/api/echo/hello` (proxy)

### Notes

- No external services (Docker, Redis, databases) are required. Redis is disabled in `configs/gateway.dev.yaml` and the in-memory rate limiter works standalone.
- The gateway exits immediately with a clear error if the config file is invalid — useful for testing config validation.
- Tests are fully self-contained (no network or external service dependencies). All tests complete in under 1 second.