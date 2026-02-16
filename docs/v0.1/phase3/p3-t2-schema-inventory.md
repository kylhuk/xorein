# P3-T2 — v0.1 Schema Inventory & Reservation Plan

## Purpose

This document enumerates the v0.1 protocol messages introduced in
[`proto/aether.proto`](proto/aether.proto:1) for P3-T2. It captures allocated field
numbers, reserved ranges, and forward-compatibility notes so downstream teams can
plan additive evolutions safely.

## Message catalog

| Message | Fields (name = number) | Reserved ranges | Notes |
| --- | --- | --- | --- |
| `CapabilitySet` | `flags = 1`, `custom_claims = 2` | `100-199` | `flags` uses `CapabilityFlag` enum; `custom_claims` stores optional bag for future dims. |
| `IdentityProfile` | `identity_id = 1`, `display_name = 2`, `signing_public_key = 3`, `profile_signature = 4`, `advertised_capabilities = 5` | `100-199` | Signature covers serialized profile for tamper-detection. |
| `MembershipState` | `member = 1`, `chat_enabled = 2`, `voice_enabled = 3`, `granted_capabilities = 4` | `100-199` | Additional toggles or reason codes can land in reserved band. |
| `ServerManifest` | `manifest_id = 1`, `owner = 2`, `server_description = 3`, `members = 4`, `server_capabilities = 5`, `published_at = 6` | `100-199` | `members` is repeated to allow unordered sets; `published_at` is unix seconds. |
| `ChatMessage` | `message_id = 1`, `sender = 2`, `room_id = 3`, `body = 4`, `sent_at = 5` | `100-199` | Future attachments can use reserved slots without renumbering. |
| `VoiceState` | `session_id = 1`, `participant = 2`, `muted = 3`, `deafened = 4`, `updated_at = 5` | `100-199` | Session extensions (codec, bitrate, etc.) fit into reserved block. |
| `SignedEnvelope` | `envelope_id = 1`, `payload_type = 2`, `payload = 3`, `signer = 4`, `signature_algorithm = 5`, `signature = 6`, `canonical_payload = 7`, `signed_at = 8` | `100-199` | `canonical_payload` pins the serialized bytes used for signing. |
| `VerificationError` | `code = 1`, `detail = 2` | `100-199` | Additional structured diagnostics can use reserved slots. |
| `EnvelopeVerification` | `envelope = 1`, `status = 2`, `errors = 3`, `canonical_bytes = 4` | `100-199` | `canonical_bytes` mirrors the deterministic representation returned to clients. |
| `AetherPlaceholder` | `placeholder_message = 1` | `100`, name `legacy_port` | Legacy scaffold kept for downgrade protection. |

### Enum catalog

All enums mirror the same reservation policy: business-use values live below 32,
`100-199` is reserved for future compatibility shims.

- `CapabilityFlag`: chat/voice/management flags; room for richer capabilities.
- `PayloadType`: identity, manifest, chat, voice payload classifiers.
- `SignatureAlgorithm`: currently ED25519 | P256; more algorithms can be added.
- `EnvelopeVerificationError`: canonical taxonomy for P3-T4 failure reasons.
- `VerificationStatus`: accepted vs rejected outcomes.

## Field numbering discipline

1. **No renumbering.** Any future removals must add `reserved` statements as
   mandated by [`docs/v0.1/phase2/proto-reservation-policy.md`](docs/v0.1/phase2/proto-reservation-policy.md:1).
2. **Growth bands.** Slots `100-199` remain intentionally unused so we can land
   feature-specifc metadata without touching core numbering. Higher bands (>=200)
   remain unallocated for now to minimize surface area.
3. **Payload typing.** `PayloadType` is the sole authority for dispatching the
   opaque `payload` bytes inside `SignedEnvelope`. New payload messages must add a
   typed enum entry before shipping.

## Compatibility considerations

- Unknown fields are tolerated across all messages, enabling staged rollouts.
- `CapabilitySet.custom_claims` uses a string map to host experimentation without
  schema churn.
- Envelope-related messages purposefully embed `IdentityProfile` structures so a
  verifier can operate without remote lookups.

This inventory, paired with the envelope spec in
[`docs/v0.1/phase3/p3-t4-envelope-spec.md`](docs/v0.1/phase3/p3-t4-envelope-spec.md:1),
completes the documented prerequisites for P3-T2.
