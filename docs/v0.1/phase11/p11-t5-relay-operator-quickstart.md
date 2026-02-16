# P11-T5 v0.1 Relay Operator Quickstart (CLI + Container)

Status: Complete for Phase 11 P11-T5. Verification commands were executed with Podman and the outputs below satisfy the evidence contract.

## Scope

This quickstart covers the headless relay runtime path (single binary, `--mode=relay`) and the container workflow artifacts captured for v0.1. It is scoped to the relay MVP defined in Phase 9 and does not claim SFU/advanced voice features are implemented.

## Prerequisites

- Podman installed and available on the host.
- Repository checked out at the v0.1 planning/execution snapshot.

## Quickstart A: relay mode (single binary)

The relay runtime is enabled via `--mode=relay` and requires explicit listen/store/health parameters.

```bash
podman run --rm --userns=keep-id -v "$PWD":/workspace:Z -w /workspace \
  docker.io/library/golang:1.24.8 bash -lc \
  'export PATH=/usr/local/go/bin:$PATH; go run ./cmd/aether --mode=relay --relay-listen 0.0.0.0:4001 --relay-store ./artifacts/generated/relay-store --relay-health-interval 30s'
```

Verification output (2026-02-16, Podman golang:1.24.8):
```
Relay runtime active mode=relay listen=0.0.0.0:4001 store=./artifacts/generated/relay-store
Relay policy: reservation_limit=256 session_timeout=2m0s max_bytes_per_sec=1000000 active=0 rejected=0 timed_out=0 established=0
Relay health status: state=ready started_at=2026-02-16T11:32:30.613337814Z next_health_in=30s
```

Expected output fields are emitted by [`runRelayMode()`](cmd/aether/main.go:97), including the active listen/store configuration and a ready-state health line.

## Quickstart B: container workflow artifacts

The relay container build workflow is defined in [`docs/v0.1/phase9/p9-t5-relay-container-workflow.md`](docs/v0.1/phase9/p9-t5-relay-container-workflow.md:1) and produces the following artifacts under `artifacts/generated/relay-container/`:

- Image ID: [`artifacts/generated/relay-container/image-id.txt`](artifacts/generated/relay-container/image-id.txt:1)
- Repo digest: [`artifacts/generated/relay-container/image-digest.txt`](artifacts/generated/relay-container/image-digest.txt:1)
- Signing command stub: [`artifacts/generated/relay-container/signing-command.txt`](artifacts/generated/relay-container/signing-command.txt:1)
- SBOM: [`artifacts/generated/relay-container/sbom/sbom.spdx.json`](artifacts/generated/relay-container/sbom/sbom.spdx.json:1)
- Publication checklist: [`artifacts/generated/relay-container/publication-checklist.txt`](artifacts/generated/relay-container/publication-checklist.txt:1)

## Security and media-scope note

- SFrame-based media E2EE remains a Phase 8 research outcome and is not enabled by the relay MVP. See the feasibility summary in [`pkg/phase8/SFrame_feasibility_note.md`](pkg/phase8/SFrame_feasibility_note.md:1).
- Relay mode does not provide SFU-level media processing or access to decrypted media in v0.1 (deferred scope remains listed in [`TODO_v01.md`](TODO_v01.md:74)).

## Troubleshooting

- `invalid --mode "..."; expected client|relay|bootstrap`
  - Use `--mode=relay` or omit `--mode` when running non-relay scenarios (see [`cmd/aether/main.go`](cmd/aether/main.go:17)).
- `invalid relay configuration: --relay-listen must be non-empty host:port`
  - Provide a concrete `--relay-listen` value (see [`cmd/aether/main.go`](cmd/aether/main.go:98)).
- `invalid relay configuration: --relay-store must be non-empty path`
  - Provide `--relay-store` for store-and-forward data (see [`cmd/aether/main.go`](cmd/aether/main.go:102)).
- `invalid relay configuration: --relay-health-interval must be greater than 0`
  - Provide a positive duration (see [`cmd/aether/main.go`](cmd/aether/main.go:106)).

## Current limitations and deferred scope

The relay MVP is limited to the v0.1 baseline (DHT bootstrap + Circuit Relay v2 + basic store-and-forward). The following remain deferred per [`TODO_v01.md`](TODO_v01.md:74):

- SFU and advanced voice optimizations
- Screen share, file transfer
- Public discovery and ecosystem expansion
- Bot platform and API shim
- Compliance hardening beyond baseline engineering controls
