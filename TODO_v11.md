# TODO v11 - Xorein Execution Plan (v1.1)

## Status
Planning artifact only. This file defines v11 scope, gates, and evidence requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- A version is complete only when all in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- No version promotion is allowed on partial completion.
- Feature cadence is spec-first: spec in vN, implementation in vN+1.
- v11 is the reset version that enforces this contract in tooling and governance.

## Version role
- Implements: `F11` (promotion gates + relay no-data-hosting boundary + naming/runtime baseline).
- Specifies: `F12` (unique identity and local backup recovery, Threema-style behavior).

## Critical scope (v11)
- Enforce version promotion gate runner and checklist automation.
- Enforce relay data-boundary policy: relay cannot be used as durable host for user content.
- Establish Xorein naming baseline (`Xorein Relay`, `Xorein Protocol`, logical servers = `Spaces`).
- Define full v12 identity and local backup spec package.
- Deliver Podman smoke baseline for relay-mode runtime checks.

## Out of scope (defer)
- Public discovery/indexers, advanced moderation, voice/video/screen share runtime features.
- Enterprise features (calendar, recording, SSO, compliance, bots marketplace).

## Dependencies and relationships
- Inputs: `TODO_v10.md`, `aether-v3.md`, `aether-addendum-qol-discovery.md`, `ENCRYPTION_PLUS.md`, `AGENTS.md`, `SPRINT_GUIDELINES.md`.
- Outputs consumed by v12: `docs/v1.1/phase4/f12-identity-backup-spec.md`, `docs/v1.1/phase4/f12-proto-delta.md`, `docs/v1.1/phase4/f12-acceptance-matrix.md`.

## Entry criteria (must be true before implementation starts)
- `v10` is in `promoted` state with evidence bundle and as-built conformance report.
- v11 deferred scope list is frozen and approved.
- v11 gate runner ownership and approvers are assigned.

## Promotion gates (must all pass)
- `G0` Scope lock and dependency map frozen.
- `G1` Compatibility policy validated (`additive-only` minor evolution constraints documented).
- `G2` Gate runner + status commands implemented and tested.
- `G3` Relay no-data-hosting policy checks implemented and tested.
- `G4` Podman relay smoke checks implemented and passing.
- `G5` v12 spec package complete and approved.
- `G6` Docs/evidence bundle complete.
- `G7` `F11` as-built conformance report completed against v11 scope and constraints.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v11/...` (or declared equivalent if paths differ)
- `make check-full`
- `scripts/v11-relay-smoke.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v11-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope and governance freeze (G0)
- [ ] `P0-T1` Freeze v11 in-scope and non-goals.
  - `ST1` Write explicit in-scope list and defer list.
  - `ST2` Add dependency graph and carry-back policy.
  - `ST3` Produce requirement-to-artifact traceability matrix.
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Acceptance: in-scope and deferred items are mutually exclusive and traceable.
  - Artifacts: `docs/v1.1/phase0/p0-scope-lock.md`, `docs/v1.1/phase0/p0-traceability-matrix.md`, `docs/v1.1/phase0/p0-gate-ownership.md`.

- [ ] `P0-T2` Freeze version-isolation promotion rules.
  - `ST1` Define binary promotion states (`open`, `blocked`, `promoted`).
  - `ST2` Define fail-close behavior (failed gate blocks advancement).
  - Acceptance: promotion policy is machine-checkable and human-readable.
  - Artifacts: `docs/v1.1/phase0/p0-promotion-contract.md`.

### Phase 1 - Gate runner implementation (G2)
- [ ] `P1-T1` Add v11 gate runner package.
  - `ST1` Implement gate state model and per-gate status file parser.
  - `ST2` Implement command entrypoint to print deterministic pass/fail summary.
  - `ST3` Add tests for missing artifacts, stale artifacts, and pass-path.
  - Acceptance: deterministic output for pass/fail conditions.
  - Artifacts: `pkg/v11/gates/*`, `pkg/v11/gates/*_test.go`.

- [ ] `P1-T2` Wire gate runner into developer workflow.
  - `ST1` Add make target(s) for v11 gate execution.
  - `ST2` Add artifact directory conventions and status stamps.
  - Acceptance: single command evaluates all v11 gates.
  - Artifacts: `Makefile`, `docs/v1.1/phase1/p1-gate-runner.md`.

### Phase 2 - Relay no-data-hosting boundary (G3)
- [ ] `P2-T1` Define allowed vs forbidden relay data classes.
  - `ST1` Allowed: transient transport/session metadata only.
  - `ST2` Forbidden: durable message bodies, attachment payloads, media frame archives.
  - Acceptance: policy table and examples are explicit.
  - Artifacts: `docs/v1.1/phase2/p2-relay-data-boundary.md`.

- [ ] `P2-T2` Implement policy checks and tests.
  - `ST1` Add config/runtime guardrails preventing forbidden persistence modes.
  - `ST2` Add negative tests proving forbidden modes fail to start.
  - `ST3` Add positive tests for allowed transient metadata paths.
  - Acceptance: tests fail if forbidden persistence is enabled.
  - Artifacts: `pkg/v11/relaypolicy/*`, `pkg/v11/relaypolicy/*_test.go`, `tests/e2e/v11/relay_boundary_test.go`.

### Phase 3 - Podman smoke baseline (G4)
- [ ] `P3-T1` Add deterministic relay smoke scenarios for Podman.
  - `ST1` Start relay container with minimal profile.
  - `ST2` Run health probe and policy probe.
  - `ST3` Capture logs and deterministic result manifest.
  - Acceptance: smoke scenario returns explicit pass/fail exit code.
  - Artifacts: `containers/v1.1/*`, `scripts/v11-relay-smoke.sh`, `docs/v1.1/phase3/p3-podman-smoke.md`.

### Phase 4 - v12 spec package (G5)
- [ ] `P4-T1` Produce unique identity spec.
  - `ST1` Define immutable global identity rules.
  - `ST2` Define "no password reset" behavior and UX warning contract.
  - `ST3` Define identity conflict and migration semantics.
  - Artifacts: `docs/v1.1/phase4/f12-identity-backup-spec.md`.

- [ ] `P4-T2` Produce local backup and recovery spec.
  - `ST1` Define `BackupID + BackupPassword` model and local storage format.
  - `ST2` Define backup encryption profile and corruption handling.
  - `ST3` Define restore-on-new-device journey with deterministic error taxonomy.
  - Artifacts: `docs/v1.1/phase4/f12-backup-recovery-flows.md`.

- [ ] `P4-T3` Produce v12 proto/test/docs acceptance matrix.
  - `ST1` Draft additive proto delta list.
  - `ST2` Define required unit/integration/e2e/perf/security tests for v12.
  - `ST3` Define docs and operator updates required for v12 closure.
  - Artifacts: `docs/v1.1/phase4/f12-proto-delta.md`, `docs/v1.1/phase4/f12-acceptance-matrix.md`.

### Phase 5 - Evidence and closure (G6)
- [ ] `P5-T1` Build v11 evidence bundle.
  - `ST1` Capture gate command outputs and test outputs.
  - `ST2` Capture Podman smoke outputs and deterministic result manifests.
  - `ST3` Capture known risks and signed deferrals.
  - `ST4` Publish `F11` as-built conformance report and promotion recommendation.
  - `ST5` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v1.1/phase5/p5-evidence-bundle.md`, `docs/v1.1/phase5/p5-risk-register.md`, `docs/v1.1/phase5/p5-as-built-conformance.md`, `docs/v1.1/phase5/p5-gate-signoff.md`, `docs/v1.1/phase5/p5-evidence-index.md`.

## Risk register (v11)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R11-1 | Promotion rules bypassed manually | Gate runner required in promotion checklist | Bypass path removed and tested |
| R11-2 | Relay boundary ambiguous | Explicit allowed/forbidden matrix | Matrix referenced by tests |
| R11-3 | Scope creep into runtime features | Deferred list in G0 and fail-close policy | No deferred item appears in code scope |

## Decision log (v11)
- `D11-1`: v11 is a reset-governance release with strict fail-close promotion.
- `D11-2`: Relay remains control-plane and transport only; no durable user-content hosting.
- `D11-3`: v12 will implement local backup only (no hosted backup service).
