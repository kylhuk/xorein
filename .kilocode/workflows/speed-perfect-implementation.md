# speed-perfect-implementation.md

Goal: deliver a minimal, correct change quickly, using fast modes for drafting and `proj-lead` for immediate review.

## Parameters (ask if missing)
- What is the exact objective and acceptance criteria?
- Which area/package is affected?
- What is the default verification command in this repo? (fallback: `go test ./...`)

## Steps
1. In `proj-lead`, restate the objective + acceptance criteria in 1-3 lines. Identify impacted files/packages using search.
2. Choose the fastest drafting mode (set once per mode via Sticky Models in Kilo Code):
   - Production Go code: switch to `go-fast` (recommended model: gpt-5.1-codex-mini, Reasoning: Low).
   - Tests: switch to `go-tests` (gpt-5.1-codex-mini, Reasoning: Low).
   - Protobuf/buf: switch to `proto-engineer` (gpt-5.3-codex, Reasoning: High).
   - CI/scripts: switch to `devops` (gpt-5.1-codex, Reasoning: Medium).
   - Docs: switch to `docs` (gpt-5-codex-mini, Reasoning: Low).
3. Draft the change in the selected mode. Run the fastest relevant local checks (format + targeted tests if possible).
4. Switch back to `proj-lead` (gpt-5.3-codex, Reasoning: ExtraHigh) and perform the review:
   - Re-open every changed file and check correctness, edge cases, and style.
   - Run verification commands (format/lint/tests/build) and paste exact output.
   - Fix any issues directly or delegate to the appropriate mode.
5. If behavior/user workflow changed, update docs (switch to `docs`).
6. Final output:
   - Summary (what changed + why)
   - Changed files
   - Commands run + exact outputs
   - Risks/assumptions (if any)
