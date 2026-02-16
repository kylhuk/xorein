# v0.2 Phase 0 - P0-T2 Compatibility and Governance Checklist

> Status: Execution artifact. Compatibility/governance readiness recorded in downstream docs and contracts.

## Purpose

Define mandatory checks for protocol and protobuf changes in v0.2 so additive evolution and governance constraints are enforced consistently.

## Source Trace

- `TODO_v02.md:189`
- `AGENTS.md`
- `.opencode/rules/proto-wire-compat.md`
- `aether-v3.md:457`

## Protobuf Evolution Checklist (P0-T2-ST1)

Use for every v0.2 schema delta:

- [ ] Change is additive-only for minor version path.
- [ ] No existing field numbers are renumbered, reused, or repurposed.
- [ ] Removed fields, if any, are marked as `reserved` (names and numbers).
- [ ] Wire types and field cardinality are unchanged for existing fields.
- [ ] Defaults/optional behavior remain backward compatible for older clients.
- [ ] New fields include deterministic handling when absent.
- [ ] Compatibility rationale is captured in task evidence.

Prohibited in v0.2 minor evolution:

- Renumbering existing fields.
- Reusing removed field numbers.
- Changing field wire types.
- Silent semantics change of existing fields without version split.

## Major-Change Trigger and Downgrade Checklist (P0-T2-ST2)

Treat a proposal as major-path if any condition is true:

1. Existing peers cannot safely parse/interpret updated behavior.
2. Existing semantic contract changes in a way that could alter safety/security outcomes.
3. Backward-compatible downgrade cannot preserve deterministic outcomes.

Required artifacts when major-path trigger is hit:

- [ ] New multistream protocol ID plan.
- [ ] Explicit downgrade negotiation behavior and failure reason taxonomy.
- [ ] AEP trigger record with review period reference.
- [ ] Evidence plan for two independent implementations.
- [ ] Migration/compatibility timeline and deprecation notes.

## Governance Review Checklist

- [ ] Protocol-first constraints are preserved over UI convenience.
- [ ] No unresolved open decision is documented as finalized architecture.
- [ ] Scope remains within v0.2 band; v0.3+ features are deferred explicitly.
- [ ] Risk log updated if tradeoff impacts performance/security/reliability priorities.

## Signoff Template

| Item | Status (Pass/Fail) | Evidence link | Reviewer | Date |
|---|---|---|---|---|
| Protobuf additive-only checks | Pass | `.opencode/rules/proto-wire-compat.md`, `buf.yaml`, `proto/aether.proto` (unchanged schema) | Compatibility automation | 2026-02-16 |
| Major-change trigger review | Pass | `.opencode/rules/proto-wire-compat.md`, `docs/v0.2/phase7/p7-t2-conformance-review.md` (governance log) | Governance review draft | 2026-02-16 |
| Downgrade/AEP requirement review | Pass (template ready; no major-path trigger) | `.opencode/rules/proto-wire-compat.md` (AEP trigger template) | Governance discipline | 2026-02-16 |
| Open-decision wording discipline | Pass | `docs/v0.2/phase7/p7-t2-conformance-review.md`, `docs/v0.2/phase0/p0-t1-scope-contract.md` | Documentation review | 2026-02-16 |

> Note: Reviewer/date columns capture this closure summary and will be replaced with formal signoff records if/when a release gate ceremony records the named participants.
