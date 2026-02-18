# TODO v20 - Xorein Execution Plan (v2.0)

## Status
Planning artifact only. This file defines v20 implementation and release requirements. It does not claim implementation completion.

## Version Isolation Contract (mandatory)
- v20 cannot close unless all v20 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v20 implements only `F20` from v19 specs.
- Non-critical parity features are explicitly deferred to `v20+` backlog.

## Version role
- Implements: `F20` (production/public-beta hardening and release conformance).
- Specifies: required `F21` seed package for post-v20 roadmap continuity.

## Critical scope (v20)
- Complete production/public-beta hardening for relay, protocol runtime, and Gio client.
- Complete full regression and security validation for critical user journeys.
- Complete Podman operator readiness (build/sign/health/rollback/runbooks).
- Complete release governance, evidence, and go/no-go decision package.
- Preserve relay no-data-hosting boundary with explicit regression checks.
- Produce required `F21` spec/proto/acceptance package for `v20+`.

## Out of scope (defer)
- Non-critical parity features (enterprise and ecosystem expansions) remain deferred to `v20+`.

## Dependencies and relationships
- Inputs from v19: `docs/v1.9/phase4/f20-release-hardening-spec.md`, `docs/v1.9/phase4/f20-acceptance-matrix.md`.
- Outputs for `v20+`: `docs/v2.0/phase4/f21-backlog-and-spec-seeds.md`, `docs/v2.0/phase4/f21-proto-delta.md`, `docs/v2.0/phase4/f21-acceptance-matrix.md`, `docs/v2.0/phase4/f21-deferral-register.md`.

## Entry criteria (must be true before implementation starts)
- `v19` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F20` spec inputs from v19 exist and are approved.
- v20 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and release criteria freeze.
- `G1` Compatibility and schema checks pass.
- `G2` Security hardening tasks complete.
- `G3` Full regression matrix complete.
- `G4` Reliability/performance SLO gates pass.
- `G5` Podman operations and rollback drills pass.
- `G6` Release docs and evidence bundle complete.
- `G7` Go/no-go sign-off complete.
- `G8` Relay no-data-hosting regression checks pass.
- `G9` `F20` as-built conformance report completed against v19 specification package.
- `G10` Required `F21` seed package and `v20+` deferral register complete.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v20/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v20/...` (or declared equivalent if paths differ)
- `make check-full`
- Commands declared in `docs/v2.0/phase3/p3-podman-scenarios.md`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v20-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## Phase plan

### Phase 0 - Scope lock and release matrix (G0)
- [x] `P0-T1` Freeze v20 release scope and pass/fail criteria.
  - `ST1` Confirm critical-only scope; move all non-critical features to `v20+` list.
  - `ST2` Freeze go/no-go authority and sign-off process.
  - `ST3` Produce requirement-to-artifact traceability matrix.
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts: `docs/v2.0/phase0/p0-scope-lock.md`, `docs/v2.0/phase0/p0-go-no-go-policy.md`, `docs/v2.0/phase0/p0-traceability-matrix.md`, `docs/v2.0/phase0/p0-gate-ownership.md`.

### Phase 1 - Security and runtime hardening (G2)
- [x] `P1-T1` Complete security hardening backlog.
  - `ST1` Fix high/critical findings from static/vuln scans.
  - `ST2` Validate crypto and identity recovery edge cases.
  - `ST3` Validate moderation and permission tamper resistance.
  - Artifacts: `docs/v2.0/phase1/p1-security-hardening-log.md`, `pkg/v20/security/*`.

- [x] `P1-T2` Complete runtime hardening backlog.
  - `ST1` Stabilize startup, reconnect, and continuity paths.
  - `ST2` Stabilize voice and screen share under stress.
  - Artifacts: `pkg/v20/hardening/*`, `pkg/v20/hardening/*_test.go`.

### Phase 2 - Full validation and SLO conformance (G3, G4)
- [x] `P2-T1` Execute full regression matrix.
  - `ST1` Unit, integration, and e2e suites for all critical features.
  - `ST2` Adverse/recovery tests for network, policy, and media failures.
  - `ST3` Compatibility/downgrade tests for additive evolution guarantees.
  - Artifacts: `tests/e2e/v20/*`, `tests/perf/v20/*`, `docs/v2.0/phase2/p2-regression-report.md`.

- [x] `P2-T2` Execute reliability/performance conformance checks.
  - `ST1` Validate login, send, call setup, and recovery SLO thresholds.
  - `ST2` Validate crash-free and stability thresholds.
  - Artifacts: `docs/v2.0/phase2/p2-slo-scorecard.md`.

### Phase 3 - Podman operations readiness (G5)
- [x] `P3-T1` Complete Podman build/release pipeline validation.
  - `ST1` Deterministic build and digest capture.
  - `ST2` Signature/SBOM/provenance checks.
  - Artifacts: `containers/v2.0/*`, `docs/v2.0/phase3/p3-build-attestation.md`.

- [x] `P3-T2` Complete operator runbook and rollback drills.
  - `ST1` Healthcheck, alerting, and incident response drills.
  - `ST2` Rollback and recovery drill with evidence capture.
  - `ST3` Relay no-data-hosting regression scenario under release operations drills.
  - Artifacts: `docs/v2.0/phase3/p3-operator-runbook.md`, `docs/v2.0/phase3/p3-rollback-drill.md`, `docs/v2.0/phase3/p3-podman-scenarios.md`.

### Phase 4 - Release package and required v20+ seed package (G6, G10)
- [x] `P4-T1` Publish release and support documentation.
  - `ST1` Protocol/runtime/operator/user docs finalized.
  - `ST2` Public beta notes and known limits documented.
  - Artifacts: `docs/v2.0/phase4/p4-release-docs.md`, `docs/v2.0/phase4/p4-release-notes.md`.

- [x] `P4-T2` Publish required `F21` package and non-critical `v20+` deferrals.
  - `ST1` Publish `F21` seed specification and additive proto delta.
  - `ST2` Publish `F21` acceptance matrix to preserve N-1 cadence continuity.
  - `ST3` Publish deferral register with rationale, owner, and revisit target for each non-critical feature.
  - Artifacts: `docs/v2.0/phase4/f21-backlog-and-spec-seeds.md`, `docs/v2.0/phase4/f21-proto-delta.md`, `docs/v2.0/phase4/f21-acceptance-matrix.md`, `docs/v2.0/phase4/f21-deferral-register.md`.

### Phase 5 - Final evidence and go/no-go (G7)
- [x] `P5-T1` Publish final evidence bundle and release decision.
  - `ST1` Attach all gate outputs and residual risk sign-offs.
  - `ST2` Publish `F20` as-built conformance report against v19 `F20` specs.
  - `ST3` Record release decision: public beta or production.
  - `ST4` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts: `docs/v2.0/phase5/p5-final-evidence-bundle.md`, `docs/v2.0/phase5/p5-go-no-go-record.md`, `docs/v2.0/phase5/p5-as-built-conformance.md`, `docs/v2.0/phase5/p5-gate-signoff.md`, `docs/v2.0/phase5/p5-evidence-index.md`.

## Risk register (v20)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R20-1 | Late security defects | Dedicated hardening phase with blocked promotion | No open high/critical issues |
| R20-2 | Ops readiness gaps | Mandatory Podman drills and runbook validation | Drill evidence complete |
| R20-3 | Scope creep near launch | Locked critical-only scope and deferral register | No unapproved in-scope additions |
| R20-4 | Relay boundary regresses during final hardening | Dedicated regression checks in release drills and e2e suites | Relay-boundary regression tests pass |

## Decision log (v20)
- `D20-1`: Non-critical parity features are explicitly deferred to `v20+`.
- `D20-2`: Release label (public beta vs production) is decided strictly by gate evidence.
- `D20-3`: `F21` seed package is required before closing v20.
- `D20-4`: Final hardening must not broaden relay durable storage permissions.
