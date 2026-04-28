# 14 — Mode: Channel (Broadcast-Epoch E2EE)

Channel mode is a variant of Crowd mode optimized for **few writers, many
readers** — analogous to a Discord announcement channel or a public server
channel where most members are read-only. It uses the same epoch key structure
as Crowd but adds per-write authentication signatures and read-only membership
semantics.

## 1. Differences from Crowd mode

| Aspect | Crowd | Channel |
|--------|-------|---------|
| Writer set | All members | Designated writers (role: `writer`) |
| Reader set | All members | All members (including read-only) |
| Key distribution | All members | Symmetric: epoch root distributed to all; sender keys derived locally |
| Per-message auth | Hybrid delivery signature | Hybrid delivery signature + embedded message sig |
| Membership auth | Implicit (key access) | RBAC role enforcement (writer list in manifest) |

Channel mode MUST inherit its epoch key structure from Crowd mode
(`13-mode-crowd.md`). This document specifies only the Channel-specific
extensions.

## 2. Writer authorization

### 2.1 Writer list

The manifest's RBAC governance data identifies which peer IDs are `writer`
role members. A `channel_message` delivery from a peer NOT in the writer list
MUST be rejected by all receiving clients.

Writer assignment is managed via the Governance family
(`/aether/governance/0.2.0`). See `51-family-governance.md`.

### 2.2 Additional per-message signature

Channel writers MUST include an embedded message signature in the message:

```
msg_canonical = b"xorein/channel/v1/msg-sign" || epoch_id || scope_id || body_ciphertext
msg_sig = hybrid_sign(identity_keys, msg_canonical)
```

This signature is included in the `Delivery.data` JSON field alongside the
message header:

```json
{
  "ciphertext_format": "channel/v1",
  "epoch_id": "<base64url>",
  "msg_sig": "<base64url hybrid>"
}
```

Recipients MUST verify `msg_sig` before processing the message. A `msg_sig`
from a non-writer peer MUST be rejected.

## 3. History manifests (signed snapshots)

Channel mode supports **signed history snapshots** to allow late-joining members
to receive a compressed history without fetching every individual message.

### 3.1 Snapshot format

A snapshot is a signed JSON document:

```json
{
  "scope_id": "<channel_id>",
  "epoch_id": "<base64url>",
  "from_seq": 0,
  "to_seq": 999,
  "message_count": 1000,
  "message_hashes": ["<sha256 hex>", ...],
  "snapshot_root": "<merkle root hex>",
  "created_at": "RFC3339Nano",
  "signature": "<base64url hybrid>"
}
```

Snapshots are signed by the server owner or an archivist node. A snapshot
without a valid hybrid signature MUST be rejected.

### 3.2 Late-join semantics

A member joining a Channel-mode server receives:

1. The current epoch sender key package (via Seal DM from owner).
2. The latest history snapshot (from archivist or server owner).
3. Individual messages from the snapshot boundary to present.

Historical messages from before the member's join epoch CANNOT be decrypted
by the new member (forward secrecy property). The snapshot provides metadata
(message count, hashes) but not decryptable content.

## 4. Key derivation

Channel mode uses identical key derivation to Crowd mode with only the KDF
labels changed:

```
sender_key_m = HKDF-SHA-256(
    epoch_root_secret,
    b"",
    "xorein/channel/v1/sender-key" || peer_id_m,
    32,
)
```

The `epoch_root_secret` derivation and rotation rules are identical to
`13-mode-crowd.md §3` and §6.

## 5. Wire format

Identical to Crowd mode except:

- `ciphertext_format` = `"channel/v1"` (not `"crowd/v1"`).
- `Delivery.data` includes the additional `msg_sig` field (§2.2).

## 6. Security properties

| Property | Channel mode |
|----------|-------------|
| Confidentiality | Group E2EE (same as Crowd) |
| Forward secrecy | Per-epoch |
| Writer authentication | Yes (hybrid delivery sig + embedded msg sig) |
| Reader anonymity | No (reader list known to epoch key holder) |
| Relay opacity | MUST be enforced |

## 7. Conformance (W3)

KATs in `pkg/spectest/channel/`:

- All Crowd-mode KATs apply (inherit).
- `writer_auth.json` — writer verification from non-writer rejected.
- `snapshot.json` — history snapshot signed and verified.
- `late_join.json` — late-joining member receives epoch key; cannot decrypt past epochs.
