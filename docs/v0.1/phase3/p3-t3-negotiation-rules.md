# P3-T3 — Capability Negotiation and Downgrade Rules (v0.1)

## Scope and implementation anchors

This specification is implemented by:

- protocol ID selection and downgrade policy in [`pkg/protocol/registry.go`](pkg/protocol/registry.go)
- capability naming/handling in [`pkg/protocol/capabilities.go`](pkg/protocol/capabilities.go)
- negotiation tests in [`pkg/protocol/registry_test.go`](pkg/protocol/registry_test.go) and [`pkg/protocol/capabilities_test.go`](pkg/protocol/capabilities_test.go)

## Feature-flag naming conventions

Capability flags MUST:

1. use `cap.` prefix
2. be lowercase
3. contain only `[a-z0-9.-]` after prefix

Canonical v0.1 flags:

- `cap.chat`
- `cap.voice`
- `cap.management`
- `cap.manifest`
- `cap.identity`
- `cap.sync`

Validation is enforced by [`ValidFeatureFlagName()`](pkg/protocol/capabilities.go:43).

## Unknown capability handling

Deterministic rules implemented by
[`NegotiateCapabilities()`](pkg/protocol/capabilities.go:108):

1. Unknown/invalid *advertised* remote flags are ignored.
2. Unknown/unsupported *required* remote flags are treated as incompatible.
3. Accepted/ignored/missing outputs are sorted deterministically.

## Incompatible capability user feedback behavior

Negotiation returns explicit user-facing feedback categories:

- `none`: all required capabilities satisfied.
- `remote-features-ignored`: optional remote capabilities were ignored.
- `upgrade-required`: required capability mismatch detected.

These map directly to [`CapabilityFeedback`](pkg/protocol/capabilities.go:74).

## Protocol downgrade decision table

| Condition | Decision | Notes |
| --- | --- | --- |
| Family mismatch | Reject | No cross-family fallback. |
| Remote major > local major | Reject | Prevents unsafe major upgrade assumption. |
| Remote major < local major | Reject by default | Allow only via explicit `allowMajorDowngrade=true`. |
| Same major, remote minor <= local minor | Accept | Deterministic minor fallback path. |
| Candidate deprecated by family anchor | Skip candidate | Enforced by deprecation guard before final match. |

This table is realized by [`VersionCompatibilityPolicy.Allows()`](pkg/protocol/registry.go:130)
and [`NegotiateProtocol()`](pkg/protocol/registry.go:177).

