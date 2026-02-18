# Phase 1 Gate Runner

## Purpose

The v11 gate runner validates machine-readable gate status artifacts so release owners can automatically
detect missing, stale, or unpromoted gates before promotion.

## Running the command

The gate runner is exposed at `cmd/v11-gate-runner`. The easiest way to execute it is via the Makefile target:

```bash
make v11-gate-runner
```

You can run the command directly and override the status directory when needed:

```bash
go run ./cmd/v11-gate-runner --status-dir artifacts/v11/gates
```

The command prints a deterministic gate summary, describes why the run passed or failed, and exits
non-zero if any of the following conditions are true:

- one or more required per-gate artifacts are missing or malformed,
- one or more gate files are older than the 24-hour freshness threshold,
- any gate is `open` or `blocked`.

## Artifact layout

The runner reads from `artifacts/v11/gates/` and expects one file per gate:

- `G0.status.json`
- `G1.status.json`
- `G2.status.json`
- `G3.status.json`
- `G4.status.json`
- `G5.status.json`
- `G6.status.json`
- `G7.status.json`

Missing files are treated as fail-close conditions.

## Status file schema

Each file is a JSON object with the following shape:

```json
{
  "gateId": "G0",
  "state": "promoted",
  "updatedAt": "2026-02-17T19:30:00Z",
  "evidenceId": "EV-v11-G0-001",
  "owner": "Runtime Lead",
  "approver": "Plan Lead",
  "notes": "Initial promotion decision"
}
```

Field meanings:

- `gateId`: identifier between `G0` and `G7`, matching the filename.
- `state`: one of `open`, `blocked`, or `promoted`. Any unrecognized value is treated as `open`.
- `updatedAt`: RFC3339 timestamp describing the last transition used for freshness checks.
- `evidenceId`, `owner`, `approver`, `notes`: optional metadata used by sign-off workflows.

The runner prints the gates in deterministic order (G0..G7) and reports the summary line-by-line.
If everything is promoted and fresh, it prints `PASS` and exits zero; otherwise it prints `FAIL` with an
explanation.
