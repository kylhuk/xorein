# Aether Implementation Plan — Addendum A

This addendum extends **“Aether Protocol & Platform — Revised Implementation Plan v3”** and specifically closes two adoption blockers: (1) “just works” quality (connectivity, calls, background), and (2) public server discovery + serious moderation. fileciteturn0file0

---

## A1) “Just works” quality (QoL + media + networking)

### A1.1 Non‑negotiable product targets (ship gates)

These are release-blockers (no “we’ll fix it later”):

1) Reliability
- Login-to-ready: ≤ 3s on warm start, ≤ 8s on cold start.
- Message send: ≥ 99.9% delivered within 5s (normal network), with clear UI states otherwise.
- Call setup: ringing within ≤ 2s (p50), ≤ 4s (p95).
- Call survival: network switch (Wi‑Fi ↔ LTE) recovers within ≤ 2s (p95) without user action.

2) Media quality
- Voice: mouth-to-ear ≤ 150ms typical; MOS target ≥ 4.0 on “good” networks.
- Screen share: 1080p30 default, 1080p60 available; first frame ≤ 2s.
- “No surprises”: if quality drops, user sees a simple reason (“Poor uplink”, “Relay in use”).

3) Background + multi-device
- Background delivery: DMs and mentions wake reliably on iOS/Android.
- Seamless device switching: join call from phone while desktop is active; switch audio output without renegotiation when possible.

### A1.2 Connectivity architecture: Connectivity Orchestrator (CO)

Current v3 plan already includes a layered discovery and NAT strategy (AutoNAT, DCUtR hole punching, circuit relay, TURN). This addendum makes it an explicit subsystem with strict SLAs and a “mini‑VPN” mode.

Core idea: treat “reachability” as a continuously-managed session, not a one-off connect.

CO responsibilities
- Detect network changes (IP churn, captive portal, NAT type changes).
- Select path per peer and per modality (chat, file, media) with fast fallback.
- Maintain keepalives and re-dial logic with backoff that is mobile-friendly.
- Provide deterministic diagnostics (reason codes, per-hop metrics).

Path ladder (attempt order)
1) Direct QUIC (UDP) + DCUtR hole punch.
2) Direct TCP (fallback for UDP-blocked networks).
3) **Aether Tunnel (mini‑VPN)**: per-peer encrypted tunnel established over the best available transport (QUIC preferred), multiplexing all streams.
4) Circuit Relay v2 (short-hop) for bootstrap and as a bridge.
5) TURN via community relay for media (WebRTC fallback).

Aether Tunnel (mini‑VPN) definition
- Goal: “if NAT is annoying, stop negotiating, just create a stable overlay path”.
- Implementation: QUIC-based tunnel with connection migration support; multiplexed streams (chat/media signaling/files) over a single session; optional relay hop.
- Security: end-to-end (Noise/MLS-compatible keying), forward secrecy, replay protection.
- Behavior: created opportunistically for peers with repeated ICE failures or unstable NATs; torn down when stable direct paths succeed.

Notes
- For WebRTC media, the tunnel can carry signaling reliably; media still prefers native WebRTC paths. When media cannot connect, CO forces TURN/SFU without user friction.

### A1.3 Media “it just works”: predictable WebRTC behavior

Implementation actions
- Standardize a strict set of codec baselines:
  - Audio: Opus 48 kHz, 10 ms frames, FEC+DTX on.
  - Video/screen: H.264 baseline + VP9; AV1 optional when both ends support.
- Aggressive ICE pre-warming:
  - Pre-gather ICE candidates on server/channel join.
  - Cache working candidate pairs per peer and per network (Wi‑Fi vs LTE).
- Deterministic topology switching:
  - 2: direct P2P
  - 3–8: mesh
  - 9+: SFU (peer-elected or relay SFU)
  - Always allow “force SFU” for users on restrictive networks.
- Audio processing defaults:
  - Use WebRTC AEC/AGC by default.
  - Noise suppression tiered: RNNoise first, upgrade to DTLN later (as already planned), but enforce stable CPU budgets (no thermal runaway).

### A1.4 Background + notifications (mobile reality)

Constraints: iOS/Android require APNs/FCM-style wakeups for reliable background delivery.

Design
- Encrypted Notification Relay (ENR): sends minimal encrypted wake payloads (no content). Relay only knows “device token X gets a ping”.
- Self-hostable + community-run defaults.
- Message content still retrieved P2P after wake; ENR is only a wake mechanism.

Implementation checklist
- iOS: CallKit integration for incoming calls; PushKit/VoIP pushes where applicable.
- Android: foreground service only when in call; otherwise rely on push + periodic job.
- Device sleep handling: keepalives throttled; resume uses CO re-dial.

### A1.5 QoL feature bundle (MVP+)

These are the “users don’t know they need them” items that make the product feel finished.

Messaging QoL
- Fast message actions: edit, reply, quote, copy, react.
- Drafts per channel; persistent unsent text.
- “Jump to new messages” and “mark as read” controls.
- Link unfurling client-side (privacy-preserving; no server fetch).

Calls QoL
- One-click device switching (mic/speaker/camera) with clear indicators.
- “Rejoin” banner on transient disconnects; never leave user in limbo.
- Live quality indicator (simple: Great / OK / Poor) + details panel for power users.

Multi-device QoL
- Session handoff: continue reading/call state on another device.
- Consistent keyboard shortcuts (desktop) and gestures (mobile).

### A1.6 Quality engineering program (how this becomes “perfect”)

Test matrix (must run in CI/nightly)
- NAT matrix: full cone, restricted, port-restricted, symmetric.
- Transport matrix: UDP ok/blocked, TCP only, captive portal, high loss.
- Mobility: IP change mid-call; network switch; background/foreground churn.

Tooling
- testground scenarios expanded specifically for call setup + recovery.
- Synthetic “canary calls” between relays and reference clients across regions.
- On-device performance budgets: CPU, memory, thermal.

Release gates
- No release if call setup p95 regresses > 10% vs previous stable.
- No release if crash-free sessions < 99.5% on supported devices.

Diagnostics (privacy-preserving)
- Local ring-buffer logs with explicit export.
- Optional opt-in anonymous metrics with k-anonymity buckets.

---

## A2) Public server visibility + search + moderation

### A2.1 Goal and constraint

Goal: anyone can discover public servers and request/join; by default servers remain private/invite-only.

Constraint: full-text search over a pure DHT is not practical at scale without indexing. Solution: opt-in directory records + community indexers, all cryptographically verifiable.

### A2.2 Public directory model

New objects
- `ServerManifest` (already exists): canonical server config.
- **`DirectoryEntry` (new)**: public, signed summary for discovery.

`DirectoryEntry` contents (public metadata only)
- Server ID + manifest hash
- Name, short description, tags, language(s), region hints
- Join policy: invite-only / request-to-join / open
- Safety labels: NSFW flag, topic category, minimum age flag
- Suggested relay/SFU nodes (optional)
- Contact for moderation (optional)
- Signature by server owner key

Publication and retrieval
- Server owner publishes `DirectoryEntry` into the DHT under deterministic keys.
- Clients can enumerate directory keys by category + time windows.
- For real search (keywords), clients query one or more **Indexers**.

### A2.3 Indexers (community-run, optional, non-authoritative)

Indexer role
- Crawl public `DirectoryEntry` records.
- Build searchable index (keywords/tags/categories/languages).
- Serve search API to clients (HTTPS + signed responses).

Trust model
- Clients ship with a small default list of indexers (community-run).
- Users can add/remove indexers.
- Indexer responses include the signed `DirectoryEntry` (verifiable), so indexers cannot forge server listings.
- Clients merge results from multiple indexers and de-duplicate by server ID.

Privacy
- Query privacy: support Tor/proxy and query padding; optionally query multiple indexers.
- Indexers never see private servers; only opt-in public metadata.

### A2.4 Joining flow (public, request-based, invite-only)

Default state: invite-only.

Modes
1) Invite-only: join requires an invite code.
2) Request-to-join: user requests access; moderators approve/deny.
3) Open: anyone can join, subject to rate limits and automated checks.

UX requirements
- “Preview” page: description, rules, member count estimate, last active, required permissions.
- Join friction controls: captcha is avoided; instead use PoW / rate limits / reputation.

### A2.5 Moderation & room management (Discord-grade)

This largely aligns with v3’s planned RBAC/moderation phase, but this addendum makes moderation a first-class protocol surface area (not just UI).

RBAC
- Default roles: Owner, Admin, Moderator, Member, Guest.
- Custom roles with permission bitmask.
- Channel overrides (allow/deny) with deterministic merge rules.

Moderation actions (protocol events)
- Message delete/redact (creates a signed `Redaction` event).
- Timeout (mute) with duration and scope.
- Kick / ban (with optional reason).
- Slow mode per channel.
- Lockdown mode (temporary “read-only”).

Important protocol note
- In a decentralized network, “deletion” is enforced by compliant clients. The protocol must define redaction as an authoritative signed event; official clients must honor it, and servers can advertise “moderation-required” in the manifest so clients can warn users if their client is non-compliant.

Audit log
- Append-only, signed moderation log, queryable by permitted roles.

Anti-abuse primitives
- Per-channel rate limits (slow mode; burst limits).
- Join throttles (per-IP is not reliable; use PoW, reputation, and invite controls).
- Local block/mute always available.

### A2.6 Roadmap adjustments (pull-forward)

To remove adoption blockers earlier, move these forward:

v0.2.x (earlier)
- Basic RBAC: Owner/Admin/Moderator/Member.
- Basic moderation: delete message, timeout, ban.
- Channel slow mode.

v0.3.x
- Public DirectoryEntry publishing + “Explore” UI (browse by category).
- Invite system + request-to-join.
- Initial indexer reference implementation (Docker) + signed search responses.

v0.4.x
- Full custom roles + channel overrides.
- Audit log + moderation policy versioning.
- Auto-moderation hooks (rate limits, keyword filters).

### A2.7 Definition of done (server discovery + moderation)

Discovery DoD
- A server can be made public and appears in at least two independent indexers.
- Client can search by keyword and tag; results are verifiable from signatures.
- Join flows cover invite-only, request-to-join, open.

Moderation DoD
- Moderator can redact messages, timeout, ban; actions replicate reliably and appear in audit log.
- Slow mode works under partition/rejoin.
- Official client enforces moderation events and clearly indicates enforcement status.

---

## A3) Deliverables (engineering)

1) Connectivity Orchestrator package
- `pkg/net/co`: path selection, tunnel management, recovery.
- Metrics/events: `ConnectivityState` with reason codes.

2) Aether Tunnel
- QUIC tunnel daemon embedded in client.
- Session migration and keepalive policy.

3) Directory + Indexer spec and reference
- Protobuf: `DirectoryEntry`, `SearchRequest/Response`.
- `cmd/indexer`: crawler + search API + signed responses.

4) Moderation protocol events
- Protobuf: `Redaction`, `Timeout`, `Ban`, `AuditLogEntry`.
- Client enforcement rules.

5) Test and release gates
- NAT/network matrix tests in nightly.
- Canary call infrastructure.

