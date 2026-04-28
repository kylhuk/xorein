# 10 — Mode: Clear

## 1. Overview

Clear mode is the explicit plaintext mode. Message bodies are readable by
infrastructure (relay nodes, archivist nodes, server manifest hosts). It is
included for completeness and for use cases where no confidentiality is
required (e.g., public announcement boards).

Clear mode is the security mode of last resort. No Xorein implementation
MUST make it the default.

## 2. Activation rules

### 2.1 DMs in Clear

A DM scope MUST NOT be created in Clear mode. The only path to Clear-mode DM
delivery is:

1. Both participants initiate an explicit mode-change negotiation.
2. Both sides send a signed `mode.changed` delivery with `new_mode="clear"`.
3. Both sides receive and verify the peer's consent before changing the
   `DMRecord.SecurityMode`.
4. The UI MUST display a prominent, persistent label on the DM scope indicating
   Clear mode is active.

### 2.2 Servers in Clear

A server manifest MAY include `"clear"` in `offered_security_modes`. Server
owners MAY create servers with `security_mode="clear"`. The manifest MUST
include a visible label in its description that the channel is unencrypted.

Clients MUST show a persistent warning when sending or viewing messages in a
Clear-mode scope.

### 2.3 Relay behavior

Clear-mode deliveries MUST be accepted by relay nodes for queuing. This is the
only mode where relay nodes may store readable plaintext. Relay nodes MUST NOT
use Clear-mode message content for any purpose other than delivery.

## 3. Wire format

Clear-mode message `Delivery.body` is a UTF-8 string of arbitrary length (max
1 MiB). The `body` field is set directly to the message text. No encryption
is applied.

The `Delivery` is signed with the sender's hybrid identity key per
`02-canonical-envelope.md §3.3`. Signature provides authenticity but not
confidentiality.

## 4. Security properties

| Property | Clear mode status |
|----------|-----------------|
| Confidentiality | None |
| Integrity | Yes (hybrid signature) |
| Sender authentication | Yes (hybrid signature) |
| Forward secrecy | No |
| Relay opacity | No — relay sees plaintext |

## 5. Downgrade prevention

Clear mode MUST NOT be selectable by infrastructure. Specifically:

- A relay node MUST NOT change the security mode of a delivery.
- A bootstrap node MUST NOT influence mode selection.
- A server owner's manifest setting DOES influence mode for server-scoped
  messages (this is intentional and disclosed to members via the manifest).

## 6. Conformance

W5 conformance requires:

- Correct activation sequence: both-party consent for DMs.
- UI label present in test screenshots (manual verification).
- Relay stores Clear-mode delivery without error.
- Relay does NOT store non-Clear delivery with plaintext body
  (`RELAY_OPACITY_VIOLATION` returned).
