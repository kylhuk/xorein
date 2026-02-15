# Capability and Security Model Specification (Discord-Parity, Secure-by-Default)

## 1. Scope

This specification defines (a) security modes (“models”) and (b) capability requirements for a Discord-like platform with a Discord-like user experience (servers, channels, threads, voice/video, streaming), while providing stronger default confidentiality for content.

This specification distinguishes between:

* Content: message bodies, attachments, media frames, user profile fields.
* Metadata: routing identifiers, server/channel membership, timestamps, presence/typing, notification triggers, rate-limit signals, and abuse-prevention telemetry.

The system SHALL provide end-to-end encryption (E2EE) for content by default where feasible, and SHALL explicitly label any capability that is not E2EE.

## 2. Normative Language

The key words “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, and “MAY” are to be interpreted as normative requirements.

## 3. Threat Model (Baseline)

1. The Service (including relays, storage/CDN, SFUs, TURN relays) SHALL be treated as untrusted for content confidentiality.
2. Endpoints (client devices) SHALL be trusted to perform cryptography correctly; compromise of an endpoint is out-of-scope for confidentiality guarantees.
3. The system SHALL minimize metadata where feasible; however, the system SHALL NOT claim to hide all metadata without additional anonymity infrastructure (not specified here).

## 4. Cryptographic Requirements (Baseline)

### 4.1 Post-quantum requirements (control plane)

1. The system SHALL support post-quantum key establishment using ML‑KEM (FIPS 203). ([NIST Computer Security Resource Center][1])
2. The system SHALL support post-quantum digital signatures using ML‑DSA (FIPS 204). ([NIST Computer Security Resource Center][2])
3. The default PQ parameter sets SHOULD be ML‑KEM‑768 and ML‑DSA‑65 (security/performance balance; implementer-defined naming is acceptable as long as it maps unambiguously to FIPS 203/204 parameter sets). ([NIST Computer Security Resource Center][1])

### 4.2 Hybrid requirement (interoperability + robustness)

1. Session establishment SHOULD be hybrid (classical + PQ) to avoid single-family dependency (e.g., X25519 + ML‑KEM‑768).
2. Implementations SHALL be crypto-agile (algorithm identifiers, versioning, and rekey support) to allow upgrades without breaking protocol state.

### 4.3 Symmetric requirements (data plane)

1. All content encryption (messages, attachments, media frames, segments) SHALL use AEAD with ≥128-bit security; 256-bit keys SHOULD be used to provide margin against Grover-type considerations.
2. The chosen AEAD, hash, and KDF primitives SHALL be standardized and widely implemented; exact selection is a profile decision (e.g., AES‑GCM or ChaCha20‑Poly1305; HKDF).

## 5. Security Modes (Models)

Each capability SHALL declare exactly one of these modes as its default, and MAY allow an explicit downgrade/upgrade if specified.

### 5.1 Seal (PQ‑Ratchet DM)

Definition: 1:1 E2EE using hybrid authenticated key establishment (X25519 + ML‑KEM‑768) and per-message forward-secure symmetric ratcheting; identity/device keys SHALL be signed with ML‑DSA‑65. ([NIST Computer Security Resource Center][1])
Service visibility: ciphertext + routing metadata only.
Security properties (content): confidentiality, integrity, endpoint authentication, forward secrecy, post-compromise recovery.
Limitations: does not hide traffic metadata; cannot protect against compromised endpoints.

### 5.2 Tree (MLS Interactive Group)

Definition: interactive group E2EE using MLS (RFC 9420) for group key agreement and state updates. ([IETF Datatracker][3])
Service visibility: ciphertext + group routing/membership metadata required for delivery (unless an additional privacy layer is introduced; not specified).
Security properties (content): confidentiality, membership authentication, forward secrecy, post-compromise security (PCS), scalable rekey ~log(N). ([IETF Datatracker][3])
PQ note: MLS is standardized; fully PQ-native MLS profiles may evolve. The system SHOULD provide PQ protection for join/rekey material via Seal-style hybrid mechanisms until a PQ-MLS profile is adopted.

### 5.3 Crowd (Sender‑Epoch Room)

Definition: room E2EE using a symmetric room “epoch key” for message encryption plus per-sender symmetric sender keys for message authentication; epoch keys rotate on a schedule and on moderation actions.
Service visibility: ciphertext + room metadata.
Security properties (content): confidentiality from the Service; authenticity for senders; “epoch-level” forward secrecy (upon rotation).
Limitations: revocation is rotation-based (a removed member MAY retain access until the next rotation); PCS is weaker than Tree.

### 5.4 Channel (Broadcast‑Epoch)

Definition: broadcast E2EE for “few writers / many readers”, using epoch keys with scheduled and event-driven rotation; writers SHALL be authenticated via sender keys and/or periodic signatures.
Service visibility: ciphertext + channel metadata.
Security properties (content): confidentiality from the Service at very large scale; strong writer authenticity; efficient fanout.
Limitations: revocation is rotation-based; not intended for high-churn “everyone can write” semantics.

### 5.5 MediaShield (E2EE Media via SFU)

Definition: real-time voice/video/screen-share using SFU/TURN for scale while encrypting media end-to-end using SFrame (RFC 9605) or an equivalent E2EE media frame mechanism; SFU can access routing metadata but SHALL NOT access frame plaintext. ([IETF Datatracker][4])
Operational note: TURN MAY be required when direct connectivity is not possible. ([IETF Datatracker][5])
Client API note (web): the system MAY use WebRTC Encoded Transform for insertable encryption in compatible environments. ([W3C][6])

### 5.6 StreamShield (Segment‑E2EE Streaming)

Definition: near-real-time one-to-many streaming using AEAD-encrypted segments/chunks distributed via relays/CDN; keys are delivered via Seal/Tree/Crowd/Channel.
Security properties (content): relays/CDN cannot read stream content.
Limitations: server-side transcoding and preview thumbnails SHALL NOT be possible without breaking strict E2EE (unless done on trusted endpoints).

### 5.7 Clear (Server‑Readable)

Definition: content is readable by the Service (still protected by transport security).
Requirement: Clear mode SHALL be explicitly labeled and SHALL NOT be the default for private conversations.

## 6. Capability Specifications (Discord-Parity Surfaces)

### 6.1 Identity, Profile, and Relationship Privacy

Capabilities:

* Account identifier (routing ID)
* Contact list / friend graph
* User profile (display name, avatar, bio)
* Server nicknames / room display identity

Requirements:

1. The Service SHALL maintain a stable routing identifier per account for delivery.
2. User profile fields (name/avatar/bio) SHOULD be stored as encrypted blobs and SHALL be decryptable only by authorized peers (e.g., contacts) when “private identity” is enabled.
3. For non-contacts in large rooms/channels, the client SHOULD display a room-scoped pseudonym unless the user explicitly opts into broader profile disclosure.
4. Profile updates SHALL be end-to-end encrypted and authenticated.

### 6.2 Direct Messages (DMs)

Default mode: Seal.
Required features (Discord-like):

* Message send/receive, replies, edits, deletes, mentions, emojis/reactions, typing indicators, read states.

Requirements:

1. DM content SHALL be E2EE under Seal.
2. Typing indicators, presence, and read states SHOULD be treated as metadata and MAY be Clear metadata (explicitly disclosed).
3. “Message Requests” (spam screening) MAY exist as a metadata policy surface similar to Discord’s DM filtering. ([Discord Support][7])

### 6.3 Small Groups (interactive)

Default mode: Tree.

Scaling guidance (normative defaults):

* Up to 200 members: Tree SHALL be supported as default.
* 200–1000 members: Tree SHOULD be used when churn and device constraints permit; otherwise Crowd MAY be used as a fallback.
* 1000–5000 members: Crowd SHOULD be default.
* > 5000 members: Channel semantics (or sharded rooms) SHOULD be default.

(These thresholds are implementation policy knobs; the UI SHALL disclose when the system uses Crowd/Channel instead of Tree.)

### 6.4 Servers (Guilds), Categories, Channels, Threads, Forums

Discord-parity targets include threads and forum channels. ([Discord Support][8])

Requirements:

1. Server/channel structure metadata (names, ordering, membership, permissions) MAY be Clear metadata for scalability and administration.
2. Content inside a channel SHALL use the channel’s declared mode: Tree, Crowd, Channel, or Clear.
3. Threads and forum posts SHALL inherit the parent channel’s security mode by default. ([Discord Support][8])
4. Pins SHALL be supported; the pinned message content SHALL remain encrypted if the channel is not Clear. ([Discord Support][9])

### 6.5 Roles, Permissions, and Moderation Actions

Discord-parity targets include roles/permissions and permission overwrites. ([Discord Support][10])

Requirements:

1. Permissions enforcement SHALL be performed by the Service (authorization) and/or verifiable client-side checks.
2. Permission state changes (role changes, kicks, bans, timeouts, channel permission changes) SHOULD be authenticated end-to-end (e.g., signed by authorized moderator devices) so clients can detect unauthorized server-side tampering.
3. Moderation actions that require content inspection SHALL NOT be available in strict E2EE channels unless users explicitly provide plaintext evidence.

### 6.6 Reactions, Stickers, Rich UI Events

Discord supports reactions and “super reactions”. ([Discord Support][11])

Requirements:

1. Reaction events SHOULD be E2EE in Seal/Tree/Crowd/Channel.
2. Aggregate reaction counts MAY be computed client-side in E2EE channels; if server-side counts are provided, that channel SHALL be labeled as leaking reaction metadata.

### 6.7 Attachments and Media Files

Discord supports file attachments and enforces upload limits. ([Discord Support][12])

Requirements:

1. Attachments SHALL be client-side encrypted prior to upload in Seal/Tree/Crowd/Channel.
2. The Service MAY store ciphertext and enforce size/type/rate limits.
3. Server-side malware scanning and content-based policies SHALL require Clear mode or explicit user-submitted plaintext scanning.

### 6.8 Search

Discord provides server-side search UX. ([Discord Support][13])

Requirements:

1. In Seal/Tree/Crowd/Channel, the Service SHALL NOT be able to provide plaintext full-text indexing.
2. The system SHOULD provide client-side search over locally cached/decrypted history.
3. If server-side search is provided, the channel SHALL be Clear and labeled accordingly.

### 6.9 Scheduled Events

Discord supports scheduled events. ([Discord Support][14])

Requirements:

1. Event metadata (title/time/RSVP) MAY be Clear metadata.
2. Event chat content SHALL follow the event channel’s declared security mode.

### 6.10 Bots, Webhooks, Integrations

Discord supports webhooks and bots broadly. ([Discord Support][15])

Requirements:

1. In strict E2EE channels, bots/webhooks SHALL NOT have implicit access to plaintext.
2. “Bot Channels” MAY exist as Clear channels with explicit labeling.
3. If a bot is granted plaintext access by users, the UI SHALL treat the bot as an endpoint (with an explicit trust warning).

## 7. Voice, Video, and Streaming (Discord-Parity)

### 7.1 Voice (real-time)

Default mode: MediaShield.

Requirements:

1. Voice channels SHALL use SFU/TURN as needed for scale; TURN MAY be required for connectivity. ([IETF Datatracker][5])
2. Media payloads SHALL be E2EE under MediaShield (SFrame or equivalent). ([IETF Datatracker][4])
3. The system SHOULD support verification codes / “privacy code” style UX for calls (optional but recommended for parity with modern expectations). Discord provides call verification UX and requires E2EE A/V for most call surfaces by March 1, 2026, while excluding Stage channels from E2EE A/V. ([Discord Support][16])

### 7.2 Video Calls (real-time)

Default mode: MediaShield (small groups).

Requirements:

1. Video SHALL use MediaShield.
2. The system SHALL enforce practical caps for interactive video. Discord currently caps server video chat at 25 video participants. ([Discord Support][17])
3. 4K60 support MAY be offered only where endpoints and network conditions permit; it SHALL NOT be claimed as a universal capability.

### 7.3 Screen Sharing (interactive)

Default mode: MediaShield.

Requirements:

1. Screen share SHALL be available in voice calls/channels using MediaShield.
2. The system SHOULD support multiple concurrent sharers with caps; Discord allows up to 50 users sharing video or screen simultaneously in a voice chat. ([Discord Support][18])

### 7.4 Streaming (Twitch-like, buffering allowed)

Default mode: StreamShield for E2EE streaming; Clear for maximum compatibility/transcoding.

Requirements:

1. For strict E2EE streaming, the system SHALL use StreamShield (encrypted segments) and SHALL NOT rely on server-side transcoding.
2. Stream previews/thumbnails for non-participants SHALL NOT be E2EE unless explicitly designed as an endpoint-decrypted feature; Discord explicitly states Go Live stream previews are NOT end-to-end encrypted. ([Discord Support][16])
3. 4K120 “near real time” MAY require significant buffering and is generally incompatible with broad device support; it SHOULD be specified as best-effort only.

## 8. Size-to-Mode Classification (Required Disclosure)

The UI and documentation SHALL disclose the effective security mode for each room/channel.

| Use case                             | Default mode                                         | Notes (security semantics)                                                                 |
| ------------------------------------ | ---------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Private chat messages                | Seal                                                 | Strongest content guarantees (FS + recovery); metadata remains visible to Service          |
| Group messages (≤50)                 | Tree                                                 | Strong interactive group guarantees (MLS) ([IETF Datatracker][3])                          |
| Group messages (≤200)                | Tree                                                 | Same as above; depends on churn/device                                                     |
| Group messages (≤1000)               | Tree (preferred) / Crowd (fallback)                  | Crowd is rotation-based revocation                                                         |
| Group messages (≤5000)               | Crowd                                                | Epoch rotation required; not per-join rekey                                                |
| Group messages (>5000)               | Channel (preferred) / sharded Crowd                  | Broadcast semantics recommended                                                            |
| Channels (few writers, many readers) | Channel                                              | Best for large audiences; rotation-based revocation                                        |
| Voice chat (HQ, real time)           | MediaShield                                          | E2EE media with SFU/TURN ([IETF Datatracker][4])                                           |
| Video calls (HQ, real time)          | MediaShield (small)                                  | Practical caps required; 4K60 is best-effort                                               |
| Screen sharing (near real time)      | MediaShield (interactive) / StreamShield (broadcast) | StreamShield if buffering allowed                                                          |
| Stream previews                      | Clear metadata/content                               | If previews exist, they SHALL be labeled (Discord-style precedent) ([Discord Support][16]) |

## 9. User-Facing Labels (Mandatory)

Every conversation surface SHALL display a mode label:

* “E2EE (Seal)”
* “E2EE (Tree / MLS)”
* “E2EE (Crowd: rotation-based)”
* “E2EE (Channel: rotation-based)”
* “E2EE Media (MediaShield)”
* “E2EE Stream (StreamShield)”
* “Not E2EE (Clear)”

For rotation-based modes (Crowd/Channel), the UI SHALL disclose: “Removed members may retain access until the next key rotation.”

## 10. Discord Parity Checklist (Capabilities to Include)

The product spec/paper SHOULD explicitly address these Discord-like capabilities and declare their security mode behavior:

* Servers/guilds, categories, channels (text/voice/stage), invites
* Roles/permissions, permission overwrites ([Discord Support][10])
* Threads, forum channels ([Discord Support][8])
* DMs, group DMs, message requests ([Discord Support][7])
* Pins ([Discord Support][9])
* Reactions (incl. “super reactions”) ([Discord Support][11])
* Attachments and upload limits ([Discord Support][12])
* Search ([Discord Support][13])
* Scheduled events ([Discord Support][14])
* Auto-moderation / content filters ([Discord Support][19])
* Webhooks/bots/integrations ([Discord Support][15])
* Voice/video/screen share limits and scaling expectations ([Discord Support][17])
* Stream previews and whether they leak content ([Discord Support][16])

---

### Reference URLs (copy/paste)

```text
NIST FIPS 203 (ML-KEM): https://csrc.nist.gov/pubs/fips/203/final
NIST FIPS 204 (ML-DSA): https://csrc.nist.gov/pubs/fips/204/final
RFC 9420 (MLS): https://datatracker.ietf.org/doc/rfc9420/
RFC 9750 (MLS Architecture): https://www.rfc-editor.org/rfc/rfc9750.html
RFC 9605 (SFrame): https://datatracker.ietf.org/doc/rfc9605/
RFC 8656 (TURN): https://datatracker.ietf.org/doc/rfc8656/
W3C WebRTC Encoded Transform: https://www.w3.org/TR/webrtc-encoded-transform/
Discord E2EE A/V (policy + exclusions): https://support.discord.com/hc/en-us/articles/25968222946071-End-to-End-Encryption-for-Audio-and-Video
Discord Go Live / Screen Share: https://support.discord.com/hc/en-us/articles/360040816151-Go-Live-and-Screen-Share
Discord Video Calls (25 cap): https://support.discord.com/hc/en-us/articles/360041721052-Video-Calls
Discord Account/Server Caps: https://support.discord.com/hc/en-us/articles/33694251638295-Discord-Account-Caps-Server-Caps-and-More
Discord Threads FAQ: https://support.discord.com/hc/en-us/articles/4403205878423-Threads-FAQ
Discord Forum Channels FAQ: https://support.discord.com/hc/en-us/articles/6208479917079-Forum-Channels-FAQ
Discord Roles and Permissions: https://support.discord.com/hc/en-us/articles/214836687-Discord-Roles-and-Permissions
Discord Search: https://support.discord.com/hc/en-us/articles/115000468588-How-to-Use-Search-on-Discord
Discord AutoMod: https://support.discord.com/hc/en-us/articles/4421269296535-AutoMod-FAQ
Discord Webhooks: https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks
Discord Scheduled Events: https://support.discord.com/hc/en-us/articles/4409494125719-Scheduled-Events
Discord Pins: https://support.discord.com/hc/en-us/articles/221421867-How-do-I-pin-messages
Discord Attachments: https://support.discord.com/hc/en-us/articles/25444343291031-File-Attachments-FAQ
Discord Reactions: https://support.discord.com/hc/en-us/articles/12102061808663-Reactions-and-Super-Reactions-FAQ
Discord Message Requests: https://support.discord.com/hc/en-us/articles/7924992471191-Message-Requests
```

[1]: https://csrc.nist.gov/pubs/fips/203/final?utm_source=chatgpt.com "FIPS 203, Module-Lattice-Based Key-Encapsulation ..."
[2]: https://csrc.nist.gov/pubs/fips/204/final?utm_source=chatgpt.com "FIPS 204, Module-Lattice-Based Digital Signature Standard"
[3]: https://datatracker.ietf.org/doc/rfc9420/?utm_source=chatgpt.com "RFC 9420 - The Messaging Layer Security (MLS) Protocol"
[4]: https://datatracker.ietf.org/doc/rfc9605/?utm_source=chatgpt.com "RFC 9605 - Secure Frame (SFrame): Lightweight ..."
[5]: https://datatracker.ietf.org/doc/rfc8656/?utm_source=chatgpt.com "RFC 8656 - Traversal Using Relays around NAT (TURN)"
[6]: https://www.w3.org/TR/webrtc-encoded-transform/?utm_source=chatgpt.com "WebRTC Encoded Transform"
[7]: https://support.discord.com/hc/en-us/articles/7924992471191-Message-Requests?utm_source=chatgpt.com "Message Requests"
[8]: https://support.discord.com/hc/en-us/articles/4403205878423-Threads-FAQ?utm_source=chatgpt.com "Threads FAQ"
[9]: https://support.discord.com/hc/en-us/articles/221421867-How-do-I-pin-messages?utm_source=chatgpt.com "How do I pin messages?"
[10]: https://support.discord.com/hc/en-us/articles/214836687-Discord-Roles-and-Permissions?utm_source=chatgpt.com "Discord Roles and Permissions"
[11]: https://support.discord.com/hc/en-us/articles/12102061808663-Reactions-and-Super-Reactions-FAQ?utm_source=chatgpt.com "Reactions and Super Reactions FAQ"
[12]: https://support.discord.com/hc/en-us/articles/25444343291031-File-Attachments-FAQ?utm_source=chatgpt.com "File Attachments FAQ"
[13]: https://support.discord.com/hc/en-us/articles/115000468588-How-to-Use-Search-on-Discord?utm_source=chatgpt.com "How to Use Search on Discord"
[14]: https://support.discord.com/hc/en-us/articles/4409494125719-Scheduled-Events?utm_source=chatgpt.com "Scheduled Events"
[15]: https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks?utm_source=chatgpt.com "Intro to Webhooks"
[16]: https://support.discord.com/hc/en-us/articles/25968222946071-End-to-End-Encryption-for-Audio-and-Video "End-to-End Encryption for Audio and Video – Discord"
[17]: https://support.discord.com/hc/en-us/articles/360041721052-Video-Calls?utm_source=chatgpt.com "Video Calls"
[18]: https://support.discord.com/hc/en-us/articles/360040816151-Go-Live-and-Screen-Share?utm_source=chatgpt.com "Go Live and Screen Share"
[19]: https://support.discord.com/hc/en-us/articles/4421269296535-AutoMod-FAQ?utm_source=chatgpt.com "AutoMod FAQ"
