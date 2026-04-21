# AGENTS.md

## Repo shape
- `xorein` is both a protocol family and a runtime binary; do not introduce breaking wire changes unless explicitly requested.
- The real runtime entrypoint is `cmd/aether/main.go`; live node wiring is in `pkg/node/service.go` and local control-plane auth/routes are in `pkg/node/control.go`.
- Protocol IDs/capabilities live in `pkg/protocol/`; protobuf source lives in `proto/aether.proto`; `gen/go/proto/aether.pb.go` is generated and must not be edited directly.
- `pkg/app/seams.go` is a scaffold/boundary file, not the current startup path.
- Treat `artifacts/generated/**` and `bin/aether` as outputs, not source.

## Source of truth
- Prefer `Makefile`, `scripts/*.sh`, `buf.yaml`, `.golangci.yml`, and `.pre-commit-config.yaml` over prose docs.
- There are no checked-in `.github/workflows/*` files in this repo; do not rely on workflow badges/docs as current CI truth.

## Commands
- Main order: `generate -> compile -> lint -> test -> scan -> build`.
- `make check-fast` = `generate compile lint`.
- `make check-full` = `generate compile lint test scan`.
- `make pipeline` = `generate compile lint test scan build`.
- `make test` runs `scripts/dhall-verify.sh` then `scripts/repro-checksums.sh`.
- `make scan` runs `govulncheck`, `gosec`, and `trivy` (via Podman) and is the real security gate.
- `make build` currently writes a placeholder `bin/aether`; it is still required before `make release-pack-verify`.
- If touching protobuf, run Buf checks via Podman (`buf lint`, `buf breaking`; generation is policy-relevant even though it is not wired into `make generate`).
- Run `pre-commit run --all-files` when checking repo hygiene.

## Gotchas
- `make compile` and `make lint` are readiness checks, not full Go compile/lint.
- `scripts/dhall-verify.sh` and `scripts/release-pack-verify.sh` expect generated artifacts and Podman.
- Keep wire/protocol evolution additive; preserve field numbers and reserve removed protobuf numbers/names.
- Never log secrets, private keys, key material, or sensitive payloads.
