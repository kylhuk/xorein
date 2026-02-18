# Protobuf / wire-compat rules

- Additive-only by default:
  - OK: add new fields (new numbers), add new messages, add new enums (new values).
  - Not OK: renumber fields, change field types, change semantics without versioning.
- Never reuse field numbers; mark removed numbers/names as `reserved`.
- Prefer optional fields and backwards-compatible defaults.
- If `buf` is present in the repo, run:
  - `buf lint`
  - `buf breaking` (against the configured baseline)
  - `buf generate` (if generation is configured)
