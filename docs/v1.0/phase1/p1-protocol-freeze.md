# Phase 1 - Protocol Freeze and Compatibility Closure

This file documents the final protocol surface, compatibility matrix, and governance-ready package for the spec release (V10-G1). All entries here follow the additive-only, single-binary compatibility rules mandated earlier.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-P1 | Protocol surface ledger for streams, capabilities, and message families | `pkg/v10/publication/publication.go` section map + `pkg/v10/network/topology.go` topology anchors |
| VA-P2 | Freeze criteria policy | `docs/v1.0/phase1/p1-protocol-freeze.md` text + release manifest `releases/VA-B1-release-manifest.md` freeze checklist |
| VA-P3 | Compatibility matrix for additive delta acceptance | `pkg/v10/governance/policies.go` compatibility classifier |
| VA-P4 | Downgrade negotiation scenarios and evidence | `pkg/v10/network/topology.go` recorded fallback map + `pkg/v10/relay/policy.go` degrade tags |
| VA-P5 | AEP readiness checklist for unresolved breaking proposals | `pkg/v10/governance/policies.go` open-decision markers and `TODO_v10.md` open-decision references |
| VA-P6 | Multi-implementation validation criteria | `pkg/v10/conformance/gates.go` gate catalog referencing scenario catwalk |

## Planned vs Implemented
- **Planned:** Produce a ledger of protocol surfaces and compatibility/downgrade controls, then lock them into release governance before proceeding.
- **Implemented:** `pkg/v10/publication` plus `pkg/v10/network` encode the ledger, while the same modules feed the CLI scenario and the release manifest. Evidence is live in this doc and the release manifest under `releases/VA-B1-release-manifest.md`, which lists the gates and compatibility rows.

## Notes
- The scenario `pkg/v10/scenario` asserts these surfaces by checking the gate catalog and compatibility helper to guarantee deterministic behavior.
