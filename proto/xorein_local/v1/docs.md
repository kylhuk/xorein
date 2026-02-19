# Local-only transport contract

This directory lives behind a daemon that binds only to platform-local transports (Unix domain sockets on POSIX, named pipes on Windows). The handshake is deterministic:

1. The client sends `HandshakeRequest` containing `ClientHello` with a device identifier, version range, nonce, and capability request list.
2. The daemon compares the requested range with the supported `negotiated_version`; mismatches are refused and logged using `RefusalReason` in `HandshakeResponse`.
3. Once the version is agreed, the daemon issues a short-lived `SessionToken` (expires via `expires_at_unix`) and echoes a nonce.
4. Subsequent RPCs carry the session token in metadata; invalid or expired tokens produce `REFUSAL_REASON_INVALID_TOKEN`.

Attach and stream semantics:

- `Attach` validates the token, optionally scopes the session to a space/channel pair, and returns deterministic refusal reasons instead of plaintext errors.
- The event stream is a single server-stream RPC with `EventStreamRequest` and `EventStreamResponse`. Clients resume by replaying the last `StreamCursor` and supply the `resume_token` from the previous response.

Refusal reason guidance:

- `REFUSAL_REASON_NON_LOCAL_BIND` is triggered whenever the daemon is configured with non-local transports (e.g., `tcp`) and must abort to prevent remote exposure.
- `REFUSAL_REASON_INVALID_TOKEN` covers missing, expired, or non-matching tokens.
- `REFUSAL_REASON_UNAUTHORIZED_CAPABILITY` is used when the client asks for a capability (danger-zone actions) that the daemon deterministically denies.
- The zero value `REFUSAL_REASON_UNSPECIFIED` is reserved for protocol evolution and should not be emitted once all handlers assign specific reasons.
