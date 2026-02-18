# Phase 1 · Compatibility policy (Planning only)

Defines the additive-only compatibility guardrails that keep v11 wiring stable while v12 planning begins.

## Additive-only constraint set
- v11 and v12 scope advances only via additive fields/messages/enums when touching protobuf or wire schemas; existing numeric tags remain `reserved` before reuse.
- No mandatory-only fields are introduced; optional/nullable primitives or wrapper/message carriers provide safe defaults for older v10 clients.
- Runtime-only behavior is additive too: new relay-state transitions can only coexist with existing transitions unless explicitly gated via feature flags that pass signal-free fallbacks.
- Any proto copy path must preserve canonical ordering so wire payloads produced by v11 still parse under v10/v12 receivers.

## Proto and wire guardrails
- All schema changes pass through `proto/aether.proto`, `buf.yaml`, and the planned delta ledger in `docs/v1.1/phase4/f12-proto-delta.md`.
- Schema diffs carry explicit compatibility comments (`// v11 additive only`) and reference the relevant gate evidence IDs before promotion review.
- v11 compatibility gate readiness is tracked via `pkg/v11/gates` status artifacts plus command evidence in `docs/v1.1/phase5/p5-evidence-index.md`; no separate v11 wire helper package is introduced in this phase.

## Non-breaking rules
- No existing field renames, number reuse, or type changes without an explicit compatibility gate; such cases remain blocked until we archive a carry-forward deprecation note and mark the old field `reserved`.
- Every added RPC or message extends with optional request/carry fields (never removing required headers) and carries a note linking back to this policy and its `EV-v11-G1-###` evidence.
- Cross-version feature discovery (v10 → v11/v12) uses feature flags stored in `proto/aether.proto` with default-off semantics so older runtimes safely ignore new capabilities.

## Validation commands (planned evidence)
| Command | Purpose | Evidence placeholder | Status |
|---|---|---|---|
| `buf lint` | Proto style/lint checklist | `EV-v11-G1-001` | pass (warning: deprecated `DEFAULT` category in `buf.yaml`) |
| `buf breaking` | Breaking-change regression | `EV-v11-G1-002` | pass (`--against '.git#branch=origin/dev'`) |

## Planned vs implemented
- **Planned:** Document the additive-only guardrails before gate runner automation enforces them, ensuring Phase 2+ teams can verify they respect proto/wire constraints.
- **Implemented:** Planning-level guardrails plus recorded lint/breaking evidence are now in place; promotion remains subject to remaining phase5 security and approval gates.
