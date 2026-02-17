# Phase 5 · P5 Webhook Contracts

## Purpose
Clarify the deterministic semantics for incoming webhook ingestion, auth, idempotency, reliability controls, and message-surface boundaries defined in Phase 5 of `TODO_v05.md`.

## Deterministic obligations
- The POST-to-channel endpoint schema, payload validation rules, and channel routing outcomes in P5-T1 ensure invalid payloads and unauthorized channels produce deterministic rejection codes in VA-W1.
- Secret generation, storage, rotation, revocation, and failure semantics from P5-T2 deliver auditable credential transitions and reason codes for misuse in VA-W2.
- Idempotency keys, duplicate suppression windows, replay handling, and retry/failure taxonomy in P5-T3 result in deterministic duplicate outcomes and producer guidance captured in VA-W3.
- Webhook rendering boundaries for formatting, emoji shortcodes, mention handling, and permission guardrails in P5-T4 guarantee webhook-origin messages match deterministic message-surface expectations documented in VA-W4.

## Deterministic contract tables
| Input / condition | Outcome + reason codes | Artifact | Validation obligation |
|---|---|---|---|
| Webhook POST ingestion and routing | Positive: `webhook.ingest.success`; Negative: `webhook.ingest.invalid` for payload or channel errors; Recovery: `webhook.ingest.retry` after correction. | VA-W1 | Link POST schema validation documentation to reason codes and scenario evidence so ingestion paths can be audited end-to-end. |
| Webhook authentication lifecycle | Positive: `webhook.auth.success`; Negative: `webhook.auth.denied` for invalid/expired secrets; Recovery: `webhook.auth.rotate` on forced rotation. | VA-W2 | Capture secret lifecycle states and failure reason codes in authentication audits referenced by release reviewers. |
| Idempotency, replay, and retry controls | Positive: `webhook.reliability.ack`; Negative: `webhook.reliability.duplicate`; Recovery: `webhook.reliability.recover` through replay-window reset. | VA-W3 | Document idempotency/replay counters and replay-window behavior so duplicate detection/resilience evidence links to `VA-X1`. |
| Message rendering + mention guardrails | Positive: `webhook.render.success`; Negative: `webhook.render.reject` when unsupported formatting/mentions appear; Recovery: `webhook.render.fallback` to sanitized output. | VA-W4 | Map rendering boundaries and guardrails to scenario packs that prove fallback states and mention suppression reason codes. |

These contract tables round out the webhook deterministic promises and feed the V5-G5 review and Phase 0 verification traceability.
