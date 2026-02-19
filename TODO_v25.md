# TODO v25 - Xorein (backend) Ciphertext Blob Store + Asset Distribution (v2.5)

## Status
Planning artifact only. This file defines v25 implementation and validation requirements. It does not claim implementation completion.

## Naming + architecture note (carry-forward)
- **Xorein**: backend node/runtime binary (`cmd/xorein`) implementing blob encryption, upload/download, replication, and verification.
- **harmolyn**: frontend UI binary (`cmd/harmolyn`) showing upload/download UX, progress, and failure reasons via the local API.
- v25 extends the “durable data plane” beyond message history segments to **non-message assets** (attachments, avatars, emojis, etc.).

## Version Isolation Contract (mandatory)
- v25 cannot close unless all v25 in-scope spec, code, tests, docs, ops artifacts, and evidence are complete.
- v25 implements only `F25` from v24 specs (additive proto only).
- v25 must also publish full `F26` spec package for v26 final closure.

## Version role
- Implements: `F25` (ciphertext blob store plane with replication, quotas, and anti-enumeration for private Spaces).
- Specifies: `F26` (system-wide release closure and “DONE” conformance gates for the full stack after v20+ architecture changes).

## Critical scope (v25)
- Implement ciphertext blob model in Xorein:
  - `BlobRef` and manifest model (content hash, size, mime, chunking)
  - encryption envelope:
    - per-blob random DEK (AEAD)
    - DEK wrapped/enveloped under Space key (Space assets) or per-DM key (DM assets)
    - deterministic verification (hash + AEAD tag validation)
  - chunking + resumable transfer for large assets
- Implement storage provider behavior:
  - default: reuse Archivist capability as blob storage provider (ciphertext-only, quota/retention enforced)
  - replication factor `r_blob` (default) with health tracking and repair under churn
  - deterministic refusal reasons (quota exceeded, retention policy, unsupported mime, too large)
- Implement retrieval behavior:
  - anti-enumeration for private Spaces:
    - do not allow probing blob existence without membership-derived secrets
  - bounded resource usage (rate limits, max concurrency, max memory)
  - client-side caching + eviction policy (mobile-safe)
- Integrate assets into product surfaces:
  - message attachments (upload → send message referencing BlobRef)
  - user avatar and Space icon as BlobRef-backed assets
  - custom emojis as BlobRef-backed assets (if feature exists)
- Implement harmolyn UX:
  - upload progress, retry, cancel, failure reasons
  - download lazy-fetch + explicit “tap to download” on constrained networks
  - offline states and coverage labeling for asset availability
- Preserve existing invariants:
  - relays still must not host durable blob data (Archivist/blob host is the durable capability)
  - no keyword leakage by default remains unchanged

## Out of scope (defer)
- Token/incentive economics for blob hosting is out of scope.
- “Public CDN mode” where operators host plaintext is out of scope.
- Remote keyword search remains out of scope.

## Dependencies and relationships
- Inputs from v24:
  - `docs/v2.4/phase4/f25-blob-store-spec.md`
  - `docs/v2.4/phase4/f25-proto-delta.md`
  - `docs/v2.4/phase4/f25-acceptance-matrix.md`
- Outputs consumed by v26:
  - `docs/v2.5/phase4/f26-final-closure-spec.md`
  - `docs/v2.5/phase4/f26-proto-delta.md` (may be empty; must exist)
  - `docs/v2.5/phase4/f26-acceptance-matrix.md`

## Entry criteria (must be true before implementation starts)
- `v24` is in `promoted` state with evidence bundle and as-built conformance report.
- Required `F25` spec inputs from v24 exist and are approved.
- v25 deferred scope list is frozen and approved.

## Promotion gates (must all pass)
- `G0` Scope lock and requirement-to-artifact traceability complete.
- `G1` Additive wire compatibility checks pass (no breaking changes).
- `G2` Blob encryption + manifest model complete (correctness + verification).
- `G3` Blob upload/download + chunking/resume complete (bounded resource usage).
- `G4` Provider replication + quota/retention + refusal reasons complete.
- `G5` Private Space anti-enumeration + privacy tests complete.
- `G6` harmolyn UX and local API surfaces complete (no-limbo asset states).
- `G7` Podman blob scenarios complete (multi-provider, churn, partial failure).
- `G8` `F26` final closure spec package complete.
- `G9` Docs and evidence bundle complete.
- `G10` Relay no-durable-blob-hosting regression checks pass.

## Mandatory command evidence (attach exact outputs in Phase 5)
- `buf lint`
- `buf breaking`
- `go test ./...`
- `go test ./tests/e2e/v25/...` (or declared equivalent if paths differ)
- `go test ./tests/perf/v25/...` (or declared equivalent if paths differ)
- `go build ./cmd/xorein`
- `go build ./cmd/harmolyn`
- `make check-full`
- `scripts/v25-blob-scenarios.sh`

## Roadmap conformance templates (mandatory)
- Gate checklist schema: `docs/templates/roadmap-gate-checklist.md`.
- Gate ownership/sign-off RACI: `docs/templates/roadmap-signoff-raci.md`.
- Evidence index schema and ID format (`EV-v25-GX-###`): `docs/templates/roadmap-evidence-index.md`.
- Deferral register schema: `docs/templates/roadmap-deferral-register.md`.

## LLM agent execution contract (mandatory)
- `P0-T1`: Dependency rule—start only after entry criteria (v24 promoted, `F25` specs approved, deferred scope frozen); Acceptance rule—complete ST1–ST4, freeze supported data classes, and mark `G0` Pass; Evidence rule—attach at least one `EV-v25-G0-###` entry plus command outputs or `not applicable`; Blocker taxonomy—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint rule—record the planned command (e.g., `make check-full`) before running and capture output.
- `P1-T1`: Dependency rule—after `G0` and `G1` compatibility checks; Acceptance rule—complete ST1–ST2 of BlobRef schema work and mark `G2` Pass; Evidence rule—≥1 `EV-v25-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `go test ./pkg/v25/blobref/...`) and capture output.
- `P1-T2`: Dependency rule—after `P1-T1`; Acceptance rule—complete ST1–ST3 encryption envelope work and mark `G2` Pass; Evidence rule—attach `EV-v25-G2-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—log the next command (e.g., `go test ./pkg/v25/blobcrypto/...`) and capture output.
- `P2-T1`: Dependency rule—after `G2`; Acceptance rule—complete ST1–ST3 provider endpoints, resumable transfer, refusal reasons, and mark `G3` Pass; Evidence rule—≥1 `EV-v25-G3-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the planned command (e.g., `go test ./pkg/v25/blobproto/...`, `tests/e2e/v25/blob_transfer_*`) and capture outputs.
- `P2-T2`: Dependency rule—after `P2-T1`; Acceptance rule—complete ST1–ST3 replication/repair policy and mark `G4` Pass; Evidence rule—attach `EV-v25-G4-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `tests/e2e/v25/replication_*`) and capture output.
- `P2-T3`: Dependency rule—after `P2-T2`; Acceptance rule—complete ST1–ST2 relay boundary enforcement and Podman scenario spec and mark `G10` Pass; Evidence rule—≥1 `EV-v25-G10-###` entry plus outputs (e.g., `tests/e2e/v25/relay_boundary_*`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the next command (`tests/e2e/v25/relay_boundary_*`) and capture output.
- `P3-T1`: Dependency rule—after `G3` and `G4`; Acceptance rule—complete ST1–ST3 messaging attachments work and mark `G5` Pass; Evidence rule—attach `EV-v25-G5-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `tests/e2e/v25/attachments_*`) and record output.
- `P3-T2`: Dependency rule—after `P3-T1`; Acceptance rule—complete ST1–ST3 asset integration and mark `G5` Pass; Evidence rule—≥1 `EV-v25-G5-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the command (e.g., `tests/e2e/v25/assets_*`) and capture output.
- `P3-T3`: Dependency rule—after `P3-T2`; Acceptance rule—complete ST1–ST3 harmolyn asset UX work and mark `G6` Pass; Evidence rule—attach `EV-v25-G6-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—note the planned command (e.g., `pkg/v25/harmolyn/ui` tests) and capture output.
- `P4-T1`: Dependency rule—after `G6`; Acceptance rule—complete ST1–ST3 for `F26` spec package and mark `G8` Pass; Evidence rule—attach `EV-v25-G8-###` entry plus outputs or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record the command (spec/publication steps) and capture output.
- `P5-T1`: Dependency rule—after all prior gates (G0–G8) and artifacts; Acceptance rule—complete ST1–ST3 (evidence bundle, conformance report, sign-offs) and mark `G9` Pass (with relay regression `G10` satisfied); Evidence rule—attach `EV-v25-G9-###` and `EV-v25-G10-###` entries plus command outputs (`go build`, `go test`, `make check-full`, `scripts/v25-blob-scenarios.sh`) or `not applicable`; Blockers—`BLOCKED:INPUT`, `BLOCKED:GATE`, `BLOCKED:EVIDENCE`, `BLOCKED:ENV`; Command hint—record planned commands (`go build ./cmd/xorein`, `go build ./cmd/harmolyn`, `make check-full`, scenario scripts) and capture outputs.

### Phase dependency and evidence map
| Phase | Top-level tasks | Must start after | Gate target(s) | Minimum evidence | Command hints |
| --- | --- | --- | --- | --- | --- |
| Phase 0 | `P0-T1` | Entry criteria (v24 promoted + `F25` specs approved) | `G0` | `EV-v25-G0-###` | Note scope commands (e.g., `make check-full`) before running and capture outputs. |
| Phase 1 | `P1-T1`, `P1-T2` | `G0`, `G1` completion | `G1`, `G2` | `EV-v25-G1-###`, `EV-v25-G2-###` | Capture blob schema/encryption test outputs (`go test ./pkg/v25/blobref`, `go test ./pkg/v25/blobcrypto`) before running. |
| Phase 2 | `P2-T1`, `P2-T2`, `P2-T3` | `G2` completion | `G3`, `G4`, `G10` | `EV-v25-G3-###`, `EV-v25-G4-###`, `EV-v25-G10-###` | Record provider/replication/boundary command outputs (`go test ./pkg/v25/blobproto`, `tests/e2e/v25/replication_*`, `tests/e2e/v25/relay_boundary_*`) before running and capture outputs. |
| Phase 3 | `P3-T1`, `P3-T2`, `P3-T3` | `G3`–`G4` completion | `G5`, `G6` | `EV-v25-G5-###`, `EV-v25-G6-###` | Plan attachment/asset/UX commands (`tests/e2e/v25/attachments_*`, `tests/e2e/v25/assets_*`, UI tests) and capture outputs. |
| Phase 4 | `P4-T1` | `G5`, `G6` completion | `G8` | `EV-v25-G8-###` | Record spec/publication commands before running and log outputs. |
| Phase 5 | `P5-T1` | Prior gates (G0–G8) | `G9`, `G10` | `EV-v25-G9-###`, `EV-v25-G10-###` | Capture final evidence commands (`go build`, `go test`, `make check-full`, `scripts/v25-blob-scenarios.sh`) before running and record outputs. |

## Phase plan

### Phase 0 - Scope lock and data-class mapping (G0)
- [x] `P0-T1` Freeze v25 blob store scope and data-class coverage.
  - `ST1` Import v24 `F25` acceptance matrix; convert to explicit go/no-go checks.
  - `ST2` Produce requirement-to-artifact traceability matrix.
  - `ST3` Freeze supported mime types and size tiers (desktop vs mobile defaults).
  - `ST4` Confirm which data classes are BlobRef-backed in v25:
    - attachments (must)
    - avatars (must)
    - Space icons (must)
    - emojis/stickers (if already present; otherwise define as v25-ready but defer UX)
  - Artifacts:
    - `docs/v2.5/phase0/p0-scope-lock.md`
    - `docs/v2.5/phase0/p0-traceability-matrix.md`
    - `docs/v2.5/phase0/p0-asset-coverage-map.md`

- [x] `P0-T2` Document proto compatibility guardrails for v25 blob store/asset flows.
  - `ST1` Enumerate the new messages/fields that must land in the v25 proto delta (chunk/manifest endpoints, asset references, capability enforcement metadata).
  - `ST2` Capture the additive-only checklist (no renumbering, no type changes, no required fields, reserve removed numbers) so future tweaks stay backward-compatible.
  - `ST3` Clarify downgrade/read-old expectations (relay/clients may see mixed-version blobs, assets, or blob metadata) with explicit guidance for old clients and roll-over reads.
  - Artifact: `docs/v2.5/phase0/p0-proto-compat-checklist.md`.

### Phase 1 - Blob model + encryption (G2)
- [x] `P1-T1` Implement BlobRef + manifest schema.
  - `ST1` Define canonical `BlobRef`:
    - `hashAlgorithm` (predefined algorithms such as BLAKE3 or SHA-256)
    - `contentHash`
    - `size`
    - `mimeType`
    - `chunkSize`
    - `chunkProfile`
    - optional `encryptedMetadataPointer`
  - `ST2` Implement manifest verification rules that surface deterministic refusal codes for chunk, size, and digest mismatches.
  - `ST3` Allow forward-extensible metadata sections that round-trip unknown keys while leaving validation untouched.
  - Artifacts:
    - `pkg/v25/blobref/*`
    - `pkg/v25/blobref/*_test.go`
    - `docs/v2.5/phase1/p1-blobref-spec.md`

- [x] `P1-T2` Implement encryption envelope.
  - `ST1` Per-blob DEK generation + AEAD encryption.
  - `ST2` DEK wrapping:
    - Space assets: wrap under Space asset key
    - DM assets: wrap under DM session key
  - `ST3` Implement key rotation hooks (no forced rotation in v25; just hooks and docs).
  - Artifacts:
    - `pkg/v25/blobcrypto/*`
    - `pkg/v25/blobcrypto/*_test.go`
    - `docs/v2.5/phase1/p1-blobcrypto-profile.md`

### Phase 2 - Provider protocol + replication (G3, G4, G10)
- [x] `P2-T1` Implement provider endpoints and bounded transfer.
  - `ST1` Add proto endpoints for:
    - `PutBlobChunk`, `GetBlobChunk`
    - `PutManifest`, `GetManifest`
    - quota/retention query (optional)
  - `ST2` Implement resumable transfer:
    - chunk presence query (privacy-safe; must not allow probing in private Spaces)
    - retry/backoff and concurrency bounds
  - `ST3` Implement deterministic refusal reasons.
  - Artifacts:
    - `pkg/v25/blobproto/*`
    - `tests/e2e/v25/blob_transfer_*`

- [x] `P2-T2` Implement replication and repair policy.
  - `ST1` Publish to `r_blob` providers; record replica set in local metadata.
  - `ST2` Verify replica health; repair under churn.
  - `ST3` Define and enforce provider selection policy:
    - prefer local/nearby providers where known
    - avoid single-AS concentration if such metadata exists (optional)
  - Artifacts:
    - `pkg/v25/replication/*`
    - `docs/v2.5/phase2/p2-replication-policy.md`
    - `tests/e2e/v25/replication_*`

- [x] `P2-T3` Enforce relay no-durable-blob-hosting boundary.
  - `ST1` Relay rejects `PutManifest`/`PutBlobChunk` uploads with deterministic refusal errors.
  - `ST2` Relay retains only manifest pointers/tokens; payload bytes never persist.
  - `ST3` Unauthorized private-space lookups return indistinguishable not-found errors for existing and missing blobs.
  - Artifacts:
    - `tests/e2e/v25/relay_blob_boundary_test.go`
    - `docs/v2.5/phase2/p2-relay-boundary-report.md`

### Phase 3 - Product integration (G5, G6)
- [x] `P3-T1` Integrate attachments into messaging flows.
  - `ST1` Send path: encrypt+upload blob(s) → send message with BlobRef(s).
  - `ST2` Receive path: lazy download with verification; deterministic failure reasons.
  - `ST3` Redaction behavior:
    - redaction removes plaintext references locally
    - does not guarantee removal from third-party caches; document explicitly
  - Artifacts:
    - `pkg/v25/attachments/*`
    - `tests/e2e/v25/attachments_*`
    - `docs/v2.5/phase3/p3-attachments-contract.md`

- [x] `P3-T2` Integrate avatars/icons/emojis as BlobRef-backed assets.
  - `ST1` User avatar set/update path + caching rules.
  - `ST2` Space icon set/update path + caching rules.
  - `ST3` Emoji asset fetch rules (if present).
  - Artifacts:
    - `pkg/v25/assets/*`
    - `tests/e2e/v25/assets_*`
    - `docs/v2.5/phase3/p3-assets-contract.md`
    - `docs/v2.5/phase3/p3-bridge-asset-policy.md`
    - `tests/e2e/v25/bridge_asset_policy_test.go`

- [x] `P3-T3` harmolyn asset UX.
  - `ST1` Upload/download progress and cancellation.
  - `ST2` “Tap to download” / “download on Wi‑Fi only” toggles (if present).
  - `ST3` Offline badges and deterministic error surfaces.
  - Artifacts:
    - `pkg/v25/harmolyn/ui/*`
    - `docs/v2.5/phase3/p3-asset-ux-contract.md`

### Phase 4 - v26 final closure spec package (G8)
- [x] `P4-T1` Produce `F26` final closure specification package.
  - `ST1` Define system-wide “DONE” criteria (no open core-feature deferrals).
  - `ST2` Define full-stack regression matrix requirements (identity → messaging → media → moderation → discovery → persistence → blob store).
  - `ST3` Define release packaging and reproducible build requirements for all binaries.
  - Artifacts:
    - `docs/v2.5/phase4/f26-final-closure-spec.md` (planning-only system specification)
    - `docs/v2.5/phase4/f26-proto-delta.md` (additive proto delta plan)
    - `docs/v2.5/phase4/f26-acceptance-matrix.md` (acceptance/evidence anchors)

### Phase 5 - Closure and evidence (G9)
- [x] `P5-T1` Publish v25 evidence bundle and promotion recommendation.
  - `ST1` Attach all command outputs and scenario manifests.
  - `ST2` Publish `F25` as-built conformance report against v24 `F25` package.
  - `ST3` Publish gate sign-off sheet and evidence index using templates.
  - Artifacts:
    - `docs/v2.5/phase5/p5-evidence-bundle.md`
    - `docs/v2.5/phase5/p5-risk-register.md`
    - `docs/v2.5/phase5/p5-as-built-conformance.md`
    - `docs/v2.5/phase5/p5-gate-signoff.md`
    - `docs/v2.5/phase5/p5-evidence-index.md`

## Risk register (v25)
| ID | Risk | Mitigation | Exit criterion |
|---|---|---|---|
| R25-1 | Blob plane enables enumeration in private Spaces | membership-derived keying + negative tests | anti-enumeration tests pass |
| R25-2 | Storage/transfer DoS | quotas + rate limits + bounded responses | abuse tests pass |
| R25-3 | Large assets break mobile UX | lazy fetch + size tiers + eviction | mobile-tier tests pass |
| R25-4 | Relay boundary regression | explicit regression probes + Podman checks | relay-boundary tests pass |
| R25-5 | Integrity regressions (tampered chunks) | hash+AEAD verification + fuzz tests | integrity tests pass |

## Decision log (v25)
- `D25-1`: Blob data is always stored as ciphertext; providers never see plaintext.
- `D25-2`: Providers are capability-based (Archivist reuse by default); no new privileged role is introduced.
- `D25-3`: Private Space blob retrieval must not leak existence to non-members (anti-enumeration invariant).
- `D25-4`: harmolyn must surface deterministic asset availability states; no silent partial loads.
