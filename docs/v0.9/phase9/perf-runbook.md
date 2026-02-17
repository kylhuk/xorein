# VA-E7 · Performance Reproducibility Runbook

## Purpose
This runbook records the deterministic setup that gate reviewers can follow to reproduce the baseline stress/perf contract targets (1000-member server planning profile, 50-person voice cascade coverage, latency-threshold checks, +50 incremental steps). Implementation in this slice includes deterministic suites under `tests/perf/v09/deterministic_suite_test.go`, the e2e suite under `tests/e2e/v09/scenario_suite_test.go`, and the CLI witness scenario.

## Repro steps
1. Checkout this v0.9 slice so deterministic helpers are present in `pkg/v09/` and the witness command is available in `cmd/aether`.
2. Run `go run ./cmd/aether --mode=client --scenario=v09-forge` and capture stdout. Expected output: `v0.9 forge: PASS`.
3. Run `go test ./tests/perf/v09/...` and capture stdout. This suite includes deterministic +50 campaign coverage, 50-person cascade coverage, and latency-threshold checks.
4. Run `go test ./tests/e2e/v09/...` and capture stdout to confirm the integrated deterministic scenario path.
5. Preserve the deterministic thresholds documented in `docs/v0.9/phase9/VA-E7-thresholds.md` as the pass/fail criteria when presenting to reviewers.

## Evidence anchors
| Metric | Target | Documentation |
|---|---|---|
| Base server membership | 1000 members | `docs/v0.9/phase5/p5-stress-contract.md` |
| Voice participants | 50 participants | `docs/v0.9/phase5/p5-stress-contract.md` |
| +50 increments | Deterministic increments validated by automated test | `tests/perf/v09/deterministic_suite_test.go`, `docs/v0.9/phase9/VA-E7-thresholds.md` |
