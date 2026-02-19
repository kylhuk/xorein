# TODO v23 - Xorein (backend) + harmolyn (frontend) Execution Plan (v2.3)

## Status
Planning artifact only. This file defines v23 implementation and release-hardening requirements for the history/search plane. It does not claim implementation completion.

## Naming + binary split note (carry-forward)
- **Xorein**: backend node/runtime binary (`cmd/xorein`) implementing history/search plane behavior and Archivist ops surfaces.
- **harmolyn**: frontend UI binary (`cmd/harmolyn`) exposing history availability/search coverage/backfill UX.
- v23 MUST keep “backend runtime has no UI deps” as an audited invariant.

## Version Isolation Contract (mandatory)
- v23 cannot close unless all v23 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v23 implements only `F23` from v22 specs.
- v23 must also publish required `F24` seed package for post-v23 roadmap continuity.

## Version role
- Implements: `F23` (history/search hardening + Archivist operator readiness + privacy/abuse conformance gates).
- Specifies: required `F24` seed package for `v23+` backlog continuity.

## Critical scope (v23)
- Complete security hardening for history/search:
  - abuse resistance (quota defaults, rate limits, refusal reasons)
  - privacy conformance (no keyword leakage by default; explicit opt-ins)
  - integrity hardening (manifest/head verification robustness)
  - replication/durability hardening (defaults + degraded-mode labeling)
- Complete reliability/performance hardening for backfill/search:
  - bounded resource usage (CPU, IO, disk growth)
  - SLO targets and regression gates
- Complete Archivist operator readiness:
  - runbooks, monitoring, storage-growth alarms
  - upgrade/rollback drills and evidence
- Complete release governance and go/no-go package specifically for history/search plane:
  - evidence bundle and conformance report
  - explicit known limits and deferrals
- Preserve relay no-long-history-hosting boundary with explicit regression checks.
- Publish required `F24` seed package and deferral register for post-v23 planning.
- **Architecture coverage audit (required):** produce an explicit map of *every persisted data class* (messages, attachments, profiles, emojis, server config, indices) → storage plane(s) → retrieval path(s) → privacy properties, and ensure nothing critical is “magic”.

## Out of scope (defer)
- New product features beyond hardening/operator readiness are deferred to `v24+`.
- Any breaking protocol change is out of scope.

## Dependencies and relationships
- Inputs from v22:
  - `docs/v2.2/phase4/f23-history-hardening-spec.md`
  - `docs/v2.2/phase4/f23-proto-delta.md`
  - `docs/v2.2/phase4/f23-acceptance-matrix.md`
- Outputs for `v23+`:
  - `docs/v2.3/phase4/f24-backlog-and-spec-seeds.md`
  - `docs/v2.3/phase4/f24-proto-delta.md`
  - `docs/v2.3/phase4/f24-acceptance-matrix.md`
  - `docs/v2.3/phase4/f24-deferral-register.md`

## Entry criteria (must be true before implementation starts)
- `v22` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F23` spec inputs from v22 exist and are approved.
- v23 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and hardening criteria freeze.
- `G1` Compatibility and schema checks pass.
- `G2` Security hardening tasks complete.
- `G3` Reliability/performance SLO gates pass.
- `G4` Archivist operator readiness drills pass.
- `G5` Full regression matrix complete (history/search specific).
- `G6` Release docs and evidence bundle complete.
- `G7` Go/no-go sign-off complete.
- `G8` Relay no-long-history-hosting regression checks pass.
- `G9` `F23` as-built conformance report completed against v22 specification package.
- `G10` Required `F24` seed package and `v24+` deferral register complete.
- `G11` Architecture coverage audit completed and approved (no “unknown persistence” left).

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v23/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v23/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- Commands declared in `docs/v2.3/phase3/p3-podman-scenarios.md`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v23-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start only after entry criteria (v22 promoted, `F23` specs approved, deferred scope frozen) are satisfied; Acceptance rule—complete ST1–ST4, update scope docs, and mark `G0` Pass; Evidence rule—attach at least one `EV-v23-G0-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—before running documentation or gate planning commands (e.g., `make check-full`), record the planned command and capture its output.
- `P0-T2`: Dependency rule—after `P0-T1` scope lock artifacts; Acceptance rule—complete ST1–ST3 of the architecture coverage audit, flag missing classes, and mark the `G11` audit row Pass; Evidence rule—attach at least one `EV-v23-G11-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (e.g., inventory scripts or doc generation) before running and capture output.
- `P1-T1`: Dependency rule—after `G0` and scope artifacts; Acceptance rule—complete ST1–ST3 abuse resistance docs/tests and mark `G2` Pass; Evidence rule—≥1 `EV-v23-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `go test ./pkg/v23/security/...`) and capture output.
- `P1-T2`: Dependency rule—after `P1-T1`; Acceptance rule—complete ST1–ST3 privacy conformance work and mark `G2` Pass; Evidence rule—≥1 `EV-v23-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the next command (e.g., `go test ./tests/e2e/v23/privacy_*`) and note output.
- `P1-T3`: Dependency rule—after `P1-T2`; Acceptance rule—complete ST1–ST2 integrity hardening and mark `G2` Pass; Evidence rule—attach `EV-v23-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the command (e.g., `go test ./pkg/v23/integrity/...`) and capture output.
- `P1-T4`: Dependency rule—after `P1-T3`; Acceptance rule—complete ST1–ST3 durability hardening artifacts and mark `G2` Pass; Evidence rule—≥1 `EV-v23-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note command (e.g., `go test ./pkg/v23/durability/...`) and capture output.
- `P2-T1`: Dependency rule—after `G2`; Acceptance rule—complete ST1–ST3 SLO definitions and scorecards, update docs/tests, and mark `G3` Pass; Evidence rule—`EV-v23-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record command (e.g., `tests/perf/v23/*` runners) and capture output.
- `P2-T2`: Dependency rule—after `P2-T1`; Acceptance rule—complete ST1–ST2 resource controls and mark `G3` Pass; Evidence rule—≥1 `EV-v23-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note command (e.g., `go test ./pkg/v23/limits/...`) and capture output.
- `P3-T1`: Dependency rule—after `G3`; Acceptance rule—complete ST1–ST3 operator runbooks/drills and mark `G4` Pass; Evidence rule—`EV-v23-G4-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—document the command (e.g., doc tooling runs) before execution and capture output.
- `P3-T2`: Dependency rule—after `P3-T1`; Acceptance rule—complete ST1–ST5 regression matrix, Podman scenarios, and mark `G5` Pass; Evidence rule—attach `EV-v23-G5-###` entry plus supporting command outputs (including those declared in `docs/v2.3/phase3/p3-podman-scenarios.md`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command(s) (e.g., Podman scenario scripts) and capture outputs.
- `P4-T1`: Dependency rule—after `G5`; Acceptance rule—publish docs/release artifacts (ST1–ST3) and mark `G6` Pass; Evidence rule—≥1 `EV-v23-G6-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record doc/publish commands and capture outputs.
- `P4-T2`: Dependency rule—after `P4-T1`; Acceptance rule—publish `F24` seeds/deferrals (ST1–ST4) and mark `G10` Pass while closing any remaining audit items (G11); Evidence rule—attach `EV-v23-G10-###` and `EV-v23-G11-###` entries plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the planned command (e.g., spec publishing script) and capture output.
- `P5-T1`: Dependency rule—after all prior phases and required seeds; Acceptance rule—complete ST1–ST4 for final evidence bundle, conformance reports, and sign-offs and mark `G7` Pass (with G8/G9 as needed); Evidence rule—attach `EV-v23-G7-###`, `EV-v23-G8-###`, and `EV-v23-G9-###` entries plus command outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned commands (`go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`) before running and capture outputs.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1`, `P0-T2` | Entry criteria (v22 promoted + `F23` specs approved) | `G0`, `G11` | `EV-v23-G0-###`, `EV-v23-G11-###` | Record scope/audit commands (e.g., `make check-full`, architecture inventory scripts) and capture outputs. |
| Phase 1 | `P1-T1`, `P1-T2`, `P1-T3`, `P1-T4` | `G0` completion | `G1`, `G2` | `EV-v23-G1-###`, `EV-v23-G2-###` | Capture abuse/privacy/integrity command outputs (`go test ./pkg/v23/*`, `tests/e2e/v23/*`). |
| Phase 2 | `P2-T1`, `P2-T2` | `G2` completion | `G3` | `EV-v23-G3-###` | Record SLO/performance commands (`tests/perf/v23/*`, `go test ./pkg/v23/limits/...`) and log outputs. |
| Phase 3 | `P3-T1`, `P3-T2` | `G3` completion | `G4`, `G5` | `EV-v23-G4-###`, `EV-v23-G5-###` | Plan runbook/regression commands (Podman scenarios per `docs/v2.3/phase3/p3-podman-scenarios.md`) and capture outputs. |
| Phase 4 | `P4-T1`, `P4-T2` | `G5` completion | `G6`, `G10`, `G11` | `EV-v23-G6-###`, `EV-v23-G10-###`, `EV-v23-G11-###` | Document spec/release commands (doc builds, seed publication scripts) before running and log outputs. |
| Phase 5 | `P5-T1` | Prior gates (G0–G6) and seeds | `G7`, `G8`, `G9` | `EV-v23-G7-###`, `EV-v23-G8-###`, `EV-v23-G9-###` | Capture final evidence commands (`go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`) and release/drill scripts before running and record outputs. |

## Phase plan

### Phase 0 - Scope lock and hardening matrix (G0)
- [x] `P0-T1` Freeze v23 hardening scope and pass/fail criteria.
  - `ST1` Import v22 `F23` acceptance matrix and convert to explicit go/no-go checks.
  - `ST2` Freeze security and privacy invariants:
    - no keyword leakage by default
    - private Space anti-enumeration remains mandatory
    - quotas/retention refusal reasons are deterministic
    - durability is explicitly labeled when replica targets are not met
  - `ST3` Produce requirement-to-artifact traceability matrix.
  - `ST4` Assign gate ownership and approvers using RACI template.
  - Artifacts:
    - `docs/v2.3/phase0/p0-scope-lock.md`
    - `docs/v2.3/phase0/p0-traceability-matrix.md`
    - `docs/v2.3/phase0/p0-gate-ownership.md`
    - `docs/v2.3/phase0/p0-hardening-matrix.md`

- [x] `P0-T2` Define architecture coverage audit checklist (for G11).
  - `ST1` List persisted data classes (minimum):
    - message timelines (history segments)
    - store-and-forward TTL (if exists)
    - attachments/blobs
    - profile fields + profile media
    - server/Space config + channels + roles/permissions metadata
    - custom emojis/stickers
    - indexes (FTS, discovery indexes)
    - moderation artifacts (reports, tombstones)
  - `ST2` For each class, record:
    - authoritative writer(s)
    - storage plane(s) (local DB, DHT, IPFS/blobstore, Archivist segments, etc.)
    - retrieval path(s) for: new device, returning device, new joiner
    - confidentiality/integrity guarantees and explicit non-guarantees
    - retention policy owner + defaults
  - `ST3` Mark “missing/unknown” items as mandatory `F24` seeds (or v23 hardening if blocking).
  - Artifacts:
    - `docs/v2.3/phase0/p0-architecture-coverage-audit.md`

### Phase 1 - Security hardening (G2)
- [x] `P1-T1` Complete history/search abuse resistance backlog.
  - `ST1` Enforce conservative default quotas/retention on Archivist.
  - `ST2` Implement request rate limits and bounded response sizing.
  - `ST3` Implement abuse telemetry (privacy-preserving counters, not content).
  - Artifacts:
    - `pkg/v23/security/*`
    - `docs/v2.3/phase1/p1-abuse-hardening-log.md`

- [x] `P1-T2` Complete privacy conformance backlog.
  - `ST1` Assert protocol does not allow keyword-bearing backfill requests by default.
  - `ST2` Ensure coverage labeling never implies full history if not present.
  - `ST3` Add explicit opt-in scaffolding (no implementation) for any future “assisted search” modes; ensure it is off-by-default and gated.
  - Artifacts:
    - `docs/v2.3/phase1/p1-privacy-conformance.md`
    - `tests/e2e/v23/privacy_*`

- [x] `P1-T3` Integrity and verification robustness hardening.
  - `ST1` Harden verification code paths against malformed manifests/heads.
  - `ST2` Ensure deterministic error mapping for all invalid cases.
  - Artifacts:
    - `pkg/v23/integrity/*`
    - `tests/e2e/v23/integrity_*`

- [x] `P1-T4` Durability/replication hardening.
  - `ST1` Verify replica-set accounting under churn and partial failures.
  - `ST2` Ensure degraded durability is surfaced as an explicit state (API + UI label).
  - `ST3` Add regression tests for “replica target unmet but service still functional”.
  - Artifacts:
    - `pkg/v23/durability/*`
    - `tests/e2e/v23/replica_*`
    - `docs/v2.3/phase1/p1-durability-conformance.md`

### Phase 2 - Reliability/performance SLO gates (G3)
- [x] `P2-T1` Define and enforce history/search SLOs.
  - `ST1` Backfill SLOs (examples: time-to-first-page; bounded retry duration).
  - `ST2` Search SLOs (query p50/p95 under DB size tiers).
  - `ST3` Archivist SLOs (ingest rate, prune cadence, disk growth bounds).
  - Artifacts:
    - `docs/v2.3/phase2/p2-slo-scorecard.md`
    - `tests/perf/v23/*`

- [x] `P2-T2` Implement bounded resource controls.
  - `ST1` CPU/IO limits for backfill verification and indexing.
  - `ST2` Disk growth alarms + safe refusal behavior when limits exceeded.
  - Artifacts:
    - `pkg/v23/limits/*`
    - `docs/v2.3/phase2/p2-resource-bounds.md`

### Phase 3 - Podman ops readiness and regression (G4, G5)
- [x] `P3-T1` Complete Archivist operator runbook and drills (`docs/v2.3/phase3/p3-archivist-operator-runbook.md`, `docs/v2.3/phase3/p3-rollback-drill.md`, `docs/v2.3/phase3/p3-incident-playbook.md`).
  - `ST1` Monitoring/alerts:
    - storage growth alarms
    - quota exhaustion alarms
    - prune lag alarms
    - replica-target unmet alarms (durability degraded)
  - `ST2` Upgrade/rollback drill with evidence capture.
  - `ST3` Incident response playbook for abuse and privacy incidents.
  - Artifacts:
    - `docs/v2.3/phase3/p3-archivist-operator-runbook.md`
    - `docs/v2.3/phase3/p3-rollback-drill.md`
    - `docs/v2.3/phase3/p3-incident-playbook.md`
  - Evidence: placeholders in `docs/v2.3/phase5/p5-evidence-index.md` using the EV-v23-G4-### and EV-v23-G5-### naming conventions until the drills execute.

- [x] `P3-T2` Execute full regression matrix for history/search plane.
  - `ST1` Offline catch-up scenarios across NAT/network chaos.
  - `ST2` Redaction tombstone regressions (local store + search + backfill).
  - `ST3` Private Space anti-enumeration regressions.
  - `ST4` Replica healing regressions under churn.
  - `ST5` Relay no-long-history-hosting regression scenario for history/search.
  - Artifacts:
    - `containers/v2.3/*`
    - `docs/v2.3/phase3/p3-podman-scenarios.md`
    - `docs/v2.3/phase3/p3-regression-report.md`

### Phase 4 - Release docs and v24+ seed package (G6, G10, G11)
- [x] `P4-T1` Publish history/search plane documentation.
  - `ST1` Operator docs: Archivist role, quotas, retention, refusal reasons, durability labeling.
  - `ST2` User docs: history availability, search coverage labels, backfill behavior.
  - `ST3` Security/privacy docs: metadata model, default guarantees, explicit non-guarantees.
  - Artifacts:
    - `docs/v2.3/phase4/p4-history-search-docs.md`
    - `docs/v2.3/phase4/p4-release-notes.md`

- [x] `P4-T2` Publish required `F24` package and `v24+` deferrals.
  - `ST1` Publish `F24` seed specification and additive proto delta (even if empty).
  - `ST2` Publish `F24` acceptance matrix.
  - `ST3` Publish deferral register with rationale, owner, and revisit target.
  - `ST4` Ensure `F24` seeds explicitly cover any missing items from the architecture coverage audit (P0-T2).
  - **Minimum `F24` seed candidates (unless already satisfied earlier in the roadmap):**
    - `F24-A` **Distributed encrypted blob store** for attachments/avatars/emojis (ciphertext-only, content-addressed, quota/retention, multi-provider pinning).
    - `F24-B` **Multi-device history policy**: per-device backfill limits + “bring-your-own device cache” guidelines (avoid every device re-downloading everything).
    - `F24-C` **Assisted search research pack** (opt-in only): evaluate privacy-preserving approaches (SSE/PIR/Bloom-filter hinting) and define explicit leakage budgets if ever implemented.
    - `F24-D` **Desktop background mode** option: (optional) local daemon attachment mode for harmolyn ↔ xorein, including auth and OS service management (if desired).
  - Artifacts:
    - `docs/v2.3/phase4/f24-backlog-and-spec-seeds.md`
    - `docs/v2.3/phase4/f24-proto-delta.md`
    - `docs/v2.3/phase4/f24-acceptance-matrix.md`
    - `docs/v2.3/phase4/f24-deferral-register.md`
    - `docs/v2.3/phase4/p4-architecture-coverage-audit-result.md` (finalized copy of P0-T2 with approvals)

### Phase 5 - Final evidence and go/no-go (G7)
- [x] `P5-T1` Publish final evidence bundle and v23 promotion decision.
  - `ST1` Attach all gate outputs and residual risk sign-offs.
  - `ST2` Publish `F23` as-built conformance report against v22 `F23` specs.
  - `ST3` Record promotion decision and any restricted-release posture if needed.
  - `ST4` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.3/phase5/p5-final-evidence-bundle.md`
    - `docs/v2.3/phase5/p5-go-no-go-record.md`
    - `docs/v2.3/phase5/p5-as-built-conformance.md`
    - `docs/v2.3/phase5/p5-gate-signoff.md`
    - `docs/v2.3/phase5/p5-evidence-index.md`

## Risk register (v23)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R23-1 | Abuse resistance insufficient | quota defaults + rate limits + drills | abuse tests/drills pass |
| R23-2 | Privacy guarantees unclear to users/operators | explicit docs + coverage labeling enforcement | doc + UX checks pass |
| R23-3 | Archivist ops instability | mandatory runbooks + rollback drills | drill evidence complete |
| R23-4 | Late regressions discovered | full regression matrix gating | regression gate passes |
| R23-5 | Relay boundary regresses | dedicated regression checks | relay-boundary tests pass |
| R23-6 | Hidden persistence gaps outside history plane | architecture coverage audit + F24 seeds | G11 passes; no unknowns remain |

## Decision log (v23)
- `D23-1`: No keyword leakage by default is a release invariant for the history/search plane.
- `D23-2`: Archivist operator readiness is mandatory; “works on my machine” is not promotable.
- `D23-3`: Coverage labeling is a safety feature; it must not overclaim history completeness.
- `D23-4`: Relays do not store long-lived history segments/manifests; durable history is an Archivist capability.
- `D23-5`: Post-v23 roadmap MUST be driven by the architecture coverage audit; any missing persistence class becomes an explicit seed or deferral (no implicit magic).
