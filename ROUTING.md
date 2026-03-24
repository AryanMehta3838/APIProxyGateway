# ROUTING.md

# Agent / Tool Routing Policy

## 1. Objective
Use the right tool for the right slice so you spend most of your time on:
- requirements
- acceptance criteria
- design decisions
- review
- prioritization

Agents should handle:
- implementation
- narrow validation
- safe iteration
- regression sweeps

## 2. Default philosophy
- Use ChatGPT to decide what to build.
- Use Codex for precise, minimal-diff backend implementation.
- Use Cursor for IDE-centric multi-file iteration, demos, and agent-assisted review.
- Use cloud agents only after the slice is well-specified.

## 3. When to use ChatGPT only
Use ChatGPT only for:
- PRD, architecture, task slicing, and test-plan work
- tradeoff analysis
- rewriting prompts for agents
- reviewing a proposed diff or PR at a high level
- deciding the next pending task
- debugging strategy before anyone edits code

Do **not** use ChatGPT only when:
- the next step is clearly implementation and acceptance criteria are already defined

## 4. When to use Codex local
Use Codex local for:
- small to medium backend slices
- exact minimal-diff changes
- focused bug fixes
- writing or updating tests close to the changed code
- refactors where precision matters more than IDE convenience
- config parsing, middleware logic, proxy logic, limiter logic

Best fit signals:
- one narrow task
- low ambiguity
- 1 to 8 likely files
- no need for long background execution
- you want deterministic, reviewable edits

Avoid Codex local when:
- the work requires broad repo exploration with frequent manual IDE interaction
- the task is long-running and you do not want your local machine tied up

## 5. When to use Codex cloud
Use Codex cloud for:
- parallel attempts on a harder backend slice
- longer-running research or implementation across isolated worktrees
- broad bug sweeps after the architecture is stable
- preparing alternate approaches for comparison
- running multi-agent work in parallel once tasks are clearly bounded

Best fit signals:
- you already have exact acceptance criteria
- you want multiple branches or variants
- work may run for a while
- you plan to review output rather than co-edit live

Avoid Codex cloud when:
- requirements are still fuzzy
- the slice is tiny and local turnaround is faster

## 6. When to use Cursor local agent
Use Cursor local agent for:
- multi-file IDE-driven changes
- creating fixtures, scripts, examples, and docs
- wiring a feature across packages when you want live steering
- polishing logs, metrics output, and reviewer-facing demo flow
- quickly running local checks while iterating in the editor

Best fit signals:
- you want to watch and steer implementation in the IDE
- the change touches multiple files but is still conceptually simple
- you may manually inspect and tweak during the task

Avoid Cursor local agent when:
- the task needs strict minimal diff and could sprawl
- a more deterministic Codex pass is better

## 7. When to use Cursor Cloud Agent
Use Cursor Cloud Agent for:
- medium or larger pre-planned slices once the design is locked
- Redis wiring, demo-flow hardening, or multi-file tasks that can run in the background
- non-urgent implementation while you review or define the next slice
- end-of-iteration sweeps and broader test execution

Best fit signals:
- the plan is already explicit
- you do not need constant live intervention
- the slice could benefit from longer autonomous execution
- you want cloud execution rather than tying up your machine

Avoid Cursor Cloud Agent when:
- task boundaries are unclear
- you have not yet defined acceptance criteria
- you need very tight diff control

## 8. When to increase reasoning effort
### Low effort
Use when:
- task is mechanical
- existing pattern already exists
- change is localized
Examples:
- add `/readyz`
- adjust config validation message
- add a single test case

### Medium effort
Use when:
- task spans a few files
- behavior must be thought through but scope is still narrow
- you need tests and implementation together
Examples:
- first proxy route
- timeout mapping
- in-memory rate limiting
- metrics middleware

### High effort
Use when:
- ambiguity remains
- failure behavior is important
- task touches architecture or multiple subsystems
- bug sweeps or broad review are needed
Examples:
- Redis-backed limiting behavior
- final regression and review pass
- deciding fail-open vs fail-closed policy
- comparing two implementation approaches

## 9. Tool assignment for this project
### ChatGPT
- all spec files
- acceptance criteria updates
- PR review summaries
- next-slice planning

### Codex local
- T01, T02, T04, T05, T06, T08, T09, T11

### Cursor local agent
- T03, T07, T12, T13

### Cursor Cloud Agent
- T10

### Codex cloud
- optional parallel variant on T10 or T14 if you want competing implementations or a deep review sweep

## 10. Escalation policy
If a slice fails twice:
1. Stop implementation.
2. Return to ChatGPT for a narrower slice or clearer acceptance criteria.
3. Retry with a smaller scope or different tool.

If a diff becomes hard to explain:
1. Split the task.
2. Revert to the last known-good state.
3. Reassign the reduced slice, usually to Codex local.

## 11. Human review policy
Human review is recommended:
- after initial skeleton
- after first successful proxy route
- after logging schema is introduced
- before Redis
- after metrics/debug endpoints
- before final portfolio/demo polish

## 12. Recommended first slice
Start with:
- typed config loader
- startup validation
- `/healthz`
- basic server bootstrap
- one dev config file

Reason:
This gives agents a stable foundation and lets you review project shape before traffic logic begins.
