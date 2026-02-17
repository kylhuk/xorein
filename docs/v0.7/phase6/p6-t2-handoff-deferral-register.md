# Phase 6 · P6-T2 Handoff & Deferral Register

## Objective
Document the operator-facing handoff package, highlight the evidence suite for each path (`cmd/aether`, `cmd/aether-push-relay`, `tests/e2e/v07/`, `containers/v07/`), and catalog deferred items mapped to future bands.

## Contract
- Include runbook pointers, verification anchors, and explicit readiness tags for every artifact that supports Phase 1–5 execution so neighboring teams know what to run or verify.
- Deferrals are grouped by target band (v0.8+, v0.9+, v1.0+, post-v1) with rationale and cross-reference to the open decisions that keep them unstable.
- The register links to `docs/v0.7/phase6/p6-t11-operator-runbooks-upgrade-notes.md` and `docs/v0.7/phase6/p6-t12-build-run-validation.md` for procedural context.

This doc keeps the handoff story cohesive by pairing every deliverable with runbook/infrastructure guidance and by surfacing the deferred scope that remains planning-only.

## Runbook links
- `cmd/aether` mode coverage is documented in [docs/v0.7/phase6/p6-t12-build-run-validation.md](./p6-t12-build-run-validation.md).
- Push relay operations are captured in [`docs/v0.7/runbooks/push-relay.md`](../runbooks/push-relay.md) and `docs/v0.7/runbooks/relay.md` anchors telemetry/reservation concerns.
- The v0.7 README summarizes the runnable demo plus the single-command e2e run (`docs/v0.7/README.md`).

## Ship-blocking gaps (if manual operator flow is required)
- **V7-BLOCK-CLIENT-RUNTIME:** `cmd/aether --mode=client` remains a deterministic scaffold and does not yet execute a true online/offline session lifecycle.  
  - **Impact:** Addendum section E manual proof steps (A/B live join-send-reconnect with history replay) are only proven through `tests/e2e/v07` harness coverage, not through a stateful client runtime.
  - **Follow-up task:** implement v0.7 runtime glue between `cmd/aether` client mode and `pkg/v07/storeforward`, `pkg/v07/historysync`, and `pkg/v07/search` (target: `V7-FU-CLIENT-RUNTIME-01`).
- **V7-BLOCK-SQLCIPHER-FTS5:** v0.7 search currently uses a deterministic in-memory index rather than an on-disk SQLCipher FTS5 engine.
  - **Impact:** filter and migration semantics are deterministic and test-covered, but not executed against a live SQLCipher backend.
  - **Follow-up task:** add SQLCipher-backed migrations + FTS5 query execution path with parity tests against current `pkg/v07/search` behavior (target: `V7-FU-SEARCH-SQLCIPHER-01`).

## Deferred scope

### v0.8+
- **Proof inner-chain anchoring:** deterministic Merkle proof auto-verification (beyond the current stubbed `proof.error` encoding) remains deferred pending alignment with the Protocol verification service release (open decision `V7-P2-ProofIntegration`).
- **Multi-tenant relay isolation:** per-tenant push registration isolation and quota tracking are postponed until the v0.8 QoS gating requirements settle (see `V7-P4-RelayQoS`).

### v0.9+
- **Search semantics audit:** fine-grained text scoring/text-body heuristics will be covered in v0.9 after the v0.7 deterministic filters settle; the current implementation sticks to substring matches.

### v1.0+
- **History delta streaming:** streaming history chunks with partial replay proofs is deferred to v1.0+ once the governance-approved retransmission conductor is available (see `V7-H1-HistoryStream`).

### Post-v1
- **Cross-version compatibility tests** (v1+ vs v0.7) are tracked separately in the release risk log and are not part of this handoff package.
