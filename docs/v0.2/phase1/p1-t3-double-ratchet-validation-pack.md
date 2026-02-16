# v0.2 Phase 1 - P1-T3 Double Ratchet Model and Validation Pack

> Status: Execution artifact. Deterministic ratchet contracts live in `pkg/v02/dmratchet` and are exercised by `pkg/v02/dmratchet/contracts_test.go`.

## Purpose

Define deterministic Double Ratchet session behavior for normal, adversarial, and degraded message delivery conditions.

## Source Trace

- `TODO_v02.md:301`
- `TODO_v02.md:312`

## Ratchet State Persistence Rules (P1-T3-ST1)

Persisted state per DM session must include at minimum:

- Root key.
- Sending chain key + counter.
- Receiving chain key + counter.
- Previous chain length metadata.
- Skipped-message key cache metadata.
- Session binding identifiers and mode epoch.

Persistence requirements:

- State write is atomic at message commit boundaries.
- Crash recovery must not roll back to a state that accepts replayed ciphertext.
- Session state is scoped by peer identity + session binding ID.

## Out-of-Order, Replay, and Duplicate Handling

- Out-of-order messages are accepted only within bounded skipped-key windows.
- Duplicate ciphertext ID within replay window is rejected deterministically.
- Replay classification is reason-coded and surfaced to diagnostics.
- If out-of-order exceeds configured bounds, return deterministic resync-required outcome.

Planned bound defaults (subject to profiling validation):

- Max skipped keys per chain: 128.
- Max total cached skipped keys per session: 512.
- Replay tracking window: bounded by message counter and retention budget.

## Resynchronization Rules

Resync is triggered when any condition is true:

1. Counter gap exceeds skipped-key bounds.
2. Required chain state missing/corrupted.
3. Session binding mismatch after transport recovery.

Resync outcomes:

- Initiate deterministic ratchet reset procedure tied to authenticated session context.
- Preserve explicit failure reason class for user-visible and diagnostic mapping.
- Do not silently downgrade encryption posture.

## Cryptographic Validation Pack (P1-T3-ST2)

### Deterministic vectors

- Key derivation determinism vectors.
- Encrypt/decrypt roundtrip vectors for sequential and out-of-order delivery.
- Replay/duplicate rejection vectors.
- Corrupted ciphertext/auth tag rejection vectors.

### Adversarial and negative-path requirements

- Tampered header fields.
- Counter rollback attempts.
- Stale prekey bootstrap handoff into ratchet.
- Corrupted persisted state recovery behavior.

### Property and fuzz requirements

- Property: valid ciphertext decrypts to original plaintext for all generated message sequences.
- Property: duplicate ciphertext never increments session counters.
- Fuzz: malformed envelopes cannot crash parser or bypass authenticity checks.
- Fuzz: random counter/order sequences preserve deterministic accept/reject outcomes.

## Evidence Targets

| Validation area | Evidence anchor |
|---|---|
| Persistence and crash recovery behavior | `pkg/v02/dmratchet/contracts.go` (`ValidatePersistedState`) |
| Out-of-order and replay handling | `pkg/v02/dmratchet/contracts.go` (`EvaluateIncomingMessage`) |
| Deterministic vector corpus | `pkg/v02/dmratchet/contracts_test.go` |
| Adversarial/fuzz requirement fulfillment | `pkg/v02/dmratchet/contracts_test.go` |
