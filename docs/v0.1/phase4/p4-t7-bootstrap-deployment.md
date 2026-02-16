# P4-T7 Bootstrap node deployment plan (three regions)

## Scope

This document defines the minimal bootstrap deployment topology and operations runbook required by **P4-T7**. It is an execution artifact, but it does not imply implementation of bootstrap runtime health checks beyond the configuration model in [`config/dhall/types.dhall`](config/dhall/types.dhall:1).

## Dependencies

- Dhall configuration model in [`config/dhall/types.dhall`](config/dhall/types.dhall:1) and defaults in [`config/dhall/default.dhall`](config/dhall/default.dhall:1).
- Environment-specific node inventory in [`config/dhall/env.dhall`](config/dhall/env.dhall:1).
- Reproducibility and release controls tracked under **P2-T8** (not a blocker for this plan, but required before production sign-off).

## Source of truth

Bootstrap node inventories are defined in [`config/dhall/env.dhall`](config/dhall/env.dhall:1) under `bootstrapEnvironments`. This plan documents the expected naming, region distribution, and addressing derived from that inventory.

## Topology plan (three regions)

Each environment runs three bootstrap nodes (EU, US, AP) with distinct listen/announce pairs and contacts. The tables below reflect the current inventory:

### Dev

| Node name | Region | Listen | Announce | Contact |
|---|---|---|---|---|
| `bootstrap-dev-eu` | `eu-west` | `/ip4/0.0.0.0/tcp/3101`, `/ip4/0.0.0.0/udp/3101/quic-v1` | `/dns4/bootstrap-dev-eu.aether.test/tcp/3101` | `noc+dev-eu@aether.test` |
| `bootstrap-dev-us` | `us-east` | `/ip4/0.0.0.0/tcp/3201`, `/ip4/0.0.0.0/udp/3201/quic-v1` | `/dns4/bootstrap-dev-us.aether.test/tcp/3201` | `noc+dev-us@aether.test` |
| `bootstrap-dev-ap` | `ap-south` | `/ip4/0.0.0.0/tcp/3301`, `/ip4/0.0.0.0/udp/3301/quic-v1` | `/dns4/bootstrap-dev-ap.aether.test/tcp/3301` | `noc+dev-ap@aether.test` |

### Staging

| Node name | Region | Listen | Announce | Contact |
|---|---|---|---|---|
| `bootstrap-staging-eu` | `eu-central` | `/ip4/0.0.0.0/tcp/3401`, `/ip4/0.0.0.0/udp/3401/quic-v1` | `/dns4/bootstrap-staging-eu.aether.test/tcp/3401` | `noc+staging-eu@aether.test` |
| `bootstrap-staging-us` | `us-west` | `/ip4/0.0.0.0/tcp/3501`, `/ip4/0.0.0.0/udp/3501/quic-v1` | `/dns4/bootstrap-staging-us.aether.test/tcp/3501` | `noc+staging-us@aether.test` |
| `bootstrap-staging-ap` | `ap-south` | `/ip4/0.0.0.0/tcp/3601`, `/ip4/0.0.0.0/udp/3601/quic-v1` | `/dns4/bootstrap-staging-ap.aether.test/tcp/3601` | `noc+staging-ap@aether.test` |

### Production

| Node name | Region | Listen | Announce | Contact |
|---|---|---|---|---|
| `bootstrap-prod-eu` | `eu-central` | `/ip4/0.0.0.0/tcp/3701`, `/ip4/0.0.0.0/udp/3701/quic-v1` | `/dns4/bootstrap-prod-eu.aether.chat/tcp/3701` | `noc+prod-eu@aether.chat` |
| `bootstrap-prod-us` | `us-east` | `/ip4/0.0.0.0/tcp/3801`, `/ip4/0.0.0.0/udp/3801/quic-v1` | `/dns4/bootstrap-prod-us.aether.chat/tcp/3801` | `noc+prod-us@aether.chat` |
| `bootstrap-prod-ap` | `ap-south` | `/ip4/0.0.0.0/tcp/3901`, `/ip4/0.0.0.0/udp/3901/quic-v1` | `/dns4/bootstrap-prod-ap.aether.chat/tcp/3901` | `noc+prod-ap@aether.chat` |

## Naming and addressing convention

- **Node name format:** `bootstrap-<environment>-<region>` (e.g., `bootstrap-staging-eu`).
- **Region codes:** derived from the `region` field in [`config/dhall/env.dhall`](config/dhall/env.dhall:1).
- **DNS names:** `<node-name>.aether.test` for dev/staging, `<node-name>.aether.chat` for production.
- **Listen/announce pairing:** listen on `0.0.0.0` for TCP/UDP ports, announce via DNS hostname on the TCP port.

## Observability signals (baseline)

Baseline observability values are defined in [`config/dhall/default.dhall`](config/dhall/default.dhall:1) under the bootstrap `metrics` and `health` blocks:

- **Metrics:** enabled with `listenAddr = "0.0.0.0:9191"`.
- **Health:** `interval = "15s"` and `expectPeers = 32`.

Operators should ensure the runtime exports these signals and that alerts are keyed on:

- Metrics endpoint availability (port bind + scrape reachability).
- Health interval adherence (stale interval indicates degraded status).
- Peers below `expectPeers` for sustained windows in production.

## Operations runbook (plan)

This runbook captures the expected operator actions and is intentionally minimal. It does not imply implementation of automation beyond the configuration model above.

### Preflight

- Ensure node entries exist in [`config/dhall/env.dhall`](config/dhall/env.dhall:1) and match the naming/region conventions above.
- Verify Dhall sources with [`scripts/dhall-verify.sh`](scripts/dhall-verify.sh:1).

### Start

- Deploy the bootstrap process using the environment-specific config rendered from `bootstrapEnvironments` (Dhall source of truth).
- Confirm the process binds to the configured listen addresses and announces the DNS address listed in the inventory table.

### Restart

- Use the environment’s standard service manager to restart the bootstrap process.
- After restart, confirm the process rebinds to the same listen/announce addresses and resumes metrics + health signaling.

### Health check

- Validate that metrics are reachable on the configured `metrics.listenAddr` and that the health interval is consistent with `health.interval`.
- Investigate sustained `expectPeers` shortfalls before scaling beyond the three-region baseline.
