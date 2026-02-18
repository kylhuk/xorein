# proto-safe-change.md

Goal: make safe protobuf changes without breaking wire compatibility.

## Parameters (ask if missing)
- Which .proto files and services/messages are changing?
- Is this explicitly allowed to be a breaking change? (default: NO)

## Steps
1. Switch to `proto-engineer` (gpt-5.3-codex, Reasoning: High).
2. Apply additive-only changes:
   - Add new fields with new numbers.
   - Use `reserved` for removed fields/names.
   - Do not change existing field types or numbers.
3. If the repo uses buf:
   - Run `buf lint`
   - Run `buf breaking` (as configured)
   - Run `buf generate` if generation is configured
   Paste exact outputs.
4. Switch to `proj-lead` (gpt-5.3-codex, Reasoning: ExtraHigh) to review schema + generated changes and run full tests.
