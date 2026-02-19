# v2.4 Phase 0 Scope Lock

Phase 0 is planning-only. This document records the frozen in/out scope for v24, notes the invariants we must preserve, and traces the handoff from v23.

## In Scope
- Explicit two-process split: `harmolyn` + `xorein` (daemon), with documentation that reflects the separation.
- Local-only API surface (UDS/named pipe) plus version negotiation policy.
- Daemon lifecycle tooling (doctor, version reporting, headless mode hooks) + documented UX for attach/detach success/failure states.
- Local API security posture (authn/authz invariants, audit logging, no UI dependencies) with evidence placeholders for `EV-v24-G0-001` and successors.
- Preservation of protocol invariants inherited from v23: relay nodes continue to refuse durable history hosting, and keyword leakage remains opt-out by default.

## Out of Scope
- Remote network exposure of the local API (see TODO_v24 section “Out of scope”).
- Public third-party control planes or plugin sandboxes unless already scoped elsewhere.

## Invariants & References
- Relay nodes must keep the no-durable-history-hosting invariant from v23 handoff (`docs/v2.3/phase4/f24-backlog-and-spec-seeds.md`).
- Keyword leakage default remains disabled; any telemetry or logging changes must be additive and opt-in per the v23 privacy plan.
- Evidence for this scope lock is tied to `EV-v24-G0-###` entries in `docs/v2.4/phase5/p5-evidence-index.md` once Phase 0 closes.
- Scope dependencies: acceptance matrix (`docs/v2.3/phase4/f24-acceptance-matrix.md`) + deferral register (`docs/v2.3/phase4/f24-deferral-register.md`).

## References & Next Steps
- The v23 handoff notes listed under “Dependencies and relationships” in `TODO_v24.md` anchor this scope lock.
- Phase 0 closes when traceability, gate ownership, and API evolution policy are documented (G0) and the local API spec path begins (G1).
