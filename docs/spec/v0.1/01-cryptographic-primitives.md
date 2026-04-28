# 01 — Cryptographic Primitives

This document defines the **locked v0.1 cryptographic profile**. Implementations
MUST use exactly these algorithms. Deviation is a conformance failure.

## 1. Identity keys — PQ-hybrid

Every Xorein node MUST maintain two co-equal signing identity key pairs:

| Purpose | Algorithm | Notes |
|---------|-----------|-------|
| Classical signing identity | Ed25519 | RFC 8032; 32-byte public key, 64-byte signature |
| Post-quantum signing identity | ML-DSA-65 | FIPS 204; 1952-byte public key, 3309-byte signature |

Both keys MUST be generated together at node creation, stored together in the
identity record, and published together in `IdentityProfile.signing_public_key`
(Ed25519) and `IdentityProfile.ml_dsa_65_public_key` (ML-DSA-65).

### 1.1 Ed25519 → X25519 key conversion

When an identity Ed25519 key is needed for Diffie-Hellman (e.g., in X3DH), the
key MUST be converted using the standard Bernstein–Hamburg birational map:

```
x25519_private = clamp(SHA-512(ed25519_private_seed)[0:32])
x25519_public  = birational_map(ed25519_public)
```

Reference: [RFC 7748 §4.1](https://www.rfc-editor.org/rfc/rfc7748#section-4.1),
[Bernstein et al. "Twisted Edwards Curves"](https://eprint.iacr.org/2008/013).

An Ed25519 key MUST NOT be used directly as an X25519 scalar without conversion.

### 1.2 PQ library requirement

Implementations MUST use a vetted ML-KEM-768 and ML-DSA-65 implementation. The
reference Go implementation uses [CIRCL](https://github.com/cloudflare/circl)
(`circl/kem/mlkem`, `circl/sign/mldsa`). Implementations MUST pass the
NIST FIPS 203 / FIPS 204 KATs (§7) before claiming conformance.

## 2. Symmetric encryption (AEAD)

| Purpose | Algorithm | Key | Nonce | Tag |
|---------|-----------|-----|-------|-----|
| Message encryption (Seal, Crowd, Channel) | ChaCha20-Poly1305 | 32 bytes | 12 bytes | 16 bytes |
| Media frame encryption (MediaShield) | AES-128-GCM | 16 bytes | 12 bytes | 16 bytes |

Reference: [RFC 8439](https://www.rfc-editor.org/rfc/rfc8439) (ChaCha20-Poly1305),
[NIST SP 800-38D](https://csrc.nist.gov/publications/detail/sp/800-38d/final) (AES-GCM).

AES-128-GCM is used for MediaShield because it is natively supported by the
MLS ciphersuite 0x0001 and its hybrid extension (§4).

### 2.1 Nonce policy

Nonces MUST be generated uniformly at random for each encryption operation.
Nonce reuse under the same key is a catastrophic failure. Implementations MUST
use a cryptographically secure random source (`crypto/rand` in Go).

Counter-based nonces MAY be used only when the key is single-use (e.g., each
Double Ratchet step produces a unique key used for exactly one message).

## 3. Key derivation (KDF)

All key derivation MUST use HKDF-SHA-256 per [RFC 5869](https://www.rfc-editor.org/rfc/rfc5869).

The labeled `Expand` interface:

```
HKDF-Expand(prk, label_string, length)
```

where `label_string` is a UTF-8 string constant. Every label used in this
protocol is defined in `pkg/crypto/labels.go` and cross-referenced here.

### 3.1 Label conventions

All labels follow the pattern `xorein/<mode>/<version>/<purpose>`:

| Label | Used in |
|-------|---------|
| `xorein/seal/v1/x3dh-master-secret` | §5.1 |
| `xorein/seal/v1/hybrid-master` | §5.1 (hybrid combiner label) |
| `xorein/seal/v1/root-key` | §5.2 |
| `xorein/seal/v1/chain-key` | §5.2 |
| `xorein/seal/v1/message-key` | §5.2 |
| `xorein/seal/v1/ratchet-step` | §5.2 |
| `xorein/tree/v1/exporter` | §4.3 (MLS exporter context) |
| `xorein/crowd/v1/sender-key` | `13-mode-crowd.md §3` |
| `xorein/crowd/v1/epoch-root` | `13-mode-crowd.md §3` |
| `xorein/channel/v1/sender-key` | `14-mode-channel.md §3` |
| `xorein/mediashield/v1` | `15-mode-mediashield.md §3` (MLS exporter label) |
| `xorein/mediashield/v1/peer/<peer_id>` | `15-mode-mediashield.md §4` (Crowd-derived per-peer key) |

## 4. Group key agreement (Tree mode)

Tree mode MUST use MLS [RFC 9420](https://www.rfc-editor.org/rfc/rfc9420) with
the following locked ciphersuite for v0.1:

```
XoreinMLS_128_HYBRID_DHKEMX25519MLKEM768_AES128GCM_SHA256_Ed25519MLDSA65
Ciphersuite ID: 0xFF01 (Xorein private-use range; pending IANA registration)
```

### 4.1 Hybrid ciphersuite construction

The v0.1 hybrid ciphersuite extends MLS_128_DHKEMX25519_AES128GCM_SHA256_Ed25519
(ID 0x0001) with the following modifications:

- **KEM**: DHKEMX25519 (as in 0x0001) combined with ML-KEM-768 via the hybrid
  combiner defined in §5.1. The combined shared secret is the input to the MLS
  key schedule.
- **AEAD**: AES-128-GCM (unchanged from 0x0001).
- **Hash**: SHA-256 (unchanged from 0x0001).
- **Signature**: Ed25519 + ML-DSA-65 hybrid (both signatures required; see §6).
- **Leaf node credentials**: MUST include both Ed25519 and ML-DSA-65 public keys
  in the MLS `Credential` extension.

Implementations SHOULD additionally accept MLS ciphersuite 0x0001 from legacy
peers and negotiate up to 0xFF01 when both sides support it.

### 4.2 MLS credential

Credentials MUST use the Xorein Ed25519 identity key as the primary MLS
credential. The ML-DSA-65 public key MUST be included in the `Credential`
extension field `ml_dsa_65_public_key`.

### 4.3 MLS exporter for MediaShield

```
MLS-Exporter(label="xorein/mediashield/v1", context=b"", length=32)
```

## 5. Asymmetric key exchange (Seal mode)

### 5.1 Hybrid X3DH

Seal mode MUST use a hybrid variant of X3DH that combines X25519 and ML-KEM-768.

**Prekey bundle** (published by responder):
- `IKr_ed` — responder's Ed25519 identity key (converted to X25519 per §1.1)
- `SPKr_25519` — signed X25519 prekey
- `OPKr_25519` — one-time X25519 prekey (if available)
- `ML_KEM_768_pk` — responder's ML-KEM-768 encapsulation public key (fresh per bundle)

**X3DH initiator computation**:
```
DH1 = X25519(IKi_25519, SPKr_25519)
DH2 = X25519(EKi_25519, IKr_25519)
DH3 = X25519(EKi_25519, SPKr_25519)
DH4 = X25519(EKi_25519, OPKr_25519)   // if OPK available; else omit
(ss_mlkem, CT_mlkem) = ML-KEM-768.Encapsulate(ML_KEM_768_pk)

x3dh_secret = DH1 || DH2 || DH3 || DH4    // classical X3DH output
hybrid_secret = HKDF-SHA-256(
    IKM = x3dh_secret || ss_mlkem,
    salt = b"",
    info = "xorein/seal/v1/hybrid-master",
    length = 32
)
```

`CT_mlkem` is transmitted to the responder in the initial message header.

**Responder decaps**: receives `CT_mlkem`, computes `ss_mlkem = ML-KEM-768.Decapsulate(CT_mlkem)`,
then reconstructs `hybrid_secret` identically.

Both X25519 and ML-KEM-768 MUST succeed; failure of either aborts the session.

| X3DH parameter | Value |
|----------------|-------|
| Identity key curve | Ed25519 → X25519 (§1.1) |
| Signed prekey | X25519 |
| One-time prekey | X25519 |
| Additional KEM | ML-KEM-768 |
| KDF | HKDF-SHA-256 (RFC 5869) |
| Hybrid combiner label | `"xorein/seal/v1/hybrid-master"` |

### 5.2 Double Ratchet

Double Ratchet MUST follow the [Signal Double Ratchet specification](https://signal.org/docs/specifications/doubleratchet/)
with:

| Parameter | Value |
|-----------|-------|
| DH curve | X25519 |
| AEAD | ChaCha20-Poly1305 |
| KDF | HKDF-SHA-256 |
| Root key label | `"xorein/seal/v1/root-key"` |
| Chain key label | `"xorein/seal/v1/chain-key"` |
| Message key label | `"xorein/seal/v1/message-key"` |
| Max skipped messages | 1000 |

The Double Ratchet provides forward secrecy and post-compromise security over
the classical X25519 component. The ML-KEM-768 component provides
store-now-decrypt-later resistance for the initial session establishment.

## 6. Hybrid signature scheme

Every signed object in this protocol uses the hybrid signature scheme defined
here. A hybrid signature consists of two independent signatures over the same
canonical payload:

```
hybrid_sig = Ed25519.Sign(ed_private, canonical_payload)
          || ML-DSA-65.Sign(mldsa_private, canonical_payload)
```

In the `SignedEnvelope` message:
- `signature` contains the Ed25519 signature (64 bytes).
- `ml_dsa_65_signature` contains the ML-DSA-65 signature (3309 bytes).
- `signature_algorithm` MUST be `SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65 = 4`.
- `signer.signing_public_key` contains the Ed25519 public key.
- `signer.ml_dsa_65_public_key` contains the ML-DSA-65 public key.

Verification MUST check both signatures. Failure of either constitutes a
`SIGNATURE_MISMATCH` error. Neither signature alone is sufficient.

### 6.1 JSON signed objects (Manifest, Invite, Delivery)

For JSON-signed objects the signature field carries a combined
`base64url(Ed25519_sig || ML-DSA-65_sig)` encoding with no padding:

```
combined = Ed25519.Sign(ed_private, canonical_json)
         || ML-DSA-65.Sign(mldsa_private, canonical_json)
signature_field = base64url_no_padding(combined)
```

Verifiers decode the combined field, split at byte 64 (Ed25519), and verify
both halves independently.

## 7. Conformance

Every implementation MUST pass the following KATs before claiming any
conformance level above W0:

### W0 — Classical primitives (required for all levels)

- RFC 7748 §6.1 explicit function test vectors for X25519 (both vectors).
- RFC 8439 §2.8.2 test vector for ChaCha20-Poly1305.
- NIST SP 800-38D Appendix B.2 test vector for AES-128-GCM.
- RFC 5869 Appendix A.1 test vector for HKDF-SHA-256.
- AEAD tamper detection: modified ciphertext or AAD MUST fail authentication.

### W0-PQ — Post-quantum primitives (required for PQ-hybrid conformance)

- NIST FIPS 203 encapsulation/decapsulation KAT for ML-KEM-768.
- NIST FIPS 204 sign/verify KAT for ML-DSA-65.
- Hybrid combiner round-trip: same `hybrid_secret` derived by initiator and
  responder from matching X3DH and ML-KEM shared secrets.
- Hybrid signature: verify that a hybrid signature over a known payload verifies
  with both Ed25519 and ML-DSA-65 independently.

### W1 — Seal mode (X3DH + Double Ratchet)

- Signal X3DH test vectors (classical only subset).
- Signal Double Ratchet test vectors.
- Relay opacity: relay queue row MUST NOT contain plaintext body.

### W2 — Tree mode (MLS)

- RFC 9420 official test vectors for the base ciphersuite (0x0001).
- TreeKEM, key schedule, Welcome/Commit, transcript hash vectors.

### W3–W6 — Crowd, Channel, MediaShield, full Discord parity

See `90-conformance-harness.md` for per-level requirements.

Reference implementations of all KATs live in `pkg/crypto/vectors_test.go`
and `pkg/spectest/<mode>/`.
