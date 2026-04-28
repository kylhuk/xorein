# 52 — Family: Voice (`/aether/voice/0.1.0`)

This document specifies the Voice family, which carries WebRTC signaling
(SDP offer/answer, ICE candidates) and MediaShield-encrypted media frames for
voice and optional video in Xorein.

## 1. Overview

The Voice family provides:

- **Voice channel presence:** join, leave, and mute-state signaling.
- **WebRTC signaling:** SDP offer/answer and ICE candidate exchange between
  participants.
- **MediaShield media delivery:** per-frame SFrame-encrypted audio/video
  payloads (see `15-mode-mediashield.md`).
- **Session lifecycle management:** ICE restart, session termination.

**Protocol ID:** `/aether/voice/0.1.0`

**Required capability:** `cap.voice`

**MediaShield requirement:** `voice.join` and `voice.frame` additionally require
`mode.mediashield`. A node without `mode.mediashield` MUST NOT be admitted to a
voice session.

**Roles that use this family:** client only. Relay nodes MUST NOT receive
`voice.frame` operations (§5, relay opacity rule for voice). Bootstrap and
archivist nodes MUST NOT participate in voice sessions.

**Topology:** Two-participant sessions use a direct peer-to-peer WebRTC
connection. Sessions with three or more participants use an SFU (Selective
Forwarding Unit) topology with a deterministically elected SFU coordinator
(§3.12).

## 2. Capability requirements

| Capability | Meaning |
|------------|---------|
| `cap.voice` | Required for all Voice family operations |
| `mode.mediashield` | Required for `voice.join` and `voice.frame`; implies SFrame key availability |

Capability strings are defined in `pkg/protocol/capabilities.go`. A node that
does not advertise `cap.voice` MUST return `MISSING_REQUIRED_CAPABILITY` and
close the stream. A node that does not advertise `mode.mediashield` MUST NOT
be sent `voice.join` or `voice.frame` operations.

## 3. Operations

| Operation | Required caps | Direction | Request payload | Response payload | Description |
|-----------|--------------|-----------|-----------------|------------------|-------------|
| `voice.join` | `cap.voice`, `mode.mediashield` | initiator → all members | `VoiceJoinRequest` | `VoiceJoinResponse` | Join a voice channel; broadcast `VoiceState` (joined) to all members |
| `voice.leave` | `cap.voice` | initiator → all members | `VoiceLeaveRequest` | `VoiceLeaveResponse` | Leave the voice channel; broadcast `VoiceState` (left) |
| `voice.mute` | `cap.voice` | initiator → all members | `VoiceMuteRequest` | `VoiceMuteResponse` | Toggle mute or deafen state; broadcast `VoiceState` update |
| `voice.offer` | `cap.voice` | initiator → responder | `VoiceSignalRequest` (type=OFFER) | `VoiceSignalResponse` | WebRTC SDP offer |
| `voice.answer` | `cap.voice` | responder → initiator | `VoiceSignalRequest` (type=ANSWER) | `VoiceSignalResponse` | WebRTC SDP answer |
| `voice.ice` | `cap.voice` | either direction | `VoiceSignalRequest` (type=ICE_CANDIDATE) | `VoiceSignalResponse` | ICE candidate (one per operation call) |
| `voice.ice_complete` | `cap.voice` | either direction | `VoiceSignalRequest` (type=ICE_COMPLETE) | `VoiceSignalResponse` | ICE gathering complete; no more candidates will be sent |
| `voice.frame` | `cap.voice`, `mode.mediashield` | sender → peer (or SFU coordinator) | `VoiceFrameRequest` | `VoiceFrameResponse` | MediaShield-encrypted SFrame payload |
| `voice.restart` | `cap.voice` | either direction | `VoiceSignalRequest` (type=RESTART) | `VoiceSignalResponse` | Request WebRTC ICE restart |
| `voice.terminate` | `cap.voice` | either direction | `VoiceSignalRequest` (type=TERMINATE) | `VoiceSignalResponse` | End the session |

All operations use the peer-stream envelope from `02-canonical-envelope.md §1`.
Voice family streams are point-to-point (initiator ↔ responder or initiator ↔
SFU coordinator). Each stream carries exactly one request-response pair.

### 3.1 `voice.join`

Joining a voice channel:

1. The joining member (initiator A) opens voice streams to all current members
   of the voice channel and sends `voice.join` containing A's `VoiceState`
   (`muted = false`, `deafened = false`).
2. Each receiving member updates its local `VoiceSession.Participants` with A's
   `VoiceParticipant` record.
3. Each receiving member broadcasts a `chat.send` delivery with
   `kind = "voice_state"` to notify the channel of the new participant.
4. If the channel has ≥ 1 existing participant, A MUST proceed immediately to
   the WebRTC signaling flow (§3.9) with each existing member.

A node that does not advertise `mode.mediashield` MUST be refused with
`VOICE_NOT_AUTHORIZED`.

### 3.2 `voice.leave`

Leaving a voice channel:

1. The leaving member sends `voice.leave` to all current session members.
2. Each receiving member removes the leaver from `VoiceSession.Participants`.
3. If the session drops to 0 participants, `VoiceSession` MUST be deleted from
   state.
4. If the leaver was the SFU coordinator and ≥ 2 participants remain, a new SFU
   coordinator MUST be elected deterministically (§3.12) and the remaining
   members MUST complete a new WebRTC signaling round with the new coordinator.
5. Key rotation MUST be triggered (per `15-mode-mediashield.md §5`) after a
   participant leaves.

### 3.3 `voice.mute`

Toggling mute or deafen state:

1. The member sends `voice.mute` with `muted` and `deafened` boolean fields.
2. Each receiving member updates the sender's `VoiceParticipant` record.
3. A deafened member MUST NOT send `voice.frame` operations.
4. A muted member MUST NOT send audio `voice.frame` operations. Video frames
   MAY continue.

### 3.4 `voice.offer`

WebRTC SDP offer from the initiator:

1. Initiator A opens a voice stream to peer B.
2. A sends `VoiceSignalRequest` with `signal_type = VOICE_SIGNAL_TYPE_OFFER`
   and `payload_bytes` containing the SDP offer JSON (see §4.3).
3. B MUST respond synchronously with `VoiceSignalResponse` containing
   `accepted = true` (SDP offer stored; `voice.answer` sent on a new stream).

### 3.5 `voice.answer`

WebRTC SDP answer from the responder:

1. After receiving `voice.offer`, B opens a new voice stream back to A.
2. B sends `VoiceSignalRequest` with `signal_type = VOICE_SIGNAL_TYPE_ANSWER`
   and `payload_bytes` containing the SDP answer JSON.
3. A responds with `VoiceSignalResponse{accepted: true}`.

### 3.6 `voice.ice`

ICE candidate exchange (one candidate per operation call):

1. Either A or B sends `VoiceSignalRequest` with
   `signal_type = VOICE_SIGNAL_TYPE_ICE_CANDIDATE` and `payload_bytes`
   containing the ICE candidate JSON (see §4.4).
2. The receiver adds the candidate to its local WebRTC peer-connection ICE
   candidate list.
3. The receiver responds with `VoiceSignalResponse{accepted: true}`.

Implementations MUST support receiving ICE candidates before and after the
remote SDP is set (trickle ICE). Candidates MUST be queued until the remote
description is available.

### 3.7 `voice.ice_complete`

ICE gathering complete:

1. A node that has finished gathering ICE candidates sends
   `VoiceSignalRequest` with `signal_type = VOICE_SIGNAL_TYPE_ICE_COMPLETE`
   and empty `payload_bytes`.
2. The remote node SHOULD use this signal to stop waiting for additional
   candidates.

### 3.8 `voice.frame`

MediaShield-encrypted media frame delivery:

1. Sender constructs an SFrame-encrypted payload per
   `15-mode-mediashield.md §3.1`.
2. Sender sends `VoiceFrameRequest` to the receiver (direct P2P) or to the SFU
   coordinator (§3.12).
3. In SFU mode, the coordinator forwards the received `VoiceFrameRequest` bytes
   unmodified to all other participants. The coordinator MUST NOT decrypt or
   inspect `sframe_payload`.
4. Each legitimate receiver decrypts per `15-mode-mediashield.md §3.3` and
   verifies the frame counter is strictly greater than the last accepted counter
   for this sender KID (replay protection).

Relay nodes MUST NOT receive `voice.frame` operations. Any relay node that
receives a `voice.frame` MUST respond with `RELAY_OPACITY_VIOLATION` and close
the stream.

### 3.9 WebRTC signaling flow (two participants)

```
A wants to connect to B:

1. A → B: voice.join          (VoiceState: joined)
2. A → B: voice.offer         (SDP offer)
3. B → A: voice.answer        (SDP answer)
4. A → B: voice.ice           (ICE candidate, repeat for each candidate)
4. B → A: voice.ice           (ICE candidate, repeat for each candidate)
5. A → B: voice.ice_complete  (A finished gathering)
5. B → A: voice.ice_complete  (B finished gathering)
6. WebRTC connection established.
7. SFrame key derived from parent security mode
   per 15-mode-mediashield.md §2.
8. A ↔ B: voice.frame         (ongoing; SFrame-encrypted per frame)
```

All signaling frames MUST be hybrid-signed by the sender (see §4.2,
`VoiceSignalFrame.signature`).

### 3.10 WebRTC signaling flow (three or more participants — SFU mode)

When ≥ 3 participants are present:

1. The SFU coordinator is elected deterministically (§3.12).
2. Each non-SFU member performs the offer/answer/ICE flow with the SFU
   coordinator only (not with every other member).
3. The SFU coordinator maintains one WebRTC peer-connection per member.
4. Media frames:
   - Each member sends `voice.frame` to the SFU coordinator.
   - The SFU coordinator forwards each received `VoiceFrameRequest` to all
     other participants without modification or decryption.
5. When a new member joins mid-session:
   - The new member performs the offer/answer/ICE flow with the SFU coordinator.
   - Key rotation is triggered (per `15-mode-mediashield.md §5`).

The SFU coordinator MUST uphold the MediaShield opacity invariant at all times:
it MUST NOT decrypt, log, or otherwise inspect `sframe_payload`.

### 3.11 ICE restart

If the WebRTC connection degrades or ICE fails:

1. Either peer sends `voice.restart` (`signal_type = VOICE_SIGNAL_TYPE_RESTART`).
2. The initiating peer generates a new SDP offer with ICE restart parameters.
3. The full offer/answer/ICE exchange repeats (§3.9 or §3.10).
4. Existing `voice.frame` delivery MAY continue with the old connection until
   the new connection is established.

### 3.12 SFU coordinator election

When ≥ 3 participants are present, the SFU coordinator is the participant with
the lexicographically smallest `peer_id` string among all current session
members. This election is deterministic and requires no explicit negotiation.

If the current SFU coordinator sends `voice.leave` or its stream disconnects,
the new coordinator is the participant with the next-smallest `peer_id`. The
remaining members MUST detect this condition (e.g., via `voice.leave` or
connection timeout) and re-establish WebRTC connections to the new coordinator.

## 4. Wire format details

All request and response payloads are JSON-encoded in `PeerStreamRequest.payload`
/ `PeerStreamResponse.payload` unless stated otherwise.

### 4.1 `VoiceJoinRequest`

```
{
  "session_id":     string,   // UUID v4; created by the joining member for this voice channel session
  "channel_id":     string,
  "server_id":      string,
  "peer_id":        string,
  "muted":          bool,
  "deafened":       bool,
  "signature":      string    // base64url hybrid sig over canonical JSON (excl. this field)
}
```

### 4.2 `VoiceSignalRequest`

All signaling operations (`voice.offer`, `voice.answer`, `voice.ice`,
`voice.ice_complete`, `voice.restart`, `voice.terminate`) use this envelope:

```
{
  "signal_id":       string,   // UUID v4; idempotency nonce for this signal
  "session_id":      string,   // voice session UUID
  "signal_type":     string,   // "OFFER" | "ANSWER" | "ICE_CANDIDATE" | "ICE_COMPLETE" | "RESTART" | "TERMINATE"
  "sender_peer_id":  string,
  "target_peer_id":  string,
  "payload_bytes":   string,   // base64url-encoded SDP JSON or ICE candidate JSON; empty for ICE_COMPLETE and TERMINATE
  "sequence":        uint64,   // per-sender monotonically increasing counter
  "sent_at":         uint64,   // unix seconds
  "expires_at":      uint64,   // unix seconds; MUST be > sent_at; max 30 seconds ahead
  "signature":       string    // base64url hybrid sig over canonical JSON (excl. this field)
}
```

Receivers MUST reject signals with `expires_at <= now` at processing time.
Receivers MUST reject signals with `sequence <= last_seen_sequence` from the
same `sender_peer_id` in the same session (replay protection).

### 4.3 SDP JSON format

The `payload_bytes` field for `voice.offer` and `voice.answer` carries
base64url-encoded JSON:

```json
{
  "type": "offer",
  "sdp":  "<RFC 3264 SDP text>"
}
```

For `voice.answer`, `type` MUST be `"answer"`. Implementations MUST parse the
SDP and enforce the codec requirements in §4.6.

### 4.4 ICE candidate JSON format

The `payload_bytes` field for `voice.ice` carries base64url-encoded JSON:

```json
{
  "candidate":       "<ICE candidate string per RFC 8839>",
  "sdp_mid":         "<media stream ID string or null>",
  "sdp_mline_index": 0
}
```

`sdp_mline_index` MUST be a non-negative integer corresponding to the media
line in the SDP. `sdp_mid` MAY be null; implementations MUST handle both forms.

### 4.5 `VoiceFrameRequest`

```
{
  "session_id":      string,   // voice session UUID
  "sender_peer_id":  string,
  "sframe_payload":  string,   // base64url(SFrame header || AES-128-GCM ciphertext) per 15-mode-mediashield.md §3.1
  "timestamp_ms":    uint64,   // sender's local capture timestamp (unix milliseconds)
  "frame_seq":       uint64,   // per-sender monotonically increasing frame counter (same as SFrame CTR)
  "signature":       string    // base64url hybrid sig over canonical JSON (excl. this field and sframe_payload)
}
```

Note: the hybrid signature covers the canonical JSON with `sframe_payload`
excluded, because `sframe_payload` is already AEAD-authenticated inside the
SFrame envelope. Receivers MUST verify the signature before attempting SFrame
decryption.

The maximum size of `sframe_payload` (after base64url decoding) is 65,535 bytes
per frame. Frames exceeding this limit MUST be rejected with `OPERATION_FAILED`.

### 4.6 Codec requirements

Implementations MUST negotiate codecs in SDP according to the following rules:

| Codec | MIME type | Role | Requirement |
|-------|-----------|------|-------------|
| Opus | `audio/opus` | Audio | MUST be offered; MUST be accepted if present in the remote offer |
| AV1 | `video/AV1` | Video | SHOULD be offered; MUST be preferred over VP8 if both are present |
| VP8 | `video/VP8` | Video | MAY be offered as fallback; MUST NOT be the sole video offer if AV1 is supported |

If no common audio codec is found after SDP negotiation, the receiver MUST
reject the offer with `VOICE_CODEC_UNSUPPORTED` and close the stream. Video is
optional in v0.1; a voice-only session with no video codec is valid.

Implementations MUST enable RTCP multiplexing (`a=rtcp-mux`) and MUST support
ICE trickle gathering (`a=ice-options:trickle`).

### 4.7 Common response envelopes

`VoiceJoinResponse`, `VoiceLeaveResponse`, `VoiceMuteResponse`:

```
{
  "accepted":   bool,
  "session_id": string,
  "error":      string    // human-readable; absent on success
}
```

`VoiceSignalResponse`:

```
{
  "accepted":  bool,
  "signal_id": string,    // mirrors VoiceSignalRequest.signal_id
  "error":     string
}
```

`VoiceFrameResponse`:

```
{
  "accepted":  bool,
  "frame_seq": uint64,    // mirrors VoiceFrameRequest.frame_seq
  "error":     string
}
```

### 4.8 `VoiceState` (proto message)

Defined in `proto/aether.proto` as `message VoiceState`:

| Field | Type | Notes |
|-------|------|-------|
| `session_id` | string | Voice session UUID |
| `participant` | `IdentityProfile` | Participant identity |
| `muted` | bool | Audio muted |
| `deafened` | bool | Audio and video muted |
| `updated_at` | uint64 | Unix seconds |

### 4.9 `VoiceSignalFrame` (proto message)

Defined in `proto/aether.proto` as `message VoiceSignalFrame`:

| Field | Type | Notes |
|-------|------|-------|
| `signal_id` | string | UUID v4 |
| `session_ref` | `VoiceSignalSessionRef` | Session and participant identifiers |
| `signal_type` | `VoiceSignalType` | See enum below |
| `sequence` | uint64 | Per-sender monotonic counter |
| `encrypted_payload` | bytes | Encrypted SDP or ICE material (encrypted by channel pipeline) |
| `sent_at` | uint64 | Unix seconds |
| `expires_at` | uint64 | Unix seconds |
| `retry_policy` | `VoiceSignalRetryPolicy` | Bounded retry parameters |

### 4.10 `VoiceSignalType` enum values

Defined in `proto/aether.proto`:

| Value | Integer | Notes |
|-------|---------|-------|
| `VOICE_SIGNAL_TYPE_UNSPECIFIED` | 0 | Invalid; MUST NOT appear in wire messages |
| `VOICE_SIGNAL_TYPE_OFFER` | 1 | SDP offer |
| `VOICE_SIGNAL_TYPE_ANSWER` | 2 | SDP answer |
| `VOICE_SIGNAL_TYPE_ICE_CANDIDATE` | 3 | Trickle ICE candidate |
| `VOICE_SIGNAL_TYPE_ICE_COMPLETE` | 4 | ICE gathering complete |
| `VOICE_SIGNAL_TYPE_RESTART` | 5 | ICE restart request |
| `VOICE_SIGNAL_TYPE_TERMINATE` | 6 | Session termination |

## 5. Security mode binding

Voice sessions are always bound to the parent server or DM's active security
mode. MediaShield key derivation depends on the parent mode:

| Parent mode | Key derivation method |
|-------------|----------------------|
| Tree (MLS) | MLS-Exporter per `15-mode-mediashield.md §2.1` |
| Crowd / Channel | HKDF from sender key per `15-mode-mediashield.md §2.2` |
| Seal (DM voice) | HKDF from Double Ratchet message key per `15-mode-mediashield.md §2.3` |

**Voice signaling** (`voice.offer`, `voice.answer`, `voice.ice`, etc.) MUST
be delivered over encrypted streams (Noise XX hop-to-hop) and MUST carry hybrid
signatures. The `payload_bytes` within signaling frames (SDP, ICE candidates)
is plaintext JSON after hop-to-hop decryption; it is not additionally encrypted
at the application layer because it contains no sensitive keying material
(SFrame keys are derived separately from the parent security mode, not from SDP).

**Voice frames** (`voice.frame`) carry SFrame-encrypted ciphertext. The
`sframe_payload` bytes MUST be opaque to any forwarding node (relay or SFU
coordinator).

**Relay opacity:** Relay nodes (`RoleRelay`) MUST NOT forward `voice.frame`
operations under any circumstances. If a relay node receives `voice.frame`, it
MUST respond with `RELAY_OPACITY_VIOLATION` and close the stream immediately.

**SFU opacity:** The SFU coordinator MUST forward `VoiceFrameRequest.sframe_payload`
bytes unmodified without decryption, logging, or inspection. Violations of this
invariant constitute a protocol error and MUST be treated as a security incident.

## 6. State persistence

| State bucket | Key | Value | Description |
|-------------|-----|-------|-------------|
| `voice` | `channel_id` | `VoiceSession` (JSON) | Active voice session per channel |

Go types from `pkg/node/types.go`:

- `VoiceSession`:
  - `ChannelID string`
  - `Participants map[string]VoiceParticipant` — keyed by `peer_id`
  - `LastFrameBy map[string]time.Time` — last frame received per peer
- `VoiceParticipant`:
  - `PeerID string`
  - `Muted bool`
  - `JoinedAt time.Time`
  - `LastFrameAt time.Time`

`VoiceSession` records are volatile: they SHOULD be written to the `voice` state
bucket for crash recovery but MUST be discarded and rebuilt from `voice.join`
broadcasts on restart. The `VoiceSession` bucket MUST NOT persist SFrame keys
or SDP material.

ICE state, WebRTC connection objects, and pending ICE candidate queues are
runtime-only state and MUST NOT be persisted.

## 7. Error codes

The following `PeerStreamError.code` values apply to the Voice family, in
addition to the generic codes in `02-canonical-envelope.md §1.3`:

| Code | Trigger |
|------|---------|
| `VOICE_CODEC_UNSUPPORTED` | No common codec after SDP negotiation (Opus not found in offer) |
| `VOICE_SESSION_NOT_FOUND` | `session_id` referenced in an operation is unknown to the receiver |
| `VOICE_NOT_AUTHORIZED` | Peer is not a member of the voice channel, or lacks `mode.mediashield` |
| `MEDIASHIELD_KEY_UNAVAILABLE` | Cannot derive MediaShield key (parent security mode not established or epoch missing) |
| `VOICE_SIGNAL_EXPIRED` | `expires_at` in `VoiceSignalRequest` is in the past at processing time |
| `VOICE_SIGNAL_REPLAY` | `sequence` is not greater than the last accepted sequence for this sender |
| `VOICE_FRAME_TOO_LARGE` | Decoded `sframe_payload` exceeds 65,535 bytes |
| `VOICE_SFU_NOT_COORDINATOR` | `voice.frame` sent to a non-SFU peer in an SFU session |
| `RELAY_OPACITY_VIOLATION` | Relay received `voice.frame`; voice frames MUST NOT traverse relay nodes |

## 8. Conformance

Implementations claiming Voice family conformance MUST pass the following KATs:

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/voice/join_leave_kat.json` | `voice.join` + `voice.leave` round-trip; `VoiceSession` state transitions |
| `pkg/spectest/voice/mute_kat.json` | `voice.mute` toggle; deafened member rejected on `voice.frame` |
| `pkg/spectest/voice/offer_answer_kat.json` | SDP offer/answer round-trip; codec list validated |
| `pkg/spectest/voice/ice_kat.json` | Trickle ICE candidate exchange; `voice.ice_complete` terminator |
| `pkg/spectest/voice/frame_kat.json` | `voice.frame` with a valid SFrame payload (using `pkg/spectest/mediashield/` vectors); frame counter replay rejection |
| `pkg/spectest/voice/restart_kat.json` | `voice.restart` triggers new offer/answer round-trip |
| `pkg/spectest/voice/terminate_kat.json` | `voice.terminate`; verify `VoiceSession` removed from state |
| `pkg/spectest/voice/sfu_election_kat.json` | 3-member session; confirm SFU coordinator is lowest `peer_id` |
| `pkg/spectest/voice/sfu_opacity_kat.json` | SFU receives encrypted `voice.frame` and forwards unmodified; SFU cannot produce plaintext from `sframe_payload` |
| `pkg/spectest/voice/relay_opacity_kat.json` | `voice.frame` sent to relay node; expect `RELAY_OPACITY_VIOLATION` |
| `pkg/spectest/voice/codec_unsupported_kat.json` | SDP offer without Opus; expect `VOICE_CODEC_UNSUPPORTED` |
| `pkg/spectest/voice/signal_replay_kat.json` | Replayed `VoiceSignalRequest` with duplicate `sequence`; expect `VOICE_SIGNAL_REPLAY` |
| `pkg/spectest/voice/mediashield_key_unavailable_kat.json` | `voice.join` when parent security mode not established; expect `MEDIASHIELD_KEY_UNAVAILABLE` |

Each KAT MUST include the full `PeerStreamRequest` and `PeerStreamResponse`
serialized bytes in the format defined by `90-conformance-harness.md`.

**SFU opacity conformance note:** The SFU opacity KAT (`sfu_opacity_kat.json`)
MUST demonstrate that the SFU coordinator node cannot produce the plaintext
media frame from the `sframe_payload` it forwards. The test SHOULD verify this
by confirming the SFU node does not hold a valid MediaShield key for the
sender's KID.

**Integration test with `pkg/spectest/mediashield/`:** The `frame_kat.json`
KAT MUST use the SFrame test vectors from `15-mode-mediashield.md §8` to
validate the full encrypt → transmit → decrypt path, including frame counter
management and replay rejection.
