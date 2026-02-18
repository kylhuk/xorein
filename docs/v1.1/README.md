# v1.1 artifacts index

This directory collects planning and implementation artifacts for v11 (v1.1). Phase 0 establishes governance baselines, while Phase 1-3 now include executable gate, policy, and smoke-check artifacts.

## Phase status
- **Phase 0 - Scope and governance (G0):** frozen in-scope/deferred lists, traceability matrix, gate ownership, and promotion contract. See `phase0/`.
- **Phase 1 - Gate runner (G2):** Gate runner implementation guidance lives in `phase1/p1-gate-runner.md`, the compatibility policy lives in `phase1/p1-compatibility-policy.md`, and both align with `pkg/v11/gates` + `cmd/v11-gate-runner`.
- **Phase 2 - Relay data boundary (G3):** Allowed/forbidden data class rationale and policy/test references are captured in `phase2/p2-relay-data-boundary.md`.
- **Phase 3 - Podman smoke (G4):** Podman relay smoke purpose, command usage, and probe expectations are now documented in `phase3/p3-podman-smoke.md`.
- **Phase 4 - v12 spec (G5):** Planning artifacts now include identity/backup spec, backup/recovery flows, the proto delta list, and the acceptance matrix in `phase4/`.
- **Phase 5:** Evidence bundle, risk register, as-built conformance notes, gate signoff checklist, and evidence index live under `phase5/` (see the templates in `docs/templates/` for expected structure).

## Templates and evidence conventions
- Gate RACI and sign-off: `docs/templates/roadmap-signoff-raci.md` (copy into your gate files).
- Evidence index IDs: use the `EV-v11-GX-###` format described in `docs/templates/roadmap-evidence-index.md` and reference them from each gate/phase doc.
- Traceability/gate checklist patterns: `docs/templates/roadmap-gate-checklist.md`.

## Working note
- All planning-artifact statements use the `Planned vs implemented` framing required by sprint governance; treat the files in `phase0/` as checkpoints rather than completed implementation evidence.
