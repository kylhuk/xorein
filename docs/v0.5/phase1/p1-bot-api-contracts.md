# Phase 1 · P1 Bot API Contracts

## Purpose
Document the deterministic contracts that fulfill the Native Bot API scope, bot events, command lifecycle, and access controls referenced by Phase 1 tasks in `TODO_v05.md`.

## Deterministic obligations
- The service namespace, RPC surface, and streaming modes must match the VA-B1 surface specification; every capability referenced in Section 1.1 is tied to an explicit RPC and metadata field.
- Bot event categories, payload schemas, ordering keys, acknowledgment semantics, and replay policy follow the taxonomy described in P1-T2 and Section 1.2, meaning reconnect, delay, and delayed-replay cases resolve to deterministic outcomes documented in VA-B2.
- Command request validation, execution acknowledgment, timeout handling, idempotency, and error taxonomy are spelled out in P1-T3 so that identical inputs produce identical reason codes and lifecycles in VA-B3.
- Authorization, credential lifecycle, audit hooks, and SecurityMode gating rules bind to the v0.4 permission baseline and P1-T4 controls; VA-B4 describes what action classes need which permission checks and audit triggers.

## Deterministic contract tables
| Input / condition | Outcome + reason codes | Artifact | Validation obligation |
|---|---|---|---|
| Bot handshake with capability advertisement | Positive: `bot.connect.success` streaming session established; Negative: `bot.connect.invalid-capability` when capability mismatch; Recovery: `bot.connect.retry` after version negotiation. | VA-B1 | Capture handshake flow in the positive-path scenario pack and ensure negative/ recovery outcomes square with audit logs referenced by `VA-X1`. |
| Event ingestion and replay handling | Positive: `bot.event.delivered` with ordered ack; Negative: `bot.event.invalid-payload` when schema fields fail validation; Recovery: `bot.event.replay` with dedup and reason baggage. | VA-B2 | Maintain event taxonomy documentation and replay tests so reason codes align with scenario coverage referenced in `VA-X1`. |
| Command invocation with context validation | Positive: `bot.command.ack-success`; Negative: `bot.command.invalid-context`; Recovery: `bot.command.retry` respecting idempotency. | VA-B3 | Validate lifecycle states in command scenario pack and tie rejection/retry codes back to audit references in `VA-X1`. |
| Authorization + SecurityMode gating | Positive: `bot.auth.granted`; Negative: `bot.auth.denied` for missing scopes or invalid creds; Recovery: `bot.auth.recover` for credential refresh or audit review. | VA-B4 | Link permission matrix and audit hooks to governance traceability and ensure open-decision reviews reference these reason classes. |

These tables feed into the V5-G0 verification matrix (`docs/v0.5/phase0/p0-t3-verification-evidence-matrix.md`) and provide the deterministic input/outcome pairs for downstream reviewers.
