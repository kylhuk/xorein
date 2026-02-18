# F13 Spaces Lifecycle Spec (v1.3 planning)

## Status
Specification artifact only. No runtime implementation is claimed in v1.2.

## Space lifecycle
1. `create_requested`
2. `created`
3. `active`
4. `archived`
5. `deleted`

## Founder/admin defaults
- Creator becomes founder and initial admin.
- Founder role is immutable unless explicit governance transfer succeeds.
- Admin list updates are auditable and deterministic.

## Visibility and join defaults
- Default visibility: `private`.
- Default join policy: `invite_only`.
- Optional policy extension points remain additive.

## Space metadata baseline
- Stable `space_id`.
- Display name, description, icon hash.
- Channel index summary references.

## Planned vs implemented
- This file defines v1.3 implementation requirements only.
