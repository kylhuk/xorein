# 42 — Family: Manifest (`/aether/manifest/0.1.0`)

This document specifies the Manifest family, which provides server descriptor
distribution and publication across the peer network.

## 1. Overview

A Manifest is the authoritative, signed descriptor for a Xorein server. It
encodes the server's identity, capability set, security posture, history
parameters, and connectivity hints (bootstrap and relay addresses). All nodes
that participate in a server MUST hold the current manifest for that server.

The Manifest family defines two peer-to-peer operations: fetching a manifest
from a node that hosts a server, and publishing an updated manifest to peers.

**Roles that use this family:**
- **Client**: Fetches manifests when joining or reconnecting; receives published
  manifest updates from server owners.
- **Relay**: SHOULD cache and serve manifests for servers they relay, to reduce
  load on server owners.
- **Bootstrap**: MAY serve manifests for servers registered with them.
- **Archivist**: MUST serve manifests for archived servers.

**Security modes:** Manifests are signed with the hybrid Ed25519 + ML-DSA-65
scheme and are mode-independent. The manifest declares which security modes the
server offers via `offered_security_modes` and which mode is currently active
via `security_mode`. The manifest itself is always transmitted and stored in
the clear (the signature provides integrity).

**Protocol ID:** `/aether/manifest/0.1.0`

**Required capability:** `cap.manifest` — MUST be present in `advertised_caps`
for all Manifest family streams.

## 2. Capability requirements

| Capability | Required for |
|------------|-------------|
| `cap.manifest` | All Manifest family operations |

The `cap.peer.manifest` capability (distinct from `cap.manifest`) indicates that
a peer can include manifest metadata in peer exchange responses (Peer family,
`40-family-peer.md §4.2`). A node that advertises `cap.manifest` SHOULD also
advertise `cap.peer.manifest` if it serves manifests.

## 3. Operations

| Operation | Required caps | Direction | Request payload type | Response payload type | Description |
|-----------|--------------|-----------|---------------------|----------------------|-------------|
| `manifest.fetch` | `cap.manifest` | initiator → responder | `ManifestFetchRequest` | `ManifestFetchResponse` | Request the current manifest for a server |
| `manifest.publish` | `cap.manifest` | server owner → peers | `ManifestPublishRequest` | `ManifestPublishResponse` | Announce an updated server manifest to peers |

The `operation` field in `PeerStreamRequest` MUST be one of the strings above.

## 4. Wire format details

All payloads are JSON-encoded and stored in `PeerStreamRequest.payload` /
`PeerStreamResponse.payload`.

### 4.1 `manifest.fetch`

**Request:** `ManifestFetchRequest` — JSON-encoded:

```
{
  "server_id": string   // UUID of the server whose manifest is requested
}
```

**Response:** `ManifestFetchResponse` — JSON-encoded:

```
{
  "manifest": { /* Manifest JSON; see §4.3 */ },
  "error":    string   // human-readable; omitted on success
}
```

Notes:
- The responder MUST return the current manifest for the requested `server_id`.
- If the responder does not hold a manifest for the given `server_id`, it MUST
  respond with `OPERATION_FAILED` and message `"manifest not found"`.
- A stale manifest (past its `expires_at`) MUST NOT be served. Instead, the
  responder MUST attempt to refresh the manifest from the server owner before
  responding, or return `OPERATION_FAILED` with message `"manifest expired"`.

### 4.2 `manifest.publish`

**Request:** `ManifestPublishRequest` — JSON-encoded:

```
{
  "manifest": { /* Manifest JSON */ }
}
```

**Response:** `ManifestPublishResponse` — JSON-encoded:

```
{
  "accepted": bool,
  "error":    string
}
```

Notes:
- The receiver MUST verify the manifest signature before accepting (§4.4).
- A manifest MUST be rejected if its `signature` fails verification.
- A manifest MUST be rejected if its `owner_peer_id` does not match the
  initiating peer's ID as established in the Noise XX handshake.
- A manifest MUST be accepted if it is newer than the locally held manifest for
  the same `server_id` (compare `updated_at` timestamps).
- If the receiver holds a manifest with the same `server_id` and the received
  manifest has an older or equal `updated_at`, the receiver MUST reject with
  `OPERATION_FAILED` and message `"manifest not newer"`.
- Upon acceptance, the receiver MUST propagate the manifest to other connected
  peers (gossip). The gossip hop limit is 3 hops from the origin.

### 4.3 `Manifest` struct

The Manifest is defined in `pkg/node/wire.go` as the `Manifest` struct. Its
JSON representation is:

```
{
  "server_id":                    string,    // UUID v4; permanent server identifier
  "name":                         string,    // human-readable server name
  "description":                  string,    // optional; human-readable description
  "owner_peer_id":                string,    // libp2p peer ID of the server owner
  "owner_public_key":             string,    // Ed25519 public key, base64url no-padding
  "owner_addresses":              [string],  // multiaddrs; sorted lexicographically
  "bootstrap_addrs":              [string],  // bootstrap node addresses; sorted
  "relay_addrs":                  [string],  // relay node addresses; sorted
  "capabilities":                 [string],  // capability flags this server requires; sorted
  "security_mode":                string,    // active mode: "seal" | "tree" | "crowd" | "channel" | "clear"
  "offered_security_modes":       [string],  // preference-ordered list of modes the server offers
  "history_retention_messages":   int,       // number of recent messages retained; 0 = unlimited
  "history_coverage":             string,    // "local-window" | archivist coverage range string
  "history_durability":           string,    // "single-node" | "quorum"
  "issued_at":                    string,    // RFC3339Nano; original creation time
  "updated_at":                   string,    // RFC3339Nano; last modification time
  "expires_at":                   string,    // RFC3339Nano; zero value means no expiry
  "signature":                    string     // hybrid signature; see §4.4
}
```

The constants `HistoryCoverageLocalWindow = "local-window"` and
`HistoryDurabilitySingleNode = "single-node"` from `pkg/node/wire.go` define
the standard values for those fields.

### 4.4 Manifest signing and verification

Manifest signing and verification follows the canonical form defined in
`02-canonical-envelope.md §3.1`.

**Canonical form for signing:**
- The `signature` field MUST be absent.
- `owner_addresses`, `bootstrap_addrs`, `relay_addrs`, and `capabilities` MUST
  be sorted lexicographically (ascending) before signing.
- `offered_security_modes` is preference-ordered and MUST NOT be sorted.
- JSON keys MUST be in lexicographic order (see `02-canonical-envelope.md §3.5`).

**Signature construction:**

```
canonical_json = UTF-8(JSON(manifest_without_signature))
combined_sig   = Ed25519.Sign(owner_ed_priv, canonical_json)
              || ML-DSA-65.Sign(owner_mldsa_priv, canonical_json)
signature      = base64url_no_padding(combined_sig)   // 3373 bytes encoded
```

The combined signature is 3373 bytes (64-byte Ed25519 || 3309-byte ML-DSA-65)
encoded as base64url without padding. See
`01-cryptographic-primitives.md §6.1`.

**Verification procedure:**
1. Reconstruct canonical JSON from received manifest fields (minus `signature`).
2. base64url-decode the `signature` field; split into 64-byte Ed25519 part and
   3309-byte ML-DSA-65 part.
3. Verify Ed25519 signature over canonical JSON using `owner_public_key`.
4. Verify ML-DSA-65 signature over canonical JSON using the owner's ML-DSA-65
   key (obtained from the owner's prekey bundle via Identity family).
5. Reject if either signature fails.
6. Reject if `owner_peer_id` does not correspond to `owner_public_key` (the
   peer ID is derived from the Ed25519 public key per libp2p conventions).

### 4.5 Manifest hash

The manifest hash is computed as:

```
canonical_json = UTF-8(JSON(manifest_without_signature))
hash           = base64url_raw(SHA-256(canonical_json))[0:32]
```

This is a 32-character base64url-no-padding prefix of the SHA-256 hash of the
canonical manifest JSON. It is used in `Invite.manifest_hash` to bind an
invite to a specific manifest version. Implementations MUST recompute this hash
from the canonical JSON and compare it to the invite's `manifest_hash` during
invite verification.

Note: the Go implementation in `pkg/node/wire.go:Manifest.Hash()` uses
`base64.RawURLEncoding` applied to the canonical JSON bytes (not the SHA-256
hash). Implementations MUST match this behavior exactly for interoperability.

## 5. Security mode binding

The Manifest family transmits the manifest in the clear. The hybrid signature
provides integrity. The Noise XX transport layer (see `30-transport-and-noise.md
§1.2`) provides hop-to-hop confidentiality.

The `security_mode` and `offered_security_modes` fields in the manifest declare
the server's security policy:

- `security_mode` is the currently active mode. Clients MUST use this mode for
  all messages to this server.
- `offered_security_modes` is a preference-ordered list. During `chat.join`, the
  initiator MAY negotiate a different mode if the server supports multiple modes.
  Negotiation follows `03-protocol-registry-and-negotiation.md §3` using the
  `NegotiateConversationSecurityMode` algorithm.
- A manifest with `security_mode: "clear"` MUST be handled with the same UI
  label requirements as Clear mode messages (`04-security-modes.md`).

Manifest updates that downgrade `security_mode` from an E2EE mode to `"clear"`
MUST be explicitly confirmed by the server owner. Nodes that receive such a
downgrade SHOULD warn connected clients.

## 6. State persistence

| State bucket | Key type | Value type | Description |
|-------------|----------|-----------|-------------|
| `servers` | `server_id` (string) | `ServerRecord` (JSON) | Includes embedded `Manifest` field |

Go types from `pkg/node/types.go`:

- `ServerRecord`: `ID`, `Name`, `Description`, `OwnerPeerID`, `SecurityMode`,
  `OfferedSecurityModes`, `CreatedAt`, `UpdatedAt`, `Members []string`,
  `Channels map[string]ChannelRecord`, `Manifest Manifest`, `Invite string`

The `Manifest` struct (from `pkg/node/wire.go`) is stored inline in
`ServerRecord.Manifest`. The full signed manifest JSON (including `signature`)
MUST be persisted so that it can be served in response to `manifest.fetch`
requests without re-signing.

Manifest updates received via `manifest.publish` MUST replace the stored
manifest if they pass verification and are newer (`updated_at` comparison).
The previous manifest SHOULD be archived in a separate log for audit purposes
(implementation-defined; not part of the wire protocol).

## 7. Error codes

| Code | Trigger |
|------|---------|
| `MANIFEST_NOT_FOUND` | `manifest.fetch` for unknown `server_id` |
| `MANIFEST_EXPIRED` | `manifest.fetch` for a manifest past its `expires_at` |
| `MANIFEST_SIGNATURE_INVALID` | `manifest.publish` signature verification failed |
| `MANIFEST_NOT_NEWER` | `manifest.publish` received manifest is not newer than local copy |
| `MANIFEST_OWNER_MISMATCH` | `manifest.publish` initiator peer ID differs from manifest `owner_peer_id` |
| `MANIFEST_INCOMPLETE` | `manifest.publish` or `manifest.fetch` response missing required fields (`server_id`, `owner_peer_id`, `owner_public_key`) |

## 8. Conformance

Implementations claiming Manifest family conformance MUST pass the following
KATs:

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/manifest/manifest_fetch_kat.json` | `manifest.fetch` round-trip; valid response |
| `pkg/spectest/manifest/manifest_not_found_kat.json` | `manifest.fetch` for unknown server → `MANIFEST_NOT_FOUND` |
| `pkg/spectest/manifest/manifest_publish_kat.json` | `manifest.publish` with valid signature → accepted |
| `pkg/spectest/manifest/manifest_publish_invalid_sig_kat.json` | `manifest.publish` with tampered signature → `MANIFEST_SIGNATURE_INVALID` |
| `pkg/spectest/manifest/manifest_publish_stale_kat.json` | `manifest.publish` with older `updated_at` → `MANIFEST_NOT_NEWER` |
| `pkg/spectest/manifest/manifest_hash_kat.json` | Canonical form hash matches expected 32-char prefix |

All KAT vectors MUST include the full manifest JSON, the canonical form bytes
used for signing, and the expected signature verification outcome. Hash vectors
MUST include the input canonical JSON, the full SHA-256 output, and the 32-char
base64url prefix.
