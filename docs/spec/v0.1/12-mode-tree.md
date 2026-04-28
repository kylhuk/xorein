# 12 — Mode: Tree (Interactive Group E2EE)

Tree mode provides end-to-end encryption for server-scoped group conversations
using a hybrid MLS (Messaging Layer Security) protocol. It is suited for
servers with up to 50 active members.

## 1. Ciphersuite

Tree mode uses the Xorein hybrid MLS ciphersuite:

```
XoreinMLS_128_HYBRID_DHKEMX25519MLKEM768_AES128GCM_SHA256_Ed25519MLDSA65
Ciphersuite ID: 0xFF01  (Xorein private-use range per RFC 9420 §17.1)
```

### 1.1 Ciphersuite parameters

| Parameter | Value |
|-----------|-------|
| KEM | Hybrid: DHKEMX25519 + ML-KEM-768 (see §1.2) |
| AEAD | AES-128-GCM |
| Hash | SHA-256 |
| Signature | Ed25519 + ML-DSA-65 hybrid (see §1.3) |
| MLS spec | RFC 9420 |

### 1.2 Hybrid KEM construction

The KEM combines DHKEMX25519 and ML-KEM-768 per the hybrid combiner from
`01-cryptographic-primitives.md §3`:

```
// Encapsulation
(ss_x25519, ct_x25519) = DHKEMX25519.Encaps(recipient_pk_x25519)
(ss_mlkem,  ct_mlkem)  = ML-KEM-768.Encapsulate(recipient_pk_mlkem768)

hybrid_ss = HKDF-SHA-256(
    IKM  = ss_x25519 || ss_mlkem,
    salt = b"",
    info = "xorein/tree/v1/kem-combine",
    L    = 32,
)
kem_output = ct_x25519 || ct_mlkem  // combined ciphertext
```

Decapsulation: reverse using both private keys; combine identically.

The KEM ciphertext in TreeKEM leaf node `encrypt` fields carries
`ct_x25519 || ct_mlkem`. Parsers MUST know the split point from the fixed
DHKEMX25519 ciphertext length.

### 1.3 Hybrid signature in MLS

All MLS messages (KeyPackage, MLSMessage, Commit, Welcome, Proposal) MUST use
the hybrid signature scheme from `01-cryptographic-primitives.md §6`:

- `signature` field in MLS messages: 3373-byte combined Ed25519 + ML-DSA-65.
- `leaf_node.signature_key` in KeyPackage: Ed25519 public key (32 bytes).
- `leaf_node.credential` extension `ml_dsa_65_public_key`: ML-DSA-65 public
  key (1952 bytes).

MLS signature verification MUST verify both components independently. Failure
of either MUST be treated as an MLS `ValidationError`.

## 2. Key lifecycle

### 2.1 KeyPackage

Each member MUST maintain a valid KeyPackage for the server's MLS group. The
KeyPackage is published via `identity.publish` and fetched by the group
administrator before adding a new member.

KeyPackage fields beyond the standard MLS spec:
- `leaf_node.credential.ml_dsa_65_public_key` — ML-DSA-65 verification key.
- `kem_encap_key_mlkem768` — ML-KEM-768 public key (for the hybrid KEM).

### 2.2 Group creation

The server owner (group administrator) creates the MLS group:

```
1. Admin fetches KeyPackages for all initial members via identity.fetch.
2. Admin creates MLS group with own KeyPackage.
3. Admin generates Welcome messages for each member.
4. Admin sends Commit + Welcome via server manifest update and 
   individual chat.send deliveries to each member.
```

### 2.3 Member add

```
1. Admin fetches newcomer's KeyPackage.
2. Admin creates MLS Proposal(Add) + Commit.
3. Admin broadcasts the Commit to all current members via chat.send.
4. Admin sends Welcome to the newcomer.
5. All members process the Commit and advance their group state.
```

### 2.4 Member remove

```
1. Admin creates MLS Proposal(Remove) + Commit.
2. Admin broadcasts the Commit to remaining members.
3. The group epoch advances; the removed member's keys are deleted
   from the ratchet tree. The removed member can no longer decrypt
   new messages.
```

### 2.5 Epoch rotation triggers

In addition to MLS's natural commit-based epoch advancement, Xorein requires
an **explicit epoch rotation** when:
- Any membership change (add or remove).
- 1000 messages sent in the current epoch.
- 7 days since the last commit.

Whichever trigger fires first causes the administrator to issue a Commit even
if no membership change is pending. This ensures forward secrecy bounds are
respected in low-activity groups.

## 3. MLS exporter for MediaShield

When a voice session is started in a Tree-mode server, the SFU coordinator
(or all-to-all in small groups) derives MediaShield keys from the MLS group:

```
mediashield_key = MLS-Exporter(
    label   = "xorein/mediashield/v1",
    context = b"",
    length  = 32,
)
```

This is called once per member per MLS epoch. On epoch change (due to member
join/leave), all MediaShield keys MUST be rotated immediately.

## 4. Wire encoding

### 4.1 MLS messages in Delivery

MLS messages (Commit, Welcome, Proposal, ApplicationMessage) are carried as the
`data` field of the `Delivery` JSON object:

```json
{
  "kind": "tree_mls_message",
  "scope_id": "<server_id>",
  "scope_type": "server",
  "body": "<base64url(aes128gcm ciphertext of application message)>",
  "data": "<base64url(proto.Marshal(MLSMessage))>",
  "ciphertext_format": "tree/v1",
  "signature": "<base64url hybrid sig>"
}
```

For Commit and Welcome messages, `body` is empty and `data` carries the MLS
framing. For ApplicationMessage (chat), `body` carries the AEAD ciphertext.

### 4.2 ApplicationMessage AEAD

```
// Encrypt a chat message payload
aad = message_header  // see 11-mode-seal.md §3.3 for header layout
aead_key = MLS application secret for this member (derived from MLS key schedule)
nonce = random(12)
ciphertext = AES-128-GCM.Seal(aead_key, nonce, aad, plaintext)
```

## 5. Group size limits

| Parameter | Value |
|-----------|-------|
| Max members | 50 |
| Max pending proposals per commit | 20 |
| Max epoch age (time-based rotation) | 7 days |
| Max epoch age (message-based rotation) | 1000 messages |

Groups that exceed 50 members MUST migrate to Crowd mode. The migration is
initiated by the server owner via a manifest update changing `security_mode`
to `"crowd"`.

## 6. Security properties

| Property | Tree mode |
|----------|-----------|
| Confidentiality | Group E2EE (AES-128-GCM per MLS key schedule) |
| Forward secrecy | Per-epoch (MLS commit) |
| Post-compromise security | Yes (MLS TreeKEM) |
| Store-now-decrypt-later | Resistant (ML-KEM-768 in KEM) |
| Relay opacity | MUST be enforced |
| Membership authentication | Yes (MLS credential + hybrid signature) |

## 7. Conformance (W2)

KATs in `pkg/spectest/tree/`:

- RFC 9420 test vectors for TreeKEM (adapted for ciphersuite 0xFF01 structure).
- RFC 9420 test vectors for the key schedule.
- RFC 9420 test vectors for Welcome and Commit message structure.
- `hybrid_kem_combine.json` — hybrid KEM combine test vector.
- `mls_hybrid_sig.json` — hybrid signature sign/verify over MLS message.
- `epoch_rotation.json` — epoch rotates after 1000 messages and after member add.
- `mediashield_exporter.json` — MLS-Exporter output for MediaShield key derivation.
