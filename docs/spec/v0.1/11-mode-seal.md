# 11 — Mode: Seal (1:1 E2EE)

Seal mode provides end-to-end encryption for direct messages using hybrid X3DH
(X25519 + ML-KEM-768) for initial key establishment and the Double Ratchet for
per-message forward secrecy.

## 1. Prekey bundle

Before a Seal session can be initiated, each node MUST publish a prekey bundle
via the Identity family (`/aether/identity/0.1.0`).

### 1.1 Prekey bundle contents

| Field | Type | Notes |
|-------|------|-------|
| `identity_key_ed25519` | bytes (32) | Ed25519 public key |
| `identity_key_ml_dsa_65` | bytes (1952) | ML-DSA-65 public key |
| `signed_prekey_x25519` | bytes (32) | X25519 signed prekey public key |
| `signed_prekey_signature` | bytes (97) | hybrid sig over SPK (see §1.2) |
| `one_time_prekeys_x25519` | []bytes(32) | Up to 100 one-time X25519 prekeys |
| `ml_kem_768_pk` | bytes | ML-KEM-768 encapsulation public key; fresh per bundle |
| `ml_kem_768_pk_signature` | bytes (97) | hybrid sig over ML-KEM-768 public key |
| `published_at` | int64 | unix milliseconds |
| `expires_at` | int64 | unix milliseconds; max 7 days from published_at |
| `bundle_signature` | bytes (97) | hybrid sig over entire bundle canonical form |

### 1.2 Signed prekey signature

```
spk_canonical = b"xorein/seal/v1/spk-sign" || signed_prekey_x25519
spk_sig_ed = Ed25519.Sign(identity_ed_priv, spk_canonical)
spk_sig_mldsa = ML-DSA-65.Sign(identity_mldsa_priv, spk_canonical)
signed_prekey_signature = spk_sig_ed || spk_sig_mldsa  // 64 + 3309 = 3373 bytes
```

Wait — per §6.1 of `01-cryptographic-primitives.md`, the combined hybrid
signature is `Ed25519_sig (64 bytes) || ML-DSA-65_sig (3309 bytes)` = 3373
bytes total. The field type annotation `bytes (97)` above was an error.
Correct:

| Field | Byte length |
|-------|------------|
| Ed25519 signature alone | 64 |
| ML-DSA-65 signature alone | 3309 |
| **Hybrid combined signature** | **3373** |

All signature fields in this document that say "hybrid sig" are 3373 bytes
(Ed25519 64B concatenated with ML-DSA-65 3309B).

### 1.3 Bundle publication

Nodes MUST publish their prekey bundle to:
1. The identity family operation `identity.publish` on directly-connected peers.
2. The Kademlia DHT under key `SHA-256("xorein/prekey/" || peer_id)`.

Bundles are valid until `expires_at`. Nodes MUST rotate their bundle at least
every 7 days. One-time prekeys MUST be replenished when the count falls below 5.

## 2. Session initiation (hybrid X3DH)

### 2.1 Initiator computation

The initiator fetches the responder's prekey bundle via `identity.fetch`.
Then:

```
// Generate ephemeral keys
EK_25519    = X25519.GenerateKeypair()
// Use one OPK if available; else fall back to 3-DH
OPK_25519   = bundle.one_time_prekeys_x25519[0]  // consumed

// X25519 DH chain
IK_25519    = birational_map(identity_ed25519_key)  // §1.1 of crypto doc
SPK_25519   = bundle.signed_prekey_x25519
IKr_25519   = birational_map(bundle.identity_key_ed25519)

DH1 = X25519(IK_25519.priv,  SPK_25519)   // initiator identity × responder SPK
DH2 = X25519(EK_25519.priv,  IKr_25519)   // initiator ephemeral × responder identity
DH3 = X25519(EK_25519.priv,  SPK_25519)   // initiator ephemeral × responder SPK
DH4 = X25519(EK_25519.priv,  OPK_25519)   // initiator ephemeral × responder OPK
                                            // (omit DH4 if no OPK available)

x3dh_secret = DH1 || DH2 || DH3 || DH4

// ML-KEM-768 encapsulation
(ct_mlkem, ss_mlkem) = ML-KEM-768.Encapsulate(bundle.ml_kem_768_pk)

// Hybrid combine
hybrid_master = HKDF-SHA-256(
    IKM  = x3dh_secret || ss_mlkem,
    salt = b"",
    info = "xorein/seal/v1/hybrid-master",
    L    = 32,
)
```

### 2.2 Responder decapsulation

Responder receives the initiator's first message with:
- `EK_25519.pub` (initiator's ephemeral X25519 public key)
- `ct_mlkem` (ML-KEM-768 ciphertext)
- `OPK_index` (index of OPK used, or absent)

Responder computes:
```
ss_mlkem    = ML-KEM-768.Decapsulate(bundle.ml_kem_768_sk, ct_mlkem)
x3dh_secret = DH1 || DH2 || DH3 || DH4  // same structure as initiator
hybrid_master = HKDF-SHA-256(x3dh_secret || ss_mlkem, b"", "xorein/seal/v1/hybrid-master", 32)
```

Both parties arrive at the same `hybrid_master`. The used OPK MUST be deleted.

### 2.3 Root key derivation

```
root_key, chain_key = HKDF-SHA-256(
    IKM  = hybrid_master,
    salt = b"",
    info = "xorein/seal/v1/root-key",
    L    = 64,  // first 32 bytes = root key, last 32 = chain key
)
```

## 3. Double Ratchet messaging

### 3.1 Ratchet initialization

Both parties initialize the Double Ratchet with `root_key` and the sender's
initial DH ratchet key as follows:

- **Initiator** (sending first): initializes with `(root_key, chain_key_i)`
  derived from `hybrid_master`. The initial ratchet public key is
  `EK_25519.pub` (already in the first message header).
- **Responder**: initializes with `(root_key, chain_key_r)`. The ratchet
  public key starts as `SPK_25519`.

### 3.2 Message encryption

For each outgoing message:

```
mk, next_chain_key = HKDF-SHA-256(chain_key, b"\x01", "xorein/seal/v1/message-key", 64)
chain_key = next_chain_key

nonce = random(12)
ciphertext = ChaCha20-Poly1305.Seal(
    key   = mk[0:32],
    nonce = nonce,
    aad   = message_header_bytes,  // see §3.3
    pt    = utf8(message_body),
)
```

### 3.3 Message header

The message header (AEAD associated data) is a fixed-layout binary structure:

```
[1-byte version = 0x01]
[4-byte ratchet step counter, big-endian]
[4-byte previous chain length, big-endian]
[32-byte sender ratchet DH public key]
[12-byte nonce]
```

Total header size: 53 bytes.

The header is transmitted in plaintext (unencrypted) but authenticated via AEAD.
The `ciphertext_format` field in the `Delivery` is set to `"seal/v1"`.

### 3.4 Max skipped messages

Implementations MUST maintain a skip list for out-of-order messages. The
maximum number of skipped messages retained is **1000**. If a message arrives
with a counter more than 1000 ahead of the current receiver counter, it MUST
be rejected (not silently dropped).

### 3.5 DH ratchet step

After receiving a message with a new ratchet public key (detected by comparing
the key in the header to the current ratchet key):

```
new_root_key, new_recv_chain_key = HKDF-SHA-256(
    IKM  = root_key,
    salt = X25519(ratchet_priv, remote_ratchet_pub),
    info = "xorein/seal/v1/ratchet-step",
    L    = 64,
)
// generate new sending ratchet key
new_send_ratchet = X25519.GenerateKeypair()
root_key = new_root_key
// derive new sending chain key
_, new_send_chain = HKDF-SHA-256(new_root_key, X25519(new_send_ratchet.priv, remote_ratchet_pub), "xorein/seal/v1/ratchet-step", 64)
```

## 4. `Delivery` wire format for Seal messages

```json
{
  "id": "<uuid>",
  "kind": "dm_message",
  "scope_id": "<dm_record_id>",
  "scope_type": "dm",
  "sender_peer_id": "...",
  "sender_public_key": "<base64url hybrid public keys>",
  "recipient_peer_ids": ["..."],
  "body": "<base64url(ciphertext)>",
  "created_at": "RFC3339Nano",
  "signature": "<base64url hybrid sig>"
}
```

The `body` field contains `base64url_no_padding(ciphertext)` where `ciphertext`
is the output of `ChaCha20-Poly1305.Seal` from §3.2.

A separate `data` field (JSON bytes) carries the message header:

```json
{
  "data": "<base64url(header)>",
  "ciphertext_format": "seal/v1",
  "ek_pub": "<base64url(EK_25519.pub)>",
  "ct_mlkem": "<base64url(ct_mlkem)>",
  "opk_index": 0
}
```

For subsequent messages (not the first in a session), `ek_pub`, `ct_mlkem`,
and `opk_index` are absent.

## 5. Prekey bundle in the identity family

Operations `identity.fetch` (request) and `identity.publish` (publish):

```json
{
  "peer_id": "...",
  "identity_key_ed25519": "<base64url>",
  "identity_key_ml_dsa_65": "<base64url>",
  "signed_prekey_x25519": "<base64url>",
  "signed_prekey_signature": "<base64url hybrid>",
  "one_time_prekeys_x25519": ["<base64url>", ...],
  "ml_kem_768_pk": "<base64url>",
  "ml_kem_768_pk_signature": "<base64url hybrid>",
  "published_at": 1234567890000,
  "expires_at": 1234567890000,
  "bundle_signature": "<base64url hybrid>"
}
```

## 6. Session teardown and rekeying

When a DM is deleted or a Seal session ends:
- All session state (root key, chain keys, skip list) MUST be deleted securely.
- A new session initiated with the same peer starts a fresh X3DH handshake.

Session state MUST be stored in SQLCipher (not in memory only) so it survives
restarts.

## 7. Security properties

| Property | Seal mode |
|----------|-----------|
| Forward secrecy | Per-message (Double Ratchet) |
| Post-compromise security | Per DH ratchet step |
| Store-now-decrypt-later | Resistant (ML-KEM-768 in initial exchange) |
| Relay opacity | MUST be enforced (relay sees only ciphertext) |
| Sender authentication | Yes (hybrid signature on Delivery) |
| Replay protection | Delivery ID deduplication per scope |

## 8. Conformance (W1)

KATs in `pkg/spectest/seal/`:

- `x3dh_classical.json` — classical X3DH with known initiator/responder keys.
- `x3dh_hybrid.json` — hybrid X3DH (X25519 + ML-KEM-768) test vector.
- `ratchet_basic.json` — Double Ratchet encrypt/decrypt round-trip.
- `ratchet_oop.json` — out-of-order messages (skip list).
- `relay_opacity.json` — relay MUST reject delivery with plaintext body.
