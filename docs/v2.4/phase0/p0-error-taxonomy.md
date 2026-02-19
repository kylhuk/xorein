# P0 Local API Error & Refusal Taxonomy

## Deterministic refusal classes
- **Transport**
  - `T-001`: `VersionMismatch` (`EV-v24-G1-001`) – client requested version range outside the daemon’s supported window. Operator action: align `harmolyn` build with the negotiated range.
  - `T-002`: `EndpointForbidden` (`EV-v24-G1-002`) – handshake arrived on a forbidden interface or the transport owner did not match the expected UID/SID. Operator action: confirm socket/pipe path and ACL settings per ST1.
- **Authentication & session**
  - `A-001`: `ReplayDetected` (`EV-v24-G1-004`) – nonce or session token reuse detected. Operator action: rotate device secret and inspect token store integrity.
  - `A-002`: `DeviceProofFailed` (`EV-v24-G1-007`) – device secret proof did not verify. Operator action: trigger device rotation and confirm the proof challenge store.
- **Stream & ordering**
  - `S-001`: `CursorMismatch` (`EV-v24-G1-005`) – reconnect cursor disagrees with server state because ACK tracking was interrupted. Operator action: persist the last ACKed cursor in the UI state store.
  - `S-002`: `OversizeFrame` (`EV-v24-G1-006`) – frame exceeded the negotiated size bound. Operator action: inspect the source payload for unexpected data or jumbo packets.
- **Authorization & capability gating**
  - `Z-001`: `CapabilityPolicy` (`EV-v24-G1-003`) – client requested a capability class the server policy forbids. Operator action: review capability policy and restart `xorein` if policy changes were applied.
  - `Z-002`: `CapabilityClassDenied` (`EV-v24-G1-008`) – explicit capability class approval was missing for the requested class. Operator action: verify the class assignment, request operator approval, and observe audit trails.
  - `Z-003`: `CapabilityRefreshRevoked` (`EV-v24-G1-009`) – session refresh attempted after capability revocation. Operator action: reconcile revocation events and ensure revocation notice propagates to refresh callers.

## Operator diagnostics guidance
- Capture the refusal ID, capability bitmap, and associated gate name (`ST1`, `ST2`, or `ST3`) in audit logs; never log plaintext user content or session secrets to preserve the no keyword leakage invariant.
- Map refusal IDs to `EV-v24-G1-###` artifacts so reviewers can point to deterministic evidence files during gating.
- Use `xorein doctor` plus system journal entries to trace related process state; refusal IDs annotate the diagnostics knot for faster triage.

## Mapping to user-safe messages
- Transport denials translate to “Daemon version mismatch” or “Local interface not permitted” statements; they never surface user data nor history snippets.
- Authentication denials expose only the refusal code and a generic remediation (e.g., “Rotate device secret and retry”).
- Capability denials mention the class name (`basic:read`, `device:manage`, etc.) plus the refusal code to keep end-users aware of the policy without leaking sensitive keywords.

## Mandatory invariants reinforced
- No remote API exposure: all entries assume a local socket/pipe, and `T-002` attacks surface only when non-local endpoints disappear.
- Capability gating: every authorization refusal references the approved capability bitmap (`Z-001`, `Z-002`, `Z-003`).
- Deterministic refusals: `EV-v24-G1-001` through `EV-v24-G1-009` cover the handshake, proof, stream, and gating classes.
- Replay/session controls: `A-001` and `Z-003` demonstrate nonce, TTL, and revocation enforcement.
- Relay no-history and no keyword leakage: operator guidance and audit log rules avoid quoting history or keywords in refusal paths.
