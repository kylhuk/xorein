# Phase 0 · P0-T3 Verification Evidence Matrix

## Purpose
Define the evidence schema and pass/fail template that V5-G0 uses to certify the deterministic contracts described throughout `TODO_v05.md`.

## Contract
- Map every `VA-*` artifact mentioned in Section 6 to a preferred evidence type (documentation, deterministic helper, scenario pack, audit record) and include the gate owner plus pass/fail declaration so V5-G0 reviewers can validate each entry.
- Reuse the reason-taxonomy language from Section 1.3 so that each entry includes a `reason-class` column, and detail how positive, negative, and recovery paths are covered per artifact.
- Include a sweep matrix that links V5-G0 items to the release-handoff dossier defined in `docs/v0.5/phase6/p6-t3-release-gate-handoff.md` so reviewers can trace evidence through the full gate flow.

### Artifact matrix by scope

#### Bot API artifacts (VA-B1..VA-B4)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-B1 | P1-T1 | Native bot handshake succeeds with streaming session established (`bot.connect.success`). | Unknown capability or version mismatch triggers `bot.connect.invalid-capability`. | Client retries handshake after version negotiation (`bot.connect.retry`). | Success → validation → retry reason-class triplet keeps handshake gating deterministic. | V5-G1 owner | `docs/v0.5/phase1/p1-bot-api-contracts.md#deterministic-contract-tables` |
| VA-B2 | P1-T2 | Accepted bot events publish `bot.event.delivered` with ordered ack. | Payload validation failures emit `bot.event.invalid-payload`. | Replay or reorder path emits `bot.event.replay` with dedupe. | Event reason classes maintain ordering, validation, and recovery semantics. | V5-G1 owner | `docs/v0.5/phase1/p1-bot-api-contracts.md#deterministic-contract-tables` |
| VA-B3 | P1-T3 | Command invocation returns `bot.command.ack-success`. | Schema/context failures surface `bot.command.invalid-context`. | Duplicate/timeouts surface `bot.command.retry` after idempotency check. | Lifecycle reason-class continuity ties acknowledgement → rejection → retry. | V5-G1 owner | `docs/v0.5/phase1/p1-bot-api-contracts.md#deterministic-contract-tables` |
| VA-B4 | P1-T4 | Authorization passes with `bot.auth.granted` and audit log entry. | Denied credentials emit `bot.auth.denied` via reason taxonomy. | Credential rotation or elevated audit entry triggers `bot.auth.recover` path. | Reason classes tie auth success, denial, and recovery to SecurityMode gating. | V5-G1 owner | `docs/v0.5/phase1/p1-bot-api-contracts.md#deterministic-contract-tables` |

#### SDK and slash artifacts (VA-S1..VA-S4)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-S1 | P2-T1 | Go SDK surfaces initialization + event/command flows (`sdk.go.success`). | Surface mismatch emits `sdk.go.invalid-surface`. | Version downgrade path uses `sdk.go.retry-compat`. | Reason classes preserve interface success, validation, and compatibility recovery. | V5-G2 owner | `docs/v0.5/phase2/p2-sdk-slash-contracts.md#deterministic-contract-tables` |
| VA-S2 | P2-T2 | Community SDK profile matched and labeled (`sdk.community.success`). | Missing feature set returns `sdk.community.incomplete`. | Support-tier fallback uses `sdk.community.degraded-retry`. | Reason-link ensures conformance claim never outpaces coverage. | V5-G2 owner | `docs/v0.5/phase2/p2-sdk-slash-contracts.md#deterministic-contract-tables` |
| VA-S3 | P2-T3 | Slash command registration accepts valid catalog items (`slash.schema.success`). | Invalid schema rejects with `slash.schema.reject`. | Concurrent updates surface `slash.schema.reconcile` recovery path. | Schema, rejection, and reconciliation share reason taxonomy. | V5-G2 owner | `docs/v0.5/phase2/p2-sdk-slash-contracts.md#deterministic-contract-tables` |
| VA-S4 | P2-T4 | Autocomplete query returns ranked suggestions (`slash.auto.success`). | Timeout or format error surfaces `slash.auto.timeout`. | Fallback to stale data emits `slash.auto.fallback`. | Reason-class continuity covers success → failure → fallback. | V5-G2 owner | `docs/v0.5/phase2/p2-sdk-slash-contracts.md#deterministic-contract-tables` |

#### Discord shim artifacts (VA-D1..VA-D4)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-D1 | P3-T1 | Supported REST endpoints translate cleanly to native (`shim.rest.success`). | Unsupported field requests return `shim.rest.unsupported`. | Native migration guidance routes `shim.rest.migrate`. | Translation reason classes document success/deferral/recovery. | V5-G3 owner | `docs/v0.5/phase3/p3-discord-shim-contracts.md#deterministic-contract-tables` |
| VA-D2 | P3-T2 | Gateway events translate via intent matrix (`shim.gateway.success`). | Heartbeat/resume errors yield `shim.gateway.session-error`. | Reconnect/resume path emits `shim.gateway.resume`. | Reason classes maintain event/heartbeat lifecycle story. | V5-G3 owner | `docs/v0.5/phase3/p3-discord-shim-contracts.md#deterministic-contract-tables` |
| VA-D3 | P3-T3 | Coverage scoring hits ≥80% of canon patterns (`shim.coverage.success`). | Unsupported feature claim requires `shim.coverage.unsupported`. | Migration plan enforces `shim.coverage.rollback`. | Reason-class continuity keeps coverage, unsupported, and rollback aligned. | V5-G3 owner | `docs/v0.5/phase3/p3-discord-shim-contracts.md#deterministic-contract-tables` |
| VA-D4 | P3-T4 | Migration playbook points to native alternatives (`shim.migration.success`). | Gaps emit `shim.migration.blocked`. | Roll-forward via native API references `shim.migration.recover`. | Reason classes document migration clarity, blocking, and fallback. | V5-G3 owner | `docs/v0.5/phase3/p3-discord-shim-contracts.md#deterministic-contract-tables` |

#### Emoji & reaction artifacts (VA-E1..VA-E4)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-E1 | P4-T1 | Emoji upload + quota enforcement succeed (`emoji.upload.success`). | Invalid asset returns `emoji.upload.invalid`. | Deletion/replacement path emits `emoji.upload.recover`. | Quota reason-class ensures accepted/rejected/recover states. | V5-G4 owner | `docs/v0.5/phase4/p4-emoji-reaction-contracts.md#deterministic-contract-tables` |
| VA-E2 | P4-T2 | Picker selection + shortcode resolution produce deterministic render (`emoji.picker.success`). | Parse collision surfaces `emoji.picker.conflict`. | Fallback text path emits `emoji.picker.fallback`. | Picker reason classes keep success/conflict/fallback traceable. | V5-G4 owner | `docs/v0.5/phase4/p4-emoji-reaction-contracts.md#deterministic-contract-tables` |
| VA-E3 | P4-T3 | Reaction add/remove toggles converge (`emoji.reaction.success`). | Duplicate or stale input returns `emoji.reaction.invalid`. | Reconciliation path emits `emoji.reaction.recover`. | Reason-class continuity covers state convergence, invalidation, and reconciliation. | V5-G4 owner | `docs/v0.5/phase4/p4-emoji-reaction-contracts.md#deterministic-contract-tables` |
| VA-E4 | P4-T4 | Governance checks allow authorized actions (`emoji.gov.success`). | Unauthorized attempts emit `emoji.gov.denied`. | Audit/appeal path surfaces `emoji.gov.recover`. | Reason taxonomy documents governance success, denial, recovery. | V5-G4 owner | `docs/v0.5/phase4/p4-emoji-reaction-contracts.md#deterministic-contract-tables` |

#### Webhook artifacts (VA-W1..VA-W4)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-W1 | P5-T1 | Webhook POST accepted and routed (`webhook.ingest.success`). | Payload errors return `webhook.ingest.invalid`. | Producer resends fixed payload via `webhook.ingest.retry`. | Ingest reason classes map success → invalidation → retry. | V5-G5 owner | `docs/v0.5/phase5/p5-webhook-contracts.md#deterministic-contract-tables` |
| VA-W2 | P5-T2 | Valid webhook secret grants `webhook.auth.success`. | Invalid/expired secrets emit `webhook.auth.denied`. | Rotation path surfaces `webhook.auth.rotate`. | Auth reason classes cover grant, denial, rotation. | V5-G5 owner | `docs/v0.5/phase5/p5-webhook-contracts.md#deterministic-contract-tables` |
| VA-W3 | P5-T3 | Idempotent submission returns `webhook.reliability.ack`. | Replay or duplicate triggers `webhook.reliability.duplicate`. | Replay window expiration surfaces `webhook.reliability.recover`. | Reliability reason classes ensure duplicate detection and recovery. | V5-G5 owner | `docs/v0.5/phase5/p5-webhook-contracts.md#deterministic-contract-tables` |
| VA-W4 | P5-T4 | Message rendering honors emoji/mention constraints (`webhook.render.success`). | Disallowed mentions or formatting return `webhook.render.reject`. | Fallback to sanitized surface uses `webhook.render.fallback`. | Render reason classes link success, rejection, sanitization. | V5-G5 owner | `docs/v0.5/phase5/p5-webhook-contracts.md#deterministic-contract-tables` |

#### Integrated validation artifacts (VA-X1..VA-X3)
| Artifact | Source task(s) | Positive path | Negative path | Recovery path | Reason-class continuity | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-X1 | P6-T1 | Scenario suite covers positive flows across all scope bullets (`validation.scenario.success`). | Scenario gap emits `validation.scenario.missing`. | Recovery scenarios cover blocked workflows `validation.scenario.recover`. | Reason taxonomy links positive coverage, missing scenarios, recovery stories. | V5-G6 owner | `docs/v0.5/phase6/p6-t1-cross-feature-scenario-pack.md#deterministic-contract-pack` |
| VA-X2 | P6-T2 | Compatibility/governance reviews declare compliant (`validation.governance.success`). | Incompatible delta surfaces `validation.governance.fail`. | Open-decision review path keeps `validation.governance.escalate`. | Governance reason classes document success, failure, and escalation. | V5-G6 owner | `docs/v0.5/phase6/p6-t2-conformance-review.md#purpose` |
| VA-X3 | P6-T3 | Release checklist satisfied and evidence linked (`validation.release.success`). | Missing evidence or unresolved dependency yields `validation.release.fail`. | Deferral path surfaces `validation.release.defer`. | Reason-class continuity preserves gate conclusion, failure, deferral states. | V5-G7 owner | `docs/v0.5/phase6/p6-t3-release-gate-handoff.md#v5-g7-checklist` |

## V5-G0 pass/fail checklist
| Checklist item | Pass/fail rule | Evidence anchor |
|---|---|---|
| Bot API, SDK/slash, shim, emoji/reaction, and webhook artifacts map to VA-B*..VA-W* with positive/negative/recovery coverage | Pass when every artifact row above lists deterministic reason codes for all three paths; fail if any path lacks explicit coverage. | `docs/v0.5/phase0/p0-t3-verification-evidence-matrix.md#artifact-matrix-by-scope` |
| V5-G0 sweep tracks trace-links from each VA artifact to release handoff dossier entries | Pass when every VA row references the release-handoff dossier; fail when trace link is missing. | `docs/v0.5/phase6/p6-t3-release-gate-handoff.md#v5-g7-checklist` |
| Positive/negative/recovery scenarios are captured for every scope bullet in release dossier | Pass when release gate scenario pack cites V6 cross-feature scenarios; fail otherwise. | `docs/v0.5/phase6/p6-t1-cross-feature-scenario-pack.md#deterministic-contract-pack` |

## Trace-link rule
- Every V5-G0 checklist item must cite the release handoff dossier (`docs/v0.5/phase6/p6-t3-release-gate-handoff.md`) and cross-reference the originating VA artifact/table entry so downstream reviewers can follow the evidence chain without inference.
