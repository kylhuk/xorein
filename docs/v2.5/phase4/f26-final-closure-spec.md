# F26 Final Closure Specification

Planning-only coverage for the `F26` final closure package that v25 publishes to prepare v26 for final-gate execution. This document captures the `G8` scope and evidence requirements without implying the v26 runtime is already delivered.

## ST1 – System-wide “DONE” criteria

| Criterion | Description | Gate | Evidence placeholder |
| --- | --- | --- | --- |
| Scope exhaustion | No remaining core features, integrations, or deferred core data classes exist beyond the documented deferral register; every open item either closed or carries a tracked handoff. | `G8` | `EV-v25-G8-001` |
| Traceable conformance | Requirement-to-artifact matrices (feature, proto, UX, ops) double-check that `F26` consumption contracts align with prior phase inputs. | `G8` | `EV-v25-G8-002` |
| Evidence readiness | Regression suites, reproducible packaging, and release notes are linked and scheduled for handoff to forensics/QA. | `G8` | `EV-v25-G8-003` |

Each criterion is gated by `G8`, so the placeholders above capture the future `EV-v25-G8-###` entries that will document the handoff readiness of `F26`.

## ST2 – Full-stack regression matrix requirements

The regression matrix ensures every major pillar is exercised before v26 signs off on final conformance:

| Pillar | Regression focus | Description | Evidence placeholder |
| --- | --- | --- | --- |
| Identity & access | Service identity & Space membership stability | Verify identity metadata (keys, tokens) survives release packaging and that Space membership checks still gate anti-enumeration surfaces. | `EV-v25-G8-004` |
| Messaging & attachments | Blob references traveling through messaging surfaces | Ensure messages with BlobRef payloads preserve integrity, redaction, and playback across relays. | `EV-v25-G8-005` |
| Media & persistence | Blob storage + compressing streaming | Confirm chunked blob transfers, resumable downloads, and quota enforcement continue to operate after packaging. | `EV-v25-G8-006` |
| Moderation & guards | Deterministic refusal reasons + relay boundary | Replay refusal semantics, relay upload rejection, and anti-enumeration guards under instrumentation. | `EV-v25-G8-007` |
| Discovery & operations | Deployment readiness, reproducible packaging | Validate that deployment artifacts (binaries, container images, release docs) are reproducible and instrumented. | `EV-v25-G8-008` |

Additional rows can be appended as `F26` planning matures, but the matrix above anchors the minimum coverage for the final closure gate.

## ST3 – Release packaging and reproducible build requirements

- **Reproducible binaries**: `cmd/xorein`, `cmd/harmolyn`, and any helper daemons must build with fixed toolchain versions (documented in `docs/v2.5/phase5/p5-release-notes.md`) and exhibit byte-for-byte outputs when hashed. Evidence will cite `EV-v25-G8-009` once the build proof is captured.
- **Artifact catalog**: Every artifact produced for `F26` (archives, release notes, evidence indexes, gate checklists) must appear in the release catalog with SHA-256 metadata and be referenced by `docs/templates/roadmap-evidence-index.md` entries; placeholder `EV-v25-G8-010` tracks that catalog.
- **Publication checklist**: Runbook steps (code freeze, proto delta review, final regressions, documentation sign-offs) are outlined and annotated with owners; the checklist itself will gain an `EV-v25-G8-011` entry when completed.

These release requirements complete the `F26` gating story for v25, locking the consumer-facing closure language that drives v26 execution.
