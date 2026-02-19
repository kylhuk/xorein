# P0 Local API Threat Model

## Threat scope and adversary personas
- **Scope**: the only exposed surface is the per-user domain socket or named pipe; the daemon never opens a TCP port and no remote API exposure is permitted. Attack surfaces are restricted to local privilege escalation or compromised clients, preserving the no keyword leakage defaults and no history hosting invariant.
- **Personas**:
  1. **Malicious local user** (same OS user tries to bypass capability checks).
  2. **Privileged process on the same host** (tries to inject requests or replay handshakes).
  3. **Stale or hijacked session** (attacker reuses old tokens via file system copies).
 4. **Insider developer** (misconfigures CLI to bypass gating).

## Attack surfaces
- **Transport tamper**: the adversary intercepts or modifies socket frames while running as the same user; mitigated by Unix domain socket permissions, Windows ACLs, and deterministic capability verification (controls encoded in `p0-local-api-spec`).
- **Replay**: reusing prior handshake or session tokens; guard with nonce tracking, TTL, and deterministic refusal codes (`EV-v24-G1-004`).
- **Capability escalation**: injecting fake capability bits; blocked by capability gating and audit logging of dangerous-bit requests (maps to deterministic refusal reasons `EV-v24-G1-003`).
- **Local privilege misuse**: a GUI process shares the socket with untrusted helper; mitigated by binding sockets to per-user directories and auditing attach semantics.

## Mitigations tied to invariants
- **Local-only transport invariant**: ST1 enforces Per-UID sockets/pipes, strict ACLs, and forbids TCP listeners (`EV-v24-G1-002`), keeping the attack surface narrow.
- **No durable history hosting**: keeping `xorein` as relay-only ensures even compromised transports cannot pull historical transcripts; audit logs avoid storing history, aligning with the planning-honest assertion.
- **No keyword leakage default**: all refusals (transport, capability, stream) log codes instead of user content, satisfying the invariant discussed in the planning scope.
- **Capability gating invariant**: ST3 forces explicit capability class approval and deterministic refusal reasons (`EV-v24-G1-003`, `EV-v24-G1-008`, `EV-v24-G1-009`), preventing stealthy elevation.
- **Device proof invariant**: ST2 requires device-secret proofing before session tokens, bounding token issuance to legitimate devices (`EV-v24-G1-007`).

## Residual risk table
| Risk ID | Description | Mitigation | Owner role | Gate/evidence pointer |
| --- | --- | --- | --- | --- |
| R1 | Local transport permissions accidentally widened (misconfigured ACLs) | Per-user directory sockets, bind hygiene, gate lock-down, and ACL drift monitoring. | Platform security team | `docs/v2.4/phase0/p0-local-api-spec.md#st1-local-transport-permutations` |
| R2 | Device secret storage compromised, enabling replayed proofs | Deterministic refusal `EV-v24-G1-007`, audit-recorded rotation paths, triggered device rotation playbook. | Device security guard | `docs/v2.4/phase0/p0-local-api-spec.md#st2-device-secret-proofing-handshake` |
| R3 | Capability policy misapplied, yielding latent privileged session | Deterministic refusals `EV-v24-G1-008`/`EV-v24-G1-009`, audit trails for each capability class, operator ACK gate. | Capability governance owner | `docs/v2.4/phase0/p0-local-api-spec.md#st3-capability-and-session-gating` |
