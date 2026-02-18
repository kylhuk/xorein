# pr-ready.md

Goal: produce a PR that is mergeable without follow-up.

## Steps
1. In `proj-lead` (gpt-5.3-codex, Reasoning: ExtraHigh), identify the verification baseline for this repo (Make targets if present).
2. Run formatting, lint, and tests (as applicable). Paste exact outputs.
3. Ensure docs/CHANGELOG are updated if behavior or developer workflow changed.
4. Provide PR notes:
   - What changed
   - Why
   - How to test (exact commands)
   - Risk/rollback notes (if relevant)
