# Phase 5 · P5-T2 Conformance Review

## Purpose
Capture compatibility, governance, and open-decision conformance evidence for V6-G5 so reviewers can verify no hidden conflicts exist before release.

## Contract
- `VA-X2` collects the compatibility/governance audit that cross-checks each `P*-T*` contract against the additive-only constraints, major-change trigger matrix, and open decisions in `TODO_v06.md`.
- The review cites `conformance.GateID`s, enumerates runtime resource constraints, and references the reason-class continuity in Section 1.3 so compliance claims explicitly map to deterministic reason codes.
- Open decisions OD6-01..OD6-03 stay marked `Open` with revisit plans; the review shows how `V6-G5` will resurface them and references the release dossier so downstream teams can continue governance oversight.

### Conformance audit checklist
| Audit item | Pass rule | Fail rule | Evidence anchor |
|---|---|---|---|
| Additive-only protobuf scope for V6-G1..G4 | Pass when every `VA-*` entry reuses existing fields or adds new ones only after `buf breaking` approval; fail when renumbering or removal occurs | V6-G5 conformance log | `docs/v0.6/phase0/p0-t2-compatibility-governance-checklist.md` |
| Reason-class continuity across gates | Pass when reason classes appear in both CA tables and release dossier; fail when any VA artifact lacks triad mapping | `docs/v0.6/phase0/p0-t3-verification-evidence-matrix.md` |
| Open decisions OD6-01..OD6-03 | Pass when table below documents why each decision remains `Open` and shows V6-G5 revisit touches; fail if any decision is marked closed without gate evidence | same doc | `docs/v0.6/phase0/p0-t2-compatibility-governance-checklist.md` |

### Open decisions coverage matrix
| Decision | Current state | V6-G5 revisit evidence | Notes |
|---|---|---|---|
| OD6-01 | `Open` | TTL tuning plan and scenario pack anchor in `VA-D1` | Documented in release dossier `VA-X3` Section 9 table |
| OD6-02 | `Open` | PoW envelope review captured via `VA-A1` helper and awaited final policy gate | Anti-abuse helper functions keep bounds explicit for future signoff |
| OD6-03 | `Open` | Reputation weighting plan referenced in `VA-R1`, tied to `VA-X2` audit | Governance audit will confirm whether anti-gaming signals remain configurable |
