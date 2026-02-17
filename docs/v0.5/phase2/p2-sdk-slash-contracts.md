# Phase 2 · P2 SDK & Slash Contracts

## Purpose
Describe the deterministic expectations for the Go SDK, community SDK conformance profiles, slash command registration, and autocomplete behavior called out in Phase 2 of `TODO_v05.md`.

## Deterministic obligations
- The Go SDK surface (initialization, event subscription, command invocation, lifecycle hooks) must align with the Native Bot API contracts from Phase 1 and treat equivalent bot workflows identically, per P2-T1 and VA-S1.
- Python, JavaScript, and Rust community SDKs must satisfy the minimum-feature matrix, support-tier labels, and version-synchronization guidance in P2-T2 so that integrations can be audited against VA-S2 without assuming Go parity.
- Slash command schema validation, catalog registration, update/removal propagation, and concurrency behavior are fully specified in P2-T3; invalid or conflicting schemas map to deterministic rejection outcomes in VA-S3.
- Slash autocomplete query envelopes, ranking budget, timeout/failure fallbacks, and empty-result behavior follow P2-T4 so that clients see repeatable suggestion ordering and failure signals documented in VA-S4.

## Deterministic contract tables
| Input / condition | Outcome + reason codes | Artifact | Validation obligation |
|---|---|---|---|
| Go SDK initialization plus event/command surfaces | Positive: `sdk.go.success` for expected events/commands; Negative: `sdk.go.invalid-surface` when capability advertisement differs; Recovery: `sdk.go.retry-compat` when version negotiation is needed. | VA-S1 | Document interface expectations and compatibility test cases so scenario pack reviewers can verify the Go client stays aligned with VA-B artifacts via `VA-X1`. |
| Community SDK conformance profile | Positive: `sdk.community.success` when Python/JS/Rust clients satisfy the minimum matrix; Negative: `sdk.community.incomplete` when required hooks are missing; Recovery: `sdk.community.degraded-retry` with explicit downgrade guidance. | VA-S2 | Map each language profile to mandatory capability checks and include reason-coded failure/resume states in the conformance review logs. |
| Slash command schema registration | Positive: `slash.schema.success`; Negative: `slash.schema.reject` for invalid option/permission combinations; Recovery: `slash.schema.reconcile` after conflicting updates. | VA-S3 | Record schema validation counterexamples and reconcile logs so coverage reviewers can audit conflict-handling reason codes. |
| Autocomplete query and fallback behavior | Positive: `slash.auto.success` for ranked suggestions; Negative: `slash.auto.timeout` under latency or validation failure; Recovery: `slash.auto.fallback` to stale data or empty states. | VA-S4 | Capture timeout/fallback flows in stability scenarios and link to integrated validation tests per `VA-X1`. |

These table entries update the verification matrix (`docs/v0.5/phase0/p0-t3-verification-evidence-matrix.md`) with deterministic input/outcome pairs tied to the phase deliverables.
