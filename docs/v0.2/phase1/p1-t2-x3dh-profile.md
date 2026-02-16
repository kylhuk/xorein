# v0.2 Phase 1 - P1-T2 X3DH Profile

> Status: Planning artifact only. No implementation completion is claimed.

## Purpose

Define the v0.2 X3DH contract for 1:1 DM session bootstrap so key lifecycle and failure handling are deterministic.

## Source Trace

- `TODO_v02.md:276`
- `aether-v3.md:82`

## Key Material Model and Prekey Classes (P1-T2-ST1)

Required key classes per identity:

1. Identity Key (IK)
   - Long-lived identity signing keypair.
   - Rotation only through explicit identity migration procedure.
2. Signed PreKey (SPK)
   - Medium-lived keypair signed by IK.
   - Used for asynchronous bootstrap stability.
3. One-Time PreKey (OPK)
   - Single-use keypairs published in inventory.
   - Consumed atomically on successful bootstrap.

Lifecycle rules:

- SPK generation includes IK signature and validity window metadata.
- SPK replacement must overlap with previous SPK validity to avoid bootstrap gaps.
- OPK inventory has minimum threshold; republish trigger starts before exhaustion.
- Expired SPK/OPK entries are invalid for handshake and must return deterministic failure reasons.

## Handshake Sequence (P1-T2-ST2)

Planned sequence:

1. Initiator fetches recipient prekey bundle from DHT.
2. Initiator verifies IK signature on SPK and bundle freshness.
3. Initiator builds initial message using IK + ephemeral key + SPK (+ OPK if present).
4. Recipient validates bundle references and message authenticity.
5. Recipient acknowledges accepted bootstrap and binds session identifiers.
6. Both peers transition to Double Ratchet session initialization.

## Deterministic Failure Taxonomy

| Code | Condition | Required behavior |
|---|---|---|
| `X3DH_BUNDLE_NOT_FOUND` | No retrievable bundle | Retry policy and user-visible next action returned |
| `X3DH_BUNDLE_EXPIRED` | SPK or bundle metadata expired | Reject handshake and request refreshed bundle |
| `X3DH_SPK_SIG_INVALID` | SPK signature fails IK verification | Hard fail; treat as authenticity violation |
| `X3DH_OPK_DEPLETED` | No OPK available when required | Fallback behavior follows profile policy, with explicit reason |
| `X3DH_REPLAY_DETECTED` | Nonce/session replay attempt | Reject and log replay class deterministically |
| `X3DH_STALE_REFERENCE` | Handshake references stale key IDs | Refetch bundle and retry once within bounded window |

## Replay and Stale-Key Handling

- Bootstrap message IDs must be unique within bounded replay window.
- Previously consumed OPKs cannot be accepted again.
- Bundle freshness checks require publish time + expiry validation.
- Retry path must not silently downgrade cryptographic guarantees.

## Verification Targets

- Positive: valid bundle handshake succeeds and transitions into ratchet init.
- Negative: invalid SPK signature, expired bundle, replayed message are rejected with taxonomy above.
- Recovery: stale reference and depleted inventory paths produce deterministic retry or fallback actions.
