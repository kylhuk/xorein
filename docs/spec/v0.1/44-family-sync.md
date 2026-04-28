# 44 — Family: Sync (`/aether/sync/0.1.0`)

This document specifies the Sync family, which enables history synchronization
between nodes across coverage gaps, reconnects, and archivist-to-client
transfers.

## 1. Overview

The Sync family provides three complementary operations: coverage advertisement
(allowing peers to discover what history a node holds), targeted message
retrieval, and archivist-push delivery for newly-joined members.

Unlike the Chat family, which delivers messages in real time, the Sync family is
oriented toward batch retrieval of already-committed ciphertext history. Sync
peers exchange coverage metadata so that gaps can be detected and filled without
full history scans.

**Roles that use this family:**
- **Client**: Uses `sync.coverage` to discover what history is available;
  uses `sync.fetch` to retrieve specific messages it missed. Receives
  `sync.push` from archivists on join.
- **Archivist**: Serves `sync.coverage` and `sync.fetch` with full-coverage
  response capability. Sends `sync.push` to newly-joined members. Archivists
  MUST advertise `cap.archivist` (in addition to `cap.sync`) to indicate
  full-coverage service guarantees.
- **Relay**: Does not participate in Sync family operations. A relay node MUST
  reject Sync family streams with `MISSING_REQUIRED_CAPABILITY`.
- **Bootstrap**: Does not participate in Sync family operations.

**Security modes:** The Sync family transmits ciphertext blobs. The syncing
node CANNOT decrypt messages it doesn't hold session keys for. Archivists MUST
NOT receive decryption keys — they store and serve ciphertext only. The relay
opacity invariant applies transitively: the archivist's store-and-forward role
is analogous to a relay but for history rather than real-time delivery.

**Protocol ID:** `/aether/sync/0.1.0`

**Required capability:** `cap.sync` — MUST be present in `advertised_caps` for
all Sync family streams.

## 2. Capability requirements

| Capability | Required for |
|------------|-------------|
| `cap.sync` | All Sync family operations |
| `cap.archivist` | Serving `sync.coverage` and `sync.fetch` with full-coverage guarantees; sending `sync.push` |

A node with `cap.sync` but without `cap.archivist` provides best-effort
coverage only (a local window of recent messages). A node with both `cap.sync`
and `cap.archivist` MUST serve all messages it has archived without gaps within
its declared coverage range.

## 3. Operations

| Operation | Required caps | Direction | Request payload type | Response payload type | Description |
|-----------|--------------|-----------|---------------------|----------------------|-------------|
| `sync.coverage` | `cap.sync` | initiator → responder | `SyncCoverageRequest` | `SyncCoverageResponse` | Exchange coverage metadata; discover what history a node holds |
| `sync.fetch` | `cap.sync` | initiator → responder | `SyncFetchRequest` | `SyncFetchResponse` | Pull specific messages from a node by message ID |
| `sync.push` | `cap.sync` | archivist → newly-joined client | `SyncPushRequest` | `SyncPushResponse` | Push history coverage to a newly-joined member |

The `operation` field in `PeerStreamRequest` MUST be one of the strings above.

## 4. Wire format details

All payloads are JSON-encoded and stored in `PeerStreamRequest.payload` /
`PeerStreamResponse.payload`.

### 4.1 `sync.coverage`

**Request:** `SyncCoverageRequest` — JSON-encoded:

```
{
  "server_id": string,   // server whose history coverage is being queried
  "from_seq":  int64,    // sequence number lower bound (inclusive); 0 means beginning
  "to_seq":    int64     // sequence number upper bound (inclusive); 0 means latest
}
```

**Response:** `SyncCoverageResponse` — JSON-encoded:

```
{
  "server_id":       string,    // mirrors request.server_id
  "available_from":  int64,     // lowest sequence number the responder has
  "available_to":    int64,     // highest sequence number the responder has
  "message_hashes":  [string],  // SHA-256 hashes (base64url) of messages in [available_from, available_to]
  "snapshot_root":   string,    // Merkle root of message_hashes; empty if < 2 messages
  "gap_ranges":      [          // gaps within the coverage range; absent if no gaps
    { "from": int64, "to": int64 }
  ]
}
```

Notes:

**Sequence numbers:** Sequence numbers are monotonically increasing 64-bit
integers assigned by the server owner to each message as it is committed. They
are not transmitted in the Delivery struct directly; archivists assign them
locally during ingestion. Sequence 0 indicates "before any message" (the empty
range).

**`message_hashes`:** Each entry is `base64url_no_padding(SHA-256(Delivery.ID ||
Delivery.created_at RFC3339Nano))`. This lets peers verify they hold the same
message for a given sequence number without exchanging full message content. The
list MUST be ordered by ascending sequence number.

**`snapshot_root`:** The Merkle root is `base64url_no_padding(SHA-256(
message_hashes[0] || message_hashes[1] || ... || message_hashes[n]))`. For a
single message, `snapshot_root` equals `message_hashes[0]`.

**Coverage model:**
- A node with `cap.archivist` MUST report the exact sequence range it has stored
  and MUST populate `gap_ranges` for any discontinuities within that range.
- A client with `HistoryCoverage = "local-window"` reports only its in-memory
  rolling window. It MUST set `available_from` and `available_to` to cover only
  the messages it currently holds, and MUST set `gap_ranges` to empty.
- A node that holds no history for `server_id` MUST respond with
  `available_from: 0, available_to: 0, message_hashes: [], snapshot_root: ""`
  rather than an error.

**Range limits:** Implementations SHOULD limit the number of hashes returned in
a single `sync.coverage` response to 10 000 entries. If the requested range
exceeds this limit, the responder MUST return `SYNC_RANGE_TOO_LARGE` and SHOULD
indicate the maximum supported range size in `error.message`.

### 4.2 `sync.fetch`

**Request:** `SyncFetchRequest` — JSON-encoded:

```
{
  "server_id":   string,    // server whose messages are being fetched
  "message_ids": [string]   // Delivery IDs to fetch; up to 200 per request
}
```

**Response:** `SyncFetchResponse` — JSON-encoded:

```
{
  "messages":       [ { /* Delivery JSON */ } ],  // found messages, in request order
  "not_found_ids":  [string],  // Delivery IDs the responder does not hold
  "error":          string
}
```

Notes:
- The responder MUST return all messages it holds whose `Delivery.ID` is in
  `message_ids`. Messages not found MUST be listed in `not_found_ids`.
- The responder MUST NOT return messages for a `server_id` that the initiator
  is not a member of. Membership is verified against the server's `members`
  list in the locally-held `ServerRecord`. Non-members MUST receive
  `OPERATION_FAILED` with message `"not a member"`.
- `message_ids` MUST contain at most 200 entries per request. Requests
  exceeding this limit MUST be rejected with `SYNC_FETCH_LIMIT_EXCEEDED`.
- The response `messages` list carries the full `Delivery` JSON including the
  `body` field. In E2EE modes, `body` is ciphertext. The initiator MUST decrypt
  locally using its held session key material.
- The archivist MUST NOT attempt to decrypt `body`. It stores and serves the
  ciphertext verbatim as it was received when the message was originally
  committed.

### 4.3 `sync.push`

**Request:** `SyncPushRequest` — JSON-encoded:

```
{
  "server_id": string,
  "messages":  [ { /* Delivery JSON */ } ]   // up to 200 messages per request
}
```

**Response:** `SyncPushResponse` — JSON-encoded:

```
{
  "accepted_count": int,   // number of messages accepted and stored
  "error":          string
}
```

Notes:
- `sync.push` is initiated by an archivist toward a newly-joined client to
  proactively deliver history. It is not a client-initiated pull.
- The receiver MUST verify each Delivery's signature before accepting it (per
  `02-canonical-envelope.md §3.4`).
- The receiver MUST apply the deduplication check (`deliveries` bucket) for
  each Delivery and silently skip already-known messages.
- The receiver SHOULD store accepted messages in the `messages` bucket.
- In E2EE modes, the receiver will accept ciphertext bodies it cannot yet
  decrypt (e.g., messages sent before it joined and received key material). These
  MUST be stored verbatim and decryption retried when key material becomes
  available.
- The archivist SHOULD pace pushes to avoid overwhelming the receiver. A push
  batch MUST NOT exceed 200 messages. The archivist SHOULD wait for a
  `SyncPushResponse` before sending the next batch.
- The archivist MUST limit the total history pushed to a newly-joined member to
  the server's `history_retention_messages` value. If 0 (unlimited), the
  archivist SHOULD push the full archived history up to a configurable cap
  (implementation-defined; SHOULD be ≥ 1 000 messages).

## 5. Security mode binding

The Sync family is designed to be mode-agnostic by carrying ciphertext
unchanged. The following mode-specific rules apply:

| Security mode | Archivist behavior | Client behavior |
|--------------|-------------------|----------------|
| `seal` | Stores ciphertext bodies; cannot decrypt | Decrypts with Seal session keys |
| `tree` | Stores MLS-encrypted ciphertext; cannot decrypt | Decrypts with MLS epoch key |
| `crowd` | Stores sender-key ciphertext; cannot decrypt | Decrypts with held sender key |
| `channel` | Stores broadcast-epoch ciphertext; cannot decrypt | Decrypts with epoch key |
| `clear` | Stores plaintext bodies | Reads bodies directly |

**Archivist key isolation invariant:** An archivist MUST NOT hold any session
decryption keys for E2EE scopes it archives. It MUST NOT request, accept, or
store decryption keys from server owners or members. This is a hard security
boundary — violation constitutes a protocol-level compromise of the E2EE
guarantees. Archivists that violate this invariant MUST be treated as
Byzantine nodes.

**Clear mode:** An archivist serving a Clear mode server stores and serves
plaintext bodies. This is intentional and declared via the server's manifest
`security_mode: "clear"`.

**Freshness:** Archived messages are served as immutable historical records.
The `Delivery.signature` field on each served message provides authenticity.
Receivers MUST verify message signatures after fetching and MUST discard
unsigned or unverifiable messages.

## 6. State persistence

| State bucket | Key type | Value type | Description |
|-------------|----------|-----------|-------------|
| `messages` | `message_id` (string) | `MessageRecord` (JSON) | Stored messages including ciphertext bodies |
| `servers` | `server_id` (string) | `ServerRecord` (JSON) | Coverage metadata in server record |
| `deliveries` | `delivery_id` (string) | `struct{}` | Deduplication set |

Go types from `pkg/node/types.go`:

- `MessageRecord`: `ID`, `ScopeType`, `ScopeID`, `ServerID`, `SenderPeerID`,
  `Body`, `CreatedAt`, `UpdatedAt`, `Deleted bool`
- `ServerRecord`: includes `Manifest.HistoryCoverage`, `Manifest.HistoryRetentionMessages`,
  `Manifest.HistoryDurability`

Coverage metadata (sequence numbers, snapshot root, gap ranges) is not
currently stored in a dedicated bucket. Implementations MUST compute coverage
information from the `messages` bucket by scanning messages for the requested
`server_id`, ordering by `created_at`, and assigning synthetic sequence numbers.

Future protocol versions MAY introduce a dedicated `coverage` bucket keyed by
`server_id`. When this extension is implemented, the bucket MUST use additive
JSON fields and MUST NOT break existing `sync.coverage` response semantics.

The `deliveries` bucket is shared with the Chat family and provides
cross-family deduplication.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `SYNC_RANGE_TOO_LARGE` | `sync.coverage` requested range exceeds 10 000 sequence numbers |
| `SYNC_FETCH_LIMIT_EXCEEDED` | `sync.fetch` `message_ids` list contains more than 200 entries |
| `SYNC_NOT_A_MEMBER` | `sync.fetch` initiator is not a member of the requested server |
| `SYNC_SERVER_NOT_FOUND` | `sync.coverage` or `sync.fetch` for unknown `server_id` |
| `SYNC_ARCHIVIST_REQUIRED` | `sync.fetch` or `sync.push` received by a non-archivist node; node lacks `cap.archivist` |
| `SYNC_SIGNATURE_INVALID` | `sync.push` message signature verification failed |

## 8. Conformance

Implementations claiming Sync family conformance MUST pass the following KATs:

| KAT file | Covers |
|----------|--------|
| `pkg/spectest/sync/coverage_empty_kat.json` | `sync.coverage` on a server with no history → empty response |
| `pkg/spectest/sync/coverage_full_kat.json` | `sync.coverage` for a server with 10 messages; hash list and snapshot root |
| `pkg/spectest/sync/coverage_gap_kat.json` | `sync.coverage` with gaps in coverage; `gap_ranges` populated |
| `pkg/spectest/sync/fetch_kat.json` | `sync.fetch` by message ID; mix of found and `not_found_ids` |
| `pkg/spectest/sync/fetch_limit_kat.json` | `sync.fetch` with 201 IDs → `SYNC_FETCH_LIMIT_EXCEEDED` |
| `pkg/spectest/sync/fetch_nonmember_kat.json` | `sync.fetch` from non-member → `SYNC_NOT_A_MEMBER` |
| `pkg/spectest/sync/push_kat.json` | `sync.push` from archivist; deduplication; `accepted_count` |
| `pkg/spectest/sync/push_invalid_sig_kat.json` | `sync.push` with tampered delivery signature → `SYNC_SIGNATURE_INVALID` |
| `pkg/spectest/sync/archivist_isolation_kat.json` | Archivist serves ciphertext body unchanged; no decryption key accepted |

All KAT files use the format defined in `90-conformance-harness.md`. Coverage
KAT vectors MUST include the message sequence, the per-message hash inputs
(`Delivery.ID || created_at`), and the expected `snapshot_root`. Archivist
isolation vectors MUST demonstrate that an archivist correctly refuses any
session key material offered over any operation.

An archivist-role conformance suite MUST additionally demonstrate:
- Full-coverage `sync.coverage` response for a server with 100 stored messages.
- Batch `sync.push` sequencing (multiple 200-message batches, each acknowledged
  before the next is sent).
- Correct `gap_ranges` reporting for a server with a deliberate sequence gap.
