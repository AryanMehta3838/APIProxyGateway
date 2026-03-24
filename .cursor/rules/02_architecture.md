# Architecture Rules

- Preserve the existing architecture unless explicitly told to change it.
- Keep concerns separated:
  - UI for rendering and user interaction
  - state for application state and transitions
  - domain logic for business rules
  - persistence for storage and retrieval
- Avoid duplicated sources of truth.
- Keep derived state derived.
- Prefer simple, explicit flows over clever abstractions.
- Do not introduce parallel systems that duplicate behavior.
- Do not add new dependencies unless clearly justified.