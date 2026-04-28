# 15 — Mode: MediaShield (SFrame Voice/Screen-Share E2EE)

MediaShield provides end-to-end encryption for real-time media (voice, video,
screen-share) using [RFC 9605 SFrame](https://www.rfc-editor.org/rfc/rfc9605).
It is layered on top of the parent scope's group key material.

## 1. Overview

MediaShield is not a standalone mode. It is activated when a voice or
screen-share session is started within a server or DM that already has a
security mode (Seal, Tree, Crowd, or Channel). The parent mode provides the
key material; MediaShield applies frame-level AEAD to each media frame before
it reaches any intermediary (SFU or relay).

## 2. Key derivation

### 2.1 From Tree mode (MLS exporter)

When the parent scope is Tree mode:

```
mediashield_key = MLS-Exporter(
    label   = "xorein/mediashield/v1",
    context = b"",
    length  = 32,
)
```

Called once per MLS epoch per member. When the MLS epoch changes (member
join/leave), `mediashield_key` MUST be rotated immediately before the next
frame is sent.

### 2.2 From Crowd/Channel mode (HKDF)

When the parent scope is Crowd or Channel:

```
mediashield_key_for_peer_p = HKDF-SHA-256(
    IKM  = crowd_sender_key_p,          // the sender's crowd sender key for this epoch
    salt = b"",
    info = "xorein/mediashield/v1/peer/" || peer_id_p,
    L    = 32,
)
```

Each participant derives their own MediaShield key from their own sender key.
Key rotation follows the Crowd/Channel epoch rotation.

### 2.3 From Seal mode (DM voice)

When the parent scope is a Seal-mode DM:

```
mediashield_key = HKDF-SHA-256(
    IKM  = current_double_ratchet_message_key,
    salt = b"",
    info = "xorein/mediashield/v1/seal-dm/" || session_id,
    L    = 32,
)
```

The Double Ratchet advances for each key derivation, providing per-session
forward secrecy.

## 3. SFrame framing (RFC 9605)

Xorein uses SFrame as specified in [RFC 9605](https://www.rfc-editor.org/rfc/rfc9605)
with the following locked parameters:

| Parameter | Value |
|-----------|-------|
| AEAD algorithm | AES-128-GCM |
| Key length | 16 bytes (derived from 32-byte MediaShield key, first 16 bytes) |
| Nonce length | 12 bytes (derived as `HKDF(mediashield_key, frame_counter, "nonce", 12)`) |
| Header format | RFC 9605 §4 (SFrame header with KID and CTR fields) |
| KID (Key ID) | Peer's 8-byte truncated SHA-256(peer_id) |
| CTR (frame counter) | 64-bit, per-sender, monotonically increasing |
| Maximum CTR before rotation | 2^48 - 1 |

### 3.1 Frame encryption

```
key    = mediashield_key[0:16]   // AES-128-GCM key
nonce  = HKDF-SHA-256(mediashield_key, frame_counter_bytes, "xorein/mediashield/v1/nonce", 12)
header = SFrame.EncodeHeader(kid=peer_kid, ctr=frame_counter)
aad    = header || rtp_header_bytes   // per RFC 9605 §4.2
ciphertext = AES-128-GCM.Seal(key, nonce, aad, media_frame_bytes)
sframe_payload = header || ciphertext
```

### 3.2 Frame counter management

Each sender maintains an independent per-KID frame counter starting at 0.
The counter MUST be monotonically increasing. Reception of a frame with a
counter ≤ the last accepted counter from the same sender MUST be rejected
(replay protection). Counter rollover above 2^48 - 1 triggers an immediate key
rotation request.

### 3.3 Frame decryption

```
(header, ciphertext) = SFrame.ParseFrame(sframe_payload)
key  = lookup_mediashield_key(header.kid)
nonce = HKDF-SHA-256(key, header.ctr_bytes, "xorein/mediashield/v1/nonce", 12)
aad  = header || rtp_header_bytes
plaintext = AES-128-GCM.Open(key, nonce, aad, ciphertext)
// verify ctr > last_ctr[kid]; else reject (replay)
```

## 4. WebRTC integration

### 4.1 SFU topology

For voice sessions with more than 2 participants, Xorein uses a
**Selective Forwarding Unit (SFU)** topology:

- One participant acts as the SFU coordinator (typically the session initiator
  or the relay node with `cap.voice`).
- The SFU receives encrypted SFrame payloads from all participants and
  forwards them to all other participants without decryption.
- SFrame encryption ensures the SFU cannot read media content.

For 2-participant voice (DM voice), a direct peer-to-peer WebRTC connection
is used (no SFU).

### 4.2 Signaling

WebRTC signaling (SDP offer/answer, ICE candidates) is carried over the Xorein
voice family stream (`/aether/voice/0.1.0`). See `52-family-voice.md` for the
full signaling operation table.

### 4.3 Codec negotiation

Supported codecs in descending preference:

| Codec | Purpose | MIME type |
|-------|---------|-----------|
| Opus | Audio (required) | `audio/opus` |
| AV1 | Video (optional) | `video/AV1` |
| VP8 | Video (optional, fallback) | `video/VP8` |

Implementations MUST support Opus. Video is optional in v0.1.

## 5. Key rotation events

MediaShield keys MUST be rotated when:

- Parent MLS epoch changes (Tree mode).
- Crowd/Channel epoch changes.
- A participant joins or leaves the voice session.
- Frame counter approaches 2^48 (48-bit threshold).

On rotation, new keys are derived from the updated parent key material. A
brief "key transition" period (max 2 seconds) allows in-flight frames with the
old key to be received and decrypted.

## 6. Voice session lifecycle

```
1. Initiator sends voice.offer (SDP) to participants via the voice family.
2. Participants respond with voice.answer (SDP).
3. ICE candidate exchange via voice.ice operations.
4. WebRTC connection established; SFrame keys derived from parent scope.
5. Media flows: voice.frame operations carry SFrame-encrypted payloads.
6. On participant leave: key rotation initiated by session coordinator.
7. Session ends: all MediaShield keys and WebRTC state discarded.
```

## 7. Security properties

| Property | MediaShield |
|----------|-------------|
| Frame confidentiality | Yes (AES-128-GCM per frame) |
| Frame integrity | Yes (AES-128-GCM auth tag) |
| Sender authentication | Yes (KID mapped to peer; hybrid sig on Delivery) |
| Replay protection | Yes (frame counter monotonicity) |
| SFU opacity | Yes (SFU cannot decrypt frames) |
| Forward secrecy | Per-epoch (parent mode provides key material) |
| Store-now-decrypt-later | Resistant (key derived from PQ-protected parent) |

## 8. Conformance (W4)

KATs in `pkg/spectest/mediashield/`:

- `sframe_encrypt_decrypt.json` — SFrame AES-128-GCM encrypt/decrypt per RFC 9605 §4.
  Uses RFC 9605 test vectors where applicable.
- `key_derivation_tree.json` — MLS-Exporter → MediaShield key.
- `key_derivation_crowd.json` — Crowd sender key → MediaShield key via HKDF.
- `nonce_derivation.json` — per-frame nonce from MediaShield key + counter.
- `counter_rollover.json` — rotation triggered at 2^48 counter.
- `replay_rejection.json` — frame with counter ≤ previous rejected.
- Integration: 2-participant voice with relay — relay sees only SFrame ciphertext.
