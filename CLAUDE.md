# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

**xorein** is both a protocol family specification and a runtime binary for a secure-by-default, P2P-first chat system with explicit security modes. It implements multiple node roles (client, relay, bootstrap, archivist) within a single binary, using libp2p for networking and multistream-select for protocol negotiation.

The system is designed with:
- **Protocol-first design**: Security modes (Seal, Tree, Crowd, Channel, Clear) are user-visible and explicitly negotiated
- **Capability-based architecture**: Feature flags (cap.chat, cap.voice, cap.dm, etc.) allow fine-grained negotiation
- **Additive-only evolution**: Wire protocol changes use multistream-select versioning; protobuf changes are additive-only
- **Single-binary multi-role design**: All node types are the same binary, differentiated by role flags and capability enablement

## Build, Test, Lint, and Run Commands

### Core Development Workflow

```bash
# Fast readiness check (compile + lint)
make check-fast

# Full pipeline (generate → compile → lint → test → scan → build)
make pipeline

# Individual stages
make generate    # protobuf compatibility checks (buf lint/breaking)
make compile     # go build ./...
make lint        # pre-commit + golangci-lint
make test        # Go tests (+ dhall verification + reproducibility)
make race        # race condition detection (go test -race)
make scan        # security scan (govulncheck + gosec + trivy)
make build       # final binary to bin/aether
make clean       # remove artifacts and bin/
```

### Running the Runtime

```bash
# Build first
make build

# Start the v0.1 runtime (default)
./bin/aether --listen 127.0.0.1:0

# Custom listen address
./bin/aether --listen 0.0.0.0:9000
```

### Container Builds

```bash
# Build relay container image
make relay-container-build

# Full relay publishing workflow (build → sign → sbom → publish checklist)
make relay-container-workflow

# Generate release pack verification bundle
make release-pack-verify
```

### Key Diagnostics

```bash
# Preflight: repo snapshot + baseline health check, exit
./bin/aether --preflight --repo-root .

# Repo snapshot only
./bin/aether --repo-snapshot --repo-root .

# Baseline health (build/test/lint/typecheck checks, exit)
./bin/aether --baseline-health --repo-root .
```

## Architecture and Design

### High-Level Design (v0.1 Spec Runtime)

The system has **four layers**:

1. **Transport** (`pkg/v0_1/transport/`): libp2p host + multistream family handler registration
2. **Protocol** (`pkg/v0_1/protocol/`): Capability negotiation, error codes, operation→cap mapping
3. **Families** (`pkg/v0_1/family/`): 13 spec-conformant protocol family handlers
4. **Runtime** (`cmd/aether/main.go` + `pkg/v0_1/runtime.go`): CLI, startup/shutdown orchestration

### Entry Points and Key Files

- **`cmd/aether/main.go`**: CLI parsing, invokes v0.1 runtime
- **`pkg/v0_1/runtime.go`**: libp2p host setup, family handler registration
- **`pkg/v0_1/family/`**: One sub-package per protocol family:
  - `chat/`, `dm/`, `groupdm/`, `voice/`, `manifest/`, `identity/`, `sync/`, `friends/`, `presence/`, `notify/`, `moderation/`, `governance/`, `peer/`
- **`pkg/v0_1/protocol/`**: Capability flags, error codes, `RequiredCapsFor`, `NegotiateCapabilities`
- **`pkg/v0_1/transport/`**: `FamilyHandler` interface, `RegisterFamily`
- **`pkg/v0_1/spectest/`**: KAT-driven spec conformance tests for all families + W6 integration
- **`docs/spec/v0.1/91-test-vectors/`**: 56 pinned JSON KAT vector files
- **`pkg/storage/store.go`**: Encrypted SQLite state store (SQLCipher) with key derivation

### Protocol Families

All 13 spec families are implemented in `pkg/v0_1/family/`. Each family:
- Implements `transport.FamilyHandler` with a spec-versioned `ProtocolID()`
- Has a corresponding KAT spectest under `pkg/v0_1/spectest/`
- Is registered in `pkg/v0_1/runtime.go`

### Capability Negotiation

Each `HandleStream` call:

1. Looks up required capabilities for the operation via `proto.RequiredCapsFor(op)`
2. Calls `proto.NegotiateCapabilities(localCaps, remoteAdvertised, required)` to compute:
   - **Accepted**: caps both sides support
   - **IgnoredRemote**: remote caps local doesn't know
   - **MissingRequired**: required caps not in local support
3. Returns `MISSING_REQUIRED_CAPABILITY` if any required cap is absent

### Security Modes (Explicit, User-Visible)

Defined in `pkg/v0_1/protocol/capabilities.go` and `pkg/protocol/capabilities.go`:

- **Seal**: 1:1 E2EE (X3DH + Double Ratchet), `mode.seal`
- **Tree**: Interactive group E2EE (MLS), `mode.tree`
- **Crowd/Channel**: Large-scale E2EE (epoch rotation)
- **MediaShield**: E2EE media frames via SFrame, `mode.mediashield`
- **Clear**: Readable by infrastructure (must be explicitly labeled)

### State Persistence

```go
storage.Load(dataDir) → Snapshot{SchemaVersion, Buckets map[string][]byte}
```

Each bucket is JSON-encoded and versioned via `SchemaVersion`. Migrations are schema-aware.

### Storage Security

- **Encryption**: SQLCipher with AES-256; key derived from:
  - 32-byte random salt (stored in `state.db.meta.json`)
  - 32-byte secret from `$XOREIN_STATE_KEY` env var OR `state.key` file (base64)
  - SHA256(salt || secret) → encryption key
- **Key verification**: `state.db.meta.json` stores a `key_check` hash to detect wrong keys early
- **Format**: `FormatVersion` = 2; handles v1→v2 migration + legacy store archival
- **SQLCipher settings**: WAL mode, foreign key constraints, 5s busy timeout, 4096-byte page size

### Peer Discovery (Layered, Resilient)

Order of discovery (as implemented in `pkg/v0_1/discovery/`):

1. **Live cache** (`cache.go`) — in-memory, 24h addr TTL / 15m mDNS TTL / no expiry for manual
2. **LAN discovery** (`mdns.go`) — custom-TXT mDNS (`_aether._udp.local`), 60s announce
3. **Bootstrap** (`bootstrap.go`) — `bootstrap.register`/`bootstrap.fetch` RPCs, capped at 200
4. **DHT walking** (`dht.go`) — kad-dht at `/aether/kad/0.1.0`, provider records, 24h TTL
5. **PEX** (`pex.go`) — peer exchange, anti-flood (5 min tentative window), max 50 returned
6. **Manual peers** (`manual.go`) — `--manual-peers` CSV, `Source=manual`, never expire
7. **Loop** (`loop.go`) — 250ms tick; gates DHT walks to 10s and bootstrap to 30s intervals

Backoff schedule per peer: immediate / 5s / 30s / 2m / 5m / 10m cap.

### Wire Protocol and Control Plane

**Peer-to-peer operations** flow through `network.Handler.HandlePeerOperation()`:

```
Stream received → unmarshalRequest() → protocol.NegotiatePeerTransport()
→ service.HandlePeerOperation(operation, payload)
→ marshalResponse() → writeStreamPayload()
```

**Control plane** (local HTTP on configurable socket/port):

- Routes for identity, peers, messages, channels, DMs, voice, relay queue, etc.
- Authentication: currently none (assume local access only)
- Socket path: `--control /tmp/aether.sock` or port (Windows fallback)

## Key Data Types and Interfaces

### Core Interfaces

```go
// Handler: implemented by Service; called when peer streams arrive
network.Handler interface {
    HandlePeerOperation(ctx, operation, payload) (response []byte, err *Error)
    Snapshot() Snapshot  // current state for diagnostics
}

// Runtime: implemented by P2PRuntime; manages libp2p lifecycle
network.Runtime interface {
    Start(ctx) error
    Close() error
    ListenAddress() string
}
```

### Types in `pkg/node/types.go`

- **Config**: node startup configuration (role, data dir, listen addr, peers, etc.)
- **Identity**: peer identity with public/private keys + profile
- **PeerRecord**: discovered peer metadata
- **MessageRecord**: chat message with scope (channel/dm), sender, timestamps
- **ChannelRecord**: server-side grouping
- **DMRecord**: direct message pair metadata
- **VoiceSession**: active WebRTC participants
- **RelayQueueEntry**: stored message for relay-mode store-and-forward

### Request/Response Envelope (in `pkg/node/wire.go`)

```go
type requestEnvelope struct {
    Operation string                  // e.g., "send_message"
    Payload []byte                    // operation-specific binary
    AdvertisedCapabilities []string   // features this peer supports
    RequiredCapabilities []string     // features needed for this operation
}

type responseEnvelope struct {
    NegotiatedProtocol string        // e.g., "/aether/chat/0.1.0"
    AcceptedCapabilities []string    // what both sides agreed on
    IgnoredCapabilities []string     // remote caps we don't support
    Payload []byte                   // operation response
    TransportError Error             // if negotiation or operation failed
}
```

## Development Workflows

### Adding a New Operation or Capability

1. **Define protocol ID** in `pkg/protocol/registry.go` if new family (e.g., `ProtocolID{Family: FamilyFoo, Version: ProtocolVersion{0, 1, 0}}`)
2. **Add FeatureFlag** in `pkg/protocol/capabilities.go` (e.g., `FeatureFoo FeatureFlag = "cap.foo"`)
3. **Add to defaults** in the same file (add to `defaultFeatureFlags` or role-specific subset)
4. **Implement handler** in `pkg/node/service.go` → `HandlePeerOperation()` switch case
5. **Marshal/unmarshal** in `pkg/node/wire.go` if using custom envelope formats
6. **Add tests** in `*_test.go` files (87 test files across the codebase)

### Protobuf Changes

1. Edit `proto/aether.proto`
2. **NEVER edit** `gen/go/proto/aether.pb.go` directly; it is generated
3. Run protobuf generation (wired into `make generate` if using buf)
4. Check breaking changes:
   ```bash
   buf lint                              # syntax
   buf breaking --against '.git#branch=main'  # wire-level compatibility
   ```
5. Follow additive-only rule: reserve removed field numbers/names
6. Regenerate if needed (buf will guide you)

### Testing Patterns

- **Unit tests**: `*_test.go` files in same package
- **KAT (spec) tests**: `pkg/v0_1/spectest/{family}/` — loader-driven against pinned JSON vectors
- **W6 integration**: `pkg/v0_1/spectest/w6/w6_test.go` — multi-family scenario tests
- **Run all**: `make test` (also runs `dhall-verify.sh` + reproducibility checks)
- **Run single**: `go test ./pkg/v0_1/spectest/governance/... -v`
- **Race detector**: `make race` or `go test -race ./...`

### Security Considerations

**DO NOT:**
- Log secrets, private keys, key material, or sensitive payloads
- Commit `.opencode.zip` or generated artifacts to main branch
- Introduce wire protocol changes without `multistream-select` versioning
- Use non-additive protobuf changes (reserved numbers must be documented)

**MUST:**
- Keep encryption at rest (SQLCipher) operational; key derivation is in `pkg/storage/store.go`
- Verify capability negotiation before processing operations
- Clear mode must be explicitly labeled and never default for private conversations
- Relay nodes must never decrypt or inspect message content (ciphertext-only)

## Repository Shape and Guardrails

- **Real runtime entrypoint**: `cmd/aether/main.go`
- **v0.1 runtime wiring**: `pkg/v0_1/runtime.go`
- **Protocol families**: `pkg/v0_1/family/{chat,dm,groupdm,voice,sync,moderation,governance,...}/`
- **Capability negotiation**: `pkg/v0_1/protocol/`
- **Local Control API**: `pkg/v0_1/control/` — HTTP server (Unix socket / Windows TCP), bearer-token auth, ~30 endpoints, SSE; spec 60
- **Discovery**: `pkg/v0_1/discovery/` — live cache, custom-TXT mDNS, Kademlia DHT (`/aether/kad/0.1.0`), bootstrap client/server, PEX, manual peers, 250ms loop; spec 31
- **NAT traversal**: `pkg/v0_1/nat/` — Circuit Relay v2 client + service, DCUtR tracer, fallback cascade, connection-type tracker; spec 32
- **Spec test vectors**: `docs/spec/v0.1/91-test-vectors/` (pinned in `pin.sha256`)
- **KAT spectest packages**: `pkg/v0_1/spectest/{family}/`
- **Protobuf source**: `proto/aether.proto`; generated code: `gen/go/proto/aether.pb.go`
- **Build outputs**: `artifacts/generated/**`, `bin/aether`; treat as outputs, not source

### CI and Security Gates

- **No `.github/workflows/*` checked in** (Makefile is source of truth)
- **Real security gates**: `govulncheck`, `gosec`, `trivy` (via `make scan`)
- **Linting**: `pre-commit` hooks + `golangci-lint` with conservative rules (no default excludes)
- **Reproducibility**: `scripts/dhall-verify.sh`, `scripts/repro-checksums.sh`, `scripts/release-pack-verify.sh`

## Additional Resources

- **Specification & Protocol**: `ENCRYPTION_PLUS.md` (security modes, capability specs, threat model)
- **Architecture & Evolution**: `aether-v3.md` (protocol design notes)
- **QoL & Discovery**: `aether-addendum-qol-discovery.md`
- **Release naming**: `docs/release-naming.md`
- **Local control API v1**: `docs/local-control-api-v1.md`
