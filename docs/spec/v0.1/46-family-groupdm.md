# 46 — Family: GroupDM (`/aether/groupdm/0.2.0`)

Group DMs provide end-to-end encrypted multi-party conversations for up to 50
participants using hybrid MLS (ciphersuite `0xFF01`) as specified in
`12-mode-tree.md`. Group DMs are scoped independently from server channels and
do not require a server manifest.

## 1. Overview

The GroupDM family manages the full lifecycle of small-group E2EE conversations:
creation, membership changes (add/remove/leave), and message delivery. All
operations are mediated by the hybrid MLS protocol. Like Tree mode in server
channels, GroupDM uses ciphersuite `0xFF01` and enforces epoch-based forward
secrecy (see `12-mode-tree.md §2.5`).

**Roles that participate:**

| Role | Participation |
|------|--------------|
| client | Creator, admin, and member of group DMs |
| relay | Stores ciphertext-only payloads for offline members; MUST NOT inspect content |
| bootstrap | Does not participate |
| archivist | Does not participate |

**Security mode:** Tree only (`mode.tree`). Clear mode is prohibited. Seal
mode is not applicable (group E2EE uses MLS, not X3DH + Double Ratchet).

**Size limits:** A GroupDM supports between 2 and 50 members (inclusive). A
group that would exceed 50 members MUST be converted to a server (Tree or Crowd
mode). The `GroupDmGrowthDecision` proto carries the deterministic outcome of
a growth check.

## 2. Capability requirements

| Capability | Role |
|-----------|------|
| `cap.group-dm` | Required on all participants for any GroupDM operation |
| `mode.tree` | Required for `groupdm.create`, `groupdm.add`, `groupdm.remove`, `groupdm.send` |
| `cap.peer.delivery` | Required on responders accepting direct message delivery |
| `cap.peer.relay` | Required on relay nodes storing offline payloads |
| `cap.identity` | Required for KeyPackage exchange during member add |

A peer that advertises `cap.group-dm` MUST also advertise `mode.tree`.

## 3. Operations

| Operation | Required caps (initiator) | Direction | Description |
|-----------|--------------------------|-----------|-------------|
| `groupdm.create` | `cap.group-dm`, `mode.tree` | initiator → each invited peer | Creator creates the MLS group, sends Welcome to each initial member |
| `groupdm.add` | `cap.group-dm`, `mode.tree` | admin → new member + existing members | Admin adds a new member; sends Commit to existing, Welcome to new member |
| `groupdm.remove` | `cap.group-dm`, `mode.tree` | admin → all remaining members | Admin removes a member; sends Commit to all remaining members |
| `groupdm.send` | `cap.group-dm`, `mode.tree` | sender → all members | Deliver an MLS ApplicationMessage to all group members |
| `groupdm.leave` | `cap.group-dm` | leaving peer → admin | Member self-removes by sending Proposal(Remove) for self |

All operations carry `security_mode = SECURITY_MODE_TREE` in the
`PeerStreamRequest`. An unexpected security mode MUST produce `MODE_INCOMPATIBLE`.

## 4. Wire format details

### 4.1 Group identifier

Each group DM has a stable `group_id` (UUID v4) assigned by the creator at
creation time. It is included in all operation payloads and is used as the MLS
`group_id` field (truncated/padded to 16 bytes for MLS compatibility).

### 4.2 `groupdm.create`

The creator:

1. Fetches a KeyPackage for each initial member via `identity.fetch`.
2. Creates an MLS group with own KeyPackage (ciphersuite `0xFF01`).
3. Generates an MLS `Commit` and individual `Welcome` messages for each
   invited member.
4. Opens a stream to each invited peer and sends `groupdm.create`.

`PeerStreamRequest.payload = proto.Marshal(GroupDmMembershipEvent)` with:
- `event_type = GROUP_DM_MEMBERSHIP_EVENT_TYPE_INVITE`
- `current_state = GROUP_DM_MEMBER_STATE_NONE`
- `next_state = GROUP_DM_MEMBER_STATE_INVITED`
- `security_mode = SECURITY_MODE_TREE`

The `data` field in the request JSON carries:

```json
{
  "group_id": "<uuid v4>",
  "group_name": "<display name; max 64 UTF-8 chars>",
  "creator_peer_id": "<base58>",
  "mls_welcome": "<base64url(proto.Marshal(MLSMessage[Welcome]))>",
  "initial_members": [
    {"peer_id": "<base58>", "display_name": "<string>"},
    "..."
  ],
  "epoch": 0,
  "created_at": "<RFC3339Nano>"
}
```

The responder MUST process the MLS Welcome, store the resulting group state,
and create a `GroupDMRecord` (see §6). On success, the responder sends:

```json
{
  "group_id": "<uuid v4>",
  "accepted": true,
  "member_state": "invited"
}
```

### 4.3 `groupdm.add`

Admin flow:

1. Admin fetches new member's KeyPackage via `identity.fetch`.
2. Admin creates `Proposal(Add)` + `Commit`.
3. Admin sends the `Commit` to all **existing** members via `groupdm.add`
   (they advance their group epoch).
4. Admin sends `Welcome` to the **new** member via a separate `groupdm.create`
   stream (with `event_type = INVITE`).

`PeerStreamRequest.payload = proto.Marshal(GroupDmMembershipEvent)` with:
- `event_type = GROUP_DM_MEMBERSHIP_EVENT_TYPE_JOIN`
- `current_state = GROUP_DM_MEMBER_STATE_MEMBER` (for existing members)
- `next_state = GROUP_DM_MEMBER_STATE_MEMBER`
- `sequence` incremented from previous event

The `data` field carries:

```json
{
  "group_id": "<uuid v4>",
  "mls_commit": "<base64url(proto.Marshal(MLSMessage[Commit]))>",
  "new_member_peer_id": "<base58>",
  "epoch": "<uint64; new epoch after commit>",
  "growth_decision": { "<GroupDmGrowthDecision fields>" }
}
```

The `growth_decision` field carries `proto.Marshal(GroupDmGrowthDecision)` with
the deterministic cap check result (see §4.8).

### 4.4 `groupdm.remove`

Admin flow:

1. Admin creates `Proposal(Remove)` + `Commit`.
2. Admin sends the `Commit` to all **remaining** members via `groupdm.remove`.
3. Admin sends a `GroupDmRekeyDecision` (see §4.7) confirming mandatory rekey.

`PeerStreamRequest.payload = proto.Marshal(GroupDmMembershipEvent)` with:
- `event_type = GROUP_DM_MEMBERSHIP_EVENT_TYPE_REMOVE`
- `member_id` = the removed peer's ID
- `next_state = GROUP_DM_MEMBER_STATE_REMOVED`

The `data` field carries:

```json
{
  "group_id": "<uuid v4>",
  "mls_commit": "<base64url(proto.Marshal(MLSMessage[Commit]))>",
  "removed_peer_id": "<base58>",
  "epoch": "<uint64; new epoch after commit>",
  "rekey_decision": { "<GroupDmRekeyDecision fields>" }
}
```

After the removed member's epoch is invalidated, they MUST NOT receive new
messages. If they were offline during removal, the relay MUST discard any
queued messages from before the epoch change once the new epoch is established.

### 4.5 `groupdm.send`

Message delivery to all group members. The sender:

1. Encrypts the message using MLS `ApplicationMessage` (AES-128-GCM with the
   MLS application secret, per `12-mode-tree.md §4.2`).
2. Opens individual streams to each online member and sends `groupdm.send`.
3. For offline members, uses `relay.store` with the ciphertext-only payload.

`PeerStreamRequest.payload = proto.Marshal(GroupDmSenderKeyEnvelope)` with:
- `group_id` = the group's UUID
- `sender_id` = sender's peer ID
- `epoch` = current MLS epoch
- `ciphertext` = AES-128-GCM ciphertext of the ApplicationMessage
- `signature` = hybrid signature over `group_id || epoch || nonce || ciphertext`
- `replay_protection_tag` = `SHA-256(group_id || epoch || nonce)` hex-encoded

The `data` field carries:

```json
{
  "group_id": "<uuid v4>",
  "message_id": "<uuid v4>",
  "mls_application_message": "<base64url(proto.Marshal(MLSMessage[ApplicationMessage]))>",
  "epoch": "<uint64>",
  "nonce": "<base64url(12-byte random nonce)>",
  "created_at": "<RFC3339Nano>"
}
```

### 4.6 `groupdm.leave`

A member who wants to leave sends a `Proposal(Remove)` for themselves to the
group admin. The admin processes the proposal, issues a `Commit`, and
distributes it via `groupdm.remove` to all remaining members.

`PeerStreamRequest.payload = proto.Marshal(GroupDmMembershipEvent)` with:
- `event_type = GROUP_DM_MEMBERSHIP_EVENT_TYPE_LEAVE`
- `member_id` = the leaving peer's own ID
- `current_state = GROUP_DM_MEMBER_STATE_MEMBER`
- `next_state = GROUP_DM_MEMBER_STATE_LEFT`

The `data` field carries:

```json
{
  "group_id": "<uuid v4>",
  "mls_proposal": "<base64url(proto.Marshal(MLSMessage[Proposal(Remove)]))>",
  "leaving_peer_id": "<base58>"
}
```

The admin MUST process the `Proposal(Remove)` and issue a `Commit` within 30
seconds. If the admin is offline, any other admin-capable member MAY do so.

### 4.7 Rekey decision (`GroupDmRekeyDecision`)

After any membership change (add, remove, leave), a `GroupDmRekeyDecision`
MUST be computed and distributed. The fields map directly to the proto message:

| Field | Value |
|-------|-------|
| `trigger` | `MEMBER_ADDED`, `MEMBER_REMOVED`, or `MEMBER_LEFT` |
| `mandatory` | Always `true` for membership changes |
| `rejoin_allowed` | `false` for removed members; `true` for voluntarily left members |
| `reason` | `REQUIRED_MEMBERSHIP_CHANGE` |
| `removed_epoch` | Epoch before the commit |
| `current_epoch` | Epoch after the commit |
| `rekey_completed` | `true` once the Commit has been processed by all members |

### 4.8 Growth decision (`GroupDmGrowthDecision`)

Before every `groupdm.add`, the admin MUST compute a `GroupDmGrowthDecision`:

| `current_members + adding_members` | `reason` | `allowed` | `convert_to_server` |
|-----------------------------------|----------|-----------|---------------------|
| ≤ 45 | `OK` | `true` | `false` |
| 46–50 | `WARNING` | `true` | `false` |
| > 50 | `CONVERT_REQUIRED` | `false` | `true` |

When `convert_to_server = true`, the admin MUST present a conversion wizard to
the user before proceeding. The `GroupDmConvertPlan` nested message carries
the deterministic wizard steps:

```json
{
  "create_server": true,
  "create_initial_channel": true,
  "migrate_member_list": true,
  "post_system_notice": true,
  "history_transferable": false,
  "disclosure_code": "GDMC-2026"
}
```

`history_transferable` is `false` because MLS epoch keys are not exportable to
server Tree/Crowd mode. Members see pre-conversion history only in their local
store.

## 5. Security mode binding

All operations in this family MUST carry `security_mode = SECURITY_MODE_TREE`.

MLS session state (TreeKEM, epoch secrets, application secrets) MUST be stored
in SQLCipher (not in memory only). Per `12-mode-tree.md §2.5`, an explicit
epoch rotation MUST be triggered after:

- Any membership change (add, remove, leave).
- 1000 messages in the current epoch.
- 7 days since the last Commit.

The removed member's leaf node MUST be blanked in the ratchet tree immediately.
The removed member's key material MUST be securely deleted from all remaining
members' stores.

Relay nodes MUST enforce opacity: `groupdm.send` payloads stored via
`relay.store` MUST be ciphertext-only. A relay that receives a plaintext
`GroupDmSenderKeyEnvelope.ciphertext` (detected by zero-length or missing
integrity tag) MUST return `RELAY_OPACITY_VIOLATION`.

## 6. State persistence

### 6.1 `group_dms` bucket

Each group DM is a `GroupDMRecord`:

```json
{
  "id": "<uuid v4; group_id>",
  "name": "<display name>",
  "creator_peer_id": "<base58>",
  "is_group_dm": true,
  "security_mode": "tree",
  "mls_group_id": "<base64url(16 bytes)>",
  "current_epoch": "<uint64>",
  "member_ids": ["<base58>", "..."],
  "admin_peer_ids": ["<base58>"],
  "created_at": "<RFC3339Nano>",
  "last_message_at": "<RFC3339Nano or null>",
  "scope_type": "groupdm"
}
```

### 6.2 `identities` bucket — MLS group sub-key

MLS group state is stored under `"group_dm_mls/<group_id>"`:

```json
{
  "group_id": "<uuid v4>",
  "ciphersuite": "0xFF01",
  "epoch": "<uint64>",
  "tree_hash": "<base64url(32 bytes)>",
  "confirmed_transcript_hash": "<base64url(32 bytes)>",
  "resumption_secret": "<base64url(32 bytes)>",
  "application_secret": "<base64url(32 bytes)>",
  "mls_state_blob": "<base64url(opaque MLS group state)>"
}
```

### 6.3 `messages` bucket

Each delivered message is a `MessageRecord` with `scope_type = "groupdm"`:

```json
{
  "id": "<message_id>",
  "scope_type": "groupdm",
  "scope_id": "<group_id>",
  "sender_peer_id": "<base58>",
  "body": "<decrypted plaintext, stored encrypted by SQLCipher>",
  "created_at": "<RFC3339Nano>",
  "mls_epoch": "<uint64>"
}
```

## 7. Error codes

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Peer does not advertise `cap.group-dm` or `mode.tree` |
| `MODE_INCOMPATIBLE` | `security_mode` in request is not `SECURITY_MODE_TREE` |
| `RELAY_OPACITY_VIOLATION` | Relay received non-encrypted GroupDM payload |
| `SIGNATURE_MISMATCH` | MLS signature or hybrid delivery signature verification failed |
| `REPLAY_DETECTED` | `replay_protection_tag` already seen for this group + epoch |
| `OPERATION_FAILED` | MLS group state error, epoch mismatch, or member not found |
| `RATE_LIMITED` | Sender exceeded per-group message rate limit (max 30 per minute) |
| `UNSUPPORTED_VERSION` | MLS ciphersuite `0xFF01` not supported by peer |

## 8. Conformance

Conformance class: **W4** (GroupDM family conformance).

KATs in `pkg/spectest/groupdm/`:

- `groupdm_create.json` — creator with 3 initial members; all Welcome messages
  accepted; all produce matching `GroupDMRecord`.
- `groupdm_add.json` — admin adds a 4th member; existing members advance epoch;
  new member processes Welcome; rekey decision emitted.
- `groupdm_remove.json` — admin removes a member; remaining members process
  Commit; removed member's keys deleted; epoch advanced.
- `groupdm_send.json` — sender encrypts ApplicationMessage; all online members
  decrypt; offline member receives via relay (ciphertext only).
- `groupdm_leave.json` — member sends Proposal(Remove); admin issues Commit;
  member state transitions to `LEFT`.
- `groupdm_growth_warning.json` — add to 47-member group; `WARNING` returned.
- `groupdm_growth_convert.json` — add that would exceed 50; `CONVERT_REQUIRED`
  returned with `GroupDmConvertPlan`.
- `groupdm_epoch_rotation.json` — 1000-message epoch rotation trigger; Commit
  issued without membership change.
- `groupdm_relay_opacity.json` — plaintext body rejected by relay with
  `RELAY_OPACITY_VIOLATION`.

Implementations MUST pass all KATs in `pkg/spectest/groupdm/` to claim GroupDM
family conformance.
