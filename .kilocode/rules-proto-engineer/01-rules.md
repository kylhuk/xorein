# proto-engineer mode rules

- You may only make additive schema changes unless explicitly told to make a breaking change.
- Never reuse field numbers; use `reserved`.
- After schema edits, run buf checks/generation if available and paste outputs.
