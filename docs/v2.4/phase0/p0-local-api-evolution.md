# v2.4 Phase 0 Local API Evolution Policy

This planning document records the additive compatibility and versioning policy for the local API between `harmolyn` and `xorein`. It preserves the planning-honest tone required of Phase 0 while pointing to the deterministic refusal/error taxonomy that will govern runtime behavior.

## Versioning rules
- **Major increments** (`v1`, `v2`, etc.) may introduce breaking changes, but each major version must negotiate via the handshake before the client is allowed to proceed. Breaking changes must be preannounced in the spec, and the daemon must refuse the connection with a deterministic reason code if the counterpart does not support the negotiated major version. Evidence placeholder: `EV-v24-G1-001`.
- **Minor increments** within the same major are strictly additive: new RPCs, message fields, or event stream topics may be introduced only if the daemon can ignore unknown fields and the client gracefully handles the absence of new server-side behavior. Backward compat requests (older client to newer daemon) must succeed even if the daemon has additive capabilities. Evidence placeholder: `EV-v24-G1-002`.
- **Patch increments** cover bug fixes that do not touch the protobuf schema (frames, gRPC options, message ordering) and are considered compatible if recorded in the `p0-local-api-spec.md` patch log.

## Deprecation handling
- RPCs/fields slated for removal must stay annotated with the `deprecated` reason, remain supported for at least one major version boundary, and emit a client-visible warning event when invoked (with a defined “next action”).
- The daemon logs each deprecated surface invocation with a deterministic code so automation can track usage; evidence of the log schema will be captured under `EV-v24-G1-003` once implemented.

## Refusal & error taxonomy
- Connections that fail version negotiation, authentication, or authorization must return a deterministic refusal code (`VERSION_MISMATCH`, `AUTH_FAILURE`, `UNAUTHORIZED_CAPABILITY`, etc.) so that callers can map to UI actions without relying on free-text messages.
- Error responses must include structured fields: `reason_code`, `service_area`, `next_action`. This taxonomy must align with `p0-error-taxonomy.md` once that document is created.
- All refusal responses are additive and do not trigger schema changes; new codes must be registered before release and documented in the traceability matrix.

## Compatibility guardrails
- The daemon must never accept more permissive API versions than explicitly negotiated; if negotiation fails, it sends a deterministic error and closes the stream (`EV-v24-G1-004` placeholder).
- API clients must be tolerant of new enumeration values and unknown fields by following protobuf best practices (unknown fields are ignored on receipt, repeated fields default to empty lists).

## Future work references
- `p0-local-api-spec.md`, `p0-local-api-threat-model.md`, and `p0-error-taxonomy.md` will elaborate on these policies once the Phase executes.
