# P3-T4 — Signed Envelope Specification v0.1

This document covers the canonical serialization boundaries, signing inputs, and
verification failure taxonomy for `SignedEnvelope` as defined in
[`proto/aether.proto`](proto/aether.proto:1).

## Canonical serialization

1. **Payload preparation**
   - Resolve `payload_type` to the concrete message.
   - Serialize the message using deterministic protobuf encoding (`--deterministic`).
   - Store the raw bytes in `SignedEnvelope.payload`.

2. **Canonical payload**
   - Construct `canonical_payload` by concatenating:
     1. `payload_type` varint tag/value from the envelope.
     2. The deterministic `payload` bytes.
     3. `signed_at` encoded as an unsigned 64-bit BE integer.
   - Persist the result verbatim inside `SignedEnvelope.canonical_payload`.

3. **Signature input**
   - Compute `signature = Sign(canonical_payload)` using the private key associated
     with `signer.signing_public_key`.
   - Record the algorithm used in `signature_algorithm`.

4. **Envelope bytes**
   - The final envelope is serialized with deterministic protobuf encoding so that
     downstream verifiers can rehydrate the same `canonical_payload` without needing
     to re-run the canonicalization routine.

## Verification process

1. **Basic structure**: parse the envelope deterministically; reject if any
   required fields are missing.
2. **Canonical replay**: recompute the canonical payload from the parsed fields and
   compare to `canonical_payload`. Mismatch ⇒ `CANONICALIZATION_MISMATCH`.
3. **Algorithm dispatch**: ensure `signature_algorithm` is recognized for the
   given signer key. Otherwise ⇒ `UNSUPPORTED_PAYLOAD_TYPE` or future codes.
4. **Signature check**: verify the signature bytes over the canonical payload.
   Failure ⇒ `SIGNATURE_MISMATCH`.
5. **Freshness**: validate `signed_at` against local policy (e.g., maximum skew).
   Failure ⇒ `EXPIRED_SIGNATURE`.
6. **Authorization**: ensure the signer exists in a trust store or manifest.
   Failure ⇒ `UNTRUSTED_SIGNER`.
7. **Payload gating**: confirm `payload_type` is supported by the recipient.
   Failure ⇒ `UNSUPPORTED_PAYLOAD_TYPE`.

## Error taxonomy mapping

| Code | Trigger | Notes |
| --- | --- | --- |
| `SIGNATURE_MISMATCH` | Signature verification failed | canonical payload bytes included in response for diagnostics |
| `UNSIGNED_PAYLOAD` | Missing signature fields | typically indicates tampering |
| `UNSUPPORTED_PAYLOAD_TYPE` | Unknown `payload_type` | clients may choose to queue for upgrade |
| `CANONICALIZATION_MISMATCH` | recomputed canonical payload differs | treat as integrity violation |
| `EXPIRED_SIGNATURE` | `signed_at` outside acceptance window | policy window defined by consumers |
| `UNTRUSTED_SIGNER` | signer key not trusted for the manifest | ensures membership enforcement |

`EnvelopeVerification` responses must include:

- `status`: `ACCEPTED` when no errors are present and signature checks succeed,
  otherwise `REJECTED`.
- `errors`: ordered list of `VerificationError` entries.
- `canonical_bytes`: the recomputed canonical payload so that upstream systems can
  archive the exact bytes that were validated or rejected.

With this specification plus the schema inventory
[`docs/v0.1/phase3/p3-t2-schema-inventory.md`](docs/v0.1/phase3/p3-t2-schema-inventory.md:1),
P3-T2 and P3-T4 prerequisites are satisfied for v0.1.

## Group-key profile decision (P3-T6)

### Decision summary

- **MLS target profile:** MLS-style group epoch model with deterministic rotation epochs
  and membership-change rekeys, represented in implementation scaffolding via
  [`KeyState`](../../pkg/phase7/bootstrap.go:12) and
  [`Bootstrapper.Rotate()`](../../pkg/phase7/bootstrap.go:46).
- **Sender Keys compatibility bridge:** accept prior sender keys for bounded
  compatibility during migration/rotation windows, represented via
  [`KeyState.LegacySender`](../../pkg/phase7/bootstrap.go:18) and enforced by
  [`Bootstrapper.SenderCompatible()`](../../pkg/phase7/bootstrap.go:73).
- **Mismatch recovery policy:** receiver-side mismatch triggers immediate rekey path
  (`Rotate`) rather than silent accept, represented in
  [`Bootstrapper.RekeyOnMismatch()`](../../pkg/phase7/bootstrap.go:91).

### Bootstrap, rotation, and member-change handling strategy

1. **Bootstrap**
   - Initialize per-participant MLS secret + sender key + signer in
     [`Bootstrapper.Bootstrap()`](../../pkg/phase7/bootstrap.go:32).
   - Initial epoch is `rotation=1` to avoid zero-value ambiguity.
2. **Rotation**
   - Create fresh MLS secret and sender key on each
     [`Bootstrapper.Rotate()`](../../pkg/phase7/bootstrap.go:46) invocation.
   - Preserve only bounded compatibility history (`<=2` legacy sender keys) to limit
     replay and stale-key acceptance windows.
3. **Member-change / mismatch rekey**
   - Any sender-key mismatch or empty candidate in
     [`Bootstrapper.RekeyOnMismatch()`](../../pkg/phase7/bootstrap.go:91) forces
     rotation and epoch advancement.
   - Message-layer signature mismatch remains hard-fail via
     [`ErrInvalidSignature`](../../pkg/phase7/pipeline.go:15), while duplicate replay
     is rejected via [`ErrDuplicateMessage`](../../pkg/phase7/pipeline.go:14).

### Candidate profile/library comparison snapshot

For v0.1, selection is constrained to repository-local validated behavior rather than
introducing external runtime dependencies mid-phase:

- **Selected now (v0.1 execution baseline):** in-repo deterministic group-key scaffold
  in [`pkg/phase7`](../../pkg/phase7/bootstrap.go) verified by
  [`TestBootstrapperRotationAndCompatibility()`](../../pkg/phase7/bootstrap_test.go:5)
  and [`TestBootstrapperRekeyOnMismatch()`](../../pkg/phase7/bootstrap_test.go:42).
- **Deferred option A (post-v0.1):** full standards-complete MLS stack integration once
  Phase-7 topic binding and network pubsub layers are fully closed.
- **Deferred option B (rejected for v0.1):** Sender Keys–only long-term profile without
  MLS-target epoch semantics, rejected because it weakens forward-compatible path to
  deterministic group-member rekey discipline.

### Threat-model notes for v0.1 text channels

- **Replay risk:** mitigated by per-sender sequence replay cache in
  [`Pipeline.Receive()`](../../pkg/phase7/pipeline.go:59) with duplicate rejection.
- **Signature forgery/tampering risk:** mitigated by mandatory Ed25519 verification in
  [`Pipeline.Receive()`](../../pkg/phase7/pipeline.go:59).
- **Stale key acceptance risk during migration:** reduced by bounded legacy key window
  in [`Bootstrapper.Rotate()`](../../pkg/phase7/bootstrap.go:66).
- **Out-of-window freshness drift:** envelope freshness enforcement remains defined at
  spec layer (`EXPIRED_SIGNATURE`) and must be enforced by downstream policy checks in
  higher-layer integration validation.
- **Known v0.1 limitations:** this phase validates deterministic unit behavior and
  dependency-safe key lifecycle scaffolding; end-to-end pubsub/topic binding behavior is
  still gated by unresolved dependency **P4-T3**.
