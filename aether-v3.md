> **SUPERSEDED BY `docs/spec/v0.1/`** — This document is a historical design
> narrative. The normative Xorein v0.1 protocol specification lives in
> `docs/spec/v0.1/`. Do not use this document as a reference for implementation.

# Aether Protocol & Platform — Revised Implementation Plan v3

> **Mission:** A fully peer-to-peer, end-to-end encrypted communication platform. If two nodes exist anywhere on Earth, the network exists. No central authority, no single point of failure, no compromise on speed or quality.

---

## Part 0: Engineering Foundations — Correctness, Reproducibility & DevSecOps

**Principle:** Every bug must be caught at compile time, not in production. Every build must be reproducible. Every deployment must be automated, secured, and auditable. One command builds everything, tests everything, scans everything, and deploys everything.

This section is the foundation. Every technology choice in this document is evaluated against these standards first.

---

### 0.1 Compile-Time Correctness: Catch Bugs Before They Exist

The goal is to make invalid states unrepresentable and incorrect code uncompilable.

**Language choice (Go) serves this goal:**
- Statically typed with compile-time type checking — no runtime type surprises.
- No implicit conversions, no null pointer dereference ambiguity (`nil` is explicit and checkable).
- `go vet` is built into the toolchain — catches common mistakes (printf format strings, unreachable code, struct tag errors) at compile time.
- The compiler is fast (~2s for a full build of most projects), enabling tight feedback loops.

**Static analysis stack (runs on every commit, blocks merge on failure):**

| Tool | Purpose | Catches |
|---|---|---|
| **golangci-lint** | Meta-linter aggregating 50+ linters | Single unified tool, runs all linters in parallel with caching |
| **staticcheck** | Deep semantic analysis (150+ checks) | Unreachable code, deprecated API usage, incorrect sync primitives, inefficient code |
| **gosec** | Security-focused static analysis | SQL injection, hardcoded credentials, weak crypto, tainted file paths, integer overflows |
| **govet** | Go team's official analyzer | Printf format mismatches, struct tag validation, copy-lock violations, unreachable code |
| **errcheck** | Unchecked error return values | Every `error` return that's silently ignored — the #1 source of Go bugs |
| **revive** | Configurable linting rules | Naming conventions, exported function docs, cyclomatic complexity |
| **go-critic** | Opinionated code reviewer | Style issues, performance anti-patterns, unnecessary type assertions |

**golangci-lint configuration (`.golangci.yml`)** is the Single Source of Truth for code quality rules. All developers and CI use the same config. No exceptions.

**Protocol Buffers (protobuf) for wire format** provides compile-time guarantees on message structure:
- Schema defines field types, numbers, and cardinality. Mismatched types = compile error.
- Generated Go code provides type-safe accessors — no runtime parsing errors for well-formed messages.
- Schema evolution rules (add-only for minor, reserved for deprecated) enforced by `buf lint` and `buf breaking` in CI.

**Additional compile-time guarantees:**
- `go build -race` in CI: detects data races at test time using the ThreadSanitizer runtime.
- All interfaces have compile-time assertion checks: `var _ Interface = (*ConcreteType)(nil)` to ensure implementations stay in sync.
- `exhaustive` linter: ensures all switch statements on enums/sum types handle every case.

---

### 0.2 Testing: Find Every Edge Case Automatically

Testing is not an afterthought. The test suite is a first-class product artifact. The testing strategy uses four layers, each catching different classes of bugs:

**Layer 1: Unit Tests (Go standard `testing` package)**
- Every exported function has unit tests. Target: ≥80% line coverage, ≥95% for crypto and protocol code.
- Table-driven tests (Go idiom) for systematic input coverage.
- `testify/assert` for clear, readable assertions.
- `go test -race ./...` runs on every commit — data races are build-breaking failures.

**Layer 2: Property-Based Testing (`rapid` — pgregory.net/rapid)**

This is the edge-case killer. Instead of writing individual test cases, you define *properties* that must hold for all inputs, and the framework generates thousands of random inputs to find counterexamples.

`rapid` is chosen over `gopter` because:
- Simpler API (one function vs. multiple packages).
- Automatic test case minimization without user code — when a failure is found, `rapid` shrinks the input to the smallest possible reproducing case.
- `rapid.MakeFuzz()` converts any property test into a standard Go fuzz target — bridging property testing and fuzzing.
- State machine testing for stateful components (DHT operations, message ordering, key ratcheting).

**Where property-based testing is mandatory:**
- **Cryptographic primitives:** Encryption round-trip (encrypt then decrypt = original), key derivation determinism, signature verification.
- **Protocol serialization:** Protobuf encode→decode round-trip, backward compatibility (v1.0 message readable by v1.1 decoder and vice versa).
- **DHT operations:** Kademlia routing table invariants, XOR distance properties, peer ordering.
- **Message ordering:** Causal ordering in GossipSub, Sender Key ratchet consistency.
- **Network topology:** SFU election algorithm convergence, mesh→SFU transition correctness.

**Layer 3: Fuzz Testing (Go 1.18+ native `testing.F`)**

Coverage-guided fuzzing runs continuously in CI (nightly, multi-hour sessions) and finds inputs that crash or misbehave:
- Protobuf deserialization: feed random bytes to every message parser.
- Multistream-select negotiation: random protocol strings and version numbers.
- SQLCipher query construction: ensure no SQL injection paths.
- Cryptographic operations: random key material, ciphertext mutation detection.

Fuzz corpus is committed to the repository. Discovered crash inputs become permanent regression tests.

**Layer 4: Integration & Network Simulation Tests**

- **testground** (Protocol Labs): Spins up N Aether nodes in isolated Docker networks with configurable latency, jitter, packet loss, and NAT simulation. Tests: discovery cascade, message delivery under partition, SFU failover, store-and-forward delivery.
- **Docker Compose test harness:** Brings up a full multi-node network (3 relays, 5 clients) for end-to-end testing of the complete protocol.
- **Chaos testing:** Random node kills, network partitions, clock skew injection. Validates graceful degradation.

**Test infrastructure rules:**
- All tests must pass before merge. No exceptions, no "skip in CI" annotations.
- Flaky tests are treated as bugs, not annoyances. A test that fails intermittently is a test that reveals a real concurrency or timing issue.
- Test coverage is tracked per-package and reported on every PR. Crypto and protocol packages have ≥95% coverage gates.

---

### 0.3 Reproducible Builds & Configuration: Single Source of Truth

Every artifact produced by the build system must be byte-for-byte reproducible. If two people run the same build at the same commit, they get identical binaries.

**Reproducible Go builds:**
- `go build -trimpath -ldflags="-s -w -X main.version=$(git describe --tags) -X main.commit=$(git rev-parse HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"` — removes local filesystem paths, embeds version metadata.
- Go modules with `go.sum` checksums ensure exact dependency versions. `GONOSUMCHECK` is never set.
- `GOFLAGS=-mod=readonly` in CI prevents accidental dependency changes.
- Binary verification: release binaries include SHA-256 checksums and are signed. Users can verify by building from source at the tagged commit and comparing hashes.

**Configuration generation with Dhall:**

All configuration files are generated from a typed, deterministic source. Dhall is chosen because it is:
- **Typed:** Configuration errors are caught at generation time, not at runtime. A relay config missing a required field is a type error in Dhall.
- **Total (not Turing-complete):** Dhall programs always terminate. Configuration generation cannot hang or loop forever.
- **Deterministic:** Same Dhall input always produces the same output. No environment variable leakage, no filesystem dependencies.
- **Go-native bindings:** `dhall-golang` (v6) unmarshals Dhall directly into Go structs.

**Configuration architecture:**

```
config/
├── types.dhall                  # Type definitions (RelayConfig, ClientConfig, BootstrapList, etc.)
├── defaults.dhall               # Default values for all configuration types
├── environments/
│   ├── development.dhall        # Dev overrides (local bootstrap, verbose logging)
│   ├── staging.dhall            # Staging overrides (test bootstrap nodes)
│   └── production.dhall         # Production config (real bootstrap, hardened settings)
├── nodes/
│   ├── relay-us-east.dhall      # Per-node overrides (inherits from environment)
│   ├── relay-eu-west.dhall
│   └── relay-ap-south.dhall
└── generate.sh                  # dhall-to-toml / dhall-to-json for each target
```

**Single Source of Truth principle:**
- Bootstrap node lists, relay configurations, default settings, and Docker Compose files are ALL derived from Dhall sources.
- The `relay.toml` that ships in the Docker image, the bootstrap list compiled into the binary, and the documentation are all generated from the same Dhall types. If they disagree, the build fails.
- Changing a bootstrap node means editing ONE Dhall file. The CI generates all downstream artifacts.

**Reproducible container images:**
- Multi-stage Docker builds with pinned base images (digest, not tag): `FROM golang:1.23.4@sha256:<digest> AS builder`.
- `ko` (Google's Go container builder) for minimal, distroless Go images — no shell, no package manager, no attack surface.
- Target-state intent: images are built in CI (not locally) once CI exists in a future implementation. The Dockerfile and build context are intended to be deterministic.
- Every image includes an SBOM (Software Bill of Materials) in SPDX format and SLSA provenance metadata.

---

### 0.4 CI/CD/CS Pipeline: One Command to Rule Them All (Target-State Blueprint)

> Repository snapshot note: this section documents planned target-state pipeline behavior. Commands shown below are placeholders and are not executable in this documentation-only checkout.

In the planned implementation, the entire build, test, scan, and deploy pipeline is intended to be invoked with a single command entrypoint:

```bash
make all    # planned target-state command placeholder
# or: task all (planned Taskfile.yml command placeholder)
```

In the planned toolchain, this would run the complete pipeline locally. In planned CI (GitHub Actions), the same `Makefile`/`Taskfile.yml` would be used — target state is no difference between local and CI builds.

**Pipeline stages (in order, each blocks on the previous):**

```
┌─────────────────────────────────────────────────────────────────────────┐
│            planned entrypoint: make all (placeholder)                    │
│                                                                         │
│  Stage 1: GENERATE                                                      │
│  ├─ dhall-to-toml: Generate all config files from Dhall sources         │
│  ├─ buf generate: Generate Go code from protobuf schemas                │
│  ├─ buf lint + buf breaking: Validate protobuf schema compatibility     │
│  └─ go generate ./...: Any other code generation                        │
│                                                                         │
│  Stage 2: COMPILE                                                       │
│  ├─ go build -trimpath ./...                                            │
│  ├─ Compile for all target platforms (cross-compile matrix)             │
│  └─ go vet ./... (built-in static analysis)                             │
│                                                                         │
│  Stage 3: LINT (Continuous Integration)                                 │
│  ├─ golangci-lint run (50+ linters including gosec, staticcheck,        │
│  │   errcheck, revive, exhaustive, go-critic)                           │
│  └─ buf lint (protobuf schema validation)                               │
│                                                                         │
│  Stage 4: TEST (Continuous Integration)                                 │
│  ├─ go test -race -cover ./... (unit + property tests + race detector)  │
│  ├─ go test -fuzz=. -fuzztime=60s ./... (fuzz for 60s per target)       │
│  └─ Integration tests (Docker Compose test harness)                     │
│                                                                         │
│  Stage 5: SECURITY SCAN (Continuous Security)                           │
│  ├─ govulncheck ./... (known Go vulnerability database)                 │
│  ├─ gosec ./... (security-focused static analysis, also in golangci)    │
│  ├─ Trivy fs . (filesystem vulnerability + license scan)                │
│  ├─ Trivy image (container image vulnerability scan)                    │
│  └─ SBOM generation (syft or Trivy, SPDX format)                       │
│                                                                         │
│  Stage 6: BUILD ARTIFACTS (Continuous Deployment)                       │
│  ├─ Build release binaries (all platforms, reproducible)                │
│  ├─ Build Docker images (ko or multi-stage, pinned base, distroless)    │
│  ├─ Sign binaries + images (cosign for containers, minisign for bins)   │
│  └─ Generate checksums (SHA-256)                                        │
│                                                                         │
│  Stage 7: DEPLOY (Continuous Deployment, tagged releases only)          │
│  ├─ Push signed images to ghcr.io                                       │
│  ├─ Publish release binaries to GitHub Releases                         │
│  ├─ Deploy relay nodes (Dhall-generated configs, rolling update)        │
│  └─ Update bootstrap DNS records if node list changed                   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Planned CI-specific extras (nightly/weekly, once CI workflows exist):**
- Extended fuzz testing (1 hour per target, nightly).
- Full network simulation (testground, 20+ nodes, various NAT configurations).
- Dependency update check (`go list -m -u all` + automated PR creation).
- License compliance scan (Trivy license mode).

**Planned GitHub Actions workflow structure (target state):**
```
.github/workflows/
├── ci.yml              # Runs on every push/PR: stages 1-5
├── release.yml         # Runs on tag push: stages 1-7
├── nightly.yml         # Nightly: extended fuzz + network simulation
└── security-audit.yml  # Weekly: full Trivy + govulncheck + dependency audit
```

**Planned local developer experience (once tooling exists):**
- `make check` — intended fast feedback loop: lint + unit tests.
- `make test` — intended full test suite including integration tests.
- `make all` — intended full pipeline parity with CI.
- Planned pre-commit hooks (via `pre-commit` framework): `gofmt`, `golangci-lint`, `go vet` on staged files only.

---

### 0.5 Technology Selection Criteria

Every library, framework, and tool in this project is evaluated against these criteria:

1. **Compile-time safety.** Does it catch errors before runtime? Strongly typed APIs preferred. Code generation from schemas preferred over hand-written serialization.

2. **Active maintenance.** Is it actively developed? When was the last release? Are issues being triaged? Abandoned dependencies are security liabilities.

3. **Permissive license.** MIT, BSD, Apache 2.0, or equivalent. GPL is acceptable for tools (not linked into the binary). No proprietary dependencies.

4. **Minimal dependency tree.** Fewer transitive dependencies = smaller attack surface = fewer supply chain risks. Heavy frameworks that pull in the world are rejected.

5. **Reproducibility.** Can it produce deterministic output? Non-deterministic tools (random orderings, timestamp-dependent behavior) are avoided or wrapped.

6. **Testability.** Can it be unit tested in isolation? Libraries with global state, init() side effects, or mandatory network access are wrapped behind interfaces.

7. **Go ecosystem fit.** Native Go is strongly preferred. CGo (C FFI) is used only when no pure-Go alternative exists (e.g., SQLCipher, noise cancellation models). Every CGo dependency is explicitly justified and isolated behind an interface for potential future replacement.

---

---

## Part I: The Aether Protocol (AEP)

The protocol is the product. The UI is just one implementation. Anyone should be able to build a client — the protocol specification is the contract.

---

### 1. Network Resilience: The Two-Node Guarantee

**Principle:** The Aether network is alive if any two nodes can find each other. There are no "special" nodes — only nodes with different capabilities enabled.

#### 1.1 Node Types (Same Binary, Different Modes)

Every Aether node runs the same Go binary. Behavior is determined by configuration flags, not different code:

| Mode | Description | Runs On |
|---|---|---|
| **Full Client** | GUI + P2P core. Participates in DHT, GossipSub, relays if configured. | Desktop, Mobile, Browser |
| **Headless Relay** | No GUI. Runs as daemon. Provides DHT bootstrap, circuit relay, store-and-forward, and optional SFU for voice. | VPS, Raspberry Pi, NAS, Docker |
| **Embedded (Lite)** | Reduced functionality. DHT client mode only. Relies on relay nodes for connectivity. | Low-power devices, browsers with limited WebRTC |

```
┌─────────────────────────────────────────────────────────────────────┐
│ Aether Node Binary (single Go binary, ~15MB)                       │
│                                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ libp2p   │  │  Pion    │  │  Crypto  │  │  Storage │           │
│  │  Host    │  │  WebRTC  │  │  Engine  │  │  Engine  │           │
│  │          │  │  (Voice/ │  │ (Signal  │  │ (SQLite  │           │
│  │ DHT      │  │  Video/  │  │  Proto)  │  │  Cipher) │           │
│  │ GossipSub│  │  Screen) │  │          │  │          │           │
│  │ Relay    │  │          │  │          │  │          │           │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘           │
│                                                                     │
│  ┌──────────────────────────────────────────────────────┐          │
│  │  Mode Switch                                         │          │
│  │  --mode=client    → Enable GUI bridge + all features │          │
│  │  --mode=relay     → Headless daemon, relay+DHT+SFU   │          │
│  │  --mode=bootstrap → Minimal DHT-only bootstrap node  │          │
│  └──────────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────────┘
```

#### 1.2 Discovery Cascade (How Two Nodes Find Each Other)

The protocol uses a **layered discovery system** where each layer is a fallback for the previous:

0. **Local Peer Cache** — On every startup, the node first tries peers it has successfully connected to before. This cache is persisted to disk (in SQLCipher) and updated on every successful connection. If even one cached peer is online, the node rejoins the DHT immediately — no bootstrap infrastructure needed. This is the single most important resilience mechanism: it means the network survives total bootstrap failure for all existing users.

1. **mDNS (Local Network)** — Zero configuration. Nodes on the same LAN discover each other instantly via multicast DNS. Works offline with zero internet. Two laptops on the same WiFi = functional network.

2. **Hardcoded Bootstrap List** — Compiled into the binary. A list of 10+ well-known multiaddresses (community-run, geographically distributed). These are just DHT entry points — they have no special authority.

3. **DNS-based Discovery** — TXT records on a well-known domain (e.g., `_aether._tcp.bootstrap.aether.chat`) resolve to current bootstrap node multiaddresses. Survives bootstrap node IP changes without binary updates.

4. **DHT Walking (Kademlia)** — Once connected to any single peer, the node populates its routing table via Kademlia's iterative lookup. Within minutes, it knows hundreds of peers across the keyspace.

5. **Rendezvous Protocol** — Peers interested in the same "server" register at a rendezvous point in the DHT. New members find existing members by querying the rendezvous point.

6. **Peer Exchange (PEX)** — Connected peers periodically exchange their peer lists. Even without DHT access, a node that connects to a single known peer can discover the broader network.

7. **Manual Peer Entry** — Last resort. User can paste a multiaddress directly (e.g., from a QR code, URL, or manual exchange). This enables networks to form even when all automated discovery fails.

**The key insight:** Discovery layers 0, 1, 6, and 7 require zero infrastructure. Returning users reconnect via their peer cache (layer 0) even if every bootstrap node is gone. New users who can exchange a multiaddress (via email, SMS, written on paper) can bootstrap from scratch. The network is truly unkillable.

#### 1.3 Headless Relay Nodes (Deep Dive)

The headless relay is the backbone that keeps the network performant and available. It's designed to be dead simple to deploy:

```bash
# One-liner deployment
docker run -d --name aether-relay \
  -p 4001:4001/tcp \
  -p 4001:4001/udp \
  -p 4002:4002/tcp \   # WebSocket for browser peers
  -p 9090:9090/tcp \   # Metrics/admin (optional)
  -v aether-data:/data \
  ghcr.io/aether/relay:latest \
  --mode=relay \
  --storage-quota=10GB \
  --relay-bandwidth=100Mbps \
  --sfu-enabled=true \
  --sfu-max-rooms=50
```

**What a relay node provides:**

| Service | Description | Resource Impact |
|---|---|---|
| **DHT Bootstrap** | Provides DHT routing table entries to new peers. Always on. | Negligible CPU/BW |
| **Circuit Relay v2** | Relays connections for peers behind strict NATs. Time-limited (max 2 min per relay, renewable). | Moderate BW |
| **Store-and-Forward** | Caches encrypted messages for offline peers. Messages encrypted to recipient's public key — relay sees nothing. | Storage (configurable quota) |
| **SFU (Optional)** | Selective Forwarding Unit for voice/video channels. Receives media streams and forwards to participants without decoding. Still E2EE via SFrame. | High CPU/BW when active |
| **TURN Relay** | WebRTC TURN fallback for peers that cannot establish direct media connections. | High BW when active |
| **History Archival** | Optionally stores encrypted channel history for servers that opt in. Enables history sync for new members. | Storage |

**Relay incentive model:**
- Reputation-based: relay nodes earn reputation in the web-of-trust.
- Server owners can designate relay nodes as "preferred" for their community.
- No cryptocurrency or tokens — pure community goodwill + practical incentives (server owners want their community to have good connectivity).

**Relay node configuration file (relay.toml):**
```toml
[node]
mode = "relay"
listen_addrs = ["/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/udp/4001/quic-v1"]
announce_addrs = ["/ip4/203.0.113.1/tcp/4001"]  # Public IP

[relay]
enabled = true
max_circuits = 1024
max_circuit_duration = "2m"
max_circuit_bandwidth = "1Mbps"  # Per circuit

[store_forward]
enabled = true
storage_path = "/data/store"
max_storage = "10GB"
message_ttl = "30d"

[sfu]
enabled = true
max_rooms = 50
max_participants_per_room = 100
preferred_codecs = ["opus", "vp9", "av1"]

[metrics]
enabled = true
listen_addr = "127.0.0.1:9090"  # Prometheus metrics
```

---

### 2. Protocol Evolution & Backward Compatibility

**Principle:** The protocol must be evolvable without breaking existing clients. A v1.0 client must be able to communicate with a v3.0 client, even if it can't use v3.0-only features.

#### 2.1 Three-Layer Versioning Strategy

The Aether Protocol uses three independent versioning mechanisms that work together:

**Layer 1: libp2p Multistream-Select (Transport/Stream Level)**

Every libp2p protocol is identified by a versioned path string. When two peers open a stream, they negotiate which version to use:

```
Peer A → Peer B: "/aether/chat/2.0.0"
Peer B → Peer A: "na"                     (doesn't support 2.0.0)
Peer A → Peer B: "/aether/chat/1.1.0"
Peer B → Peer A: "/aether/chat/1.1.0"     (agreed!)
```

Each Aether sub-protocol has its own version:
- `/aether/chat/1.x.x` — Text messaging
- `/aether/voice/1.x.x` — Voice signaling
- `/aether/sync/1.x.x` — History synchronization
- `/aether/manifest/1.x.x` — Server manifest replication
- `/aether/identity/1.x.x` — Identity and prekey exchange
- `/aether/bot/1.x.x` — Bot API protocol
- `/aether/file/1.x.x` — File transfer

A node advertises ALL versions it supports. Peers negotiate downward to the highest mutually supported version. This means a v2.0 node can still communicate with a v1.0 node on the v1.0 protocol, while using v2.0 features with peers that support it.

**Layer 2: Protobuf Schema Evolution (Message Level)**

All Aether wire messages are encoded in Protocol Buffers (proto3). Protobuf provides inherent forward and backward compatibility:

- **New fields can always be added.** Old clients ignore unknown fields.
- **Field numbers are never reused or changed.** Deprecated fields are marked `reserved`.
- **The wire format is stable.** Binary compatibility is guaranteed across all protobuf versions.

```protobuf
// v1.0 message
message ChatMessage {
  string id = 1;
  bytes author_pubkey = 2;
  bytes ciphertext = 3;
  int64 timestamp = 4;
  bytes signature = 5;
}

// v1.1 message (backward compatible — old clients ignore new fields)
message ChatMessage {
  string id = 1;
  bytes author_pubkey = 2;
  bytes ciphertext = 3;
  int64 timestamp = 4;
  bytes signature = 5;
  // v1.1 additions
  string reply_to = 6;           // new: threaded replies
  repeated bytes attachments = 7; // new: file attachments
  bytes edit_of = 8;             // new: message edits
}

// v2.0 message (new required capability — needs protocol negotiation)
message ChatMessageV2 {
  // Breaking change: new encryption scheme
  // This goes through multistream-select as /aether/chat/2.0.0
  // Old clients negotiate down to /aether/chat/1.x.x
}
```

**Compatibility rules for Aether protobuf schemas:**
1. Minor versions (1.0 → 1.1): Add fields only. Never remove, rename wire types, or reuse field numbers.
2. Major versions (1.x → 2.0): New multistream protocol ID. Old protocol ID remains supported for N+2 major versions.
3. All deprecated fields are marked `reserved` with a comment explaining when they were deprecated and the minimum version that no longer needs them.

**Layer 3: Capability Negotiation (Feature Level)**

After connecting, peers exchange a `Capabilities` message listing supported features:

```protobuf
message Capabilities {
  uint32 protocol_version = 1;    // Highest supported AEP version
  repeated string features = 2;    // Feature flags
  // e.g., ["e2ee.sender_keys", "e2ee.treekem", "voice.sframe",
  //        "voice.av1", "screen.simulcast", "bot.discord_shim"]
  map<string, string> extensions = 3;  // Vendor extensions
}
```

This enables fine-grained feature detection without protocol version bumps. A client that supports AV1 screen sharing can detect whether the remote peer does too, and fall back to VP9 if not.

#### 2.2 Protocol Governance

- The protocol specification lives in a public Git repository as a set of versioned RFC documents.
- Breaking changes require an AEP (Aether Enhancement Proposal) with a minimum 90-day review period.
- At least two independent implementations must validate a new protocol version before it's finalized.
- The specification is licensed under CC-BY-SA to prevent proprietary forks that break compatibility.

---

### 3. Voice, Video & Screen Sharing — Performance Architecture

**Non-negotiable target: voice mouth-to-ear latency ≤ 150ms, screen share glass-to-glass ≤ 100ms at 1080p60.**

Research confirms these targets are achievable with WebRTC. Direct P2P typically achieves 60–120ms mouth-to-ear latency. SFU forwarding adds ~20–40ms. The key is eliminating unnecessary processing and keeping the media pipeline tight.

#### 3.1 Topology Selection (Automatic, Per-Channel)

The protocol dynamically selects the optimal topology based on participant count:

```
Participants    Topology           Reason
─────────────   ──────────────     ────────────────────────────
2               Direct P2P         Lowest latency. No intermediary.
3–8             Full Mesh P2P      Each peer connects to every other.
                                   Manageable bandwidth at this scale.
9–50            Peer SFU           Best-connected peer elected as SFU.
                                   Still E2EE via SFrame.
50+             Relay SFU          Dedicated relay node acts as SFU.
                                   Cascading SFU for 200+ participants.
```

**SFU peer election algorithm:**
1. All peers in the voice channel report their upstream bandwidth and latency to each other.
2. The peer with the highest `(upstream_bandwidth_kbps / avg_latency_ms)` score becomes the SFU candidate.
3. Election requires majority agreement (>50% of participants).
4. If the SFU disconnects, re-election triggers within 2 seconds.
5. If a dedicated relay node is available and has better metrics, it's always preferred.

#### 3.2 Audio Pipeline (Latency Budget)

Every millisecond matters. Here's the target budget for the audio pipeline:

```
┌────────────────────────────────────────────────────────────────────┐
│ Audio Latency Budget: Target ≤ 150ms mouth-to-ear                  │
│                                                                    │
│ Capture + OS buffer           10ms                                 │
│ Noise suppression (DTLN)       5ms  (2.5ms lookahead + processing) │
│ Opus encode (10ms frames)     10ms  (frame size + encode time)     │
│ Network transit               20-80ms (depends on geography)       │
│ Jitter buffer                 20-50ms (adaptive)                   │
│ Opus decode                    1ms                                 │
│ Playback + OS buffer          10ms                                 │
│ ─────────────────────────────────                                  │
│ Total:                        76-166ms                             │
│                                                                    │
│ P2P direct:   ~80ms typical                                        │
│ Via SFU:      ~100-120ms typical (SFU adds ~20-40ms forwarding)    │
│ Via relay:    ~120-150ms typical                                   │
└────────────────────────────────────────────────────────────────────┘
```

**Opus encoder configuration:**
- Frame size: 10ms (lowest latency, slightly more overhead than 20ms, but worth it for feel)
- Bitrate: Adaptive, 24–128 kbps. Default 48kbps for voice, 96kbps for music mode.
- Application: `VOIP` mode (optimized for speech, enables FEC and DTX)
- FEC: Enabled. Adds ~1kbps overhead but recovers from up to 20% packet loss.
- DTX: Enabled. Stops transmission during silence — saves bandwidth, reduces noise.
- Bandwidth: Fullband (20Hz–20kHz) for maximum quality.
- Sample rate: 48kHz (Opus native, avoids resampling).

**Jitter buffer:**
- Adaptive target: starts at 50ms, adjusts between 20–200ms based on measured jitter.
- Packet loss concealment (PLC): Opus built-in PLC extrapolates missing frames.
- NetEQ-inspired: aggressive adaptation — shrinks quickly when network improves, grows slowly when degrading.

#### 3.3 Noise Cancellation Stack

Research shows three tiers of open-source options:

| Solution | Quality | Latency | License | Status |
|---|---|---|---|---|
| **RNNoise** (Xiph/Mozilla) | Good — suppresses steady noise well, struggles with transients | ~10ms | BSD-3 | Maintained (updated Jan 2025) |
| **DTLN** (dtln-rs by Datadog) | Very Good — dual-stage LSTM, scored well in DNS Challenge | ~5ms (2.5ms lookahead) | MIT | Open-sourced, portable Rust + WASM |
| **DeepFilterNet2** (Univ. Erlangen) | Excellent — state-of-the-art, ERB gains + deep filtering | ~5ms | MIT | Research, actively developed |

**Recommended approach:**
1. **MVP (v0.1–v0.3):** Use RNNoise via C FFI. It's proven, lightweight (~10% of one CPU core), and dead simple to integrate. Works even on weak hardware.
2. **v0.4+:** Upgrade to DTLN (Rust library via FFI, or native Go port). Significantly better quality than RNNoise, comparable to commercial solutions. The `dtln-rs` project from Datadog provides a Rust library with WASM and native targets.
3. **v1.0+:** Evaluate DeepFilterNet2 or train a custom model using the DNS Challenge dataset. Ship multiple models and let users select based on their hardware capability.

All noise cancellation runs **on-device only**. Audio is processed before encryption — the relay/SFU never sees clean audio.

#### 3.4 Screen Sharing — Maximum Quality Architecture

The protocol places zero artificial limits on screen sharing quality. The limiting factors are exclusively the sender's hardware encoder and upload bandwidth, and the receiver's download bandwidth and decoder.

**Encoder selection (priority order):**

| Priority | Encoder | Codec | Notes |
|---|---|---|---|
| 1 | NVENC (NVIDIA) | H.264/HEVC/AV1 | Lowest CPU. NVIDIA GPUs since GTX 600. |
| 2 | QuickSync (Intel) | H.264/HEVC/AV1 | Most Intel CPUs since 3rd gen. |
| 3 | VideoToolbox (Apple) | H.264/HEVC | All macOS/iOS devices. |
| 4 | VA-API (Linux/AMD) | H.264/HEVC/AV1 | AMD GPUs, some Intel. |
| 5 | libvpx (software) | VP9 | CPU-heavy but universal fallback. |
| 6 | libaom (software) | AV1 | Best compression, high CPU. Only for beefy machines. |

**Quality presets:**

| Preset | Resolution | FPS | Bitrate | Codec | Use Case |
|---|---|---|---|---|---|
| Low | 720p | 15 | 1–2 Mbps | VP9 | Weak hardware/network |
| Standard | 1080p | 30 | 3–6 Mbps | VP9/H.264 | Default for most users |
| High | 1080p | 60 | 6–15 Mbps | VP9/AV1 | Gaming, fast content |
| Ultra | Native | 60 | 15–50+ Mbps | AV1/HEVC | No limits. Peer's HW is the ceiling. |
| Auto | Adaptive | Adaptive | Adaptive | Best available | Adjusts in real-time based on conditions |

**Simulcast:** The sender encodes at up to 3 quality layers simultaneously. Each viewer receives the highest layer their connection supports. The SFU (if used) selects layers per-viewer — it never transcodes.

**E2EE for media:** SFrame (Secure Frames) encryption. Each participant encrypts their media frames with their own key before they hit the network. The SFU forwards encrypted frames without any ability to inspect them. Key distribution happens through the existing Sender Keys mechanism used for text channels.

---

### 4. E2EE Architecture (ENCRYPTION_PLUS scaling-aware baseline)

- **Security modes are explicit per conversation surface** (never a silent on/off switch):
  - **Seal:** 1:1 E2EE (X3DH + Double Ratchet).
  - **Tree:** small interactive groups/channels E2EE (MLS).
  - **Crowd:** large interactive rooms E2EE (rotation-based sender-epoch; weaker removal semantics until rotation).
  - **Channel:** broadcast E2EE (few writers / many readers; rotation-based epoch; weaker removal semantics until rotation).
  - **Clear:** server-readable content (explicitly labeled; not the default for private conversations).
- **Default mode selection is automatic** and based on size + semantics; initial default thresholds (policy knobs) are:
  - Tree ≤200, Tree→Crowd fallback 200–1000 (device/churn dependent), Crowd 1000–5000, Channel semantics >5000.
- **Mode transitions are first-class**: a mode change creates a new `mode_epoch_id`, is announced to the channel, and is recorded as an auditable governance event.
- **History continuity is best-effort across epochs**: existing members retain access to prior epochs; new joiners may see “locked history” unless admins enable re-sharing/re-encryption workflows.
- **Voice/Video/Screen:** SFrame per-participant media E2EE (MediaShield). Keys are delivered via the active conversation mode’s key-distribution mechanism.
- **At rest:** SQLCipher (AES-256) for all local data.

---

## Part II: The Reference Client

### 5. UI Framework Decision — Deep Analysis

We evaluated seven immediate-mode and lightweight GUI frameworks against Aether's requirements: cross-platform (desktop + mobile + web), high performance, aesthetic flexibility, and tight integration with the Go core.

#### 5.1 Framework Comparison

| Framework | Language | Platforms | iOS | Android | WASM | Chat App Suitability | Notes |
|---|---|---|---|---|---|---|---|
| **Gio UI** | Go | Win, Mac, Linux, iOS, Android, WASM, FreeBSD, OpenBSD | ✅ Native | ✅ Native | ✅ | Good — building widgets from primitives, active text/IME improvements | v0.9.0 (pre-1.0). GPU-accelerated (Vulkan/Metal/DX/GLES). MIT/Unlicense. |
| **Dear ImGui** | C++ | Win, Mac, Linux (+ backends) | ❌ No native support | ❌ No native support | ⚠️ Via Emscripten | Poor — explicitly designed for dev tools, not end-user apps. Lacks text editing, IME, accessibility. | Industry standard for game/debug tooling. Not suitable for consumer apps. MIT. |
| **Hello ImGui** | C++ (wraps Dear ImGui) | Win, Mac, Linux, iOS, Android, Emscripten | ✅ Via SDL | ✅ Via SDL | ✅ | Poor — inherits Dear ImGui's tool-oriented widget set. Better platform coverage but same fundamental limitations. | Adds cross-platform scaffolding to Dear ImGui. MIT. |
| **Dear ImGui Bundle** | C++ (wraps Hello ImGui) | Win, Mac, Linux, iOS, Android, Emscripten | ✅ Via Hello ImGui | ✅ Via Hello ImGui | ✅ | Fair — adds implot, node editor, markdown renderer, color text edit. Still tool-oriented at core. | Kitchen sink of ImGui ecosystem. Warns against use for "fully fledged applications." MIT. |
| **egui + eframe** | Rust | Win, Mac, Linux, Android, WASM | ❌ No iOS support | ⚠️ Experimental (soft keyboard issues) | ✅ | Fair — good widget set, but mobile text editing is a known weakness. eframe explicitly warns about mobile limitations. | v0.31+ (pre-1.0). wgpu-backed. Active community. MIT. |
| **Nuklear** | C (ANSI C89) | Desktop only (no built-in platform layer) | ❌ | ❌ | ⚠️ Via sokol | Poor — extremely minimal. Single-header, no platform integration. You provide everything. | Public domain. Good for embedded debug panels, not complex apps. Stagnant development. |
| **raygui** | C (built on raylib) | Win, Mac, Linux, Android, WASM | ❌ No iOS (open PR, experimental) | ✅ | ✅ | Poor — designed for game tool UIs. ~25 basic controls. No rich text, no virtual scrolling, no complex layouts. | zlib license. Great for small tool UIs, not messaging apps. |

#### 5.2 Detailed Elimination Reasoning

**Dear ImGui / Hello ImGui / Dear ImGui Bundle (C++ ecosystem):**
Omar Cornut (Dear ImGui's creator) explicitly states it is "designed for content creation tools and visualization/debug tools, as opposed to UI for the average end-user." While Hello ImGui and Dear ImGui Bundle extend platform support to mobile, the fundamental widget set remains tool-oriented: no rich text editing, no virtual scrolling for chat history, no proper IME support for CJK input, no accessibility APIs. Building a Discord-quality chat UI on Dear ImGui would mean reimplementing most of what a proper framework gives you. Additionally, all three require C++ FFI from Go, introducing build complexity, two memory management models, and CGo overhead on every UI call.

**egui + eframe (Rust):**
egui is the strongest alternative to Gio — it's immediate mode, GPU-accelerated, and has a growing ecosystem. However, it has two critical disqualifiers for Aether: (1) no iOS support at all, and (2) eframe's own documentation warns that "mobile text editing is not as good as for a normal web app" and the on-screen keyboard integration "doesn't always work." For a messaging app where text input is the primary interaction, this is a dealbreaker. Additionally, Rust-Go FFI is non-trivial and would add significant build complexity.

**Nuklear (C):**
Nuklear is a raw single-header library with zero platform integration. It provides immediate mode widgets and nothing else — no window creation, no input handling, no rendering backend. You must provide all of this yourself. For a complex cross-platform app, this means writing thousands of lines of platform glue before rendering a single button. The project has also seen minimal development activity in recent years. Eliminated due to insufficient scope.

**raygui (C / raylib):**
raygui is designed for small game development tools. It provides ~25 basic controls (buttons, sliders, dropdowns) with no support for complex layouts, virtual scrolling, or rich text. More critically, raylib has no iOS support — there's an open PR (#3880) that has been in progress since 2023, with the project maintainer noting that "the code structure required for this iOS implementation is hardly compatible with raylib structure." raygui's 145KB RAM footprint is impressive but irrelevant when the widget set can't support a messaging UI.

#### 5.3 Recommendation: Gio UI (Primary) + Web Client (Secondary)

**Gio UI is the clear winner for this project.** Here's why it wins over every alternative:

1. **Same language as the core (Go).** No FFI bridge. No serialization overhead. No second build system. The core P2P library and the UI compile into one binary. This eliminates an entire class of bugs (memory safety across FFI boundaries, build system fragmentation, platform-specific CGo issues).

2. **True cross-platform including iOS.** Gio is the only immediate-mode framework in this comparison that genuinely supports all target platforms — Windows, macOS, Linux, Android, iOS, and WebAssembly — from a single codebase. The `gogio` tool packages for each platform natively. The September 2025 newsletter (v0.9.0) shows active platform fixes including Android 16KB page size support (required for Google Play by November 2025).

3. **Immediate mode fits real-time communication.** UI is a function of state. When the network state changes (new message, peer connected, voice activity), the next frame reflects it. No callback hell, no state synchronization bugs, no widget tree invalidation. This is how a real-time communication app *should* work.

4. **GPU-accelerated rendering.** Gio renders via compute shaders (Vulkan, Metal, DirectX, OpenGL ES). Vector-based text rendering without baking to textures — supports smooth scaling, animation, and resolution independence. 60fps on mobile with minimal CPU usage.

5. **Aesthetic flexibility.** Gio gives you raw drawing primitives. You're not fighting a pre-existing widget theme — you build the exact aesthetic you want. For a "terminal meets glass" Discord alternative, this is ideal.

6. **Active development with improving text handling.** The January and September 2025 newsletters show ongoing work on text wrapping, font selection, IME improvements, and rich text APIs. The project is specifically working on the areas most critical for a messaging app.

7. **Aligns with open protocol philosophy.** Since the protocol is open, anyone can build a client. The Go + Gio reference client is small, readable, and easy for contributors to hack on. Flutter, Qt, or native platform clients can come later from the community.

**Trade-offs acknowledged:**
- Gio is pre-1.0 (v0.9.0). API may change between minor versions, though the September 2025 newsletter notes breaking changes are "trivial and obscure, unlikely to affect real code."
- Smaller widget library than Flutter. Custom components needed for: virtual scrolling chat list, rich message rendering, emoji picker, voice participant grid.
- Text rendering and IME (input method editors for CJK languages) are improving but not yet at Flutter/native levels.
- Accessibility support is improving but not yet at Flutter/native levels.

**Mitigation:** Build the UI as a thin layer over a well-defined internal API (`pkg/ui/`). The internal API defines what the UI *does* (show messages, render voice grid, display settings). Gio implements *how*. If Gio proves inadequate for a specific platform (e.g., iOS text input), a native wrapper can handle that platform's input while Gio handles rendering. The architecture makes renderer replacement possible without rewriting application logic.

**Secondary: Web Client**
- Go compiled to WASM + Gio's WASM backend covers the browser.
- For users who need a lighter-weight browser experience, a standalone web client using the same WASM core but with a Svelte/React frontend is a future option.
- The protocol is what matters — the web client is just another implementation.

#### 5.3 UI Design Philosophy

The aesthetic is: **terminal meets glass**. Think if Hyper Terminal and Discord had a baby raised by a brutalist architect.

- **Dark by default.** Deep charcoal (#1a1a2e) base, not pure black. Subtle transparency where supported.
- **Monospace for system/code, proportional for chat.** JetBrains Mono for metadata, Inter for message body.
- **Minimal chrome.** No unnecessary borders, shadows, or decorations. Content fills the space.
- **Color as information.** User accent colors, role colors, status indicators — color means something, it's not decoration.
- **Keyboard-first.** Every action reachable via keyboard. Vim-style navigation option.
- **Nerdy details.** Show peer count, DHT routing table size, connection type (direct/relay), encryption status, latency. Power users love seeing the guts. Casual users can hide it.
- **Motion is functional.** Animations are fast (100-200ms), purposeful, and skippable. No gratuitous bounce or slide.

---

## Part III: Revised Version Roadmap

### Addendum A Merge: QoL Reliability and Discovery/Moderation Pull-Forward

This section merges Addendum A into roadmap planning. It is a planning-level override only, not an implementation claim.

#### A) Quality ship gates (release blockers)
- Login-to-ready SLO: warm start ≤3s, cold start ≤8s.
- Delivery SLO: ≥99.9% of messages delivered within 5s on normal networks, with explicit degraded-state UX.
- Call setup SLO: ringing within ≤2s (p50), ≤4s (p95).
- Recovery SLO: Wi‑Fi↔LTE network switch recovers active calls within ≤2s (p95) without user action.
- Stability gates: crash-free sessions ≥99.5% on supported devices; no release if call setup p95 regresses >10% versus previous stable.

#### B) Connectivity subsystems now explicit
- **Connectivity Orchestrator (CO):** continuous reachability management, path selection/fallback, recovery, and deterministic reason-coded diagnostics.
- **Aether Tunnel:** opportunistic QUIC-based encrypted overlay session for unstable/failed NAT scenarios, with migration + multiplexed streams.

#### C) Discovery contracts: `DirectoryEntry` + indexer trust model
- Public discovery uses signed `DirectoryEntry` records (opt-in only), published via deterministic DHT keys.
- Keyword search uses optional, community-run indexers; indexers are non-authoritative and must return verifiable signed records.
- Clients support multi-indexer querying and local de-duplication; no special authoritative node class is introduced.

#### D) Moderation protocol events + enforcement semantics
- Moderation is a protocol contract (not UI-only): signed events for redaction/delete, timeout, ban, and slow mode.
- Official clients must enforce signed moderation events and expose enforcement state; auditability remains append-only and signed.

#### E) Pull-forward roadmap overrides (scope timing)
- **v0.2:** basic RBAC + baseline moderation + slow mode.
- **v0.3:** directory/explore/indexer + invite/request-to-join.
- **v0.4:** advanced moderation (policy versioning + auto-mod) and full custom role/override model.
- **v0.6:** hardening/scaling of discovery, moderation reliability, and anti-abuse systems.

#### F) Planned QoL experience invariants (planning contract, not implementation claim)
- **Global no-limbo UX invariant:** no critical user journey may end in ambiguous waiting; users must always see current state, deterministic reason, and next recovery action.
- **Unified connection health/recovery clarity:** startup, messaging, sync, and calls use one canonical health model and recovery progression language.
- **Recovery-first call experience:** transient failures prioritize automatic recovery and rejoin paths before hard-fail outcomes.
- **Deterministic reason taxonomy for user-visible states:** official clients map protocol/runtime conditions to stable reason classes and consistent user-facing explanations.
- **Unread/mention/notification coherence as a first-class contract:** unread badges, mention state, push/local notifications, and read markers converge deterministically.
- **Cross-device continuity contract:** draft state, read position, and call handoff behavior are explicitly specified as conflict-safe continuity semantics.
- **Hidden-delight micro-interactions are planned requirements:** auto-heal path transitions, smart device-switch prompts, and exact attention resume are roadmap-level UX contracts.

#### G) Planned roadmap guidance: journey gates and QoL scorecards
- Planning gates must evaluate end-to-end user journeys (login-to-ready, message send under degradation, call join/rejoin, network switch mid-call, wake-and-open from mention, and cross-device resume).
- Each journey gate must include a QoL scorecard with deterministic pass/fail evidence for limbo avoidance, recovery success, reason-taxonomy coverage, coherence of unread/mention/notification state, and continuity correctness.
- This guidance remains planning-only in this document and does not claim delivered behavior.
- Any protocol-surface evolution introduced for QoL objectives remains bound by additive-only minor evolution and major-path governance requirements.

#### H) Guardrails unchanged
- Protocol-first framing remains mandatory: protocol/spec contract is the product.
- Compatibility/governance remains unchanged: protobuf minor updates are additive-only; major changes require new multistream IDs + downgrade negotiation + AEP process + multi-implementation validation.
- Open decisions remain open, including mobile-notification wake centralization risk.

### v0.1.0 — "First Light" (MVP)
**Duration: 6–8 weeks**

**Goal:** Working P2P client. Create identity, create/join server, text chat, voice chat. Two people on a LAN should be able to talk within 5 minutes of first launch.

#### Core Infrastructure
- [ ] Go project scaffold: monorepo with `cmd/aether` (client), `cmd/relay` (headless), `pkg/` (shared libraries)
- [ ] `Makefile` / `Taskfile.yml` with full pipeline: `make all` runs generate → compile → lint → test → scan → build
- [ ] `.golangci.yml` with full linter suite (staticcheck, gosec, errcheck, revive, exhaustive, go-critic)
- [ ] GitHub Actions CI: `ci.yml` (every push), `nightly.yml` (extended fuzz), `security-audit.yml` (weekly)
- [ ] Pre-commit hooks: `gofmt`, `golangci-lint`, `go vet` on staged files
- [ ] Dhall configuration sources with `types.dhall`, `defaults.dhall`, environment overrides
- [ ] `buf` setup for protobuf: schema validation, breaking change detection, Go code generation
- [ ] Property-based test scaffolding with `rapid` for crypto and protocol packages
- [ ] Docker multi-stage build with pinned base images, distroless runtime, SBOM generation
- [ ] libp2p host initialization: TCP + QUIC transports, Noise encryption, Yamux muxer
- [ ] Kademlia DHT (private network, custom protocol prefix `/aether/kad/1.0.0`)
- [ ] GossipSub v1.1 with custom topic naming (`/aether/srv/<id>/ch/<id>`)
- [ ] AutoNAT + Circuit Relay v2 + DCUtR hole punching
- [ ] mDNS for LAN discovery
- [ ] Local peer cache: persist known peers to SQLCipher, try cached peers before bootstrap on startup
- [ ] 3 bootstrap nodes deployed (US, EU, Asia) — simple Docker containers running `--mode=bootstrap`
- [ ] Protobuf message definitions v1.0 for: ChatMessage, ServerManifest, VoiceState, Identity, Capabilities
- [ ] Multistream protocol registration for all sub-protocols with `/1.0.0` versions

#### Identity & Accounts
- [ ] Ed25519 keypair generation + BIP39 mnemonic recovery
- [ ] Profile: display name, avatar (local image), accent color, bio
- [ ] Profile published as signed protobuf to DHT
- [ ] SQLCipher local database with migration system
- [ ] Settings UI: profile editing, audio devices, network info

#### Server & Text Chat
- [ ] Community Manifest creation (signed protobuf, published to DHT)
- [ ] Join server via `aether://join/<id>` deeplink
- [ ] GossipSub-based text channels with Sender Keys E2EE
- [ ] Message rendering: markdown subset, timestamps, replies
- [ ] Message history sync from peers on join (last 500 messages)
- [ ] Discord-like UI layout: server rail → channel sidebar → chat → member list

#### Voice
- [ ] Pion WebRTC integration: Opus 48kHz, 10ms frames
- [ ] Signaling via GossipSub (encrypted SDP/ICE exchange)
- [ ] Full mesh voice (up to 8 participants)
- [ ] SFrame E2EE for voice streams
- [ ] Speaking indicator (VAD), self-mute, self-deafen
- [ ] Audio device selection, push-to-talk option
- [ ] Voice connected bar (persistent UI element showing current channel)

#### Headless Relay (MVP)
- [ ] `--mode=relay` flag: runs without GUI
- [ ] Provides DHT bootstrap + Circuit Relay v2
- [ ] Basic store-and-forward for offline DMs (encrypted, 7-day TTL)
- [ ] Docker image published to ghcr.io
- [ ] `relay.toml` configuration file

#### Gio UI Shell
- [ ] App window with dark theme
- [ ] Server list sidebar (icons, unread indicators)
- [ ] Channel list (text #, voice 🔊)
- [ ] Message list with virtual scrolling
- [ ] Message composer (Enter to send, Shift+Enter for newline)
- [ ] Voice channel participant grid
- [ ] Responsive layout (desktop/tablet/mobile breakpoints)

---

### v0.2.0 — "Connections" (DMs, Friends, Presence)
**Duration: 4–5 weeks**

- [ ] X3DH key agreement + Double Ratchet for 1:1 DMs
- [ ] Prekey bundles published to DHT
- [ ] DM transport: direct stream or store-and-forward via DHT
- [ ] Group DMs (Sender Keys, up to 50 members)
- [ ] Friend requests via public key / QR code / aether:// link
- [ ] Presence system: online, idle (10min auto), DND, invisible
- [ ] Custom status text
- [ ] Friends list UI with online/offline/pending tabs
- [ ] Notification system (in-app badges, unread counts)
- [ ] @mentions: `@user`, `@role`, `@everyone`, `@here`
- [ ] Basic RBAC baseline: Owner/Admin/Moderator/Member
- [ ] Baseline moderation protocol events + enforcement: redaction/delete, timeout, ban
- [ ] Channel slow mode (deterministic per-channel enforcement)

---

### v0.3.0 — "Clarity" (Voice Quality, Noise Cancellation, Screen Share)
**Duration: 5–6 weeks**

- [ ] RNNoise integration via C FFI for noise cancellation
- [ ] Opus adaptive bitrate (16–128 kbps based on network)
- [ ] Adaptive jitter buffer (20–200ms)
- [ ] FEC + DTX enabled
- [ ] Peer SFU election for 9+ participant voice channels
- [ ] SFU mode in relay nodes (`--sfu-enabled=true`)
- [ ] Screen capture: platform-native (entire screen or window selection)
- [ ] Hardware encoder detection and selection (NVENC/QuickSync/VideoToolbox/VA-API)
- [ ] Quality presets: Low/Standard/High/Ultra/Auto
- [ ] Simulcast encoding (up to 3 layers)
- [ ] Screen share viewer: fullscreen, PiP, zoom/pan
- [ ] P2P file transfer (inline in chat, up to 25MB, chunked)
- [ ] Image inline preview, file attachment cards
- [ ] Public `DirectoryEntry` publication (opt-in) + DHT retrieval contract
- [ ] Explore/Discover browsing + server preview for public listings
- [ ] Invite system + request-to-join flow
- [ ] Initial community-run indexer reference (`cmd/indexer`, Docker) + signed/verifiable search responses

---

### v0.4.0 — "Dominion" (RBAC, Permissions, Moderation)
**Duration: 4–5 weeks**

- [ ] Permission bitmask system (server-level + channel overrides)
- [ ] Role CRUD: create, edit, delete, reorder
- [ ] Role properties: name, color, icon, permissions, hoisted, mentionable
- [ ] Channel categories (collapsible groups)
- [ ] Channel types: text, voice, announcement, stage
- [ ] Channel security modes + disclosure + auditable mode transitions (Tree/Crowd/Channel/Clear)
- [ ] Full custom roles + deterministic channel permission overrides
- [ ] Advanced moderation policy: versioning, migration rules, and rollback semantics
- [ ] Auto-moderation hooks (rate limits, keyword filters, extensible policy pipeline)
- [ ] Audit log expansion (signed entries + policy traceability for authorized roles)

---

### v0.5.0 — "Pulse" (Bots, Emoji, Extensibility)
**Duration: 5–7 weeks**

- [ ] Native Bot API (gRPC): events + commands
- [ ] Bot SDK for Go (first-class), with community SDKs for Python/JS/Rust
- [ ] Slash commands with autocomplete
- [ ] Discord API compatibility shim (`aether-discord-shim`)
  - REST API subset: messages, channels, guilds, members, roles
  - WebSocket Gateway: event translation
  - Coverage target: 80% of common discord.py/discord.js patterns
  - Documented limitations + migration guide
- [ ] Custom emoji system (server owners upload, max 50 per server)
- [ ] Emoji picker, `:shortcode:` syntax
- [ ] Reactions on messages
- [ ] Incoming webhooks (POST → message in channel)

---

### v0.6.0 — "Sentinel" (Hardening, Safety, Anti-Spam)
**Duration: 4–5 weeks**

- [ ] Discovery/indexer hardening: freshness validation, poisoning resistance, deterministic dedupe/merge
- [ ] Multi-indexer query strategy hardening (privacy-preserving query options, fallback behavior)
- [ ] Moderation reliability hardening under partition/rejoin with deterministic event replay
- [ ] Proof-of-Work on identity creation (~5s computation, Hashcash-style)
- [ ] Per-user rate limiting (5 msgs/5s, enforced locally)
- [ ] GossipSub peer scoring (flood protection, penalize misbehavior)
- [ ] Web-of-trust reputation system
- [ ] Block user (local mute), report to server moderators
- [ ] Optional per-server content filters hardening (keyword/regex, on-device ML image classification)

---

### v0.7.0 — "Archive" (Offline Delivery, Search, History)
**Duration: 4–5 weeks**

- [ ] Robust DHT store-and-forward (30-day TTL, k=20 replication)
- [ ] History sync protocol with Merkle tree verification
- [ ] History migration across security-mode epochs (optional re-encryption / History Capsule workflow)
- [ ] Configurable history retention per server
- [ ] "Archivist" role: peers that volunteer full history storage
- [ ] SQLCipher FTS5 full-text search (scoped by channel/server/DMs)
- [ ] Search filters: from user, date range, has file, has link
- [ ] Push notification relay (encrypted payload → FCM/APNs, relay sees nothing)
- [ ] Desktop native notifications

---

### v0.8.0 — "Echo" (Polish, Themes, Rich Content)
**Duration: 3–4 weeks**

- [ ] Threaded replies (sub-conversations)
- [ ] OpenGraph/Twitter card link previews (fetched client-side)
- [ ] Message pinning + personal bookmarks
- [ ] Theme system: Dark, Midnight, Light, AMOLED Black + custom JSON themes
- [ ] Accessibility: screen reader support, high contrast, keyboard nav
- [ ] i18n: English, Spanish, German, French, Japanese, Chinese, Portuguese
- [ ] Upgrade noise cancellation from RNNoise to DTLN

---

### v0.9.0 — "Forge" (Performance, Scale)
**Duration: 4–5 weeks**

- [ ] IPFS integration for persistent file hosting (pinning by server owners)
- [ ] Large server optimization: hierarchical GossipSub, lazy member loading
- [ ] Scale-driven security-mode transitions + sharding guidance for huge interactive channels
- [ ] Cascading SFU mesh for 200+ participant voice
- [ ] Performance profiling and optimization across all platforms
- [ ] Stress testing: 1000-member server, 50-person voice, latency benchmarks
- [ ] Relay node performance optimization and load testing
- [ ] Battery optimization on mobile (background activity reduction)

---

### v1.0.0 — "Genesis" (Public Release)
**Duration: 4–6 weeks**

- [ ] External security audit (E2EE, P2P networking, key management, DHT security)
- [ ] Aether Protocol Specification v1.0 published as open standard
- [ ] User guide, admin guide, developer guide, API reference
- [ ] Landing page and comparison site
- [ ] App store submissions (Google Play, App Store, Microsoft Store, Flathub)
- [ ] Expand bootstrap infrastructure to 10+ global nodes
- [ ] Community relay node program launch
- [ ] Reproducible builds for binary verification

---

## Part IV: Post-v1.0

### v1.1 — "Bridges"
- Matrix, IRC, and Discord bidirectional bridges.

### v1.2 — "Canvas"
- Collaborative whiteboard, polls, calendar/events.

### v1.3 — "Agora"
- Forum channels, wiki channels, server templates.

### v1.4 — "Horizon"
- Client plugin system, app directory, reproducible builds.

---

## Timeline Summary

| Version | Codename | Duration | Cumulative |
|---|---|---|---|
| v0.1.0 | First Light (MVP) | 6–8 weeks | 6–8 weeks |
| v0.2.0 | Connections | 4–5 weeks | 10–13 weeks |
| v0.3.0 | Clarity | 5–6 weeks | 15–19 weeks |
| v0.4.0 | Dominion | 4–5 weeks | 19–24 weeks |
| v0.5.0 | Pulse | 5–7 weeks | 24–31 weeks |
| v0.6.0 | Sentinel | 4–5 weeks | 28–36 weeks |
| v0.7.0 | Archive | 4–5 weeks | 32–41 weeks |
| v0.8.0 | Echo | 3–4 weeks | 35–45 weeks |
| v0.9.0 | Forge | 4–5 weeks | 39–50 weeks |
| v1.0.0 | Genesis | 4–6 weeks | 43–56 weeks |

**~10–14 months with a team of 2–4 developers.**

---

## Key Technical Risks & Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| **NAT traversal failures** | High | 5-layer fallback: STUN → ICE → hole punch (DCUtR) → circuit relay → manual peer. Community relay nodes. |
| **Gio UI maturity** | High | Build UI as thin layer over internal API (`pkg/ui/`). Gio v0.9.0 is actively maintained with platform fixes. If Gio proves inadequate for a platform, swap renderer without rewriting logic. All 6 alternatives evaluated (Dear ImGui ecosystem, egui, Nuklear, raygui) have worse trade-offs for this use case. |
| **Voice quality with P2P SFU** | High | Aggressive Opus tuning (10ms frames, FEC, DTX). Automatic topology switching. Relay SFU fallback. LiveKit's architecture (Go+Pion, Apache 2.0) as reference implementation. |
| **E2EE key management at scale** | High | TreeKEM for 50+ member channels. Formal verification of crypto primitives. External audit before v1.0. |
| **DHT reliability for store-and-forward** | Medium | k=20 replication, configurable TTL, relay nodes as super-peers for storage. Graceful degradation with user notification. |
| **Discord bot shim completeness** | Medium | 80/20 rule: cover 80% of functionality used by 80% of bots. Document unsupported features. Native API always preferred. |
| **Protocol evolution coordination** | Medium | Three-layer versioning. 90-day review for breaking changes. N+2 backward compatibility guarantee. |
| **Sybil attacks** | Medium | PoW on identity, reputation system, web-of-trust, server-configurable join requirements. |
| **Go WASM binary size** | Medium | Tree-shake unused code. Compress with Brotli. Target <5MB for initial load. Use lazy loading for non-critical modules. |
| **App store approval** | Low-Med | Comply with all platform guidelines. Emphasize safety features. Community relay infrastructure. |

---

## Open Decisions

All items below remain unresolved candidate options and must not be treated as finalized architecture until explicit governance ratification.

1. **Project name** — "Aether" is a working title. Must be trademarkable and have an available `.chat` or `.app` domain.
2. **Relay incentive model** — Start with reputation-based (no tokens). Revisit if relay supply is insufficient.
3. **Maximum tested server size** — Needs load testing. Architectural target: 10,000 members. Voice target: 200 simultaneous in one channel.
4. **Governance** — Candidate model (not finalized): protocol governance could be handled via a nonprofit foundation (similar to Let's Encrypt for ACME). Bootstrap/relay infrastructure funding remains open, with community support and grants (e.g., NLnet, NGI) as candidate paths.
5. **Mobile notification relay** — Must be carefully designed to avoid becoming a centralization point. Open-source relay software, easy to self-host, community-run default instances.

**Funding & Licensing:** Aether is fully open source — no premium features, no paywalled tiers, no monetization. The project sustains through community contributions, grants from organizations that fund open-source privacy/P2P infrastructure (NLnet, NGI, Open Technology Fund), and donations. All code is licensed permissively (MIT or similar). The protocol specification is CC-BY-SA to prevent proprietary incompatible forks.
