# Phase 0 scope lock (P0-T1)

This planning document freezes v23 hardening invariants, relay boundary defaults, and the gate pass/fail expectations that feed `G0`. It does **not** claim execution or implementation; it remains a planning artifact until all entry criteria and STs are satisfied.

## Hardening invariants (v23 history/search plane)
- **No keyword leakage by default.** Search/backfill requests and logging retain only coverage labels; any assisted or keyword-bearing access requires an explicit, opt-in contract in `F24` seed artifacts.
- **Private Space anti-enumeration stays mandatory.** Metadata surfaced to owners, relays, or Archivists can never be used to enumerate private Spaces or their memberships without explicit consent workflows.
- **Deterministic quotas and refusal reasons.** Archivist/relay quota enforcement always returns documented refusal codes and telemetry markers so failures stay auditable.
- **Durability state labeling.** Replica-set shortfalls (partial sync, degraded availability) show up as a first-class durability state in APIs, UIs, and alerts so operators know history is still soft-limited.
- **Relay boundary freeze.** Relays continue to host no long-lived history segments/manifests or mined indexes; all durable history resides in Archivist/Archivist-like storage.

## Gate expectations
- `G0` passes when the scope lock artifacts and traceability commitments listed below are published, approved by the named approvers, and referenced by evidence entries `EV-v23-G0-###`.
- Any deviations or unresolved questions feed into a `BLOCKED:GATE` note until the scope lock is re-approved or documented as a planned `F24` seed.

## Scope-to-artifact commitments
| Requirement | Artifact(s) | Notes |
| --- | --- | --- |
| Hardening invariants (ST1–ST4) | `docs/v2.3/phase0/p0-hardening-matrix.md` | Enumerates G0/G11 criteria and default guarantees. |
| Traceability of invariants to tasks | `docs/v2.3/phase0/p0-traceability-matrix.md` | Maps STs to downstream artifacts so nothing is unaddressed. |
| Gate ownership assignments | `docs/v2.3/phase0/p0-gate-ownership.md` | Lists owners/approvers for G0–G11. |
| Architecture coverage readiness | `docs/v2.3/phase0/p0-architecture-coverage-audit.md` | Ensures every persistence plane is inventoried; missing items become `F24` seeds. |

## Evidence expectations
- Record at least one `EV-v23-G0-###` entry per hardening invariant (quota default, privacy gate, relay boundary, durability labeling) in the evidence index once the artifacts above exist.
- All `EV-v23-G0-###` entries should include the commands or review notes (e.g., design review, doc approvals) that demonstrate the scope lock artifact was produced.

## Observability of the relay boundary
- The assumption that relays host no long history must be re-affirmed in the scope lock review and referenced from the hardening matrix.
- Any potential exception (e.g., new Archivist-like cache on the relay) must be documented as a flagged `F24` seed; otherwise `G0` cannot pass.

This scope lock remains open until the Phase 0 STs are complete and `G0` evidence entries are recorded. The architecture coverage audit in `G11` is a dependent deliverable.
