# 49 — Family: Notify (`/aether/notify/0.2.0`)

The Notify family delivers in-app notification events and @mention tokens to
peers. Notifications are ephemeral attention signals — they surface new DMs,
reactions, friend requests, and @mentions. They are not encrypted message
bodies; they are lightweight metadata pointers that direct a peer to fetch the
full content via the appropriate family.

## 1. Overview

Notifications solve the offline-visibility problem: when a peer is temporarily
unreachable, a reduced-metadata notification is queued via the relay so the
peer knows something happened when it reconnects. The security rule (§5) limits
what can be carried in a relay-stored notification.

The @mention sub-protocol (`notify.mention`) is a separate operation because
mention delivery is scoped to specific peer IDs and requires `cap.mentions` in
addition to `cap.notify`. Non-mention notifications do not require `cap.mentions`.

**Roles that participate:**

| Role | Participation |
|------|--------------|
| client | Sender and recipient of all notify operations |
| relay | Stores reduced-payload notifications for offline recipients; MUST NOT inspect body_preview |
| bootstrap | Does not participate |
| archivist | Does not participate |

**Security mode:** Notification operations are metadata-level and carry no
encrypted content (the full message body is fetched separately via the DM or
chat family). For non-Clear scopes, `body_preview` MUST itself be encrypted
(see §4.5). The `security_mode` field in `PeerStreamRequest` for this family
MUST match the scope's security mode so the responder can validate it.

## 2. Capability requirements

| Capability | Role |
|-----------|------|
| `cap.notify` | Required on both initiator and responder for `notify.deliver` and `notify.ack` |
| `cap.mentions` | Required additionally for `notify.mention` |
| `cap.peer.delivery` | Required on responders accepting direct delivery |
| `cap.peer.relay` | Required on relay nodes storing offline notifications |

## 3. Operations

| Operation | Required caps (initiator) | Direction | Description |
|-----------|--------------------------|-----------|-------------|
| `notify.deliver` | `cap.notify` | sender → recipient | Deliver a `NotificationEvent` to the recipient |
| `notify.mention` | `cap.notify`, `cap.mentions` | sender → mentioned peer | Deliver a `MentionToken` + notification to a @mentioned peer |
| `notify.ack` | `cap.notify` | recipient → sender (or relay) | Acknowledge receipt; prevents re-delivery |

## 4. Wire format details

### 4.1 `NotificationEvent` fields

`proto.Marshal(NotificationEvent)` with:

| Field | Type | Notes |
|-------|------|-------|
| `kind` | NotificationKind | `DM`, `GROUP_DM`, `FRIEND`, `PRESENCE`, `MODERATION`, `SYSTEM` |
| `actor_id` | string | peer ID of the peer who triggered the notification |
| `subject_id` | string | scope ID (dm_record_id, group_id, server_id, etc.) |
| `unread` | bool | `true` if this notification is unread at generation time |
| `unread_count` | uint32 | total unread count in the scope at generation time |
| `source_id` | string | message ID, delivery ID, or event ID that caused the notification |
| `source_url` | string | optional deep-link into the scope (e.g. `xorein://dm/<scope_id>`) |
| `mode_epoch_id` | string | MLS epoch ID or Seal session ID; used to validate decryption context |

The `data` JSON field carries additional delivery context not in the proto:

```json
{
  "notification_id": "<uuid v4>",
  "type": "<mention|dm|reaction|system|friend_request>",
  "scope_type": "<dm|groupdm|channel|server|friend>",
  "scope_id": "<uuid v4 or peer ID>",
  "from_peer_id": "<base58>",
  "body_preview": "<encrypted preview, max 100 bytes encoded; see §4.5>",
  "created_at_unix_ms": 1234567890000,
  "delivery_attempt": 1
}
```

### 4.2 `body_preview` encryption and truncation

**Rule**: `body_preview` MUST be truncated to at most 100 UTF-8 bytes of
plaintext before encryption. If the plaintext is longer, truncate at a
UTF-8-safe boundary at or before byte 100 and append `"…"` (U+2026, 3 bytes)
if the truncation point is before the end.

**Encryption** (for non-Clear scopes):

```
preview_key = HKDF-SHA-256(
    IKM  = current_scope_encryption_key,
    salt = b"",
    info = "xorein/notify/v1/preview-key",
    L    = 32,
)
nonce = random(12)
ciphertext = ChaCha20-Poly1305.Seal(
    key   = preview_key,
    nonce = nonce,
    aad   = UTF-8(notification_id),
    pt    = truncated_preview_bytes,
)
body_preview = base64url_no_padding(nonce || ciphertext)
```

Where `current_scope_encryption_key` is:
- For **Seal** (DM): the current Double Ratchet message key, derived
  from the active chain key for the DM session.
- For **Tree** (GroupDM, server): the MLS application secret for the current
  epoch.
- For **Crowd/Channel** (server): the current epoch sender key.
- For **Clear**: `body_preview` is the plaintext truncated preview (no
  encryption applied).

**Relay fallback**: when a notification is stored in the relay queue for an
offline recipient, `body_preview` MUST be omitted entirely from the stored
payload. The relay stores only: `notification_id`, `type`, `scope_type`,
`scope_id`, `from_peer_id`, `created_at_unix_ms`, `delivery_attempt`. The
`body_preview` field is stripped before the payload is handed to `relay.store`.

### 4.3 `notify.deliver` request payload

`PeerStreamRequest.payload = proto.Marshal(NotificationEvent)`.

Full JSON for the request `data` field (direct delivery):

```json
{
  "notification_id": "<uuid v4>",
  "type": "dm",
  "scope_type": "dm",
  "scope_id": "<dm_record_id>",
  "from_peer_id": "<base58>",
  "body_preview": "<base64url(nonce || ciphertext)>",
  "created_at_unix_ms": 1234567890000,
  "delivery_attempt": 1
}
```

Relay-stored payload (all preview removed):

```json
{
  "notification_id": "<uuid v4>",
  "type": "dm",
  "scope_type": "dm",
  "scope_id": "<dm_record_id>",
  "from_peer_id": "<base58>",
  "created_at_unix_ms": 1234567890000,
  "delivery_attempt": 1
}
```

### 4.4 `notify.mention` — @mention delivery

When a sender's message contains one or more @mentions, the sender MUST send a
`notify.mention` to each mentioned peer ID. This is a separate stream from the
message delivery itself.

`PeerStreamRequest.payload = proto.Marshal(MentionToken)` with:
- `type` = `USER`, `ROLE`, `EVERYONE`, or `HERE`
- `raw` = the raw text of the mention token (e.g. `"@alice"`, `"@everyone"`)

The `data` field carries:

```json
{
  "mention_id": "<uuid v4>",
  "mention_type": "<user|role|everyone|here>",
  "target_peer_id": "<base58; for USER mentions>",
  "target_role_id": "<uuid v4; for ROLE mentions>",
  "display_text": "<string; the display text of the mention>",
  "scope_id": "<uuid v4; the channel or DM scope where the mention occurred>",
  "scope_type": "<channel|groupdm|dm>",
  "source_message_id": "<uuid v4; the message containing the mention>",
  "from_peer_id": "<base58>",
  "created_at_unix_ms": 1234567890000
}
```

**Mention detection** is performed by the sender's client BEFORE sending the
message via the chat or DM family. The sender MUST:

1. Parse the message body for `@<token>` patterns.
2. For each `@user` mention: resolve to a peer ID from the server's member list
   or friends list.
3. For `@everyone` or `@here`: send `notify.mention` to all server members (or
   all online members for `@here`).
4. Send `notify.mention` to each resolved peer ID via the Notify family.
5. Send the actual message via the chat/DM family (order is not mandated; both
   SHOULD be sent within the same 100ms window).

`@everyone` and `@here` mention delivery MAY be rate-limited by the server's
role policy (see `cap.moderation`). If the sender's role does not permit
`@everyone`, the server's local policy MUST suppress the `notify.mention`
delivery and the sender's client MUST display an authorization error.

### 4.5 `notify.ack` — acknowledgement

After a recipient processes a notification (marks it read, or fetches the
full content via the appropriate family), it MUST send `notify.ack` to prevent
re-delivery on the next connection.

`PeerStreamRequest.payload` is empty. The `data` field carries:

```json
{
  "notification_id": "<uuid v4>",
  "acknowledged_at_unix_ms": 1234567890000
}
```

The responder (the original sender, or the relay if the notification was relay-
stored) records the acknowledgement and removes the notification from the
pending delivery queue.

### 4.6 Notification routing decision

Before delivering a notification, the sender MUST evaluate the recipient's
current presence state:

| Recipient state | Delivery path | `body_preview` |
|----------------|---------------|----------------|
| Online (direct) | `notify.deliver` directly | Included (encrypted) |
| Offline or unreachable | `relay.store` + `notify.deliver` on reconnect | Stripped (not included) |

If the relay is unavailable, the notification is queued locally in the sender's
`pending_notifications` in-memory list and retried on the next connection.

## 5. Security mode binding

The `security_mode` field in `PeerStreamRequest` for `notify.deliver` and
`notify.mention` MUST match the scope's negotiated security mode:

- DM notifications: `SECURITY_MODE_SEAL`
- GroupDM notifications: `SECURITY_MODE_TREE`
- Server channel notifications (Crowd/Channel): `SECURITY_MODE_CROWD` or
  `SECURITY_MODE_CHANNEL`
- Server channel notifications (Clear): `SECURITY_MODE_CLEAR`

A mismatch MUST result in `MODE_INCOMPATIBLE`. This prevents an attacker from
downgrading a Seal-scoped notification to Clear mode to extract `body_preview`.

For relay-stored notifications, `body_preview` MUST be stripped (see §4.2
relay fallback rule). A relay that receives a notification with `body_preview`
set for a non-Clear scope MUST strip the field before storing. It MUST NOT
return an error — stripping is silent.

## 6. State persistence

Notifications are **ephemeral** and are not persisted to the SQLCipher store
beyond the relay queue TTL.

The sender maintains a local in-memory `pending_notifications` list:

```
pendingNotification {
    NotificationID  string
    RecipientID     string
    Payload         []byte   // relay-stripped proto bytes
    CreatedAt       time.Time
    ExpiresAt       time.Time  // max 24 hours
    Attempts        int
    LastAttemptAt   time.Time
}
```

Acknowledged notifications are removed from the list immediately. Unacknowledged
notifications are retried up to 3 times with exponential backoff (initial: 5s,
max: 60s). After `ExpiresAt`, unacknowledged notifications are discarded.

Relay queue TTL for stored notifications: 24 hours. Notifications in the relay
queue that have passed their TTL MUST be discarded on the next relay cleanup
cycle.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `MISSING_REQUIRED_CAPABILITY` | Peer does not advertise `cap.notify` (or `cap.mentions` for `notify.mention`) |
| `MODE_INCOMPATIBLE` | `security_mode` in request does not match the scope's security mode |
| `REPLAY_DETECTED` | `notification_id` already acknowledged for this scope |
| `OPERATION_FAILED` | Malformed payload, missing required fields |
| `RATE_LIMITED` | Sender has exceeded the notification rate limit |

Rate limits:

| Operation | Limit |
|-----------|-------|
| `notify.deliver` | 60 per peer per minute |
| `notify.mention` | 10 per message (each mention to a distinct target counts separately) |
| `@everyone` / `@here` | Subject to server role policy; see moderation family |

## 8. Conformance

Conformance class: **W7** (Notify family conformance).

KATs in `pkg/spectest/notify/`:

- `notify_deliver_direct.json` — online recipient; `body_preview` included and
  encrypted; verify round-trip.
- `notify_deliver_relay.json` — offline recipient; `body_preview` stripped;
  verify relay stores metadata-only payload.
- `notify_deliver_clear.json` — Clear-mode scope; `body_preview` is plaintext
  (no encryption); verify `SECURITY_MODE_CLEAR` in request.
- `notify_mention_user.json` — `@user` mention; verify `MentionToken` and
  notification delivered to target peer ID.
- `notify_mention_everyone.json` — `@everyone`; verify broadcast to all
  server member IDs.
- `notify_mention_role.json` — `@role`; verify delivery to all peers with
  the target role ID.
- `notify_ack.json` — recipient acks; verify notification removed from
  pending queue; second delivery attempt suppressed.
- `notify_replay.json` — duplicate `notification_id`; verify `REPLAY_DETECTED`.
- `notify_mode_mismatch.json` — DM notification sent with
  `SECURITY_MODE_CLEAR`; verify `MODE_INCOMPATIBLE`.
- `notify_preview_truncation.json` — message body > 100 bytes; verify
  `body_preview` truncated at UTF-8 boundary with `…` appended.
- `notify_preview_encryption.json` — known scope key; verify
  `body_preview = base64url(nonce || ChaCha20-Poly1305(preview_key, nonce, notification_id, plaintext))`.
- `notify_expiry.json` — unacknowledged notification past 24-hour TTL;
  verify discarded from pending queue.

Implementations MUST pass all KATs in `pkg/spectest/notify/` to claim Notify
family conformance.
