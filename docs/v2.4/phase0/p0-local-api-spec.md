# P0 Local API Specification

## Planning-honest scope and assumptions
- This doc covers the **local-only API** between `harmolyn` and `xorein`; expose no listeners beyond the per-user domain socket or named pipe. Any claim about future remote surfaces, plugin sandboxes, or out-of-band transports is explicitly out of scope.
- Assumes the relay invariant stays in place: the daemon never hosts long-term history, and keyword leakage defaults remain opt-out/no keyword enumeration.
- Capability gating is mandatory: every sensitive operation carries a named capability bit expressed in the handshake and enforced server-side.

## ST1: Local transport permutations
- **Permitted transports**: daemon binds per-user sockets in `/run/xorein/$USER/` (Unix domain sockets permission `0700` directory, socket `0600`) on Linux/macOS, and `\\.\\pipe\\xorein\\$USER\\session` with ACLs granting only the owning user and approved administrators on Windows. Listener names derive deterministically from the authenticated user id; no TCP, UDP, or network-facing listeners exist.
- **Handshake binding proof**: `ClientHello(versionRange, nonce, capabilityMask)` must arrive on the expected socket or pipe path; the daemon confirms the transport is owned by the same UID (or SID) and rejects siblings or remote attachment attempts (`EV-v24-G1-002`).
- **Version negotiation**: incompatible client ranges result in deterministic refusal `EV-v24-G1-001`; negotiation selects the highest supported version inside the overlap and records it in the session token material.
- **Permission hygiene**: newly created sockets reset ACLs/permissions before advertising readiness, cleanup removes stale pipes, and no fallback remote listener is opened unless a separate governance gate authorizes it.

## ST2: Device-secret proofing handshake
- **Device proof frame**: after version negotiation the client sends a `DeviceProof(deviceSecretId, challengeSignature, nonceSalt)` block. Daemon validates the signature against the stored device secret, updates the replay tracker, and seeds the session key material. Any mismatch or replay is refused with `EV-v24-G1-007` before capability gates run.
- **Scoped binding**: proof verification depends on the local transport check; once a proof is accepted the service issues a session token tied to that socket/pipe, preventing reuse from alternate transports or clones.
- **Rotation paths**: stale device secrets require manual rotation recorded in audit trails; repeated proof refusals highlight potential compromise without leaking keywords or history.

## ST3: Capability and session gating
- **Capability classes**: capabilities are grouped into `basic:read`, `device:manage`, `dangerous:export`, and `admin:configure`. Each class receives explicit server-side approval per user/device pair, and class requests are recorded in audit logs before a session begins.
- **Gating semantics**: capability mask bits split into `explicitApproval` and `delegatedSession` groups. Handshake refuses capability bits that violate the server policy with `EV-v24-G1-003`, while class-level approvals surface explicit denial `EV-v24-G1-008`; both refusals record the capability class and the policy gate.
- **Session enforcement**: session tokens embed the approved capability bitmap, TTL, and gating state. Any request referencing an unapproved capability—or refresh attempts after a revocation event—returns deterministic refusal `EV-v24-G1-009`, ensuring `harmolyn` cannot continue with stale privileges.
- **Audit visibility**: every capability denial emits the capability class, refusal code, and gate name to audit logs; logs omit keyword or history transcripts per invariant.

## Event stream contract and ordering
- **Stream structure**: single multiplexed, ordered protobuf stream carrying timeline events, connectivity updates, and capability change notices. Each frame includes type, cursor, and `capabilityBitmap` (if the event promotes or revokes rights).
- **Deterministic ordering**: timeline events maintain server-assigned cursors; UI must ACK progress to resume on reconnect. Reconnect resumes from last acknowledged cursor; missing ACK leads to deterministic refusal/resumption reason (`EV-v24-G1-005`).
- **Flow control**: daemon enforces bounded frames and rejects oversized commits (`EV-v24-G1-006`), keeping the stream backpressure-aware.

## Refusal taxonomy linkage
- Transport refusals: incompatible version (`EV-v24-G1-001`), forbidden interface (`EV-v24-G1-002`), capability mismatch (`EV-v24-G1-003`).
- Auth/session refusals: replay attempts (`EV-v24-G1-004`), cursor resume failures (`EV-v24-G1-005`), oversized frames (`EV-v24-G1-006`), device proof failures (`EV-v24-G1-007`), denied capability class (`EV-v24-G1-008`), revoked refresh gating (`EV-v24-G1-009`).
- Each refusal code maps to deterministic downstream handling in `p0-error-taxonomy.md` and evidence artifacts in the `EV-v24-G1` family for gate proof.
