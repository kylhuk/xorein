# TODO v24 - Xorein (backend daemon) + harmolyn (frontend) Local API + Process Split (v2.4)

## Status
Planning artifact only. This file defines v24 implementation and validation requirements. It does not claim implementation completion.

## Naming + architecture note (carry-forward, clarified)
- **Xorein**: backend node/runtime binary (`cmd/xorein`). Must be fully headless-capable and must not depend on any UI packages.
- **harmolyn**: frontend UI binary (`cmd/harmolyn`). Must not import protocol runtime packages directly; it consumes Xorein via a **local-only API**.
- v24 converts the “library embed” posture into an explicit **two-process** model (by default on desktop):
  - harmolyn launches or attaches to a per-user Xorein daemon.
  - The daemon listens on a local transport only (Unix domain socket / Windows named pipe).
  - Local API is versioned and is **never** exposed on a public network interface.

## Version Isolation Contract (mandatory)
- v24 cannot close unless all v24 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v24 implements only `F24` from v23 `F24` seed package (additive proto only, where proto applies).
- v24 must also publish full `F25` spec package for v25 implementation.

## Version role
- Implements: `F24` (local-daemon split, local API contract, multi-client attach, and packaging/runbooks).
- Specifies: `F25` (ciphertext blob/asset distribution plane for attachments, avatars, emojis, and other non-message data).

## Critical scope (v24)
- Implement a local-only API between harmolyn and Xorein:
  - transport: UDS / named pipe (no TCP listener by default)
  - authentication: per-user access + session auth token derived from local device secret
  - authorization: API surface enforces the same permission checks as “in-process” callers (no bypass)
  - streaming events: timeline updates, call state, backfill progress, connectivity reasons
  - compatibility: additive-only evolution within major API version; explicit API version negotiation
- Implement Xorein daemon lifecycle:
  - one daemon per OS user profile (default)
  - harmolyn launches if missing, reuses if present
  - deterministic shutdown semantics and stale-socket recovery
- Implement harmolyn attach mode:
  - bootstraps/attaches to Xorein, shows deterministic “backend not available” UX
  - no-limbo error + next-action taxonomy for backend attach failures
- Implement operator/developer tooling:
  - `xorein doctor` diagnostics
  - `xorein api-version` reporting
  - optional `xorein --headless` UX-free client mode for scripting and e2e
- Preserve existing network behavior and invariants:
  - relay no-long-history-hosting invariant remains unchanged
  - Archivist capability remains opt-in (v22+)
  - no keyword leakage by default remains a privacy invariant
- Produce `F25` spec package for v25.

## Out of scope (defer)
- Exposing the local API over a remote network interface is out of scope.
- A public “remote control” API for third-party untrusted apps is out of scope.
- A plugin sandbox model is out of scope unless previously planned elsewhere.

## Dependencies and relationships
- Inputs from v23 (must exist before v24 starts):
  - `docs/v2.3/phase4/f24-backlog-and-spec-seeds.md`
  - `docs/v2.3/phase4/f24-proto-delta.md`
  - `docs/v2.3/phase4/f24-acceptance-matrix.md`
  - `docs/v2.3/phase4/f24-deferral-register.md`
- Outputs consumed by v25:
  - `docs/v2.4/phase4/f25-blob-store-spec.md`
  - `docs/v2.4/phase4/f25-proto-delta.md`
  - `docs/v2.4/phase4/f25-acceptance-matrix.md`
- Binary deliverables (v24):
  - `cmd/xorein` (backend node/daemon)
  - `cmd/harmolyn` (frontend UI)
  - optional helper: `cmd/xoreinctl` (only if `xorein` CLI becomes too heavy; otherwise keep subcommands in `xorein`)

## Entry criteria (must be true before implementation starts)
- `v23` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F24` spec inputs from v23 exist and are approved.
- v24 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and requirement-to-artifact traceability complete.
- `G1` Local API spec and version-negotiation contract complete.
- `G2` Xorein daemon mode complete (transport, auth, lifecycle, logging/metrics).
- `G3` harmolyn attach mode complete (all critical journeys functional via API only).
- `G4` Security test matrix for local API complete (authz, replay, injection, downgrade/compat).
- `G5` Multi-client and recovery scenarios complete (daemon reuse, crash recovery, stale socket).
- `G6` Podman and desktop harness scenarios complete (end-to-end across both binaries).
- `G7` `F25` spec package complete.
- `G8` Docs and evidence bundle complete.
- `G9` “Backend has no UI deps” invariant audited and enforced in CI.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v24/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v24/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- `scripts/v24-daemon-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v24-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start once entry criteria (v23 promoted, `F24` specs approved, deferred scope frozen) are satisfied; Acceptance rule—complete ST1–ST5, publish scope/docs, and mark `G0` Pass; Evidence rule—attach at least one `EV-v24-G0-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—record upcoming scope commands (e.g., `make check-full`) and log outputs.
- `P0-T2`: Dependency rule—after `P0-T1`; Acceptance rule—complete ST1–ST4 for local API spec/no-go states and mark `G1` Pass; Evidence rule—attach `EV-v24-G1-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note planned command (spec rendering, threat model checks) and capture output.
- `P1-T1`: Dependency rule—after `G1`; Acceptance rule—complete ST1–ST4 local API server scaffolding artifacts and mark `G2` Pass; Evidence rule—≥1 `EV-v24-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record command (e.g., `go test ./proto/xorein_local/...`, `buf lint`) and capture output.
- `P1-T2`: Dependency rule—after `P1-T1`; Acceptance rule—complete ST1–ST3 daemon lifecycle work, docs, and mark `G2` Pass; Evidence rule—attach `EV-v24-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the next command (e.g., `go test ./pkg/v24/daemon/...`, `scripts/xorein-doctor.sh`) and log output.
- `P1-T3`: Dependency rule—after `P1-T2`; Acceptance rule—complete ST1–ST2 boundary enforcement tests, publish CI rules, and mark `G2` and `G9` Pass; Evidence rule—≥1 `EV-v24-G2-###` and `EV-v24-G9-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (e.g., boundary enforcement script) and capture output.
- `P2-T1`: Dependency rule—after `G2`; Acceptance rule—complete ST1–ST3 bootstrap/attach workflow artifacts and mark `G3` Pass; Evidence rule—`EV-v24-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note command (e.g., `go test ./pkg/v24/harmolyn/attach/...`) before running and capture output.
- `P2-T2`: Dependency rule—after `P2-T1`; Acceptance rule—complete ST1–ST4 journey migrations to API-only and mark `G3` Pass; Evidence rule—attach `EV-v24-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (e.g., `tests/e2e/v24/journeys_*`) and capture output.
- `P3-T1`: Dependency rule—after `G3`; Acceptance rule—complete ST1–ST3 security tests (auth, replay, injection) and mark `G4` Pass; Evidence rule—≥1 `EV-v24-G4-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—document the upcoming command (`tests/e2e/v24/localapi_security_*`, `tests/fuzz/v24/*`) before running and log output.
- `P3-T2`: Dependency rule—after `P3-T1`; Acceptance rule—complete ST1–ST3 multi-client scenarios, crash/recovery, and mark `G5` Pass; Evidence rule—attach `EV-v24-G5-###` entry plus supporting command outputs (including `scripts/v24-daemon-scenarios.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (`scripts/v24-daemon-scenarios.sh`, `tests/e2e/v24/multiclient_*`) and record output.
- `P4-T1`: Dependency rule—after `G5`; Acceptance rule—complete ST1–ST5 for `F25` spec package and mark `G7` Pass; Evidence rule—≥1 `EV-v24-G7-###` entry plus documentation outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record spec publishing commands and capture outputs.
- `P5-T1`: Dependency rule—after all prior gates (G0–G7) and invariants; Acceptance rule—complete ST1–ST3 for evidence bundle, conformance report, sign-off, and mark `G8` Pass (with G9 audit + dependency enforcement); Evidence rule—attach `EV-v24-G8-###` plus `EV-v24-G9-###` entries plus command outputs (`go build`, `go test`, `make check-full`, `scripts/v24-daemon-scenarios.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record planned commands (`go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`, scenario scripts) before running and capture outputs.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1`, `P0-T2` | Entry criteria (v23 promoted + `F24` specs approved) | `G0`, `G1` | `EV-v24-G0-###`, `EV-v24-G1-###` | Record API/scope commands (`make check-full`, spec rendering) before running and capture outputs. |
| Phase 1 | `P1-T1`, `P1-T2`, `P1-T3` | `G1` completion | `G2`, `G9` | `EV-v24-G2-###`, `EV-v24-G9-###` | Capture daemon/local API commands (`go test ./proto/xorein_local/...`, boundary enforcement scripts). |
| Phase 2 | `P2-T1`, `P2-T2` | `G2` completion | `G3` | `EV-v24-G3-###` | Log attach/journey commands (`go test ./pkg/v24/harmolyn/attach/...`, `tests/e2e/v24/journeys_*`) before running and record outputs. |
| Phase 3 | `P3-T1`, `P3-T2` | `G3` + Phase 2 artifacts | `G4`, `G5` | `EV-v24-G4-###`, `EV-v24-G5-###` | Plan security/multi-client commands (`tests/e2e/v24/localapi_security_*`, `tests/fuzz/v24/*`, `scripts/v24-daemon-scenarios.sh`) and capture outputs. |
| Phase 4 | `P4-T1` | `G5` completion | `G7` | `EV-v24-G7-###` | Record spec publishing commands and doc builds before running and capture outputs. |
| Phase 5 | `P5-T1` | Previous gates (G0–G7) | `G8`, `G9` | `EV-v24-G8-###`, `EV-v24-G9-###` | Capture evidence commands (`go build`, `go test`, `make check-full`, `scripts/v24-daemon-scenarios.sh`) and log outputs. |

## Phase plan

### Phase 0 - Scope lock and API surface freeze (G0, G1)
- [ ] `P0-T1` Freeze v24 scope to the local-daemon split and local API.
  - `ST1` Import v23 `F24` acceptance matrix; convert to explicit go/no-go checks.
  - `ST2` Produce requirement-to-artifact traceability matrix (spec → code → tests → docs).
  - `ST3` Freeze non-goals (no remote API exposure; no plugin sandbox).
  - `ST4` Assign gate ownership and approvers using RACI template.
  - `ST5` Declare API evolution policy:
    - major: breaking changes only with explicit negotiation
    - minor: additive-only messages/fields and additive services
    - patch: bugfix only (no schema changes)
  - Artifacts:
    - `docs/v2.4/phase0/p0-scope-lock.md`
    - `docs/v2.4/phase0/p0-traceability-matrix.md`
    - `docs/v2.4/phase0/p0-gate-ownership.md`
    - `docs/v2.4/phase0/p0-local-api-evolution.md`

- [ ] `P0-T2` Write the local API specification and transport security profile.
  - `ST1` Specify transport targets:
    - Linux/macOS: Unix domain socket path conventions + file perms
    - Windows: named pipe path + ACL requirements
  - `ST2` Specify authentication handshake:
    - client presents `ClientHello` with nonce + claimed API version range
    - daemon responds with `ServerHello` + selected version + session nonce
    - client proves possession of local device secret-derived key (HMAC or ECDH-based)
    - daemon returns short-lived `SessionToken` for subsequent RPC metadata
  - `ST3` Specify authorization invariants:
    - all mutating RPCs require an authenticated session
    - sensitive RPCs (identity export, key material, bridge config) require explicit “danger zone” capability bit
  - `ST4` Specify event stream contract:
    - single multiplexed stream or multiple streams (must be bounded and backpressure-aware)
    - deterministic reconnection/resume (cursor-based) for UI restart
  - Artifacts:
    - `docs/v2.4/phase0/p0-local-api-spec.md`
    - `docs/v2.4/phase0/p0-local-api-threat-model.md`
    - `docs/v2.4/phase0/p0-error-taxonomy.md`

### Phase 1 - Xorein daemon runtime (G2, G9)
- [ ] `P1-T1` Implement local API server scaffolding in Xorein.
  - `ST1` Define local API protobuf surface (separate from network proto):
    - `proto/xorein_local/v1/*.proto`
    - codegen target: `gen/go/xorein_local/v1/*`
  - `ST2` Implement daemon listener for:
    - UDS (unix) / named pipe (windows)
    - explicit refusal if attempting to bind to non-local interface
  - `ST3` Implement session handshake + token validation middleware.
  - `ST4` Implement structured audit logs for sensitive RPCs (no plaintext content):
    - include request type, SpaceID/ChannelID where relevant, outcome reason code
  - Artifacts:
    - `proto/xorein_local/v1/*`
    - `pkg/v24/localapi/*`
    - `pkg/v24/localapi/*_test.go`
    - `docs/v2.4/phase1/p1-local-api-as-built.md`

- [ ] `P1-T2` Implement daemon lifecycle management.
  - `ST1` One-per-user locking:
    - lockfile or OS-native singleton
    - deterministic refusal reason for “already running”
  - `ST2` Stale socket recovery and crash recovery:
    - detect orphan sockets and remove safely
    - ensure harmolyn can restart daemon deterministically
  - `ST3` Implement `xorein doctor`:
    - report versions, socket status, basic health probes, last crash marker
  - Artifacts:
    - `pkg/v24/daemon/*`
    - `pkg/v24/daemon/*_test.go`
    - `scripts/xorein-doctor.sh` (optional)
    - `docs/v2.4/phase1/p1-daemon-lifecycle.md`

- [ ] `P1-T3` Enforce “no UI deps” and package boundaries in CI.
  - `ST1` Add CI rule: `cmd/xorein` and `pkg/xorein/**` must not import Gio packages.
  - `ST2` Add CI rule: `cmd/harmolyn` must not import protocol runtime internals directly (only local API client + thin glue).
  - Artifacts:
    - `scripts/ci/enforce-boundaries.sh`
    - `docs/v2.4/phase1/p1-boundary-enforcement.md`

### Phase 2 - harmolyn attach mode (G3)
- [ ] `P2-T1` Implement harmolyn bootstrap/attach workflow.
  - `ST1` Attach sequence:
    - detect running daemon (socket probe)
    - if missing: spawn `cmd/xorein --daemon` and wait for readiness
    - connect and complete handshake; record session token
  - `ST2` Deterministic UX for failure states:
    - `DAEMON_START_FAILED`, `DAEMON_INCOMPATIBLE`, `AUTH_FAILED`, `SOCKET_PERMISSION_DENIED`
    - each state includes a “next action” (retry, open logs, repair, reset)
  - `ST3` Implement graceful detach and reattach on daemon restart.
  - Artifacts:
    - `pkg/v24/harmolyn/attach/*`
    - `pkg/v24/harmolyn/attach/*_test.go`
    - `docs/v2.4/phase2/p2-attach-ux-contract.md`

- [ ] `P2-T2` Move critical user journeys to API-only execution.
  - `ST1` Identity journeys: create, restore from backup, device list/revoke (if present).
  - `ST2` Space/channel journeys: discover, join, list channels, send/receive messages.
  - `ST3` History journeys: local timeline, local search, backfill (if enabled from v22), coverage labels.
  - `ST4` Media journeys: voice + screen share controls must route via API (even if media engine remains inside daemon).
  - Artifacts:
    - `pkg/v24/harmolyn/ui/*`
    - `tests/e2e/v24/journeys_*`

### Phase 3 - Security + multi-client validation (G4, G5)
- [ ] `P3-T1` Add security tests for local API.
  - `ST1` Unauthorized access tests:
    - wrong user perms (UDS/named pipe)
    - invalid session token
  - `ST2` Replay and downgrade tests:
    - reuse old handshake nonce
    - attempt API version downgrade below minimum supported
  - `ST3` Injection and fuzzing:
    - malformed protobuf frames
    - oversized payload bounds
  - Artifacts:
    - `tests/e2e/v24/localapi_security_*`
    - `tests/fuzz/v24/*`
    - `docs/v2.4/phase3/p3-security-report.md`

- [ ] `P3-T2` Multi-client attach and recovery scenarios.
  - `ST1` Two harmolyn instances attach to one daemon (read-only concurrency allowed; writes serialized/deterministic).
  - `ST2` Daemon crash mid-call: harmolyn reconnects and shows deterministic recovery states.
  - `ST3` Stale socket simulation and auto-repair.
  - Artifacts:
    - `tests/e2e/v24/multiclient_*`
    - `scripts/v24-daemon-scenarios.sh`
    - `docs/v2.4/phase3/p3-daemon-scenarios.md`

### Phase 4 - v25 spec package (G7)
- [ ] `P4-T1` Produce ciphertext blob/asset distribution specification package (`F25`).
  - `ST1` Define supported data classes:
    - attachments (files)
    - avatars (user + Space)
    - custom emojis and stickers (if present)
    - optional: pinned media/thumbs (ciphertext-only)
  - `ST2` Define `BlobRef` model:
    - content hash, size, mime, chunking parameters
    - encryption envelope and access control
  - `ST3` Define storage provider capability model (re-use Archivist or define separate capability).
  - `ST4` Define retrieval and anti-enumeration requirements for private Spaces.
  - `ST5` Define acceptance matrix with perf and quota requirements.
  - Artifacts:
    - `docs/v2.4/phase4/f25-blob-store-spec.md`
    - `docs/v2.4/phase4/f25-proto-delta.md`
    - `docs/v2.4/phase4/f25-acceptance-matrix.md`

### Phase 5 - Closure and evidence (G8)
- [ ] `P5-T1` Publish v24 evidence bundle and promotion recommendation.
  - `ST1` Attach all command outputs and scenario manifests.
  - `ST2` Publish `F24` as-built conformance report against v23 `F24` package.
  - `ST3` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.4/phase5/p5-evidence-bundle.md`
    - `docs/v2.4/phase5/p5-risk-register.md`
    - `docs/v2.4/phase5/p5-as-built-conformance.md`
    - `docs/v2.4/phase5/p5-gate-signoff.md`
    - `docs/v2.4/phase5/p5-evidence-index.md`

## Risk register (v24)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R24-1 | Local API creates a new bypass path | auth middleware + “same checks as in-process” + adversarial tests | security matrix passes |
| R24-2 | Cross-platform local transport complexity | UDS + named pipe only; minimal surface; platform tests | platform attach tests pass |
| R24-3 | Daemon lifecycle flakiness | singleton lock + stale socket recovery + crash scenarios | multi-client scenarios pass |
| R24-4 | UI/backend dependency drift | CI boundary enforcement | boundary CI gates pass |

## Decision log (v24)
- `D24-1`: Local API transport is non-network (UDS/named pipe) by default; TCP is not enabled by default.
- `D24-2`: harmolyn consumes Xorein exclusively via local API in v24 (no direct runtime imports).
- `D24-3`: Sensitive RPCs require explicit capability bits and are audit-logged (without plaintext).
- `D24-4`: API evolution is additive-only within major; negotiation is mandatory to avoid silent breakage.
