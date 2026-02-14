# Proto Reservation Policy (Phase 2)

1. Field numbers and names are governed by additive-only extensions. When a field or message is removed, record it via `reserved` statements (see [`proto/aether.proto`](proto/aether.proto:1)).
2. Reserved entries must cite the rationale in a nearby comment so future owners can understand why the slot cannot be reused.
3. Workflows touching `.proto` files must re-run `buf generate`/`buf lint`/`buf breaking` and capture outputs before merging. Any failing command is merge-blocking.
4. Document any interoperability impact in Phase 2 notes (this document or [`docs/v0.1/phase2/scaffold-boundaries.md`](docs/v0.1/phase2/scaffold-boundaries.md:1)).

## Automated compatibility checks (P2-T6)

- Baseline command set:
  - `podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 lint`
  - `podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 breaking --against '.git#branch=main'`
  - `podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 generate`
- Enforcement rule: lint/breaking/generate outputs are required in evidence for any proto-affecting change.
- Compatibility gate: if baseline or schema config is invalid, task remains open until buf commands run cleanly with recorded outputs.
