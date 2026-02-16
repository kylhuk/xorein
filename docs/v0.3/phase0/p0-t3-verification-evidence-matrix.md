# v0.3 Phase 0 - P0-T3 Verification Evidence Matrix

> Status: Execution artifact. Verification expectations for every in-scope intent now map to deterministic evidence anchors and `pkg/v03` tests.

## Purpose

Predefine the evidence bucket for each v0.3 scope item so completion claims remain verifiable and bound to code/tests, documentation, and governance artifacts.

## Evidence Matrix (one positive + one negative/adverse + optional recovery proof per scope)

| Scope ID | Positive evidence | Negative/adverse evidence | Recovery evidence |
|---|---|---|---|
| S3-01 | `pkg/v03/voice/contracts_test.go` noise-adaptation vector | `pkg/v03/voice/contracts_test.go` invalid input handling | `pkg/v03/voice/contracts_test.go` fallback handoff scenario |
| S3-02 | `pkg/v03/voice/contracts_test.go` bitrate ladder test | `pkg/v03/voice/contracts_test.go` ceiling/ground detection | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` ABR recovery table |
| S3-03 | `pkg/v03/voice/contracts_test.go` buffer sizing test | `pkg/v03/voice/contracts_test.go` late-arrival/overflow classification | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` jitter-recovery narrative |
| S3-04 | `pkg/v03/voice/contracts_test.go` parity success case | `pkg/v03/voice/contracts_test.go` burst-loss detection | `docs/v0.3/phase6/p6-t1-integrated-scenario-pack.md` voice recovery path |
| S3-05 | `pkg/v03/voice/contracts_test.go` 9+ participant handoff | `pkg/v03/voice/contracts_test.go` tie-break conflict path | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` election fallback narrative |
| S3-06 | `pkg/v03/voice/contracts_test.go` relay enable path | `pkg/v03/voice/contracts_test.go` disabled-mode rejection | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` relay transition table |
| S3-07 | `pkg/v03/screenshare/contracts_test.go` multi-platform capture paths | `pkg/v03/screenshare/contracts_test.go` capture permission denial | `docs/v0.3/phase2/p2-screen-share-baseline.md` capture restart guidance |
| S3-08 | `pkg/v03/screenshare/contracts_test.go` capability-match tests | `pkg/v03/screenshare/contracts_test.go` unsupported-hardware rejection | `docs/v0.3/phase2/p2-screen-share-baseline.md` encoder downgrade path |
| S3-09 | `pkg/v03/screenshare/contracts_test.go` preset-to-parameter matrix | `pkg/v03/screenshare/contracts_test.go` degraded preset fallback | `docs/v0.3/phase2/p2-screen-share-baseline.md` preset recovery notes |
| S3-10 | `pkg/v03/screenshare/contracts_test.go` layer handshake | `pkg/v03/screenshare/contracts_test.go` layer failure diagnostics | `docs/v0.3/phase2/p2-screen-share-baseline.md` simulcast failover table |
| S3-11 | `pkg/v03/screenshare/contracts_test.go` control-state transitions | `pkg/v03/screenshare/contracts_test.go` control conflict resolution | `docs/v0.3/phase2/p2-screen-share-baseline.md` viewer recovery guidance |
| S3-12 | `pkg/v03/transfer/contracts_test.go` full transfer success | `pkg/v03/transfer/contracts_test.go` partial failure detection | `docs/v0.3/phase3/p3-file-media-baseline.md` resume/retry table |
| S3-13 | `pkg/v03/transfer/contracts_test.go` inline rendering coverage | `pkg/v03/transfer/contracts_test.go` metadata rejection | `docs/v0.3/phase3/p3-file-media-baseline.md` clear vs E2EE disclosure note |
| S3-14 | `pkg/v03/discovery/contracts_test.go` publish/retrieve path | `pkg/v03/discovery/contracts_test.go` stale entry handling | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` recovery guidance |
| S3-15 | `pkg/v03/discovery/contracts_test.go` browse success vector | `pkg/v03/discovery/contracts_test.go` stale preview signal | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` preview recovery table |
| S3-16 | `pkg/v03/discovery/contracts_test.go` invite lifecycle success | `pkg/v03/discovery/contracts_test.go` request rejection cases | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` policy fallback guidance |
| S3-17 | `pkg/v03/indexer/contracts_test.go` signed-response acceptance | `pkg/v03/indexer/contracts_test.go` invalid signature rejection | `docs/v0.3/phase5/p5-indexer-verification-baseline.md` de-dup/retry guidance |

## Evidence Requirements

1. Every verification entry above must point to both a doc (phase baseline or release) and a corresponding `pkg/v03` test file before V3-G6 can close.
2. Evidence packages include raw command outputs (`buf lint`, `go test`, `go build`, etc.) stored under phase6 docs (`p6-t1`, `p6-t2`, `p6-t3`).
