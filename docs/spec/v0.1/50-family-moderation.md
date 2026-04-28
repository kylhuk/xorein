# 50 — Family: Moderation (`/aether/moderation/0.2.0`)

This document specifies the Moderation family, which provides server-level
moderation actions (kick, ban, mute, slow-mode, message deletion) for Xorein
servers.

## 1. Overview

The Moderation family enables authorized nodes to enforce server-level
behavioural policies: membership removal (kick/ban), timed mute, channel
slow-mode rate limiting, and remote message deletion.

All moderation actions MUST be performed only by nodes whose `peer_id` appears
in the server's moderator list as determined by the Governance family
(`51-family-governance.md`). The root of trust for the moderator list is the
server manifest's RBAC section, which the Governance family maintains.

**Protocol ID:** `/aether/moderation/0.2.0`

**Required capability:** `cap.moderation`

**Roles that use this family:** client (moderator-or-above role), relay
(forwarding only, no action processing).

**Security modes:** Moderation events MUST be delivered inside the server's
active security mode. `ModerationEvent` and `SlowModeDecision` payloads are
hybrid-signed (Ed25519 + ML-DSA-65) by the acting moderator before being
broadcast as `Delivery` objects to all server members.

## 2. Capability requirements

| Capability | Role | Meaning |
|------------|------|---------|
| `cap.moderation` | Required | Enables all Moderation family operations |
| `cap.slow-mode` | Required for `moderation.mute` and `moderation.slow_mode` | Slow-mode enforcement support |
| `cap.rbac` | Implicit | Moderator list lookup; the receiver MUST verify actor role via the Governance family |

All capability strings are defined in `pkg/protocol/capabilities.go`. A node
that does not advertise `cap.moderation` MUST return
`MISSING_REQUIRED_CAPABILITY` and close the stream.

## 3. Operations

| Operation | Required caps | Request payload | Response payload | Description |
|-----------|--------------|-----------------|------------------|-------------|
| `moderation.kick` | `cap.moderation` | `ModerationRequest` | `ModerationResponse` | Remove a member from the server; broadcast `ModerationEvent` (action=`REDACT` lifecycle) to all members |
| `moderation.ban` | `cap.moderation` | `ModerationRequest` | `ModerationResponse` | Permanently ban a peer ID; add to server ban list in manifest; broadcast `ModerationEvent` (action=`BAN`) |
| `moderation.unban` | `cap.moderation` | `ModerationRequest` | `ModerationResponse` | Remove a ban entry; update manifest ban list; broadcast `ModerationEvent` |
| `moderation.mute` | `cap.moderation`, `cap.slow-mode` | `MuteRequest` | `ModerationResponse` | Mute a member for `duration_ms` (cannot send messages); broadcast `ModerationEvent` (action=`TIMEOUT`) |
| `moderation.slow_mode` | `cap.moderation`, `cap.slow-mode` | `SlowModeRequest` | `SlowModeResponse` | Set per-channel minimum delay between messages per sender; broadcast `SlowModeDecision` |
| `moderation.delete_message` | `cap.moderation` | `DeleteMessageRequest` | `ModerationResponse` | Remote-delete a message by ID; broadcast delete `Delivery` to all members |

### 3.1 Authorization rules

Before executing any operation, the receiver MUST:

1. Look up the actor's `RoleState` for the server via the Governance family.
2. Confirm the actor holds role `moderator`, `admin`, or `owner`
   (`V02_ROLE_MODERATOR`, `V02_ROLE_ADMIN`, or `V02_ROLE_OWNER` in the proto).
3. For `moderation.ban` and `moderation.unban`, require at minimum
   `V02_ROLE_ADMIN`.
4. Confirm the actor's role outranks the target's role
   (`rbac.CanActOnTarget(actor, target)` MUST return `true`). A moderator MUST
   NOT kick or ban another moderator, admin, or owner.

If any check fails, the receiver MUST respond with `OPERATION_FAILED` and
reason `POLICY_REASON_MODERATION_FORBIDDEN_TARGET` or
`POLICY_REASON_MODERATION_INVALID_SIGNER` as appropriate, and MUST NOT apply
the action.

### 3.2 `moderation.kick`

Kicks a member from the server:

1. Receiver verifies authorization (§3.1).
2. Receiver removes `target_peer_id` from the server's member list.
3. Receiver constructs and hybrid-signs a `ModerationEvent`
   (`action = MODERATION_ACTION_REDACT`).
4. Receiver broadcasts the signed `ModerationEvent` as a `Delivery` to all
   current server members.
5. Kicked member MUST be notified via the `Delivery` before the server closes
   the connection.

A kicked member MAY rejoin via the server invite flow.

### 3.3 `moderation.ban`

Bans a peer ID from the server:

1. Receiver verifies authorization (§3.1); MUST be `admin` or `owner`.
2. Receiver adds `target_peer_id` to the server ban list
   (`ServerRecord.BannedPeerIDs`).
3. Receiver updates the server manifest to include the ban list and
   re-publishes it (triggering `manifest.publish` on the manifest family).
4. Receiver constructs and hybrid-signs a `ModerationEvent`
   (`action = MODERATION_ACTION_BAN`).
5. Receiver broadcasts the signed `ModerationEvent` as a `Delivery`.

Any node receiving a join request (`chat.join`) from a banned peer ID MUST
reject it with `OPERATION_FAILED` and reason `"banned"`.

### 3.4 `moderation.unban`

Removes a ban:

1. Receiver verifies authorization (§3.1); MUST be `admin` or `owner`.
2. Receiver removes `target_peer_id` from `ServerRecord.BannedPeerIDs`.
3. Receiver updates and re-publishes the server manifest.
4. Receiver constructs and broadcasts a `ModerationEvent`.

### 3.5 `moderation.mute`

Mutes a member for a finite duration:

1. Receiver verifies authorization (§3.1).
2. Receiver records a mute entry for `target_peer_id`: expiry =
   `now + duration_ms`.
3. While a valid mute entry exists, the receiver MUST reject any
   `chat.send` from `target_peer_id` with `OPERATION_FAILED` and reason
   `"muted"`.
4. Receiver constructs and hybrid-signs a `ModerationEvent`
   (`action = MODERATION_ACTION_TIMEOUT`, `duration_ms` set).
5. Receiver broadcasts the `ModerationEvent` as a `Delivery`.

Mute state is ephemeral (in-memory only); it is NOT persisted across restarts.
The maximum permitted `duration_ms` is 2,592,000,000 (30 days).

### 3.6 `moderation.slow_mode`

Sets a per-channel slow-mode delay:

1. Receiver verifies authorization (§3.1).
2. Receiver updates `ChannelRecord.SlowModeMS` for `channel_id`.
3. Receiver constructs and hybrid-signs a `SlowModeDecision`.
4. Receiver broadcasts the `SlowModeDecision` as a `Delivery` to all members.

Setting `min_delay_ms = 0` disables slow-mode for the channel. The maximum
`min_delay_ms` is 21,600,000 (6 hours).

Slow-mode enforcement: each message receipt on a slow-mode channel MUST be
evaluated with `policy.EvaluateSlowMode()`. A message arriving before the
actor's `last_activity_unix + min_delay_ms` MUST be rejected with
`POLICY_REASON_SLOW_MODE_ACTIVE`. Nodes with role `moderator` or above
receive a bypass (`POLICY_REASON_SLOW_MODE_BYPASS`).

### 3.7 `moderation.delete_message`

Remote-deletes a message:

1. Receiver verifies authorization (§3.1).
2. Receiver marks the message `Deleted = true` in `MessageRecord`.
3. Receiver broadcasts a delete `Delivery` with `kind = "delete"` and
   `scope_id = message_id` to all server members.
4. Receiver constructs and hybrid-signs a `ModerationEvent`
   (`action = MODERATION_ACTION_DELETE`).

Receiving nodes MUST update their local `MessageRecord` to `Deleted = true`
when they receive the delete `Delivery`. Deleted message bodies MUST NOT be
surfaced to users.

## 4. Wire format details

All request and response payloads are JSON-encoded and placed in
`PeerStreamRequest.payload` / `PeerStreamResponse.payload`.

### 4.1 Common request envelope: `ModerationRequest`

```
{
  "actor_peer_id":  string,   // peer ID of the moderator performing the action
  "target_peer_id": string,   // peer ID of the subject
  "scope_id":       string,   // server ID (or channel ID for channel-scoped ops)
  "reason":         string,   // optional human-readable reason (max 512 bytes)
  "signature":      string    // base64url hybrid sig over canonical JSON (excl. this field)
}
```

### 4.2 `MuteRequest`

Extends `ModerationRequest` with:

```
{
  ...,
  "duration_ms": uint64    // mute duration; MUST be > 0 and <= 2,592,000,000
}
```

### 4.3 `SlowModeRequest`

```
{
  "actor_peer_id":   string,
  "channel_id":      string,
  "min_delay_ms":    uint64,    // 0 = disable; max 21,600,000
  "effective_from":  uint64,    // unix milliseconds; MUST be >= now
  "signature":       string
}
```

### 4.4 `DeleteMessageRequest`

```
{
  "actor_peer_id": string,
  "message_id":    string,
  "server_id":     string,
  "channel_id":    string,
  "reason":        string,     // optional
  "signature":     string
}
```

### 4.5 Common response: `ModerationResponse`

```
{
  "accepted": bool,
  "reason":   string,   // PolicyReason string; see proto PolicyReason enum
  "error":    string    // human-readable; absent on success
}
```

### 4.6 `SlowModeResponse`

```
{
  "accepted":        bool,
  "channel_id":      string,
  "min_delay_ms":    uint64,
  "effective_from":  uint64,
  "error":           string
}
```

### 4.7 `ModerationEvent` (proto message)

Defined in `proto/aether.proto` as `message ModerationEvent`:

| Field | Type | Notes |
|-------|------|-------|
| `actor_id` | string | Peer ID of the moderator |
| `target_id` | string | Peer ID of the subject |
| `action` | `ModerationAction` | `REDACT`, `DELETE`, `TIMEOUT`, or `BAN` |
| `allowed` | bool | `true` if the action was permitted |
| `reason` | `PolicyReason` | Outcome classification |
| `event_id` | string | UUID v4; idempotency nonce |
| `signer_id` | string | Peer ID of the signing node |
| `signature_algorithm` | `SignatureAlgorithm` | MUST be `HYBRID_ED25519_ML_DSA_65` |
| `signature` | bytes | Ed25519 signature (32 bytes) |
| `verification_status` | `VerificationStatus` | `ACCEPTED` or `REJECTED` |
| `enforcement_required` | bool | Receiver MUST apply the action if `true` |
| `non_compliance_reason` | string | Populated when `verification_status = REJECTED` |

### 4.8 `SlowModeDecision` (proto message)

Defined in `proto/aether.proto` as `message SlowModeDecision`:

| Field | Type | Notes |
|-------|------|-------|
| `allowed` | bool | Whether the triggering message was permitted |
| `reason` | `PolicyReason` | `SLOW_MODE_ACTIVE`, `SLOW_MODE_BYPASS`, or `SLOW_MODE_PASS` |
| `actors` | repeated `SlowModeActorState` | Per-actor last-activity timestamps |
| `scope_id` | string | Channel ID |
| `event_id` | string | UUID v4 |
| `actor_id` | string | Peer ID of the sender being evaluated |
| `evaluated_at_unix` | uint64 | Unix seconds |
| `replay` | bool | `true` if this is a replay-reconciliation event |
| `reconciled` | bool | `true` if divergent slow-mode state was resolved |

`SlowModeDecision` is broadcast to all members when slow-mode settings change.
Per-message slow-mode evaluations are NOT broadcast; they are local decisions.

## 5. Security mode binding

Moderation family operations MUST execute within the server's active security
mode. The action-level payload (request JSON) MUST be encrypted under the
server scope key before being placed in `PeerStreamRequest.payload`, matching
the security mode requirements of the channel family.

The broadcast `ModerationEvent` and `SlowModeDecision` `Delivery` objects MUST
be encrypted under the server's active security mode (Seal, Tree, Crowd, or
Channel) and MUST carry a valid hybrid signature.

Clear mode MUST NOT be used for moderation event delivery in servers that offer
non-Clear security modes. A server configured for Clear mode MAY deliver
`ModerationEvent` without payload encryption, but the hybrid signature MUST
still be present.

Relay nodes MUST forward moderation `Delivery` objects without decrypting them,
upholding the relay opacity invariant (`40-family-peer.md §4.4`).

## 6. State persistence

| State bucket | Key | Value | Description |
|-------------|-----|-------|-------------|
| `servers` | `server_id` | `ServerRecord` (JSON) | Contains `BannedPeerIDs []string` and per-`ChannelRecord.SlowModeMS uint64` |

Go types from `pkg/node/types.go`:

- `ServerRecord.Members []string` — current member peer IDs; `kick` removes
  the target.
- `ServerRecord.BannedPeerIDs []string` — ban list; persisted in the `servers`
  bucket and reflected in the server manifest.
- `ChannelRecord.SlowModeMS uint64` — per-channel minimum delay in
  milliseconds; 0 = disabled.

Mute entries are ephemeral (in-memory `map[string]time.Time`). They are NOT
written to the persistent state store.

The ban list MUST be included in the server manifest's canonical JSON so that
joining peers receive it during the join flow.

## 7. Error codes

The following `PeerStreamError.code` values apply to the Moderation family, in
addition to the generic codes in `02-canonical-envelope.md §1.3`:

| Code | Trigger |
|------|---------|
| `MODERATION_UNAUTHORIZED` | Actor's role does not permit the requested action |
| `MODERATION_FORBIDDEN_TARGET` | Actor cannot act on the target due to role precedence |
| `MODERATION_MISSING_SIGNATURE` | `ModerationEvent` signature field absent or empty |
| `MODERATION_INVALID_SIGNER` | Signer peer ID not in moderator list |
| `MODERATION_INVALID_SIGNATURE` | Hybrid signature verification failed |
| `SLOW_MODE_INVALID_DURATION` | `min_delay_ms` exceeds 21,600,000 or `duration_ms` exceeds 2,592,000,000 |
| `BAN_LIST_FULL` | Server ban list has reached the implementation limit (max 4096 entries) |
| `MESSAGE_NOT_FOUND` | `moderation.delete_message` target message ID unknown to this node |

## 8. Conformance

Implementations claiming Moderation family conformance MUST pass the following
known-answer tests (KATs) in `pkg/spectest/moderation/`:

| KAT file | Covers |
|----------|--------|
| `kick_kat.json` | Successful kick; verify member removed and `ModerationEvent` broadcast |
| `ban_unban_kat.json` | Ban + unban round-trip; manifest ban list update |
| `mute_kat.json` | Mute for 5 seconds; message rejected during mute, accepted after |
| `slow_mode_kat.json` | Set slow-mode to 2000 ms; second message within window rejected; third accepted |
| `delete_message_kat.json` | `moderation.delete_message`; verify `Deleted = true` on receiver |
| `unauthorized_kat.json` | Actor with `member` role attempts kick; expect `MODERATION_UNAUTHORIZED` |
| `forbidden_target_kat.json` | Moderator attempts to kick an admin; expect `MODERATION_FORBIDDEN_TARGET` |
| `signature_mismatch_kat.json` | Tampered `ModerationEvent` signature; expect `MODERATION_INVALID_SIGNATURE` |

Each KAT MUST include the full `PeerStreamRequest` and `PeerStreamResponse`
serialized bytes in the format defined by `90-conformance-harness.md`.

A relay-role conformance suite MUST additionally demonstrate that relay nodes
forward moderation `Delivery` objects without inspecting or decrypting the
payload.
