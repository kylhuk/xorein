# Verify Repo Go-Client Reachability

## TL;DR
> **Summary**: Prove the repo’s Go client can reach and communicate with the current server using the live libp2p runtime, then package the result as a scoped verdict: repo Go client yes, external interoperability not claimed.
> **Deliverables**: static protocol audit, separate-process smoke harness, live client/server exchange evidence, negative negotiation evidence, final verdict artifact.
> **Effort**: Short
> **Parallel**: YES - 2 waves
> **Critical Path**: Static protocol audit → subprocess smoke harness → live smoke execution → verdict gate

## Context
### Original Request
Please check the repository and verify that any client supporting the protocol is now able to reach and communicate with this server here.

### Interview Summary
- The live entrypoint is `cmd/aether/main.go`; it boots `network.NewP2PRuntime(...)` and injects it into `node.NewService(...)`.
- `pkg/network/runtime.go` starts a libp2p host and registers handlers for canonical peer protocol IDs.
- `pkg/network/transport.go` negotiates `/aether/peer/0.1.0` and peer capability flags before dispatch.
- The user wants the verification scoped to the repo’s Go client/runtime path, with both static review and live smoke evidence.

### Metis Review (gaps addressed)
- Keep the claim narrow: prove repo Go-client reachability only.
- Require a separate-process smoke check; in-process tests alone are not enough.
- Avoid cross-implementation interoperability promises unless a dedicated external harness exists.

## Work Objectives
### Core Objective
Verify that the repo’s Go client can connect to the current server, negotiate the canonical peer protocol, and complete at least one real request/response exchange.

### Deliverables
- Static audit of protocol IDs, capabilities, and runtime wiring
- Separate-process smoke harness or equivalent live-process test
- Positive and negative negotiation evidence
- Final verdict artifact with the claim scoped to the repo Go client

### Definition of Done (verifiable conditions with commands)
- `go test ./pkg/protocol ./pkg/network ./pkg/node` passes.
- A live server process starts and emits a ready signal.
- The repo Go client completes one successful peer operation against the live server.
- Unsupported protocol/capability negotiation still fails closed.
- Final verdict says “repo Go client reachability verified” and does not claim external interoperability.

### Must Have
- Canonical protocol family/version stays `/aether/peer/0.1.0` for the verified path.
- The live smoke uses the real runtime, not a stub.
- Evidence is captured as logs/artifacts under `.sisyphus/evidence/`.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No transport redesign.
- No UI or docs-only work.
- No claim that arbitrary external clients interoperate unless separately proven.
- No reliance on in-process tests as the only proof of reachability.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after + live smoke
- QA policy: every task has agent-executed happy-path and failure-path scenarios
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`

## Execution Strategy
### Parallel Execution Waves
Wave 1: static protocol audit, separate-process smoke harness
Wave 2: live smoke execution and verdict packaging

### Dependency Matrix (full, all tasks)
- T1 → T3, F1-F4
- T2 → T3, F1-F4
- T3 → F1-F4

### Agent Dispatch Summary (wave → task count → categories)
- Wave 1 → 2 tasks → deep, deep
- Wave 2 → 1 task → unspecified-high
- Final verification → 4 review agents → oracle, unspecified-high, unspecified-high, deep

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Static protocol and client-surface audit

  **What to do**: Review `cmd/aether/main.go`, `pkg/network/runtime.go`, `pkg/network/transport.go`, `pkg/protocol/registry.go`, `pkg/protocol/capabilities.go`, `pkg/protocol/transport.go`, `pkg/network/runtime_test.go`, and `README.md` to confirm the repo’s Go client and server share the same canonical protocol ID, feature flags, and negotiated error paths. Capture the exact call path used by the client to reach the live server.

  **Must NOT do**: Expand the claim to external-client interoperability; invent new wire formats; rely on docs alone.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: the task needs exact protocol/wire inspection and precise claim boundaries.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - too easy to miss protocol details.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T3, F1-F4 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `cmd/aether/main.go:53-180` - real entrypoint, runtime wiring, and ready output.
  - `pkg/network/runtime.go:79-186` - libp2p host startup, handler registration, and negotiated response handling.
  - `pkg/network/transport.go:120-189,579-595` - repo Go client dial/stream path and canonical peer protocol IDs.
  - `pkg/protocol/registry.go:10-276` - canonical `/aether` registry, parsing, and negotiation.
  - `pkg/protocol/capabilities.go:8-104,158-261` - canonical feature-flag sets and peer transport flags.
  - `pkg/protocol/transport.go:32-94` - peer transport negotiation and fail-closed errors.
  - `pkg/network/runtime_test.go:35-240` - live runtime, negotiation success, and negative-path tests.
  - `README.md:65-75` - stated compatibility model and breaking-change rules.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `go test ./pkg/protocol ./pkg/network ./pkg/node` passes.
  - [ ] The audit confirms the verified client/server path uses `/aether/peer/0.1.0` and `cap.peer.*` negotiation.
  - [ ] Unsupported protocol and unsupported capability cases are already executable or explicitly added to the test surface.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Confirm canonical protocol path
    Tool: Bash
    Steps: run `go test ./pkg/protocol ./pkg/network ./pkg/node -run 'TestNegotiatePeerTransport|TestP2PRuntimeNegotiatesPeerTransport'`
    Expected: tests pass and the negotiated protocol is `/aether/peer/0.1.0`.
    Evidence: .sisyphus/evidence/task-1-static-audit.log

  Scenario: Confirm rejection paths
    Tool: Bash
    Steps: run `go test ./pkg/network -run 'TestP2PRuntimeRejectsUnsupportedCapability|TestP2PRuntimeRejectsUnsupportedProtocolVersion'`
    Expected: tests pass by rejecting the unsupported negotiation paths with deterministic errors.
    Evidence: .sisyphus/evidence/task-1-static-audit-negatives.log
  ```

  **Commit**: NO | Message: `n/a` | Files: `n/a`

- [x] 2. Build the separate-process smoke harness

  **What to do**: Add a narrow live-process harness (preferably in `cmd/aether/main_test.go` or a sibling test file) that launches the real server binary or `go run ./cmd/aether`, waits for the ready line, captures the emitted listen address, and uses the repo Go client (`network.NewClient`) to call at least one peer operation such as `peer.info` against the live server.

  **Must NOT do**: Treat `pkg/network/runtime_test.go` as sufficient proof; use stubs instead of the real server process; add cross-implementation fixtures or redesign the transport.

  **Recommended Agent Profile**:
  - Category: `deep` - Reason: subprocess orchestration and end-to-end reachability need careful integration handling.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - the harness must exercise the real process boundary.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: T3, F1-F4 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - `cmd/aether/main.go:53-180` - server startup, ready line, and runtime wiring.
  - `cmd/aether/main_test.go:157-176` - existing runtime-start tests to extend.
  - `pkg/network/runtime_test.go:35-145` - current live Go-client success/rejection patterns.
  - `pkg/network/transport.go:109-189` - Go client request/response path.
  - `pkg/node/service_test.go:977-977` - existing `network.NewClient(...).Call(...)` usage.
  - `README.md:65-75` - the compatibility model that the smoke harness must validate.

  **Acceptance Criteria** (agent-executable only):
  - [ ] A separate-process harness exists and starts the real server process.
  - [ ] The harness waits for readiness and captures the live listen address.
  - [ ] The repo Go client completes one request/response exchange and records the negotiated protocol.
  - [ ] Unsupported negotiation still fails closed in the same or adjacent negative test.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Live server/client success
    Tool: Bash
    Steps: start the server with a temp data dir and `--listen 127.0.0.1:0`; wait for `xorein runtime ready`; run the repo Go client against the emitted listen address and issue `peer.info`.
    Expected: client receives a response, negotiated protocol equals `/aether/peer/0.1.0`, and exit code is 0.
    Evidence: .sisyphus/evidence/task-2-smoke-harness.log

  Scenario: Live negative negotiation
    Tool: Bash
    Steps: rerun the smoke harness with a request that requires an unsupported capability or protocol version.
    Expected: negotiation fails closed with `unsupported-capability` or `unsupported-protocol`.
    Evidence: .sisyphus/evidence/task-2-smoke-harness-negative.log
  ```

  **Commit**: NO | Message: `n/a` | Files: `n/a`

- [x] 3. Run live smoke and package the verdict

  **What to do**: Execute the static audit and the separate-process smoke harness, collect the evidence logs, and write the final verdict artifact. The verdict must explicitly say whether the repo Go client can reach and communicate with the server, and it must not generalize that result to external implementations.

  **Must NOT do**: Broaden the claim beyond the repo Go client; treat in-process tests as sufficient; leave the verdict unscoped.

  **Recommended Agent Profile**:
  - Category: `unspecified-high` - Reason: this task is evidence packaging and scope control rather than code discovery.
  - Skills: `[]` - no specialized skill required.
  - Omitted: `quick` - the evidence must be assembled from real runs.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: none | Blocked By: T1, T2

  **References** (executor has NO interview context - be exhaustive):
  - `cmd/aether/main.go:53-180`
  - `pkg/network/runtime_test.go:35-145`
  - `cmd/aether/main_test.go:157-176`
  - `README.md:65-75`
  - `.sisyphus/evidence/task-1-static-audit.log`
  - `.sisyphus/evidence/task-2-smoke-harness.log`
  - `.sisyphus/evidence/task-2-smoke-harness-negative.log`

  **Acceptance Criteria** (agent-executable only):
  - [ ] Live smoke output shows a running server and a successful Go-client request/response.
  - [ ] Negative-path output shows unsupported negotiation fails closed.
  - [ ] Final verdict text is explicitly scoped to the repo Go client.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Package the evidence
    Tool: Bash
    Steps: collect the static audit log and smoke logs, then write the verdict summary artifact.
    Expected: the summary states “repo Go client reachability verified” and cites the evidence files.
    Evidence: .sisyphus/evidence/task-3-verdict.md

  Scenario: Enforce scope
    Tool: Bash
    Steps: review the verdict against the collected evidence.
    Expected: any unsupported cross-implementation claim is omitted and the report stays scoped to the repo Go client.
    Evidence: .sisyphus/evidence/task-3-scope-check.log
  ```

  **Commit**: NO | Message: `n/a` | Files: `n/a`

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> Do NOT auto-proceed after verification.
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — unspecified-high
- [x] F3. Real Manual QA — unspecified-high
- [x] F4. Scope Fidelity Check — deep

  **QA Scenarios** (MANDATORY - final gate incomplete without these):
  ```
  Scenario: Parallel final review
    Tool: task
    Steps: run the four review agents in parallel against `.sisyphus/plans/protocol-reachability.md`, then collect their verdicts.
    Expected: all four reviewers approve; any rejection blocks completion until the plan is revised and re-reviewed.
    Evidence: .sisyphus/evidence/final-review.log

  Scenario: Review retry after blocker
    Tool: task
    Steps: if any reviewer reports a blocker, update the plan, rerun the four-agent review, and confirm every reviewer now approves.
    Expected: no unresolved blockers remain before handoff.
    Evidence: .sisyphus/evidence/final-review-retry.log
  ```

## Commit Strategy
- No git commit for the planning pass.
- Execution should save logs under `.sisyphus/evidence/` only.

## Success Criteria
- Static protocol and client-surface evidence is collected.
- A separate-process live smoke proves the repo Go client can connect and exchange a real peer request/response.
- Negative negotiation still fails closed.
- Final verdict is narrowly scoped to the repo Go client and does not claim broader interoperability.
