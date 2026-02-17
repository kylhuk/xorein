# Open Decisions Proposals

Status: draft recommendations for maintainer review.

This document proposes defaults for all currently open decisions in `open_decisions.md`.
Nothing here is final until you accept it and we update source planning docs.

## How to use

1. For each row, mark one of: Accept, Modify, or Defer.
2. For accepted items, copy the chosen default into the source TODO/roadmap file.
3. Keep unresolved items explicitly open with a revisit gate.

## Roadmap-level decisions (`aether-v3.md`)

| ID | Proposed default | Tradeoff | Suggested timing | Decision |
|---|---|---|---|---|
| RM-01 (name) | Keep working name `Aether` for v0.9 planning; run trademark/domain check in parallel; prepare fallback codename if blocked. | Unresolved branding risk remains until legal check completes. | Before v0.9 external-facing docs freeze. | Name for the chat app: "Harmolyn". Name for the P2P protocoll and backend: "xorein" |
| RM-02 (relay incentives) | Stay non-token: reputation + operator allowlist + reliability scoring. | Slower relay growth vs economic incentives, but much lower governance/legal complexity. | Decide now for v0.9 ops design. | No economic incentives. This is okay. People who want to host will have to pay the fee for the VPS themselves which is fine. Most things should be P2P anyways and one or two master servers should be enough at the beginning. |
| RM-03 (scale targets) | Set provisional target: 5k-member server and 200 concurrent voice participants as v0.9 baseline test goal. | May be conservative; can be raised later with evidence. | Before v0.9 perf test plan lock. | We need to test multiple constellations, because depending on the number of participants, different encryption methods are used. We need to know the maximum possible. For that test this protocol thoroughly with always 50 more people until you reach the hard limit of each encryption method. This way, we can then specify how large a group can be, until the protocol suffers and needs to switch to a less secure method. |
| RM-04 (governance/funding) | Use interim maintainer-RFC governance through v0.9; defer foundation/legal structure to v1.0 track. | Less institutional certainty now, faster execution now. | Decide now; revisit at v1.0 planning gate. | Governance will be at one point a consortium. Right now it will be just open source, available on a public repository under AGPL. We need to have a minimal legal text probably, just so we are not liable if anything gets destroyed/damaged. |
| RM-05 (mobile relay topology) | Multi-provider/federated notification relays with failover; no single mandatory central relay. | More implementation complexity, lower centralization risk. | Must decide before v0.9 mobile wake-policy work. | One core feature is, that this whole backend is fully decentralized. This is very important to manage before we have v1.0. The whole network can be run by just having one server online, as it is only used to coordinate clients. In the beginning, this will be me hosting it. But at one point, somneone else can take over so that the network never dies. |

## Carry-forward version decisions (OD3-OD7)

| ID | Proposed default | Tradeoff | Suggested timing | Decision |
|---|---|---|---|---|
| OD3-01 | Directory entry freshness: soft TTL 24h, stale grace to 72h with explicit stale label. | More stale content tolerated for availability. | Before v0.9 discovery-retention policies. | I am not sure if something like this is even needed, as the data of the server should live inside the clients/users. So as long as there is a single user online that was recently on this server, the information is retained. If the user is offline and stays offline, the data will be gone. |
| OD3-02 | Ranking tie-break: relevance -> trust score -> recency -> deterministic ID lexical tie-break. | More predictable but less "adaptive" ranking. | Before v0.9 search/discovery tuning. | You decide for now |
| OD3-03 | Keep RNNoise fallback mandatory through v0.9; reevaluate removal post-v0.9 with quality telemetry. | Keeps dual-path maintenance cost. | Decide now; revisit at v1.0 media review. | Accepted |
| OD3-04 | Privacy default: single-indexer query with rotation; multi-index parallel querying opt-in only. | Slightly slower discovery by default. | Before v0.9 trust/privacy lock. | Accepted for now, but please come up with a better idea if discovery is slow. This needs to be quite fast or we will lose users |
| OD4-01 | Manual moderator action wins over auto-mod in races; preserve full audit trail. | Potentially reduced automation efficacy in edge races. | Before v0.9 moderation interaction updates. | No, both are equal. First come, first served. Ignore the second action, if it was withing a very low timeframe. Don't ignore it, if it was manually done 5 seconds later (virtual number, I don't know what number would be ideal) |
| OD4-02 | Policy rollback horizon default 24h, with privileged override up to 7 days. | Limits rollback blast radius, may constrain long outages. | Before v0.9 policy/version handling changes. | As long as we are in alpha state, policy rollback should be a large window. It should be reduced in Beta. And even more reduced when we are live. |
| OD4-03 | Start conservative auto-mod thresholds (precision-first) with shadow-mode tuning before hard enforcement. | May miss more abusive content initially. | Before v0.9 moderation rollout. | Accepted |
| OD5-01 | Bot event delivery profile: at-least-once with idempotency keys and 24h replay window. | Consumers must handle duplicates. | Before v0.9 reliability controls that touch bots/events. | Accept |
| OD5-02 | Keep Discord compatibility explicitly subset-scoped; defer long-tail endpoints to compatibility backlog. | Some migration friction for advanced Discord usage. | Keep deferred unless v0.9 scope explicitly expands. | Accept |
| OD5-03 | Emoji retention: server-owned assets retained until explicit delete; archive on server archival. | More storage usage vs safer history continuity. | Before v0.9 retention/storage policy finalization. | All emojis are retained, even old versions, unless they get manually deleted by an admin for abuse reasons (illegal content). |
| OD5-04 | Webhook signing: HMAC-SHA256 + key-id header + 90-day rotation baseline. | Requires key management discipline. | Before v0.9 webhook/relay hardening. | Accept |
| OD5-05 | SDK governance: conformance suite + tiered labels (official/verified/community). | Added process overhead for SDK contributors. | Before v0.9 external SDK guidance publication. | Accept |
| OD6-01 | Discovery hardening TTL profile: soft 12h, hard 48h, jittered refresh. | More refresh churn under unstable networks. | Before v0.9 discovery resilience work. | Accept |
| OD6-02 | PoW adaptation: bounded floor/ceiling with periodic (time-windowed) adjustments per client class. | May underfit outlier devices initially. | Before v0.9 anti-abuse tuning. | Accept |
| OD6-03 | Sparse-graph reputation: blend global baseline with local evidence, cap max trust influence. | Slower trust differentiation for strong local communities. | Before v0.9 trust weighting lock. | Accept |
| OD7-01 | Replica placement default: 3 replicas across distinct relay operators/regions (anti-affinity). | Higher replication/storage cost. | Before v0.9 persistent hosting defaults. | Accept |
| OD7-02 | Merkle chunk size default: 256 KiB. | Not optimal for every payload profile, but balanced baseline. | Before v0.9 large-file/history sync tuning. | Accept |
| OD7-03 | Scoped search ranking: relevance + recency decay + permission-context priority. | More predictable, less personalization. | Before v0.9 search optimization work. | Accept |
| OD7-04 | Relay topology default: hub-and-spoke with optional mesh fallback under partition. | Hub relays can become hotspots. | Before v0.9 relay/cascade topology decisions. | Accept |

## Active decisions feeding v0.9 planning (OD8-OD9)

| ID | Proposed default | Tradeoff | Suggested timing | Decision |
|---|---|---|---|---|
| OD8-01 | Thread depth policy: soft display depth cap 6, deeper replies collapsed with explicit expand affordance. | Deep threads need extra clicks. | Decide now for v0.9 UX continuity. | This seems not ideal. There is no chat app providing more than 1 or max 2 levels. Keep it at max 2 for now. |
| OD8-02 | Metadata precedence: OG canonical fields first; Twitter fields fill missing OG data; deterministic conflict tie-break logged. | Some platform-specific card styling may be reduced. | Decide now before v0.9 link-preview refinements. | Accept |
| OD8-03 | High-contrast policy: override only tokens that fail contrast thresholds; preserve compliant custom tokens. | Mixed visual style under partially compliant themes. | Decide now before theme/a11y expansion. | Accept |
| OD8-04 | Locale fallback granularity: language-level fallback first (for example `es-AR -> es-ES`), else `en-US`. | Regional nuance may be lost in fallback. | Decide now before v0.9 locale growth. | Accept |
| OD8-05 | DTLN runtime policy: prefer DTLN by default, auto-fallback to RNNoise when compute/power constraints exceed threshold. | Variable quality across constrained devices. | Decide now before v0.9 realtime perf work. | Accept |
| OD9-01 | Server pin retention default: minimum 30-day retention, refresh-on-access extension. | Higher storage pressure on small operators. | Must decide before TODO_v09 implementation starts. | Accept |
| OD9-02 | GossipSub hierarchy depth default: 2 tiers baseline, allow 3 tiers for very large deployments only. | Might under-optimize extreme scale at first. | Must decide before OD9 topology tasks. | Accept |
| OD9-03 | Cascade split/merge aggressiveness: conservative thresholds with hysteresis to avoid oscillation. | Slower reaction to rapid load spikes. | Must decide before voice-cascade tuning. | Accept |
| OD9-04 | Optimization go/no-go confidence: require >=10% p95 gain and no critical reliability regression. | Some useful but smaller wins may be deferred. | Must decide before perf gate criteria lock. | Accept |
| OD9-05 | Overload priority order: control plane -> active media -> store-forward -> bulk sync. | Lower-priority backlog may grow during incidents. | Must decide before relay overload handling implementation. | Accept |
| OD9-06 | Mobile battery policy: adaptive default; battery-saver mode aggressively reduces non-critical background activity. | Potential UX latency in low-power mode. | Must decide before mobile runtime tuning work. | Accept |
| OD9-07 | Wake-policy topology: multi-provider wake path with client failover, no single mandatory provider. | More integration complexity. | Must decide before mobile notification architecture lock. | Accept |

## Suggested decision order

1. Decide RM-05 and OD9-07 together (same topology risk).
2. Decide OD9-02/03/05 together (topology and overload interactions).
3. Decide OD8-02/03/04/05 together (client UX/runtime coherence).
4. Resolve remaining carry-forward OD3-OD7 items that can affect v0.9 defaults.
