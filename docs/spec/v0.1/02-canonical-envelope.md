# 02 — Canonical Envelope

This document specifies the serialization, signing, and verification rules for
every signed data structure in the Xorein protocol. Two envelope formats are
used:

- **Native protobuf peer-stream envelope** — `PeerStreamRequest` /
  `PeerStreamResponse` — the canonical framing for all peer-to-peer libp2p
  stream operations.
- **Protobuf `SignedEnvelope`** — used for signed binary payloads inside stream
  operations and the conformance API.
- **JSON signed objects** — `Manifest`, `Invite`, `Delivery` — used for the
  local HTTP control API and stored documents. These are also transported
  as the `payload` bytes inside stream envelopes.

## 1. Peer-stream envelope (primary wire format)

Every libp2p stream between two Xorein nodes uses the following framing:

```
[4-byte big-endian uint32 payload-length]
[proto.Marshal(PeerStreamRequest or PeerStreamResponse)]
```

Max payload length: **8 388 608 bytes** (8 MiB). A stream that advertises a
larger length MUST be rejected immediately.

### 1.1 `PeerStreamRequest`

Defined in `proto/aether.proto` as `message PeerStreamRequest`.

| Field | Type | Notes |
|-------|------|-------|
| `operation` | string | e.g. `"peer.info"`, `"chat.send"`, `"dm.send"` |
| `payload` | bytes | protobuf-encoded, operation-specific |
| `advertised_caps` | []string | capability flags this node supports |
| `required_caps` | []string | capability flags required from responder |
| `protocol_id` | string | `/aether/<family>/<major>.<minor>.<patch>` |
| `security_mode` | SecurityMode | active mode for this operation |
| `request_id` | string | UUID v4; idempotency nonce |

### 1.2 `PeerStreamResponse`

Defined in `proto/aether.proto` as `message PeerStreamResponse`.

| Field | Type | Notes |
|-------|------|-------|
| `negotiated_protocol` | string | agreed `/aether/<family>/<version>` |
| `accepted_caps` | []string | intersection of both sides' capabilities |
| `ignored_caps` | []string | remote caps the responder does not support |
| `payload` | bytes | protobuf-encoded, operation-specific |
| `error` | PeerStreamError | populated on negotiation or operation failure |
| `request_id` | string | mirrors `PeerStreamRequest.request_id` |

### 1.3 `PeerStreamError` codes

Defined in `proto/aether.proto` as `message PeerStreamError`.

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Remote cannot satisfy a required capability |
| `UNSUPPORTED_VERSION` | No mutually acceptable protocol version |
| `UNSUPPORTED_OPERATION` | Operation string not recognized |
| `SIGNATURE_MISMATCH` | Payload signature failed verification |
| `MODE_INCOMPATIBLE` | Requested security mode rejected by policy |
| `RELAY_OPACITY_VIOLATION` | Relay received non-encrypted payload for E2EE scope |
| `OPERATION_FAILED` | Logical operation error (details in `error.message`) |
| `RATE_LIMITED` | Sender exceeded rate limit; back off per `error.message` |

When the `error` field is populated, the `payload` field MUST be empty. When
there is no error, the `error` field MUST be absent (zero-value proto field).

## 2. Protobuf `SignedEnvelope`

Defined in `proto/aether.proto` as `message SignedEnvelope`. Used as the
payload inside `PeerStreamRequest.payload` / `PeerStreamResponse.payload`
when the operation carries a signed binary artifact (e.g., identity announce,
manifest broadcast, prekey publication).

### 2.1 Canonical payload construction

```
canonical_payload =
    payload_type_varint              // tag+value of the payload_type field
    || deterministic_proto(payload)  // proto.Marshal with deterministic option
    || uint64_BE(signed_at)          // 8 bytes, big-endian, unix milliseconds
```

The concatenation is reproducible from the envelope fields alone. A verifier
MUST recompute it from the parsed fields and compare byte-for-byte against the
`canonical_payload` field.

### 2.2 Signing (PQ-hybrid)

```
signature           = Ed25519.Sign(ed_private, canonical_payload)
ml_dsa_65_signature = ML-DSA-65.Sign(mldsa_private, canonical_payload)
signature_algorithm = SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65
```

Both signatures MUST be present. A classical-only node (transitional period
only) MAY use `SIGNATURE_ALGORITHM_ED25519 = 1` with only `signature` set;
PQ-hybrid nodes MUST reject classical-only signatures once PQ support is
negotiated (see §2.3).

### 2.3 Verification procedure

1. **Parse.** Deserialize the envelope. Reject if required fields (`payload`,
   `signer`, `signature`, `canonical_payload`) are missing → `UNSIGNED_PAYLOAD`.
2. **Algorithm.** Confirm `signature_algorithm`.
   - `HYBRID_ED25519_ML_DSA_65`: both `signature` and `ml_dsa_65_signature`
     MUST be present; proceed to step 3 for both.
   - `ED25519`: single `signature` only; ML-DSA-65 check skipped (classical path).
   - Unknown algorithm → `UNSUPPORTED_PAYLOAD_TYPE`.
3. **Canonical replay.** Recompute `canonical_payload` from parsed fields.
   Compare byte-for-byte. Mismatch → `CANONICALIZATION_MISMATCH`.
4. **Signature(s).** Verify `signature` (Ed25519) over `canonical_payload`
   using `signer.signing_public_key`. If HYBRID: also verify `ml_dsa_65_signature`
   (ML-DSA-65) over `canonical_payload` using `signer.ml_dsa_65_public_key`.
   Either failure → `SIGNATURE_MISMATCH`.
5. **Freshness.** Check `signed_at` (unix milliseconds) against acceptance
   window: ±5 minutes for stream operations, 7 days for stored manifests.
   → `EXPIRED_SIGNATURE`.
6. **Authorization.** Confirm signer is trusted for the payload type
   (e.g., manifest signer matches `owner_peer_id`). → `UNTRUSTED_SIGNER`.
7. **Payload type.** Confirm `payload_type` is known. → `UNSUPPORTED_PAYLOAD_TYPE`.

### 2.4 Verification error codes

| Code | Trigger |
|------|---------|
| `SIGNATURE_MISMATCH` | Ed25519 or ML-DSA-65 check failed |
| `UNSIGNED_PAYLOAD` | Required signature fields absent |
| `UNSUPPORTED_PAYLOAD_TYPE` | Unknown `payload_type` or `signature_algorithm` |
| `CANONICALIZATION_MISMATCH` | Recomputed canonical payload differs from stored |
| `EXPIRED_SIGNATURE` | `signed_at` outside acceptance window |
| `UNTRUSTED_SIGNER` | Signer not trusted for this payload type |

All verification responses MUST include `status` (`ACCEPTED` / `REJECTED`),
an ordered list of `VerificationError` entries, and `canonical_bytes` (the
recomputed payload, for archival diagnostics).

## 3. JSON signed objects

`Manifest`, `Invite`, and `Delivery` are signed JSON objects used in:
- The local HTTP control API (request/response bodies).
- Peer stream payloads (JSON-encoded, then embedded as `PeerStreamRequest.payload` bytes).
- Archivist and relay queue storage.

The JSON-signed layer uses the hybrid signature scheme from
`01-cryptographic-primitives.md §6`.

### 3.1 `Manifest` canonical form

```json
{
  "server_id": "...",
  "name": "...",
  "description": "...",
  "owner_peer_id": "...",
  "owner_public_key": "...",
  "owner_addresses": [ "sorted", "lexicographic" ],
  "bootstrap_addrs": [ "sorted" ],
  "relay_addrs": [ "sorted" ],
  "capabilities": [ "sorted" ],
  "security_mode": "...",
  "offered_security_modes": [ "preference-ordered" ],
  "history_retention_messages": 0,
  "history_coverage": "...",
  "history_durability": "...",
  "issued_at": "RFC3339Nano",
  "updated_at": "RFC3339Nano"
}
```

Rules:
- The `signature` field MUST be absent from the canonical form used for signing.
- The slice fields `owner_addresses`, `bootstrap_addrs`, `relay_addrs`, and
  `capabilities` MUST be sorted (lexicographic ascending) before signing.
- `offered_security_modes` is NOT sorted; order is preserved as given.

```
canonical_json = UTF-8(JSON(canonical_manifest))  // keys sorted, no extra whitespace
combined_sig   = Ed25519.Sign(ed_priv, canonical_json)
              || ML-DSA-65.Sign(mldsa_priv, canonical_json)
signature      = base64url_no_padding(combined_sig)
```

### 3.2 `Invite` canonical form

```json
{
  "server_id": "...",
  "owner_peer_id": "...",
  "owner_public_key": "...",
  "server_addrs": [ "sorted" ],
  "bootstrap_addrs": [ "sorted" ],
  "relay_addrs": [ "sorted" ],
  "manifest_hash": "...",
  "expires_at": "RFC3339Nano"
}
```

Signed with the same identity keys as the manifest. `manifest_hash` binds the
invite to a specific manifest version.

### 3.3 `Delivery` canonical form

```json
{
  "id": "...",
  "kind": "...",
  "scope_id": "...",
  "scope_type": "...",
  "server_id": "...",
  "sender_peer_id": "...",
  "sender_public_key": "...",
  "recipient_peer_ids": [ "..." ],
  "body": "...",
  "created_at": "RFC3339Nano"
}
```

For encrypted deliveries (Seal/Tree/Crowd/Channel modes), `body` contains the
AEAD ciphertext encoded as base64url with no padding. Plaintext bodies are only
valid for Clear mode and MUST be rejected by relays (§3.4).

The `signature`, `data`, and `muted` fields are excluded from the canonical form.

### 3.4 Delivery verification

When a node receives a `Delivery`:

1. Decode `sender_public_key` (base64url combined hybrid public key or legacy Ed25519 key).
2. Reconstruct canonical JSON from the received fields.
3. Verify the combined signature per `01-cryptographic-primitives.md §6.1`.
4. Reject if `sender_peer_id` does not correspond to `sender_public_key`.
5. Reject if the delivery is a replay (duplicate `id` for the same scope).
6. For relay operations: reject if the scope mode is non-Clear and the body
   is not encrypted (`RELAY_OPACITY_VIOLATION`).

### 3.5 JSON key sorting

JSON objects within canonical forms MUST serialize keys in lexicographic
(byte-level, UTF-8) order. This applies recursively to nested objects.
Implementations MUST NOT rely on the order of unknown fields — unknown
fields are ignored at parse time and excluded from canonical form.

## 4. Manifest hash

`Manifest.Hash()` computes:

```
hash = base64url_raw(SHA-256(canonical_manifest_JSON))[0:32]
```

This is a content-addressable 32-character prefix used in invites and for
cache invalidation. The full hybrid signature provides integrity.

## 5. Additive evolution rules

- New fields added to a signed object MUST use `omitempty` JSON tags.
- Fields present at signing MUST be present at verification.
- New fields added after signing are excluded from canonical form and are
  therefore ignored by verifiers (Go `json.Unmarshal` ignores unknown fields).
- Field removal is a breaking change; field numbers/names MUST be reserved.
- New proto fields in `PeerStreamRequest`/`PeerStreamResponse` MUST use
  field numbers outside the `reserved 100 to 199` range and must be
  additive (no semantic changes to existing fields).
