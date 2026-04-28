# 91 — Test Vectors

Known-answer test (KAT) vectors for the Xorein v0.1 protocol.

## Vector format

All files use this JSON schema:

```json
{
  "description": "What is being tested",
  "source": "RFC XXXX §Y.Z | NIST SP XXX | Constructed",
  "inputs": { "<field>": "<lowercase hex>" },
  "expected_output": { "<field>": "<lowercase hex or null>" }
}
```

- All byte values are lowercase hex, no `0x` prefix.
- Empty byte strings are `""`.
- `null` expected outputs are **implementation-derived** — they must be computed
  from the verified reference Go implementation and replaced with concrete hex
  values before the corresponding wave gate is met (see `90-conformance-harness.md`).

## Pin file

`pin.sha256` is a SHA-256 manifest of all vector files:

```
<sha256hex>  <filename>
```

CI regenerates and verifies this file. **Do not hand-edit it.** Update it by
running `scripts/pin-vectors.sh` after adding or modifying a vector file.

## Files

### RFC-derived (exact values, immediately verifiable)

| File | Source | Wave |
|------|--------|------|
| `primitive_chacha20_poly1305.json` | RFC 8439 §2.8.2 | W0 |
| `primitive_aes_128_gcm.json` | NIST SP 800-38D B.2 | W0 |
| `primitive_hkdf_sha256.json` | RFC 5869 A.1 | W0 |
| `primitive_x25519.json` | RFC 7748 §6.1 | W0 |

### Implementation-derived (null outputs — must be computed and pinned)

| File | Description | Wave |
|------|-------------|------|
| `seal_kdf_labels.json` | Root-key, message-key, ratchet-step KDF | W1 |
| `seal_ratchet_basic.json` | Double Ratchet encrypt/decrypt round-trip | W1 |
| `crowd_sender_key.json` | Crowd sender key per-member HKDF | W3 |
| `crowd_epoch_chain.json` | Epoch chain derivation | W3 |
| `channel_kdf_label.json` | Channel sender key label | W3 |
| `mediashield_nonce.json` | Per-frame nonce derivation | W4 |
| `mediashield_sframe.json` | Full SFrame AES-128-GCM encrypt | W4 |
| `storage_kdf_sha256.json` | Storage key derivation (SHA-256 path) | W0 |
| `storage_key_check.json` | Storage key verification round-trip | W0 |

### Family KAT vectors (operation-level, implementation-derived)

| File | Family | Wave |
|------|--------|------|
| `governance_*.json` | Governance (RBAC) — 10 scenarios | W5 |
| `moderation_*.json` | Moderation — 8 scenarios | W5 |
| `sync_*.json` | Sync archivist — 8 scenarios | W5 |
| `voice_*.json` | Voice (MediaShield) — 13 scenarios | W5 |
| `chat_kat.json` | Chat — join/send/history | W5 |
| `dm_kat.json` | DM — send and rate-limit | W5 |
| `friends_kat.json` | Friends — request/accept/rate-limit | W5 |
| `groupdm_kat.json` | GroupDM — create/send/non-member | W5 |
| `identity_kat.json` | Identity — register/fetch | W5 |
| `manifest_kat.json` | Manifest — publish/fetch/not-found | W5 |
| `notify_kat.json` | Notify — push/drain/rate-limit | W5 |
| `presence_kat.json` | Presence — announce/query/stale-version | W5 |

## How to compute and pin implementation-derived vectors

```bash
# Run the reference implementation's vector generation command (Phase 2):
go run ./scripts/gen-vectors/main.go --output docs/spec/v0.1/91-test-vectors/

# Pin the results:
bash scripts/pin-vectors.sh docs/spec/v0.1/91-test-vectors/
```

The `gen-vectors` tool reads each JSON file, computes the expected output
using `pkg/crypto` primitives, and fills in the `null` fields. It must be
run on a clean build with `make compile` passing. The output is deterministic;
re-running it produces the same hex values.
