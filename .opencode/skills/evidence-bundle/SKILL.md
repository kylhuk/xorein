---
name: evidence-bundle
description: Produces a paste-ready verification bundle (commands + outputs) for reviews/PRs.
---
# Evidence bundle

Use this skill at the end of any task that changes code.

1. List changed files (group by area: code/tests/proto/ci/docs).

2. Provide commands + exact outputs you observed, in this order:
   - Formatter (e.g., gofmt/goimports)
   - Lint (if configured)
   - Tests (unit/integration)
   - Build/run smoke test (if applicable)

3. If you could not run a command:
   - State WHY (tool missing, CI-only, permissions).
   - Reduce confidence in the claim accordingly.
