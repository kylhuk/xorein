# v0.3 Phase 6 - P6-T1 Integrated Scenario Pack

> Status: Execution artifact. Scenario pack now links every scope bullet to documented positive/adverse/recovery evidence.

## Purpose

Collect integrated scenario descriptions that hit every scope bullet (S3-01..S3-17) across positive, adverse, degraded, and recovery paths to enable traceable V3-G6 validation.

## Scenario Coverage Summary

| Scenario | Scope Buckets Covered | Evidence Reference |
|---|---|---|
| Voice quality under packet loss | S3-01..S3-06 | `pkg/v03/voice/contracts_test.go`, `docs/v0.3/phase1/p1-voice-sfu-baseline.md` |
| Screen share capture + viewer control failure | S3-07..S3-11 | `pkg/v03/screenshare/contracts_test.go`, `docs/v0.3/phase2/p2-screen-share-baseline.md` |
| File transfer resume + preview fallback | S3-12..S3-13 | `pkg/v03/transfer/contracts_test.go`, `docs/v0.3/phase3/p3-file-media-baseline.md` |
| Discovery + explore + invite flow with stale data | S3-14..S3-16 | `pkg/v03/discovery/contracts_test.go`, `docs/v0.3/phase4/p4-discovery-admission-baseline.md` |
| Indexer signed response + de-dup verification | S3-17 | `pkg/v03/indexer/contracts_test.go`, `docs/v0.3/phase5/p5-indexer-verification-baseline.md` |

## Evidence Pattern

1. Each scenario includes (a) positive proof, (b) adverse/failure proof, (c) recovery guidance, and (d) corresponding `pkg/v03` test file.
2. Raw command outputs (`go test`, `buf lint`, `pkg/v03` scenario logs) are archived alongside this doc via `docs/v0.3/phase6/p6-t2-conformance-review.md`.
3. Scenario pack references OD3 open decisions where relevant (e.g., TTL, ranking, privacy posture) without resolving them.
