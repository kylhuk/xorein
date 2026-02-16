# P11-T5 v0.1 User First-Launch Quickstart (CLI Harness)

Status: Complete for Phase 11 P11-T5. Verification commands were executed with Podman and the outputs below satisfy the evidence contract.

## Scope

This quickstart documents the CLI scenario harness used to validate the v0.1 first-contact baseline. It does **not** launch the GUI shell; GUI flow validation is tracked separately in [`docs/v0.1/phase10/p10-t6-first-launch-validation.md`](docs/v0.1/phase10/p10-t6-first-launch-validation.md:1).

## Prerequisites

- Podman installed and available on the host.
- Repository checked out at the v0.1 planning/execution snapshot.

## Quickstart (first-contact scenario)

Run the scenario harness with Podman so the outputs align with the evidence model.

```bash
podman run --rm --userns=keep-id -v "$PWD":/workspace:Z -w /workspace \
  docker.io/library/golang:1.24.8 bash -lc \
  'export PATH=/usr/local/go/bin:$PATH; go run ./cmd/aether --scenario=first-contact --first-contact-runs=3'
```

Verification output (2026-02-16, Podman golang:1.24.8):
```
First-contact baseline generated: runs=3 passed=3 failed=0 pass_rate=1.00 output=artifacts/generated/first-contact
Duration metrics: target=5m0s mean_ms=0 median_ms=0
Run 01 success=true target_met=true duration=1.257471ms
Run 02 success=true target_met=true duration=128.782µs
Run 03 success=true target_met=true duration=114.892µs
```

Outputs are written under `artifacts/generated/first-contact/` by default (see artifacts below).

## First-contact flags

- `--first-contact-runs`: number of repeated first-contact runs to execute (default: `3`).
- `--first-contact-output`: output directory for generated summary and per-run artifacts (default: `artifacts/generated/first-contact`).
- `--first-contact-target`: target duration per run used for target-met reporting (default: `5m`).

Example:

```bash
go run ./cmd/aether --scenario=first-contact --first-contact-runs=5 --first-contact-output=artifacts/generated/first-contact --first-contact-target=2m
```

## Evidence anchors (existing artifacts)

- Baseline summary: [`artifacts/generated/first-contact/summary.md`](artifacts/generated/first-contact/summary.md:1)
- Per-run JSON: [`artifacts/generated/first-contact/run-01.json`](artifacts/generated/first-contact/run-01.json:1), [`artifacts/generated/first-contact/run-02.json`](artifacts/generated/first-contact/run-02.json:1), [`artifacts/generated/first-contact/run-03.json`](artifacts/generated/first-contact/run-03.json:1)
- Regression mirror (P11-T2): [`artifacts/generated/regression/report.txt`](artifacts/generated/regression/report.txt:1)

## Security and media-scope note

- SFrame-based media E2EE is a research deliverable only; it is **not** enabled or validated by the v0.1 CLI harness. Track the feasibility work and constraints in [`pkg/phase8/SFrame_feasibility_note.md`](pkg/phase8/SFrame_feasibility_note.md:1).
- Any SFU-level or server-side media processing expectations remain out of scope for v0.1 (see deferred items in [`TODO_v01.md`](TODO_v01.md:74)).

## Troubleshooting

- `unknown scenario "..."; valid scenarios: create-server, join-deeplink, first-contact`
  - Use `--scenario=first-contact` (see [`cmd/aether/main.go`](cmd/aether/main.go:56)).
- `first-contact scenario failed: ...`
  - Inspect the `run-*.json` artifacts for failure fields; the harness emits owner-tagged failure details in [`artifacts/generated/first-contact/`](artifacts/generated/first-contact/run-01.json:1).
- `invalid --mode "..."; expected client|relay|bootstrap`
  - If passing `--mode`, use one of the supported values in [`cmd/aether/main.go`](cmd/aether/main.go:17).

## Current limitations and deferred scope

The quickstart is limited to the v0.1 CLI harness and excludes deferred v0.2+ scope from [`TODO_v01.md`](TODO_v01.md:74):

- DM protocol with X3DH and Double Ratchet
- Presence, friends, notifications
- SFU and advanced voice optimizations
- Screen share, file transfer
- RBAC moderation and audit features
- Bot platform and API shim
- Public discovery and ecosystem expansion
- Push notification relays
- Compliance hardening beyond baseline engineering controls
