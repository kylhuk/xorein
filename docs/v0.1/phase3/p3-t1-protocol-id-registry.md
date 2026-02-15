# P3-T1 — v0.1 Multistream Protocol ID Registry

## Canonical namespace

All v0.1 protocol IDs use the canonical multistream namespace:

- `/aether/<family>/<major>.<minor>`

The registry is implemented in [`pkg/protocol/registry.go`](pkg/protocol/registry.go) and
validated by [`pkg/protocol/registry_test.go`](pkg/protocol/registry_test.go).

## Registered v0.1 IDs

| Family | Canonical ID | Registry symbol |
| --- | --- | --- |
| Chat | `/aether/chat/0.1` | `chatV01` |
| Voice signaling | `/aether/voice/0.1` | `voiceV01` |
| Manifest | `/aether/manifest/0.1` | `manifestV01` |
| Identity | `/aether/identity/0.1` | `identityV01` |
| Sync | `/aether/sync/0.1` | `syncV01` |

## Deterministic downgrade/selection strategy (v0.1)

Negotiation behavior is implemented by [`NegotiateProtocol()`](pkg/protocol/registry.go:177)
and [`VersionCompatibilityPolicy.Allows()`](pkg/protocol/registry.go:130):

1. Family mismatch is rejected.
2. Offers with a higher major version than local candidate are rejected.
3. Major downgrade is disallowed by default and only allowed by explicit policy
   override (`allowMajorDowngrade=true`).
4. Minor downgrade is allowed by default (`allowMinorDowngrade=true`) with
   `minimumMinor=1` guard.
5. Deprecated candidates are skipped using
   [`DeprecationGuard.IsDeprecated()`](pkg/protocol/registry.go:161).

This strategy is intentionally identical across all registered families for
v0.1 to avoid family-specific ambiguity.

