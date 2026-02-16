# Phase 5 - Task 6: Profile Privacy & Metadata Minimization

## Objective
Minimize the amount of profile data exposed by default and classify the remaining metadata according to privacy sensitivity so we can harden signing, publishing, and diagnostics flows.

## Field Sensitivity Matrix
| Field        | Sensitivity  | Notes |
|--------------|--------------|-------|
| `identity`   | Restricted   | Hex-encoded Ed25519 public key. Only exposed when cryptographically required. |
| `display_name` | Public    | User-facing handle. Sanitized before display. |
| `bio`        | Personal     | Optional, defaults to empty. Never included in redacted surfaces. |
| `avatar_url` | Personal     | Optional, defaults to empty. Allowed only when explicitly provided. |
| `version`    | Operational  | Implementation metadata. Shared to detect downgrade/invalidation. |
| `updated_at` | Operational  | RFC3339Nano timestamp in UTC. Used for deterministic ordering. |

## Default-Off Strategy for Optional Fields
- `DefaultProfileOptionalFields()` now resets `Bio` and `AvatarURL` to empty strings unless explicitly set during profile edit.
- `RedactedProfileMetadata()` emits only `identity`, `display_name`, `version`, and `updated_at`.
- Publish/resolve caching layers operate on redacted copies when diagnostics or membership views need metadata.

## Metadata Minimization Checklist
1. **Field classification enforced** via `ProfileFieldSensitivityMap()` copy semantics (tests ensure map cannot be mutated externally).
2. **Optional fields cleared by default** (`DefaultProfileOptionalFields()`), ensuring Sender Keys/relay diagnostics never inherit stale user-provided data.
3. **Redacted view coverage** for diagnostics: `RedactedProfileMetadata()` excludes personal fields and produces deterministic timestamps.
4. **Store/cache invariants** verified through `ProfileStorePublishUpdateDeterminism` and `ProfileStoreCacheInvalidationOnAcceptedUpdate` to prevent historical leakage.
5. **Documentation alignment**: this note is referenced by TODO tracking to demonstrate completion.

## Evidence
- Code: [`pkg/phase5/profile.go`](../../../pkg/phase5/profile.go)
- Tests: [`pkg/phase5/profile_test.go`](../../../pkg/phase5/profile_test.go)
- Verification command output captured via Podman (see execution log in this run).

