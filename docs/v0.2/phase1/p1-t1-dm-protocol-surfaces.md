# v0.2 Phase 1 - P1-T1 DM Protocol Surfaces

> Status: Planning artifact only. No implementation completion is claimed.

## Purpose

Capture the v0.2 planning contract for DM protocol identifiers, schema surfaces, and capability negotiation behavior.

## Task Coverage

| Task | Coverage summary | Evidence anchor |
|---|---|---|
| P1-T1-ST1 | DM protocol identifiers are namespaced, versioned, and include downgrade behavior requirements. | `EVID-V2-P1-T1-ST1` |
| P1-T1-ST2 | DM handshake and envelope schema extensions remain additive-safe and compatibility-governed. | `EVID-V2-P1-T1-ST2` |
| P1-T1-ST3 | `SecurityMode` and `mode_epoch_id` negotiation is explicit with deterministic unsupported-mode outcomes and no silent downgrade. | `EVID-V2-P1-T1-ST3` |

## Planning Constraints

- Minor-version evolution remains additive-only.
- Any incompatible behavior follows major-path governance (new multistream ID + downgrade procedure + AEP path).
- Security-mode labeling remains explicit on DM and Group DM surfaces.
