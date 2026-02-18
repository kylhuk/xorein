# Phase 0 - Promotion contract (Planning only)

This artifact defines the version-isolation promotion policy for v11, including state machine, fail-close rule, and the future machine-checkable source of truth hosted under `artifacts/v11/gates/`.

## Promotion states
- `open`: Gate is available for work; evidence stream is being gathered.
- `blocked`: Gate has failed or pending issues; downstream gates must stop until blockers are resolved.
- `promoted`: Gate passed with documented approval and evidence anchors; promotion flow can advance.

Each gate transitions only forward once its evidence ID (EV-v11-GX-###) is recorded and approvals are captured.

## Fail-close rule
- Any gate whose deterministic checklist or command output fails sets the promotion state to `blocked`.
- Blocked gates block downstream gates (no partial promotion) until the responsible owner clears the issue and captures updated evidence IDs and approver timestamps.
- Manual overrides are not permitted; the gate runner under `pkg/v11/gates/` (Phase 1) enforces this rule and refuses to emit `promoted` without upstream clarity.

## Machine-checkable source of truth
- The definitive gate state manifests live under `artifacts/v11/gates/*.status.json` (planned) and is read/written by the gate runner commands described in Phase 1.
- Each status file will include `gateId`, `state`, `updatedAt`, `evidenceId`, `owner`, `approver`, and `notes` fields in JSON, enabling automation to assert fail-close behavior before any release promotes.
- `docs/v1.1/phase5/p5-gate-signoff.md` will cite these files once produced, closing the human-readable/evidence loop.

## Planned vs implemented
- **Planned:** Capture the state definitions, fail-close rule, and SoT reference before gate runner code is authored.
- **Implemented:** Pending. Actual transitions are recorded only when `artifacts/v11/gates/*.status.json` files are produced by Phase 1/5 workflows.
