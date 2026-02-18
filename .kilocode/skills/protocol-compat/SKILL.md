---
name: protocol-compat
description: Checklist for safe wire/protobuf evolution (additive changes, buf checks, downgrade thinking).
---
# Protocol compatibility checklist

Use this skill whenever you touch .proto files, message envelopes, capability negotiation, or anything that affects interoperability.

1. Is the change additive-only?
   - Adding new message/field/enum value: OK.
   - Renumbering/changing types/removing without reserve: NOT OK.

2. Field-number discipline
   - Never reuse a removed field number.
   - Add `reserved <numbers>;` and `reserved "<names>";` entries for removed items.

3. Backwards/forwards behavior
   - Old client receiving new fields: should ignore unknowns and behave safely.
   - New client receiving old messages: must have defaults/fallbacks.

4. Validation
   - If buf is used, run: `buf lint`, `buf breaking`, `buf generate`.
   - Ensure generated code is updated and committed if the repo expects it.

5. Documentation
   - If public contracts changed, update protocol docs and include a migration note.
