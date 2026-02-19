# TODO v26 - Final Full-Stack Closure (“DONE”) for Xorein + harmolyn (v2.6)

## Status
Planning artifact only. This file defines the terminal v2.x closure requirements. It does not claim implementation completion.

## Terminal release note (v2.6 is the “DONE” gate)
- v26 is the **terminal closure** for the v2.x roadmap segment introduced after the v20+ architecture changes.
- v26 has one job: prove that the full system is deployable, testable, and operable with:
  - Xorein backend (relay/bootstrap/client/archivist capabilities)
  - harmolyn UI client (API-attached)
  - supporting services already introduced in prior versions (indexer, push relay, bridge runtime, TURN, etc.)
- v26 does **not** require a “next version seed package”. Any future work becomes a new roadmap track after v2.6.

## Version Isolation Contract (mandatory)
- v26 cannot close unless all v26 in-scope hardening, regression, packaging, docs, and evidence are complete.
- v26 implements only `F26` from v25 specs (proto delta may be empty but must be audited).
- v26 is terminal: no `F27` seed package is required.

## Version role
- Implements: `F26` (system-wide closure, end-to-end conformance, and release evidence for all binaries).
- Specifies: none (terminal).

## Critical scope (v26)
- Full-stack regression closure for all critical user journeys:
  - identity create/restore/multi-device (if present)
  - Space discovery/join + RBAC + moderation
  - chat (send/receive/read markers/reactions/edits/deletes) + persistence + backfill + search coverage
  - voice + screen share setup/recovery (where supported)
  - attachments/assets via blob store (upload/download/verify/caching)
  - bridges (if shipped) and bot/webhook paths (if shipped)
- Operational closure:
  - relay operator runbooks (no durable storage; scaling; upgrades; rollback)
  - archivist/blob provider runbooks (quota/retention; storage alarms; upgrade/rollback)
  - indexer and push relay runbooks (if present)
  - TURN reference deployment validation (if present)
- Packaging/release closure:
  - reproducible builds for `cmd/xorein` and `cmd/harmolyn` (and any shipped auxiliary binaries)
  - SBOM + signing + provenance capture
  - published protocol + local API documentation as the “as-built” truth for v2.6
- Security closure:
  - vulnerability scanning + dependency hygiene
  - crypto and privacy invariants re-validated:
    - no keyword leakage by default
    - relay no-durable-history and no-durable-blob-hosting boundaries
    - private Space anti-enumeration for history and blobs
    - local API cannot be used to bypass permissions
- Final “DONE” evidence bundle:
  - gate checklists + sign-off sheet + evidence index
  - as-built conformance reports for `F23`..`F26` as relevant
  - explicit deferral register: must be empty for **core** features; only “future enhancements” allowed

## Out of scope (defer)
- New product features are out of scope.
- Breaking protocol changes are out of scope.
- “Nice-to-have” parity improvements are deferred to a post-v2.6 roadmap.

## Dependencies and relationships
- Inputs from v25:
  - `docs/v2.5/phase4/f26-final-closure-spec.md`
  - `docs/v2.5/phase4/f26-proto-delta.md`
  - `docs/v2.5/phase4/f26-acceptance-matrix.md`
- v26 produces final release artifacts in:
  - `docs/v2.6/*`
  - `containers/v2.6/*`

## Entry criteria (must be true before implementation starts)
- `v25` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F26` inputs exist and are approved.
- All “carry-forward deferrals” from v23/v24/v25 are either:
  - explicitly cancelled (out of scope forever), or
  - explicitly scheduled into v26 as closure work items, or
  - explicitly declared “post-v2.6 enhancements” (must not be core)

## Promotion gates (must all pass)
- `G0` Scope lock and terminal “DONE” criteria freeze.
- `G1` Compatibility and schema checks pass (`buf breaking`, API version checks).
- `G2` Security hardening tasks complete (no open high/critical issues).
- `G3` Full regression matrix complete (end-to-end across all binaries).
- `G4` Reliability/performance SLO gates pass (startup, send, call setup, backfill, blob transfer).
- `G5` Operator readiness drills pass (relay + archivist/blob + indexer/push relay where shipped).
- `G6` Packaging + reproducible build verification complete (signed artifacts + SBOM).
- `G7` Docs and evidence bundle complete (protocol + local API + operator + user docs).
- `G8` Boundary regression checks pass:
  - relay no-long-history-hosting
  - relay no-durable-blob-hosting
  - private Space anti-enumeration invariants
- `G9` Final as-built conformance report complete against v25 `F26` acceptance package.
- `G10` Final “DONE” sign-off complete (terminal).

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v26/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v26/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- `scripts/v26-release-drills.sh`
- `scripts/v26-repro-build-verify.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v26-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start only once entry criteria (v25 promoted, `F26` inputs approved, carry-forward deferrals classified) are satisfied; Acceptance rule—complete ST1–ST4, freeze terminal deferral policy, and mark `G0` Pass; Evidence rule—attach at least one `EV-v26-G0-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—record planned scope commands (e.g., `make check-full`) before running and capture output.
- `P1-T1`: Dependency rule—after `G0`; Acceptance rule—complete ST1–ST3 security hardening actions and mark `G2` Pass (also revalidate `G8` boundaries per ST1–ST3); Evidence rule—≥1 `EV-v26-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., static analysis/vuln scan outputs) and capture output.
- `P1-T2`: Dependency rule—after `P1-T1`; Acceptance rule—complete ST1–ST3 boundary regression probes and mark `G8` Pass; Evidence rule—attach `EV-v26-G8-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned boundary commands (`tests/e2e/v26/boundaries/*`) and capture outputs.
- `P2-T1`: Dependency rule—after `G2` and `G8`; Acceptance rule—complete ST1–ST6 regression matrix (identity → assets) and mark `G3` Pass; Evidence rule—≥1 `EV-v26-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `tests/e2e/v26/*`) and capture outputs.
- `P2-T2`: Dependency rule—after `P2-T1`; Acceptance rule—complete ST1–ST4 performance/reliability scorecard and mark `G4` Pass; Evidence rule—attach `EV-v26-G4-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (e.g., `tests/perf/v26/*`) and capture outputs.
- `P3-T1`: Dependency rule—after `G3` and `G4`; Acceptance rule—complete ST1–ST2 relay runbook/drill work and mark `G5` Pass; Evidence rule—≥1 `EV-v26-G5-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (relay drill scripts) and capture outputs.
- `P3-T2`: Dependency rule—after `P3-T1`; Acceptance rule—complete ST1–ST2 archivist/blob runbooks and mark `G5` Pass; Evidence rule—attach `EV-v26-G5-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (archivist drill scripts) and capture outputs.
- `P3-T3`: Dependency rule—after `P3-T2`; Acceptance rule—complete ST1–ST2 aux services readiness and mark `G5` Pass; Evidence rule—≥1 `EV-v26-G5-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (aux service validation) and capture outputs.
- `P4-T1`: Dependency rule—after `G5`; Acceptance rule—complete ST1–ST3 reproducible build pipeline and mark `G6` Pass; Evidence rule—attach `EV-v26-G6-###` entry plus outputs (including `scripts/v26-repro-build-verify.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—before running build verification, record the command (e.g., `scripts/v26-repro-build-verify.sh`) and capture output.
- `P4-T2`: Dependency rule—after `P4-T1`; Acceptance rule—complete ST1–ST4 documentation set and mark `G7` Pass; Evidence rule—attach `EV-v26-G7-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record documentation build commands before running and capture outputs.
- `P5-T1`: Dependency rule—after all earlier phases and gates (G0–G7); Acceptance rule—complete ST1–ST4 for final evidence bundle, conformance report, terminal deferral register, sign-off, and mark `G9`/`G10` Pass; Evidence rule—attach `EV-v26-G9-###` and `EV-v26-G10-###` entries plus command outputs (`go build`, `go test`, `make check-full`, `scripts/v26-release-drills.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned commands (`go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`, `scripts/v26-release-drills.sh`) before running and capture outputs.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1` | Entry criteria (v25 promoted + `F26` inputs approved) | `G0` | `EV-v26-G0-###` | Note scope commands (`make check-full`) before running and capture outputs. |
| Phase 1 | `P1-T1`, `P1-T2` | `G0` completion | `G2`, `G8` | `EV-v26-G2-###`, `EV-v26-G8-###` | Capture security/boundary commands (`tests/e2e/v26/boundaries/*`, static analysis) and log outputs. |
| Phase 2 | `P2-T1`, `P2-T2` | `G2`, `G8` completion | `G3`, `G4` | `EV-v26-G3-###`, `EV-v26-G4-###` | Record regression/perf commands (`tests/e2e/v26/*`, `tests/perf/v26/*`) before running and capture outputs. |
| Phase 3 | `P3-T1`, `P3-T2`, `P3-T3` | `G3`, `G4` completion | `G5` | `EV-v26-G5-###` | Plan operator/drill commands (relay/archivist/aux runbooks) and capture outputs. |
| Phase 4 | `P4-T1`, `P4-T2` | `G5` completion | `G6`, `G7` | `EV-v26-G6-###`, `EV-v26-G7-###` | Capture reproducible build and doc commands (`scripts/v26-repro-build-verify.sh`, doc builds) before running and log outputs. |
| Phase 5 | `P5-T1` | Prior gates (G0–G7) complete | `G9`, `G10` | `EV-v26-G9-###`, `EV-v26-G10-###` | Record final evidence commands (`go build`, `go test`, `make check-full`, `scripts/v26-release-drills.sh`) before running and capture outputs. |

## Phase plan

### Phase 0 - Scope lock and terminalization (G0)
- [x] `P0-T1` Freeze v26 scope to terminal closure.
  - `ST1` Import `F26` acceptance matrix and turn each row into a gate check.
  - `ST2` Produce requirement-to-artifact traceability matrix.
  - `ST3` Build the terminal deferral policy:
    - core deferrals must be zero at closure
    - enhancement deferrals (post-v2.6) must be explicitly labeled and separated
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts:
    - `docs/v2.6/phase0/p0-scope-lock.md`
    - `docs/v2.6/phase0/p0-traceability-matrix.md`
    - `docs/v2.6/phase0/p0-terminal-deferral-policy.md`
    - `docs/v2.6/phase0/p0-gate-ownership.md`

### Phase 1 - Hardening and boundary regression (G2, G8)
- [x] `P1-T1` Security hardening closure.
  - `ST1` Run static analysis + vuln scans; fix high/critical.
  - `ST2` Re-validate crypto invariants and key lifecycle edge cases.
  - `ST3` Re-validate local API authz and sensitive RPC audit logs.
  - Artifacts:
    - `docs/v2.6/phase1/p1-security-hardening-log.md`
    - `pkg/v26/security/*`

- [x] `P1-T2` Boundary regression suite. Evidence: `tests/e2e/v26/boundaries/*` + `docs/v2.6/phase1/p1-boundary-report.md`.
  - `ST1` Relay no-long-history-hosting probe (must remain true).
  - `ST2` Relay no-durable-blob-hosting probe (must remain true).
  - `ST3` Private Space anti-enumeration probes for history and blobs.
  - Artifacts:
    - `tests/e2e/v26/boundaries/*`
    - `docs/v2.6/phase1/p1-boundary-report.md`

### Phase 2 - Full regression + SLO gates (G3, G4)
- [x] `P2-T1` Execute full end-to-end regression matrix.
  - `ST1` Identity + backup restore journeys.
  - `ST2` Space join + RBAC + moderation journeys.
  - `ST3` Messaging + persistence + backfill + search journeys.
  - `ST4` Media journeys (voice + screen share) where supported.
  - `ST5` Attachment/asset journeys (blob upload/download + caching + verification).
  - `ST6` Bridge/bot/webhook journeys (if shipped).
  - Artifacts:
    - `tests/e2e/v26/*`
    - `docs/v2.6/phase2/p2-regression-report.md`

- [x] `P2-T2` Execute performance/reliability scorecard (tests/perf/v26/perf_v26_scorecard_test.go + docs/v2.6/phase2/p2-slo-scorecard.md).
  - `ST1` Startup time and reconnect stability thresholds.
  - `ST2` Message send latency p50/p95 (local + with relay).
  - `ST3` Backfill throughput and bounded disk growth under churn.
  - `ST4` Blob transfer throughput and CPU/memory bounds.
  - Artifacts:
    - `tests/perf/v26/*`
    - `docs/v2.6/phase2/p2-slo-scorecard.md`

### Phase 3 - Operator readiness + drills (G5)
- [x] `P3-T1` Relay operator readiness verification.
  - `ST1` Validate deployment configs, healthchecks, alerts.
  - `ST2` Upgrade and rollback drills with evidence capture.
  - Artifacts:
    - `docs/v2.6/phase3/p3-relay-runbook.md`
    - `docs/v2.6/phase3/p3-relay-rollback-drill.md`

- [x] `P3-T2` Archivist/blob provider operator readiness verification.
  - `ST1` Storage growth alarms and quota tuning.
  - `ST2` Upgrade and rollback drills; corruption recovery guidance.
  - Artifacts:
    - `docs/v2.6/phase3/p3-archivist-runbook.md`
    - `docs/v2.6/phase3/p3-archivist-rollback-drill.md`

- [x] `P3-T3` Indexer/push relay/TURN readiness verification (if shipped).
  - `ST1` Runbooks and health endpoints validated.
  - `ST2` Failure drills (indexer down, push relay down, TURN down) validated with deterministic client UX.
  - Artifacts:
    - `docs/v2.6/phase3/p3-aux-services-runbook.md`
  - Status note: Planning-only runbook captures shipped vs not-shipped paths, deterministic reason taxonomy, and EV mappings before any auxiliary drill executes.

### Phase 4 - Packaging + docs + reproducibility (G6, G7)
- [x] `P4-T1` Reproducible build pipeline for all shipped binaries.
  - Status note: Implemented deterministic rebuild + hash capture path with optional baseline comparison via `scripts/v26-repro-build-verify.sh`; SBOM/signing inputs are defined in companion phase4 artifacts.
  - `ST1` Deterministic build instructions and tooling captured.
  - `ST2` SBOM + signatures + provenance generated.
  - `ST3` Verify reproducibility in CI (bit-for-bit or declared acceptable deltas).
  - Artifacts:
    - `docs/v2.6/phase4/p4-repro-build.md`
    - `docs/v2.6/phase4/p4-sbom.md`
    - `docs/v2.6/phase4/p4-signing.md`
    - `scripts/v26-repro-build-verify.sh`

- [x] `P4-T2` Publish final documentation set (as-built truth).
  - `ST1` Protocol documentation: network proto + history plane + blob plane.
  - `ST2` Local API documentation (daemon + attach) including version negotiation and threat model.
  - `ST3` Operator docs: relay, archivist/blob, indexer, push relay, TURN.
  - `ST4` User docs: onboarding, backup, privacy model, offline/history/backfill behavior.
  - Artifacts:
    - `docs/v2.6/phase4/p4-protocol-docs.md`
    - `docs/v2.6/phase4/p4-local-api-docs.md`
    - `docs/v2.6/phase4/p4-operator-docs.md`
    - `docs/v2.6/phase4/p4-user-docs.md`

### Phase 5 - Final evidence + terminal sign-off (G9, G10)
- [x] `P5-T1` Publish final evidence bundle and "DONE" decision record.
  - `ST1` Attach command outputs and drill manifests.
  - `ST2` Publish final as-built conformance report against `F26`.
  - `ST3` Publish terminal deferral register (must contain **no core deferrals**).
  - `ST4` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.6/phase5/p5-final-evidence-bundle.md`
    - `docs/v2.6/phase5/p5-as-built-conformance.md`
    - `docs/v2.6/phase5/p5-terminal-deferrals.md`
    - `docs/v2.6/phase5/p5-gate-signoff.md`
    - `docs/v2.6/phase5/p5-evidence-index.md`

## Risk register (v26)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R26-1 | Hidden cross-version deferrals remain | terminal deferral policy + architecture coverage audits | core deferrals = 0 |
| R26-2 | Packaging/reproducibility fails across platforms | CI reproducible build verification | reproducibility checks pass |
| R26-3 | Operator readiness gaps | mandatory drills + runbooks | drills complete with evidence |
| R26-4 | Boundary regressions | dedicated boundary probes | boundary probes pass |
| R26-5 | Full regression too shallow | explicit matrix + scorecards | all matrix rows pass |

## Decision log (v26)
- `D26-1`: v26 is terminal for v2.x; no further seed package is required.
- `D26-2`: A “DONE” release requires zero open deferrals for core features; enhancements are explicitly separated.
- `D26-3`: Release evidence must be sufficient for an external reviewer to reproduce the build and validate conformance.
