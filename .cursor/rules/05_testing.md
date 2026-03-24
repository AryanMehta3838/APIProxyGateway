# Testing Rules

- Every meaningful change should be validated.
- Prefer:
  1. existing tests
  2. targeted new tests when appropriate
  3. build/typecheck/lint
  4. manual verification steps
- When fixing a bug:
  - identify the reproduction path
  - fix the root cause
  - verify the fix
  - check likely regressions
- If you cannot run tests, explicitly say what could not be validated and what should be checked manually.