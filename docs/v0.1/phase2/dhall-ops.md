# Dhall Ops (Phase 2)

1. Type definitions live under [`config/dhall/types.dhall`](config/dhall/types.dhall:1) so all environments share the same schema.
2. Defaults (`config/dhall/default.dhall`) load the schema and set deterministic placeholder values.
3. Environment-specific overrides (`config/dhall/env.dhall`) currently just reuse defaults; future variants can `merge` additional records.
4. Generation/verification script scaffolds reside in [`scripts/dhall-verify.sh`](scripts/dhall-verify.sh:1) and should run from the repository root to match Podman-in-CI expectations.

## Generation and verification stage (P2-T7)

- Stage command (containerized):
  - `podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc './scripts/dhall-verify.sh'`
- Expected stage output:
  - Confirms Dhall source files are present and emits deterministic placeholder verification text.
- Gate rule:
  - Missing script, missing Dhall source files, or non-zero stage exit is merge-blocking for config-surface changes.
