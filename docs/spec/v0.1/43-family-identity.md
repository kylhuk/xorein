# 43 — Family: Identity (`/aether/identity/0.1.0`)

This document specifies the Identity family, which manages cryptographic prekey
bundle distribution for end-to-end encrypted sessions.

## 1. Overview

The Identity family provides the key-distribution substrate required for
initiating Seal (1:1 E2EE) sessions and, in future protocol versions, for
Tree and Crowd mode participant enrollment. It defines two operations: fetching
a peer's current prekey bundle, and publishing an updated prekey bundle.

Prekey bundles contain all public key material required to initiate an
asymmetric key exchange with a peer without requiring that peer to be online at
the time of session initiation. Successful bundle fetch is a prerequisite for
Seal mode DM sessions (see `11-mode-seal.md §1`).

**Roles that use this family:**
- **Client**: Publishes its own bundle on startup; fetches bundles before
  initiating Seal sessions.
- **Relay**: MAY cache and serve prekey bundles as a proxy to reduce traffic to
  clients; MUST NOT modify bundles.
- **Bootstrap**: MAY serve cached bundles fetched from peers.
- **Archivist**: SHOULD cache bundles for archived peers.

**Security modes:** The Identity family is mode-independent. Prekey bundles
enable E2EE modes; they are not themselves mode-scoped. All bundle content is
transmitted and stored in the clear because it is public-key material; the
hybrid signature provides integrity.

**Protocol ID:** `/aether/identity/0.1.0`

**Required capability:** `cap.identity` — MUST be present in `advertised_caps`
for all Identity family streams.

## 2. Capability requirements

| Capability | Required for |
|------------|-------------|
| `cap.identity` | All Identity family operations |

## 3. Operations

| Operation | Required caps | Direction | Request payload type | Response payload type | Description |
|-----------|--------------|-----------|---------------------|----------------------|-------------|
| `identity.fetch` | `cap.identity` | initiator → responder | `IdentityFetchRequest` | `IdentityFetchResponse` | Fetch the current prekey bundle for a peer |
| `identity.publish` | `cap.identity` | client → peers | `IdentityPublishRequest` | `IdentityPublishResponse` | Publish or update this node's prekey bundle |

The `operation` field in `PeerStreamRequest` MUST be one of the strings above.

## 4. Wire format details

All payloads are JSON-encoded and stored in `PeerStreamRequest.payload` /
`PeerStreamResponse.payload`.

### 4.1 `identity.fetch`

**Request:** `IdentityFetchRequest` — JSON-encoded:

```
{
  "peer_id":     string,   // peer ID whose bundle is requested
  "opk_consumed": bool     // true if the requester will consume a one-time prekey
}
```

The `opk_consumed` field informs the publisher that a one-time prekey (OPK)
has been consumed so it can replenish its OPK supply. A requester MUST set
`opk_consumed: true` if and only if it intends to use an OPK from the returned
bundle for session initiation.

**Response:** `IdentityFetchResponse` — JSON-encoded:

```
{
  "bundle":          { /* PrekeyBundle JSON; see §4.3 */ },
  "remaining_opks":  int,    // number of OPKs remaining after this fetch
  "error":           string  // human-readable; omitted on success
}
```

Notes:
- If `opk_consumed` is true, the responder MUST mark one OPK as consumed in its
  local state before responding. The returned bundle MUST include that OPK in
  the `one_time_prekeys_x25519` list so the requester can use it.
- If `opk_consumed` is true and no OPKs remain, the responder MUST respond with
  `OPERATION_FAILED` and message `"no one-time prekeys available"`. The
  initiator MUST fall back to using only the signed prekey (see
  `11-mode-seal.md §2.1`).
- If the responder does not hold a bundle for `peer_id`, it MUST respond with
  `OPERATION_FAILED` and message `"bundle not found"`.
- A bundle past its `expires_at` MUST NOT be served. The responder MUST return
  `OPERATION_FAILED` and message `"bundle expired"`.

### 4.2 `identity.publish`

**Request:** `IdentityPublishRequest` — JSON-encoded:

```
{
  "bundle": { /* PrekeyBundle JSON */ }
}
```

**Response:** `IdentityPublishResponse` — JSON-encoded:

```
{
  "accepted": bool,
  "error":    string
}
```

Notes:
- The receiver MUST verify `bundle.bundle_signature` before accepting. See §4.4.
- A bundle MUST be rejected if its `published_at` is not strictly newer than the
  currently stored bundle for the same peer ID.
- A bundle MUST be rejected if its `expires_at` is more than 7 days after
  `published_at`. (Maximum bundle validity: 7 days.)
- Upon acceptance, the receiver MUST store the bundle in the `identities` bucket
  keyed by the bundle's `identity_key_ed25519` peer ID derivation.
- The receiver SHOULD propagate the new bundle to other directly-connected peers
  that have advertised `cap.identity` (gossip, 3-hop limit).

### 4.3 `PrekeyBundle` struct

The PrekeyBundle is defined in `11-mode-seal.md §1.1` and extended here for
wire transport purposes. The JSON representation is:

```
{
  "identity_key_ed25519":     string,    // Ed25519 public key, base64url no-padding (32 bytes)
  "identity_key_ml_dsa_65":   string,    // ML-DSA-65 public key, base64url no-padding (1952 bytes)
  "signed_prekey_x25519":     string,    // X25519 signed prekey public key, base64url (32 bytes)
  "signed_prekey_signature":  string,    // hybrid sig over SPK, base64url (3373 bytes)
  "one_time_prekeys_x25519":  [string],  // list of X25519 one-time prekey public keys, base64url; up to 100
  "ml_kem_768_pk":            string,    // ML-KEM-768 encapsulation public key, base64url
  "ml_kem_768_pk_signature":  string,    // hybrid sig over ML-KEM-768 public key, base64url (3373 bytes)
  "published_at":             int64,     // unix milliseconds; bundle issuance time
  "expires_at":               int64,     // unix milliseconds; MUST be ≤ published_at + 7 days
  "bundle_signature":         string     // hybrid sig over canonical bundle, base64url (3373 bytes)
}
```

Field encoding:
- All byte fields are base64url-encoded without padding (RFC 4648 §5).
- `published_at` and `expires_at` are 64-bit integers (unix milliseconds).
- All hybrid signatures are 3373 bytes (64-byte Ed25519 || 3309-byte ML-DSA-65);
  see `01-cryptographic-primitives.md §6.1`.

### 4.4 Bundle signing and verification

**Signed prekey signature** (`signed_prekey_signature`):

```
spk_canonical       = b"xorein/seal/v1/spk-sign" || signed_prekey_x25519
spk_sig_ed          = Ed25519.Sign(identity_ed_priv, spk_canonical)
spk_sig_mldsa       = ML-DSA-65.Sign(identity_mldsa_priv, spk_canonical)
signed_prekey_signature = spk_sig_ed || spk_sig_mldsa   // 3373 bytes
```

**ML-KEM-768 public key signature** (`ml_kem_768_pk_signature`):

```
mlkem_canonical         = b"xorein/seal/v1/mlkem-sign" || ml_kem_768_pk
mlkem_sig_ed            = Ed25519.Sign(identity_ed_priv, mlkem_canonical)
mlkem_sig_mldsa         = ML-DSA-65.Sign(identity_mldsa_priv, mlkem_canonical)
ml_kem_768_pk_signature = mlkem_sig_ed || mlkem_sig_mldsa   // 3373 bytes
```

**Bundle signature** (`bundle_signature`) covers the canonical bundle form:

```
canonical_bundle = JSON-encode({
  "identity_key_ed25519":     ...,
  "identity_key_ml_dsa_65":   ...,
  "signed_prekey_x25519":     ...,
  "signed_prekey_signature":  ...,
  "one_time_prekeys_x25519":  [...],   // lexicographically sorted
  "ml_kem_768_pk":            ...,
  "ml_kem_768_pk_signature":  ...,
  "published_at":             ...,
  "expires_at":               ...
  // bundle_signature excluded
})
bundle_sig_ed     = Ed25519.Sign(identity_ed_priv, canonical_bundle)
bundle_sig_mldsa  = ML-DSA-65.Sign(identity_mldsa_priv, canonical_bundle)
bundle_signature  = bundle_sig_ed || bundle_sig_mldsa   // 3373 bytes
```

JSON keys in the canonical form MUST be in lexicographic order.
`one_time_prekeys_x25519` MUST be sorted lexicographically before signing.

**Verification procedure for a received bundle:**
1. Verify `signed_prekey_signature` over `b"xorein/seal/v1/spk-sign" ||
   signed_prekey_x25519` using `identity_key_ed25519` (Ed25519) and
   `identity_key_ml_dsa_65` (ML-DSA-65). Both MUST pass.
2. Verify `ml_kem_768_pk_signature` over `b"xorein/seal/v1/mlkem-sign" ||
   ml_kem_768_pk` using both identity keys. Both MUST pass.
3. Reconstruct the canonical bundle JSON (step above). Verify `bundle_signature`
   over it using both identity keys. Both MUST pass.
4. Check `expires_at > published_at`. Reject if not.
5. Check `expires_at - published_at ≤ 7 days`. Reject if exceeded.
6. Check `published_at ≤ now + 5 minutes` (allow small clock skew). Reject
   future-dated bundles.
7. Confirm `one_time_prekeys_x25519` count ≤ 100. Reject if exceeded.

### 4.5 OPK management

One-time prekeys provide forward secrecy by ensuring each session initiation
uses a unique ephemeral key. The following rules govern OPK lifecycle:

- A bundle MUST contain at least 1 OPK and at most 100 OPKs.
- When a node receives a fetch with `opk_consumed: true`, it MUST remove one
  OPK from its local bundle state and MUST track its remaining OPK count.
- When the remaining OPK count falls below 5, the node SHOULD immediately
  generate 20 new OPKs, update its bundle, increment `published_at` to
  `now`, and publish the updated bundle via `identity.publish` to all
  directly-connected peers.
- OPK replenishment MUST NOT reuse OPK key material. Each generated OPK MUST
  be a fresh X25519 key pair. The private key MUST be stored locally and MUST
  be zeroized from memory after first use.
- If no OPKs are available and a fetch arrives with `opk_consumed: true`, the
  responder MUST return error `"no one-time prekeys available"`. The initiator
  SHOULD retry after a short delay to allow time for replenishment.

## 5. Security mode binding

The Identity family is mode-independent. Prekey bundles are public-key material
and are always transmitted in the clear (hop-to-hop Noise XX protects
confidentiality in transit). The hybrid signature scheme ensures that a
compromised classical-only key cannot forge a valid bundle.

Prekey bundles enable:
- **Seal mode**: X25519 signed prekey + OPK + ML-KEM-768 for hybrid X3DH
  session initiation. See `11-mode-seal.md §2`.
- **Tree mode (future)**: Identity keys used for MLS credential verification.
- **Crowd/Channel mode (future)**: Identity keys used for sender-key epoch
  authentication.

A node MUST NOT use a prekey bundle whose `bundle_signature` fails verification,
regardless of security mode. Doing so would allow a man-in-the-middle to inject
attacker-controlled key material.

## 6. State persistence

| State bucket | Key type | Value type | Description |
|-------------|----------|-----------|-------------|
| `identity` | singleton | `Identity` (JSON) | Local node identity (keys, profile) |

Note: prekey bundles for remote peers are cached in memory (not persisted by
default). Implementations MAY persist remote bundles in an `identities` bucket
(keyed by peer ID) for resilience across restarts. If persisted, bundles MUST
be re-verified on load and discarded if expired.

The local node's prekey bundle is derived from `Identity.PrivateKey` and related
key material held in the `identity` state bucket. The Identity struct (from
`pkg/node/types.go`) is:

```
{
  "id":          string,    // internal UUID
  "peer_id":     string,    // libp2p peer ID (base58btc)
  "public_key":  string,    // Ed25519 public key, base64url
  "private_key": string,    // Ed25519 private key, base64url (MUST be kept secret)
  "created_at":  string,    // RFC3339Nano
  "profile":     { "display_name": string, "bio": string }
}
```

The ML-DSA-65 key pair and X25519 prekeys are derived from or stored alongside
`Identity`. The exact storage of PQ private keys is implementation-specific but
MUST be encrypted at rest using the same SQLCipher key derivation as all other
state (see `pkg/storage/store.go`).

## 7. Error codes

| Code | Trigger |
|------|---------|
| `BUNDLE_NOT_FOUND` | `identity.fetch` for a peer with no known bundle |
| `BUNDLE_EXPIRED` | `identity.fetch` for a bundle past its `expires_at` |
| `NO_OPK_AVAILABLE` | `identity.fetch` with `opk_consumed: true` and no OPKs remaining |
| `BUNDLE_SIGNATURE_INVALID` | `identity.publish` bundle verification failed (any signature) |
| `BUNDLE_NOT_NEWER` | `identity.publish` received bundle `published_at` ≤ stored bundle |
| `BUNDLE_TTL_EXCEEDED` | `identity.publish` `expires_at - published_at` > 7 days |
| `BUNDLE_FUTURE_DATED` | `identity.publish` `published_at` > now + 5 minutes |
| `BUNDLE_OPK_LIMIT` | `identity.publish` bundle contains more than 100 OPKs |

## 8. Conformance

Implementations claiming Identity family conformance MUST pass the following
KATs:

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/identity/bundle_publish_kat.json` | `identity.publish` with valid bundle → accepted |
| `pkg/spectest/identity/bundle_fetch_kat.json` | `identity.fetch` without OPK consumption |
| `pkg/spectest/identity/bundle_fetch_opk_kat.json` | `identity.fetch` with `opk_consumed: true`; OPK count decremented |
| `pkg/spectest/identity/bundle_no_opk_kat.json` | `identity.fetch` with `opk_consumed: true` and zero OPKs → `NO_OPK_AVAILABLE` |
| `pkg/spectest/identity/bundle_invalid_sig_kat.json` | `identity.publish` with tampered `bundle_signature` → `BUNDLE_SIGNATURE_INVALID` |
| `pkg/spectest/identity/bundle_expired_kat.json` | `identity.fetch` for expired bundle → `BUNDLE_EXPIRED` |
| `pkg/spectest/identity/bundle_ttl_kat.json` | `identity.publish` with `expires_at` > 7 days out → `BUNDLE_TTL_EXCEEDED` |
| `pkg/spectest/identity/spk_sig_kat.json` | Signed prekey signature construction and verification |
| `pkg/spectest/identity/mlkem_sig_kat.json` | ML-KEM-768 public key signature construction and verification |

All KAT files use the format defined in `90-conformance-harness.md`. Bundle KAT
vectors MUST include the full PrekeyBundle JSON, all intermediate canonical
forms, and the expected verification outcome. OPK management vectors MUST
include the before/after OPK count state.
