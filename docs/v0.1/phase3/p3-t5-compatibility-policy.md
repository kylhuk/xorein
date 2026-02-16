# P3-T5 — v0.x Compatibility and Deprecation Policy

## Policy statement (normative for v0.x)

1. **Minor versions are additive-only.**
   - Allowed: adding new fields/messages/enums.
   - Not allowed: field renumbering, wire-type mutation, semantic repurposing.
2. **No field-number reuse.**
   - Removed fields/names MUST be marked `reserved`.
3. **No required fields in v0.x minors.**
   - Absent fields must be safely defaultable.
4. **Unknown-field tolerance is mandatory.**
   - Newer senders must not break older receivers.
5. **Breaking behavior requires new major protocol IDs.**
   - Major transitions MUST include documented downgrade behavior.

This policy aligns with [`docs/v0.1/phase1/protocol-constraints.md`](docs/v0.1/phase1/protocol-constraints.md)
and is implemented in negotiation/deprecation primitives under
[`pkg/protocol/registry.go`](pkg/protocol/registry.go).

## Prohibited schema/protocol changes (v0.x)

- Reusing field numbers or removed names.
- Changing existing field wire types.
- Converting optional/defaultable fields into required semantics.
- Reinterpreting existing enum values with incompatible meaning.
- In-place breaking changes to protocol selection without a new major ID.

## Protocol major-bump trigger conditions

A major bump is required when any of the following are true:

1. Existing major-minor peers cannot interoperate safely after change.
2. Required capability semantics change in a way older peers cannot emulate.
3. Envelope verification semantics change incompatibly for existing payloads.
4. Downgrade path cannot be deterministic and testable.

## Deprecation process notes (v0.x)

1. Declare per-family deprecation anchors.
2. Keep deprecated IDs parseable but non-preferred.
3. Enforce skip behavior during negotiation.
4. Provide explicit upgrade-required user feedback when no compatible required
   set exists.

Deprecation skip mechanics are implemented by
[`DeprecationGuard.IsDeprecated()`](pkg/protocol/registry.go:161) and consumed by
[`NegotiateProtocol()`](pkg/protocol/registry.go:177).

## Code review checklist linkage

This policy is now a required reference for protocol-touching reviews alongside
the Phase 1 review template:

- [`docs/v0.1/phase1/review-template.md`](docs/v0.1/phase1/review-template.md)
- [`docs/v0.1/phase1/protocol-constraints.md`](docs/v0.1/phase1/protocol-constraints.md)

