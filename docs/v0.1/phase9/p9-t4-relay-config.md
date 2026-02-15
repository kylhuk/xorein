# P9-T4 Relay Configuration Model and Generation Flow

Status: ✅ Completed during v0.1 execution slice on 2026-02-15.

## 1. Dhall Schema

Relay configuration is now modeled directly in Dhall to keep all node traits deterministic and source controlled. The schema lives in [`config/dhall/types.dhall`](config/dhall/types.dhall) and introduces the following records:

| Record | Purpose |
| --- | --- |
| `Relay.Node` | Static identity for each relay instance (name, region, environment, role, listen/announce multiaddrs, free-form tags). |
| `Relay.Limits` | Admission controls for the circuit relay policy (max circuits, per-session duration string, per-circuit bandwidth cap). |
| `Relay.StoreForward` | Store-and-forward policy flags (enabled, storage path, deterministic quota counts, TTL text). |
| `Relay.Sfu` | Selective forwarding unit toggles and sizing fields (rooms, participants). |
| `Relay.Metrics` | Observability enablement and listen address for Prometheus scraping. |
| `Relay.Health` | Interval string for liveness/health hints surfaced via logs. |
| `Relay.Config` | Aggregate structure that combines Node + service policies. |
| `Relay.Environment` | Multi-node collection keyed by deployment environment name. |

The base single-node template is defined in [`config/dhall/default.dhall`](config/dhall/default.dhall) under `relay.base`. This template ensures the following defaults:

* TLS-neutral TCP + QUIC listen multiaddrs plus a WebSocket control address.
* Circuit relay protection of 512 concurrent reservations, 2-minute max duration, 1 Mbps per circuit.
* Store-and-forward enabled with 8k message quota, 512 MiB storage cap, and 30-day TTL.
* SFU enabled with 50 rooms and 100 participants per room (offline-friendly baseline).
* Metrics and health probes are explicit and environment-agnostic.

Environment overlays reside in [`config/dhall/env.dhall`](config/dhall/env.dhall). They define concrete node instances for `dev`, `staging`, and `production` along with environment-specific ports and announce addresses. Each environment list references the shared base template via `mkRelay`, guaranteeing consistent defaults while allowing region- and env-specific port overrides.

## 2. Generated Config Validation

Generation is currently a placeholder stage until downstream consumers exist, but the deterministic verification gate is already wired. The command (also referenced in [`docs/v0.1/phase2/dhall-ops.md`](docs/v0.1/phase2/dhall-ops.md)) is:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 \
  sh -lc './scripts/dhall-verify.sh'
```

Output observed during this task (Command evidence):

```text
dhall verification placeholder: config sources present
```

Gate rules:

1. All Dhall sources (`types.dhall`, `default.dhall`, `env.dhall`) must exist.
2. The verification script must run inside Podman to mirror future CI behavior.
3. Non-zero exit or missing files fails the gate.

## 3. Next Steps / Consumers

* Relay runtime wiring (P9) will hydrate configs from Dhall once TOML/JSON exporters are added.
* Build/ops tasks will eventually render per-environment manifests and bake them into container images.
* Additional environments can be added by extending `relayEnvironments` in `config/dhall/env.dhall` following the documented pattern.

This document, combined with the Dhall updates and verification evidence, satisfies P9-T4 deliverables: schema definition and validation checklist.
