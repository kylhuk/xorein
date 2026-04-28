# 90 — Conformance Harness

This document defines the conformance levels (W0–W6) for Xorein v0.1, the
pass/fail criteria for each level, the known-answer test (KAT) vector format,
the required integration scenarios, and the release gate.

A Xorein implementation is considered conformant at wave **Wn** if it passes
all tests for W0 through Wn inclusive and has no regressions at any lower level.

## 1. Conformance levels

| Level | Name | Must pass for |
|-------|------|---------------|
| **W0** | Crypto bedrock | Any Xorein operation at all |
| **W1** | Seal mode | 1:1 E2EE DMs |
| **W2** | Tree mode | Interactive group E2EE (≤50 members) |
| **W3** | Crowd/Channel mode | Large-scale E2EE (>50 members) |
| **W4** | MediaShield | E2EE voice/video frames |
| **W5** | Discovery & connectivity | Real P2P without manual peer injection |
| **W6** | Full Discord-parity | All 13 protocol families, interop gate |

### W0 — Crypto bedrock

**Scope**: Cryptographic primitives that underpin every higher-level operation.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W0-CHACHA | ChaCha20-Poly1305 encrypt+decrypt | RFC 8439 §2.8.2 |
| W0-AES-GCM | AES-128-GCM encrypt+decrypt | NIST SP 800-38D B.2 |
| W0-HKDF | HKDF-SHA-256 extract+expand | RFC 5869 A.1 |
| W0-X25519 | X25519 scalar multiplication (2 vectors) | RFC 7748 §6.1 |
| W0-AEAD-FAIL | Tampered ciphertext → `ErrDecryptFailed` | Constructed |
| W0-PQ-KEM | ML-KEM-768 encap/decap round-trip | NIST FIPS 203 KAT |
| W0-PQ-SIG | ML-DSA-65 sign/verify round-trip | NIST FIPS 204 KAT |
| W0-HYBRID-SIG | Hybrid Ed25519+ML-DSA-65 sign/verify | Constructed |
| W0-HYBRID-KEM | X25519+ML-KEM-768 combined shared secret | Constructed |

**Pass criterion**: All assertions in `pkg/crypto/vectors_test.go` and
`docs/spec/v0.1/91-test-vectors/primitive_*.json` pass.

**Fail criterion**: Any assertion fails or panics; any AEAD decryption of
valid ciphertext fails; ErrDecryptFailed not returned on tampered ciphertext.

---

### W1 — Seal mode (1:1 E2EE)

**Scope**: Hybrid X3DH session initiation, Double Ratchet messaging, and the
DM family (`/aether/dm/0.2.0`). Depends on W0.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W1-X3DH-CLASSICAL | Classical X3DH with 4 DH operations | Constructed |
| W1-X3DH-HYBRID | Full hybrid X3DH (X25519+ML-KEM-768) | Constructed |
| W1-RATCHET-BASIC | Double Ratchet encrypt/decrypt round-trip | Constructed |
| W1-RATCHET-OOP | Out-of-order messages (skip list ≤1000) | Constructed |
| W1-RATCHET-REPLAY | Replayed delivery ID rejected | Constructed |
| W1-RELAY-OPACITY | Relay queue contains no plaintext body | Integration |
| W1-PREKEY-EXPIRE | Expired prekey → error, not silent failure | Integration |
| W1-KDF-ROOT | KDF labels: root-key, chain-key derivation | Constructed |
| W1-KDF-MSG | KDF labels: message-key derivation | Constructed |
| W1-DOWNGRADE | Downgrade attempt → structured NegotiationError | Integration |

**Pass criterion**: All `pkg/spectest/seal/` tests pass. Integration test
`TestRelayOpacitySeal` confirms relay queue payload does not contain plaintext.

**Fail criterion**: DH output diverges between initiator and responder; skip
list exceeds 1000; replayed message accepted; relay queue contains plaintext.

---

### W2 — Tree mode (MLS hybrid group E2EE)

**Scope**: Hybrid MLS ciphersuite 0xFF01, TreeKEM, key schedule,
Welcome/Commit/Proposal, and the GroupDM family (`/aether/groupdm/0.2.0`).
Depends on W1.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W2-TREEKEM | TreeKEM encap/decap (adapted for 0xFF01) | RFC 9420 vectors |
| W2-KEYSCHED | MLS key schedule | RFC 9420 vectors |
| W2-WELCOME | Welcome message parse/verify | RFC 9420 vectors |
| W2-COMMIT | Commit message parse/verify | RFC 9420 vectors |
| W2-HYBRID-KEM-COMBINE | Hybrid KEM combiner output | Constructed |
| W2-HYBRID-SIG-MLS | Hybrid signature over MLS message | Constructed |
| W2-EPOCH-ROTATE-MSG | Epoch rotates after 1000 messages | Integration |
| W2-EPOCH-ROTATE-ADD | Epoch rotates on member add | Integration |
| W2-EPOCH-ROTATE-REMOVE | Epoch rotates on member remove | Integration |
| W2-MEDIASHIELD-EXPORT | MLS-Exporter → MediaShield key | Constructed |
| W2-MAX-MEMBERS | 51st member add rejected | Integration |

**Pass criterion**: RFC 9420 vector suite passes with ciphersuite 0xFF01
adaptations. 4-node integration test: create → add → remove → check group
key transitions.

**Fail criterion**: TreeKEM output diverges; epoch not rotated on membership
change; removed member can decrypt post-remove messages.

---

### W3 — Crowd/Channel mode (epoch sender keys)

**Scope**: Epoch root key, sender key HKDF derivation, epoch chain, rotation
triggers, legacy window enforcement, and the Chat family (`/aether/chat/0.1.0`)
for server-scope messages. Depends on W1.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W3-SENDER-KEY | `xorein/crowd/v1/sender-key` HKDF derivation | Constructed |
| W3-EPOCH-CHAIN | Non-membership epoch chain derivation | Constructed |
| W3-EPOCH-ROTATE-MEMBER | Fresh random epoch on member removal | Integration |
| W3-EPOCH-LEGACY | Epoch ≤2 generations old accepted | Integration |
| W3-EPOCH-EXPIRED | Epoch >2 generations old rejected with `EPOCH_EXPIRED` | Integration |
| W3-RELAY-OPACITY | Chat message relay queue: ciphertext only | Integration |
| W3-CHANNEL-WRITER | Non-writer `channel_message` rejected | Integration |
| W3-CHANNEL-SNAPSHOT | History snapshot signed and verified | Integration |
| W3-CHANNEL-KDF | `xorein/channel/v1/sender-key` label | Constructed |

**Pass criterion**: 5-node integration test: large group with epoch rollover
at 1000 messages. Legacy window enforced exactly at 2; epoch 3 generations
old rejected.

**Fail criterion**: Epoch chain derivation diverges; removed member's sender
key still decrypts post-rotation messages; legacy window > 2 accepted.

---

### W4 — MediaShield (SFrame voice/video E2EE)

**Scope**: RFC 9605 SFrame, MediaShield key derivation from each parent scope,
per-frame nonce derivation, frame counter replay protection, and the Voice
family (`/aether/voice/0.1.0`). Depends on W2 (MLS exporter) and W3.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W4-SFRAME-ENCRYPT | SFrame AES-128-GCM encrypt | RFC 9605 §4 |
| W4-SFRAME-DECRYPT | SFrame AES-128-GCM decrypt | RFC 9605 §4 |
| W4-KEY-DERIVE-TREE | MLS-Exporter → MediaShield key | Constructed |
| W4-KEY-DERIVE-CROWD | Crowd sender key → MediaShield key HKDF | Constructed |
| W4-KEY-DERIVE-SEAL | DR message key → MediaShield key HKDF (Seal DM voice) | Constructed |
| W4-NONCE-DERIVE | Per-frame nonce from MediaShield key + counter | Constructed |
| W4-COUNTER-ROLLOVER | Rotation triggered at counter = 2^48 | Constructed |
| W4-REPLAY-REJECT | Frame with counter ≤ previous → rejected | Constructed |
| W4-SFU-OPACITY | SFU relay sees only SFrame ciphertext | Integration |
| W4-EPOCH-ROTATE | MediaShield key rotated on parent epoch change | Integration |

**Pass criterion**: 2-participant voice call integration test. SFU relay node
cannot read audio frames; frame counter monotonicity enforced; key rotation
completes within 2 seconds of parent epoch change.

**Fail criterion**: SFU relay can decrypt frame content; replayed frame
accepted; nonce reuse across frames.

---

### W5 — Discovery and connectivity

**Scope**: mDNS, Kademlia DHT, PEX, bootstrap, Circuit Relay v2, and DCUtR.
Depends on W0 (Noise identity). Peer-to-peer connectivity without manual peer
injection.

**Must pass:**

| Test ID | Description | Source |
|---------|-------------|--------|
| W5-MDNS-ANNOUNCE | Node announces on LAN via mDNS | Integration |
| W5-MDNS-DISCOVER | Node discovers LAN peer via mDNS | Integration |
| W5-DHT-PROVIDE | Node publishes provider record to DHT | Integration |
| W5-DHT-FIND | Node finds peer via DHT lookup | Integration |
| W5-PEX | Peer exchange delivers peer records | Integration |
| W5-RELAY-RESERVE | Circuit Relay v2 reservation granted | Integration |
| W5-DCUTR-UPGRADE | DCUtR upgrades relayed connection to direct | Integration |
| W5-FALLBACK | No direct path → relay fallback within 10s | Integration |
| W5-CLEAR-MODE | Clear-mode message delivered; relay stores readable body | Integration |
| W5-DOWNGRADE-PREVENT | Clear mode not silently downgraded from Seal | Integration |

**Pass criterion**: 3-node integration test: two nodes behind simulated NAT
discover each other via DHT, establish Circuit Relay v2, then DCUtR-upgrade to
direct. Seal-mode DM completes end-to-end.

**Fail criterion**: Peers cannot connect without manual address injection;
DCUtR attempt not made before relay fallback; Clear mode accepted in Seal scope.

---

### W6 — Full Discord-parity (release gate)

**Scope**: All 13 protocol families fully operational across all applicable
security modes. Multi-client interop test with two independent implementations.
Depends on W0–W5.

**Must pass** (all families exercised):

| Family | Required scenario |
|--------|-----------------|
| `/aether/peer/0.1.0` | PEX + relay store/drain round-trip |
| `/aether/chat/0.1.0` | Server channel message send/receive (Crowd mode) |
| `/aether/voice/0.1.0` | 3-participant SFU voice call (MediaShield) |
| `/aether/manifest/0.1.0` | Signed invite creation, distribution, join |
| `/aether/identity/0.1.0` | Prekey publish, OPK replenishment cycle |
| `/aether/sync/0.1.0` | Own-device sync round-trip (Seal) |
| `/aether/dm/0.2.0` | Seal DM round-trip across two Noise sessions |
| `/aether/groupdm/0.2.0` | Tree-mode group DM: create, add, send, leave |
| `/aether/friends/0.2.0` | Full request → accept → remove lifecycle |
| `/aether/presence/0.2.0` | Status update disseminated within 90s |
| `/aether/notify/0.2.0` | @mention notification delivered and acked |
| `/aether/moderation/0.2.0` | Kick event propagated; kicked peer loses access |
| `/aether/governance/0.2.0` | Role assignment; non-writer rejects in Channel mode |

**Additional W6 requirements:**

- Two independent implementations (e.g., Go reference + a third-party client)
  complete a Seal DM round-trip using only the spec (no shared code).
- 24-hour fuzz pass on all fuzz targets in §6 with no crashes.
- `make pipeline` is clean (generate + compile + lint + test + race + scan + build).
- `scripts/spec-lint.sh` reports zero errors.
- All KAT vector files in `91-test-vectors/` are pinned in
  `91-test-vectors/pin.sha256` and CI verifies the pin.

**Pass criterion**: All family scenarios succeed. Interop between two
independent implementations confirmed.

**Fail criterion**: Any family scenario fails; interop fails; fuzz crash found;
pipeline not clean.

---

## 2. KAT vector format

All vector files in `docs/spec/v0.1/91-test-vectors/` use this JSON schema:

```json
{
  "description": "Human-readable description of what is being tested",
  "source": "RFC XXXX §Y.Z | NIST SP XXX | Constructed",
  "inputs": {
    "<field>": "<hex-encoded bytes or scalar value>"
  },
  "expected_output": {
    "<field>": "<hex-encoded bytes or scalar value>"
  }
}
```

All byte fields are lowercase hex-encoded without `0x` prefix.

**Constructed vectors** are computed from the reference Go implementation
(`pkg/crypto`, `pkg/spectest`) and pinned. Their `"source"` field reads
`"Constructed — computed from reference implementation"`. They are not less
normative than RFC-derived vectors; they serve as regression anchors for
independent implementations.

**Pin file** (`91-test-vectors/pin.sha256`): one line per vector file, format:
```
<sha256hex>  <filename>
```

CI MUST verify every file's SHA-256 against the pin before running vectors.
The pin file is updated only when adding new vectors or correcting an error in
an existing one (with a commit message explaining the change).

---

## 3. Running the conformance suite

```bash
# Full pipeline (compile + lint + all tests + race + scan + build):
make pipeline

# KAT vectors only (W0 primitives):
go test ./pkg/crypto/... -run TestVectors

# Mode-specific KATs (once implemented per wave):
go test ./pkg/spectest/seal/...         # W1
go test ./pkg/spectest/tree/...         # W2
go test ./pkg/spectest/crowd/...        # W3
go test ./pkg/spectest/channel/...      # W3
go test ./pkg/spectest/mediashield/...  # W4

# Integration tests (requires two running node processes):
go test ./pkg/node/ -run TestIntegration -v -timeout 120s

# Race condition check:
make race

# Security scan:
make scan

# Spec self-consistency:
bash scripts/spec-lint.sh
```

---

## 4. Multi-node integration tests

Integration tests use real libp2p hosts with a 40 ms discovery interval.
They live in `pkg/node/service_test.go` and follow the naming pattern
`TestIntegration<Scenario>`.

Required integration scenarios per wave (minimum):

| Wave | Scenario | Nodes | Key assertion |
|------|---------|-------|---------------|
| W1 | Seal DM round-trip | 2 | Relay queue row contains no plaintext |
| W2 | Tree group: create → add → remove | 4 | Group key transitions correct; removed member locked out |
| W3 | Crowd large group, epoch rollover | 5 | Epoch counter increments; legacy bound ≤2 |
| W4 | MediaShield voice call via relay | 2 | Relay sees only SFrame ciphertext |
| W5 | NAT traversal: DCUtR upgrade | 3 | Direct connection established after relay bootstrap |
| W6 | Full Discord-parity: all families | 4 | All operations succeed end-to-end |

Relay queue introspection: the test reads relay state directly and asserts that
`payload` bytes do not contain the plaintext message body as a substring.

---

## 5. Negative tests

Each mode MUST have the following negative tests:

- Tampered ciphertext → AEAD authentication failure.
- Replayed `Delivery.id` → duplicate suppression.
- Expired prekey (Seal) → `identity.fetch` required.
- Downgrade attempt → `NegotiationError` with code `mode-incompatible`.
- Wrong sender key epoch (Crowd/Channel) → `EPOCH_EXPIRED`.
- SFrame frame counter regression (MediaShield) → replay rejection.
- Non-writer `channel_message` in Channel mode → rejected.
- MLS member remove: ex-member cannot decrypt next `ApplicationMessage`.

---

## 6. Fuzz targets

| Wave | Target | File |
|------|--------|------|
| W0 | Canonical envelope | `pkg/node/wire_fuzz_test.go` |
| W1 | Ratchet message header | `pkg/spectest/seal/ratchet_fuzz_test.go` |
| W2 | MLS message | `pkg/spectest/tree/mls_fuzz_test.go` |
| W4 | SFrame frame | `pkg/spectest/mediashield/sframe_fuzz_test.go` |

A 24-hour fuzz pass on all four targets is required at the W6 release gate.

---

## 7. Release gate (v0.1.0)

The v0.1.0 tag MUST NOT be created until all of the following are true:

1. `make pipeline` exits 0.
2. All W0–W6 KATs pass.
3. All W0–W6 integration scenarios pass.
4. All negative tests pass.
5. 24-hour fuzz pass shows no crashes on all four targets.
6. `91-test-vectors/pin.sha256` is committed and CI-verified.
7. Two independent implementations complete a Seal DM round-trip.
8. The spec commit SHA is referenced in the v0.1.0 release notes.
9. `scripts/spec-lint.sh` reports zero errors.
10. `buf lint` and `buf breaking` pass on `proto/aether.proto`.
