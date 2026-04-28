# Draft: Quick Start Guide

## Requirements (confirmed)
- Build the server binary.
- Run the server.
- Explain what clients need to access it.

## Technical Decisions
- Use `make build` as the primary build command.
- Treat `cmd/aether/main.go` as the runtime entrypoint.
- Separate **runtime network access** from **local control-plane access**.

## Research Findings
- `make build` builds `bin/aether`.
- The runtime accepts `--mode`, `--config`, `--data-dir`, `--listen`, `--bootstrap-addrs`, `--manual-peers`, `--relay-addrs`, and `--control`.
- The control API is local-only and bearer-token protected via `DataDir/control.token`.
- Other nodes/clients reach the runtime through the P2P listener, not the control API.

## Open Questions
- None blocking.

## Scope Boundaries
- INCLUDE: build, run, client connectivity, and admin access notes.
- EXCLUDE: protocol internals, release packaging, and deep troubleshooting.

## Quick Start Guide Draft

### 1) Build the server binary
```bash
make build
```

This produces `bin/aether`.

### 2) Run the server
```bash
./bin/aether --mode relay --data-dir ./data --listen 0.0.0.0:12345
```

If you prefer, you can run the binary directly from source:
```bash
go run ./cmd/aether --mode relay --data-dir ./data --listen 0.0.0.0:12345
```

Useful flags:
- `--mode`: `client`, `relay`, `bootstrap`, or `archivist` (use `relay` or `bootstrap` for a node others connect to)
- `--data-dir`: where runtime state and the control token are stored
- `--listen`: the runtime/P2P listen address
- `--bootstrap-addrs`: seed peers for initial discovery
- `--manual-peers`: explicit peers to connect to
- `--relay-addrs`: relay addresses, if needed
- `--control`: local control endpoint override

### 3) What clients need to access it

There are two different access paths:

#### A. Other nodes / network clients
Make the runtime listen on an address reachable from your clients, then provide peer/bootstrap information.

Example:
```bash
./bin/aether --mode relay --listen 0.0.0.0:12345 --bootstrap-addrs <bootstrap-peer>
```

Notes:
- The runtime listener is separate from the control API.
- Use `--bootstrap-addrs`, `--manual-peers`, and/or `--relay-addrs` so clients can find the node.

#### B. Local admin / control clients
The control API is **not** for remote clients.

Requirements:
- Run on the same machine (or via the Unix socket / loopback endpoint).
- Read the token from `./data/control.token`.
- Send `Authorization: Bearer <token>`.

### 4) Minimal example
```bash
mkdir -p ./data
make build
./bin/aether --mode bootstrap --data-dir ./data --listen 0.0.0.0:12345
```

Then give clients:
- the reachable listen address
- any bootstrap/peer addresses they need
- for local control access only: the `control.token` file path
