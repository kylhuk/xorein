# Xorein Backend Server Rewrite

## TL;DR
> **Summary**: Replace the current HTTP/JSON peer runtime with a protocol-native p2p server, while preserving the local-only control plane and making persistence, discovery, relay, and verification real.
> **Deliverables**: p2p runtime bootstrap, protocol negotiation, durable storage, discovery/bootstrap, server/join/delivery flows, relay/history support, local control API continuity, real test/build gates.
> **Effort**: XL
> **Parallel**: YES - 3 waves
> **Critical Path**: Runtime bootstrap → transport/protocol core → peer flows/persistence → relay/control-plane continuity → verification

## Context
### Original Request
Xorein needs to be a fully working server for serving the innovative protocol, which is p2p, to the clients.

### Interview Summary
- The live runtime starts in `cmd/aether/main.go` and currently boots `pkg/node.Service`.
- `pkg/node/service.go` serves peer traffic over HTTP/JSON today, while `pkg/node/control.go` is the local-only authenticated control plane.
- `pkg/protocol/*` and `proto/aether.proto` already define the intended protocol/version/capability surface, but they are not the live transport.
- Prior planning is spread across the v0.1/v0.3 docs, local control API docs, release naming, and the later QoL/discovery addendum.
- The target architecture is a replacement runtime, not a polish pass on the current HTTP peer server.
- Test strategy is TDD.

### Metis Review (gaps addressed)
- Lock the end-state architecture: protocol-native p2p server, not HTTP peer runtime.
- Keep the local control API local-only unless the plan explicitly changes that boundary.
- Preserve additive-only wire evolution and explicit downgrade/version handling.
- Avoid scope creep into client UI, docs cleanup, or unrelated product features.

### Implementation Decisions
- **Seam boundary**: the peer runtime must be extracted behind an explicit transport/runtime seam; the local control plane stays in `pkg/node/control.go` and only reaches the runtime through service interfaces, not direct peer-handler reuse.
- **Transport stack**: peer traffic uses the protocol-native p2p stack with `pkg/protocol` as the negotiation source of truth and protobuf-defined payloads on the wire.
- **Storage backend**: runtime state uses the encrypted durable storage path selected in the repo (`SQLCipher`-backed persistence), with explicit schema versioning and quarantine semantics.
- **Mode semantics**:
  - `client` = primary p2p client/runtime mode.
  - `relay` = store-and-forward / connectivity role.
  - `bootstrap` = discovery entrypoint role.
  - `archivist` = long-lived history/capability overlay role.
  - Startup mode is the CLI selection; role/capability is what the runtime advertises on the wire.
- **Role boundaries**: this plan covers `client`, `relay`, `bootstrap`, and `archivist`; the blob-provider capability stays deferred/out of scope unless a later plan explicitly adds it.

### Architecture Checkpoint
- **Transport seam**: peer runtime code lives behind a dedicated p2p transport layer owned by `pkg/network/*`; `pkg/node/control.go` remains the local-only operator boundary.
- **Storage seam**: state/migration code lives behind a dedicated storage package (planned under `pkg/storage/*`); `pkg/node/state.go` becomes the node-facing facade.
- **Storage feasibility checkpoint**: confirm the SQLCipher-backed storage package exists in-tree or create the package boundary in the same task before attempting migration.
- **Proto generation**: `proto/aether.proto` is the source of truth, `gen/go/proto/aether.pb.go` is generated-only, and `buf generate` is required whenever proto assets change.

### Ownership Split
| Concern | Owner | Notes |
|---|---|---|
| Peer runtime | `pkg/network/*` | libp2p/transport, stream handlers, peer-facing serving |
| Local control | `pkg/node/control.go` | loopback/socket-only `/v1` operator surface |
| Orchestration | `pkg/node/service.go` | wires runtime, control plane, discovery, and state |
| Persistence | `pkg/storage/*` + `pkg/node/state.go` facade | SQLCipher-backed durable store and migration |
| Protocol negotiation | `pkg/protocol/*` | multistream IDs, capabilities, version compatibility |
| Wire schema | `proto/aether.proto` | source of truth for generated protobufs |

### Migration Order
- T3 establishes the storage seam and migration contract before any peer-flow cutover.
- T2 establishes transport/protocol negotiation before transport-facing handlers are moved.
- T5 rewires server manifest/join/delivery flows after T2/T3 exist.
- T4, T6, and T7 then build on the stabilized seams.
- T8 stays last so verification reflects the final runtime shape.
- T4 and T7 may temporarily use the current `pkg/node/state.go` facade while T3 lands, but they must not introduce a parallel persistence model.

### Compatibility Matrix
| Edge case | Expected outcome |
|---|---|
| Unsupported protocol version | Deterministic rejection |
| Wrong storage key / unreadable store | Explicit failure |
| Corrupt state file / store | Quarantine and recovery path |
| Partial migration | Old state intact or quarantined backup |

### Failure Policy
- **Protocol upgrade failure**: reject the incompatible peer, preserve the prior manifest/state, and do not partially activate the new wire shape.
- **Storage migration failure**: keep the old state intact or quarantine the failed attempt; never leave a half-migrated store as the only copy.
- **Control-plane continuity failure**: if the new runtime cannot preserve the documented local `/v1` surface, abort the cutover and keep the local control path stable.

## Work Objectives
### Core Objective
Make the backend operate as a real p2p server: nodes can start, negotiate protocol versions/capabilities, discover peers, create/join/preview servers, exchange deliveries, retain state safely, and keep the local control plane working.

### Deliverables
- P2P runtime bootstrap and node lifecycle
- Protocol negotiation and transport wiring
- Durable, recoverable storage
- Discovery/bootstrap/peer exchange
- Server manifest/join/delivery flows
- Relay/store-forward/history semantics
- Local-only control API continuity
- Real verification gates and integration coverage

### Definition of Done (verifiable conditions with commands)
- `go test ./...` passes.
- `go test -race ./...` passes for the backend packages.
- Protocol changes pass `buf lint` and `buf breaking` via Podman when `.proto` assets are touched.
- `pre-commit run --all-files` passes.
- `make check-full` and `make pipeline` pass once task 8 has replaced placeholder gates with real ones.
- The runtime starts in each supported mode and exposes the local control API only on loopback/socket.
- A multi-node integration scenario proves discovery, join, delivery, restart recovery, and relay drain.

### Must Have
- Replace the peer-facing HTTP runtime with the p2p transport/runtime target.
- Preserve the local-only authenticated control plane.
- Keep wire evolution additive and versioned.
- Ensure persistence survives restart and can recover from corruption.
- Provide explicit failure behavior for invalid signatures, unsupported protocol versions, and unauthorized access.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No client UI work.
- No docs-only cleanup as a substitute for implementation.
- No hidden transport/security assumptions.
- No unversioned wire breaks or field renumbering.
- No direct edits to `gen/go/proto/aether.pb.go`; regenerate protobuf artifacts instead.
- No expansion into public discovery/indexers/moderation unless explicitly re-scoped later.
- No blob-provider implementation in this plan.

## Verification Strategy
> ZERO HUMAN INTERVENTION during automated verification; final handoff still requires explicit user approval.
- Test decision: TDD + integration-first verification for the p2p paths.
- QA policy: Every task has agent-executed scenarios.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`
- Test file convention: package-local unit tests live alongside the touched code (`cmd/aether/*_test.go`, `pkg/node/*_test.go`, `pkg/protocol/*_test.go`); multi-node scenarios use a dedicated integration suite under `tests/integration/` or equivalent package-local integration tests.
- Per-wave gate pattern: use package-scoped tests and focused integration checks for the touched surface in each task; reserve the full repo commands for Task 8/final integration.

## Execution Strategy
### Parallel Execution Waves
> Maximize parallelism where dependencies allow; this graph lands at 4 / 3 / 1 tasks because storage, discovery, and verification each gate multiple downstream items.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: runtime bootstrap, transport core, persistence foundation, discovery foundation
Wave 2: peer protocol flows, relay/history semantics, local control API continuity
Wave 3: verification harness, Makefile/CI gate realignment, full-system regression

### Dependency Matrix (full, all tasks)
- T1 → T5, T7, T8
- T2 → T4, T5, T6, T8
- T3 → T5, T6, T7, T8
- T4 → T5, T6, T8
- T5 → T8
- T6 → T8
- T7 → T8
- T8 → none

### Agent Dispatch Summary (wave → task count → categories)
- Wave 1 → 4 tasks → deep, ultrabrain, unspecified-high
- Wave 2 → 3 tasks → deep, unspecified-high
- Wave 3 → 1 task → unspecified-high, deep

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Replace runtime bootstrap and mode wiring

  **What to do**: Rework `cmd/aether/main.go` and the service startup path so the end-state runtime boots the p2p stack instead of the current HTTP peer server, while preserving the local-only control plane and explicit mode selection (`client|relay|bootstrap|archivist`). Introduce a clear runtime boundary between the public p2p server and the local operator surface.

  **Must NOT do**: Keep the current HTTP peer server as the production serving path; add client UI behavior; introduce undocumented modes.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: this task defines the new process/runtime boundary and startup shape.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - too architecture-heavy.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T5, T7, T8 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `cmd/aether/main.go:51-130` - current entrypoint, mode flags, and startup flow.
  - `pkg/node/service.go:123-201` - current service bootstrap and runtime lifecycle.
  - `docs/release-naming.md:1-10` - runtime/control naming split to preserve.
  - `docs/local-control-api-v1.md:1-42` - local-only operator contract.
  - `docs/v0.1/phase2/scaffold-boundaries.md:10-19,31-45` - boundary expectations for runtime ownership.

  **Acceptance Criteria** (agent-executable only):
  - [ ] Runtime startup is driven through the new p2p bootstrap path, not the HTTP peer runtime.
  - [ ] `--mode`, `--data-dir`, `--listen`, and `--control` still validate and produce deterministic startup errors.
  - [ ] The local control endpoint remains reachable only on loopback/socket with bearer auth.
  - [ ] If the transport boundary package does not exist yet, the task creates the boundary package before wiring runtime startup.
  - [ ] Startup-mode mapping (`client|relay|bootstrap|archivist`) and on-wire role/capability advertising are both covered by tests.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Start in relay mode
    Tool: Bash
    Steps: run the runtime in relay mode with a temp data dir and inspect readiness output.
    Expected: process starts, prints ready state, and exposes the local control endpoint.
    Evidence: .sisyphus/evidence/task-1-relay-start.log

  Scenario: Reject invalid startup config
    Tool: Bash
    Steps: run with an empty data dir or invalid mode.
    Expected: startup fails before serving any peer traffic, with a deterministic error.
    Evidence: .sisyphus/evidence/task-1-invalid-startup.log
  ```

  **Commit**: YES | Message: `feat(node): rework backend bootstrap` | Files: `cmd/aether/main.go`, `pkg/node/*`

- [x] 2. Implement p2p transport and protocol negotiation

  **What to do**: Build the peer transport core on the protocol registry/capability layer: negotiate multistream IDs, validate feature flags, regenerate protobuf artifacts when `proto/aether.proto` changes, and move the peer-serving surface onto the target p2p transport. Keep protocol/version handling additive and explicit.

  **Must NOT do**: Bypass `pkg/protocol`; hard-code protocol strings in handlers; invent new wire rules outside the registry.

  **Recommended Agent Profile**:
  - Category: `ultrabrain` - Reason: this is the hardest protocol/transport boundary in the plan.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - not a trivial patch.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T4, T5, T6, T8 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/protocol/registry.go:10-273` - canonical multistream registry, version negotiation, and deprecation guard.
  - `pkg/protocol/capabilities.go:8-218` - feature flags, capability negotiation, and security-mode negotiation.
  - `proto/aether.proto:7-167` - protobuf enum surface and reserved-number discipline.
  - `docs/v0.1/phase1/protocol-constraints.md:8-26` - additive-only evolution and no required fields.
  - `aether-v3.md:39-47,101-145` - protobuf/compile-time and reproducibility expectations.

  **Acceptance Criteria** (agent-executable only):
  - [ ] A peer can negotiate a supported protocol ID and feature set without falling back to ad hoc strings.
  - [ ] Unsupported protocol versions/capabilities fail closed with a deterministic error.
  - [ ] Negotiation is covered by unit tests for both positive and negative cases.
  - [ ] `buf generate`/`buf lint`/`buf breaking` are part of the proto-touch workflow and generated Go is never edited directly.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Successful capability negotiation
    Tool: Bash
    Steps: run the protocol negotiation tests covering matched versions and feature flags.
    Expected: negotiation selects the expected protocol and accepted capabilities.
    Evidence: .sisyphus/evidence/task-2-negotiation-pass.log

  Scenario: Reject incompatible version
    Tool: Bash
    Steps: run the protocol negotiation tests with an unsupported major or required capability.
    Expected: negotiation fails closed with a deterministic rejection.
    Evidence: .sisyphus/evidence/task-2-negotiation-fail.log
  ```

  **Commit**: YES | Message: `feat(protocol): add p2p negotiation core` | Files: `pkg/protocol/*`, `proto/aether.proto`

- [x] 3. Replace single-file persistence with durable storage

  **What to do**: First confirm the SQLCipher-backed storage package/boundary exists (or create it as part of this task), then move runtime state off the ad hoc JSON file into the selected durable storage backend, preserving schema versioning, corruption recovery, identity backup/restore, and replay/dedup semantics. Treat the storage package/boundary check as a hard preflight step. Include migration from the current JSON state into the encrypted store, with explicit rollback/recovery behavior.

  **Must NOT do**: Persist secrets in logs; silently drop migration data; keep whole-state rewrites as the only durability mechanism.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: storage/recovery design and migration edge cases are central here.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `visual-engineering` - not a UI task.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T5, T6, T7, T8 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/node/state.go:13-200,202-260` - current persistence model, schema versioning, and corruption quarantine.
  - `docs/v0.1/phase2/p2-t10-sqlcipher-decision.md:12-80` - selected encrypted storage baseline and migration/key-management constraints.
  - `docs/v0.1/phase2/scaffold-boundaries.md:10-19,31-45` - ownership boundaries around storage/package seams.
  - `aether-v3.md:101-145` - reproducible build and configuration expectations.

  **Acceptance Criteria** (agent-executable only):
  - [ ] The storage package/boundary exists before migration logic is attempted.
  - [ ] If the storage boundary package does not exist yet, the task creates the boundary package before migration work starts.
  - [ ] Restarting the runtime preserves identity, peers, servers, messages, and relay queues.
  - [ ] Legacy JSON state migrates into the new storage backend without losing records or breaking recovery.
  - [ ] Partial migration failure leaves either the old state intact or a quarantined backup that can be recovered deterministically.
  - [ ] Corrupt storage is quarantined and replaced deterministically.
  - [ ] Wrong-key / unreadable-storage failure is explicit and test-covered.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Persistence survives restart
    Tool: Bash
    Steps: create a server and message, stop the runtime, restart it from the same data dir, and inspect state.
    Expected: state is restored without loss of the created records.
    Evidence: .sisyphus/evidence/task-3-restart-persistence.log

  Scenario: Corrupt state is quarantined
    Tool: Bash
    Steps: corrupt the storage file and start the runtime.
    Expected: the corrupt payload is quarantined and a fresh state is created or recovery fails explicitly.
    Evidence: .sisyphus/evidence/task-3-corrupt-state.log
  ```

  **Commit**: YES | Message: `feat(storage): add durable backend for node state` | Files: `pkg/node/state.go`, storage backend files

- [x] 4. Implement discovery, bootstrap, and peer exchange

  **What to do**: Replace the current HTTP polling discovery path with the target peer discovery model: local peer cache → LAN discovery → bootstrap list/DNS → DHT walking → rendezvous → peer exchange → manual peers, with resilient reconnect/backoff behavior. Preserve the layered discovery order from the roadmap while making it run on the new runtime.

  **Must NOT do**: Hard-code a single bootstrap path; introduce public discovery/indexers in this plan; make discovery depend on client UI.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: discovery is a distributed-systems boundary with many failure modes.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - not a small local change.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T5, T6, T8 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/node/service.go:1427-1562,1850-1879` - current discovery loop, bootstrap targets, relay targets, and peer fetch logic.
  - `docs/v0.1/phase1/scope-contract.md:3-23` - v0.1 discovery/relay scope and explicit non-goals.
  - `aether-addendum-qol-discovery.md:42-69,153-196` - layered discovery and optional public discovery context.
  - `docs/v0.1/phase9/p9-t4-relay-config.md:5-34` - Dhall-shaped bootstrap/relay config expectations.

  **Acceptance Criteria** (agent-executable only):
  - [ ] The discovery cascade covers the layered order from the roadmap, or any intentionally deferred layer is explicitly documented in-plan.
  - [ ] Manual peers, bootstrap nodes, and peer exchange all function in the new runtime.
  - [ ] Discovery handles unavailable peers without crashing and preserves known-peer state.
  - [ ] Discovery behavior is covered by deterministic tests and at least one multi-node integration scenario.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Manual peer and bootstrap discovery
    Tool: Bash
    Steps: start two nodes, add a manual peer or bootstrap address, and verify peer exchange.
    Expected: both nodes learn each other and persist the relationship.
    Evidence: .sisyphus/evidence/task-4-discovery-pass.log

  Scenario: Bootstrap unavailable
    Tool: Bash
    Steps: start a node with an offline bootstrap target.
    Expected: the node stays up, retries/backoffs, and preserves existing local state.
    Evidence: .sisyphus/evidence/task-4-discovery-backoff.log
  ```

  **Commit**: YES | Message: `feat(node): implement peer discovery path` | Files: `pkg/node/*`

- [x] 5. Move manifest, join, and delivery flows onto the p2p transport

  **What to do**: Replace the current peer-facing HTTP endpoints with the protocol-native message flow for server manifests, previews, joins, and deliveries. Preserve signed invites/manifests/deliveries and ensure the runtime still validates signatures, hashes, and membership changes.

  **Must NOT do**: Leave peer traffic dependent on `/_xorein/*` as the production path; weaken signature verification; alter field numbers or payload semantics without versioning.

  **Recommended Agent Profile**:
  - Category: `ultrabrain` - Reason: this is the main protocol-serving behavior.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - too much wire/protocol impact.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: T8 | Blocked By: T1, T2, T3, T4

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/node/service.go:421-631,1038-1137,1624-1808,1882-2125` - current manifest, join, delivery, relay, and peer handler behavior.
  - `pkg/node/wire.go:14-85` - current signed wire structs and history semantics.
  - `proto/aether.proto:25-167` - payload types, security modes, and verification enums.
  - `pkg/protocol/registry.go:41-170,251-273` - version negotiation rules to preserve.
  - `pkg/protocol/capabilities.go:14-218` - capability and security-mode negotiation.

  **Acceptance Criteria** (agent-executable only):
  - [ ] Server creation emits a signed manifest/invite that peers can preview and join over the new transport.
  - [ ] Delivery application verifies signatures, rejects mismatches, and persists accepted events.
  - [ ] The current HTTP peer endpoints are not required for the production path.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Preview and join a server
    Tool: Bash
    Steps: create a server on one node, preview the invite from another, then join it.
    Expected: preview succeeds, join succeeds, and both nodes converge on the server state.
    Evidence: .sisyphus/evidence/task-5-join-flow.log

  Scenario: Reject invalid delivery signature
    Tool: Bash
    Steps: mutate a signed delivery and send it through the peer transport.
    Expected: the runtime rejects it and does not persist the mutation.
    Evidence: .sisyphus/evidence/task-5-invalid-delivery.log
  ```

  **Commit**: YES | Message: `feat(node): move server flows onto p2p transport` | Files: `pkg/node/*`, protocol runtime files

- [x] 6. Implement relay, archive, and history semantics

  **What to do**: Make relay/store-and-forward and history retention real in the new runtime. Enforce ciphertext-only relay behavior, quotas/TTL, history window labels, and archiving semantics that align with the role/capability model.

  **Must NOT do**: Expose decrypted media to relays; treat local history as replicated archive; allow relay queues to grow without bound.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: this is a durability and policy surface, not just a handler change.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `visual-engineering` - not a UI task.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: T8 | Blocked By: T1, T2, T3, T4

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/node/service.go:1542-1562,1664-1808,2511-2578` - relay drain/store, delivery application, and history pruning.
  - `pkg/node/wire.go:19-24` - local-window vs single-node history semantics.
  - `docs/v0.1/phase9/p9-t4-relay-config.md:5-34` - relay and bootstrap configuration model.
  - `docs/v0.1/phase11/p11-t5-relay-operator-quickstart.md:5-67` - relay MVP and deferred scope boundaries.

  **Acceptance Criteria** (agent-executable only):
  - [ ] Relay nodes store only the intended forwardable payloads and drain them to the right peers.
  - [ ] History retention is bounded and labels match the manifest safety labels.
  - [ ] Relay/history behavior is exercised by tests and a restart scenario.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Store and drain relay traffic
    Tool: Bash
    Steps: send a delivery through a relay node, then drain it from the recipient.
    Expected: relay stores then returns the queued delivery, and the recipient applies it exactly once.
    Evidence: .sisyphus/evidence/task-6-relay-drain.log

  Scenario: Enforce history window
    Tool: Bash
    Steps: generate messages beyond the configured limit.
    Expected: oldest messages are pruned and the history labels remain consistent.
    Evidence: .sisyphus/evidence/task-6-history-prune.log
  ```

  **Commit**: YES | Message: `feat(node): add relay and history semantics` | Files: `pkg/node/*`

- [x] 7. Preserve and align the local control API

  **What to do**: Keep the documented local operator API working on the new runtime, including token auth, local-only enforcement, SSE events, and the documented `/v1` surface. Make sure the control plane reports the new runtime state accurately.

  **Must NOT do**: Open the control API to remote callers; break the `/v1` surface without a migration; make it dependent on the peer transport being HTTP.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: this is an operational surface with security boundaries.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - more than a local tweak.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: T8 | Blocked By: T1, T3, T5, T6

  **References** (executor has NO interview context - be exhaustive):
  - `pkg/node/control.go:19-166` - local-only auth, SSE, and /v1 routing.
  - `docs/local-control-api-v1.md:1-42` - operator contract and error model.
  - `docs/release-naming.md:3-10` - runtime/control naming split.
  - `pkg/node/service.go:270-303,305-367,743-965,881-963` - snapshot, identity, notifications, and events backing the control surface.

  **Acceptance Criteria** (agent-executable only):
  - [ ] The local-control surface remains aligned with the documented `/v1` contract only.
  - [ ] The local control API remains reachable on loopback/socket only and rejects remote access.
  - [ ] `/v1/state` and `/v1/events` reflect the new runtime correctly.
  - [ ] Identity backup/restore and server/message controls work against the new backend.
  - [ ] Contract tests cover `/v1/state`, `/v1/events`, token auth, and loopback-only enforcement.
  - [ ] A dedicated regression test covers remote control rejection and bearer-token auth failure.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Local operator access
    Tool: Bash
    Steps: authenticate with the control token and query `/v1/state` and `/v1/events`.
    Expected: requests succeed locally and emit SSE updates.
    Evidence: .sisyphus/evidence/task-7-control-local.log

  Scenario: Remote control rejection
    Tool: Bash
    Steps: attempt to hit the control API from a non-loopback address or with a wrong token.
    Expected: request is rejected with the documented error code.
    Evidence: .sisyphus/evidence/task-7-control-remote.log
  ```

  **Commit**: YES | Message: `feat(node): preserve local control API` | Files: `pkg/node/control.go`, `pkg/node/service.go`

- [x] 8. Realign verification, build, and regression gates

  **What to do**: Replace placeholder verification assumptions with real backend gates: make the build/test pipeline exercise the new runtime, add backend integration tests, enforce race and negative-path coverage, and wire in Buf/pre-commit checks where protocol assets are touched. The real gates should be the concrete commands `go test ./...`, `go test -race ./...`, `pre-commit run --all-files`, Podman-wrapped `buf lint`, Podman-wrapped `buf breaking`, and the release-pack verification flow; `make build` must become a thin wrapper around the real Go build/packaging step that emits the runnable `bin/aether` binary and any release artifacts.

  **Must NOT do**: Leave `make compile`, `make lint`, or `make build` as pretend gates; let protocol changes bypass Buf checks; rely on manual verification only.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: this is a repo-wide verification and tooling task.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `visual-engineering` - not UI-related.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: none | Blocked By: T1-T7

  **References** (executor has NO interview context - be exhaustive):
  - `Makefile:14-115` - current stage order, placeholder build/compile/lint, and real test/scan hooks.
  - `scripts/dhall-verify.sh:4-64` - Podman-based config verification.
  - `scripts/release-pack-verify.sh:4-149` - release verification prerequisites and signature workflow.
  - `.golangci.yml:1-16` - current lint policy.
  - `.pre-commit-config.yaml:1-12` - repo hygiene gates.
  - `aether-v3.md:51-98,149-215` - testing/CI expectations and reproducibility goals.

  **Acceptance Criteria** (agent-executable only):
  - [ ] The backend-specific test suite fails on the intended negative cases and passes on the happy path.
  - [ ] Repository gates run the real commands needed to prove the backend (including race/integration/Buf when applicable).
  - [ ] Release-pack verification still works after the backend changes.
  - [ ] `make build` is a real gate, not a placeholder stub.
  - [ ] The exact repo gates are `go test ./...`, `go test -race ./...`, `pre-commit run --all-files`, Podman-wrapped `buf lint`, Podman-wrapped `buf breaking`, `make check-full`, `make pipeline`, and `make release-pack-verify`.
  - [ ] `make check-full` and `make pipeline` are rewritten or explicitly wrapped around the real commands above; they must not remain placeholder gates.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Full backend verification
    Tool: Bash
    Steps: run the repo’s real verification commands for tests, race checks, pre-commit, and release-pack validation.
    Expected: all gates pass on a clean backend implementation.
    Evidence: .sisyphus/evidence/task-8-full-verify.log

  Scenario: Verification failure on regression
    Tool: Bash
    Steps: break a protocol or persistence invariant and rerun the focused gate.
    Expected: the gate fails for the right reason and reports a useful error.
    Evidence: .sisyphus/evidence/task-8-regression.log
  ```

  **Commit**: YES | Message: `test(repo): make backend verification real` | Files: `Makefile`, `scripts/*`, tests

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> This is a post-implementation handoff gate: automated verification runs first, then the user reviews the consolidated result before completion.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — unspecified-high
- [x] F3. Real Manual QA — unspecified-high (+ playwright if UI)
- [x] F4. Scope Fidelity Check — deep

## Commit Strategy
- Use one commit per task once its task-specific tests and QA scenario pass.
- Keep commits small and task-shaped; do not bundle unrelated runtime, storage, or verification changes.
- No docs-only or placeholder commits.

## Success Criteria
- The backend starts as a protocol-native p2p server, not the current HTTP peer runtime.
- The local control API still works locally and is rejected remotely.
- Peers can discover each other, create/join servers, exchange deliveries, and recover state after restart.
- Relay/history behavior and protocol negotiation are covered by tests.
- The repo’s verification path is real enough to prove the backend, not just the scaffolding.
- Blob-provider support stays deferred and is not counted as part of backend success for this plan.
