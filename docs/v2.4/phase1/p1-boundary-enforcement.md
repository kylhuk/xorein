# Boundary enforcement (P1-T3)

## ST1 – backend Gio imports blocked (G9)
- Added `scripts/ci/enforce-boundaries.sh`, which scans `cmd/xorein` and `pkg/xorein/**` for any `gioui.org/...` import literal and fails the run if a match is found.
- **Evidence command:** `bash scripts/ci/enforce-boundaries.sh`

## ST2 – harmolyn limited to local API glue (G9)
- The same script enumerates `cmd/harmolyn` imports and rejects any reference to known daemon/runtime packages such as `pkg/xorein`, `pkg/v24/daemon`, the legacy `pkg/v{18..23}` slices, `pkg/phase5`, or `pkg/protocol`.
- **Evidence command:** `bash scripts/ci/enforce-boundaries.sh`

## ST3 – CI gate + tests (G9)
- CI and release builders must run `scripts/ci/enforce-boundaries.sh` before promotion; it reports the offending file/line and exits with a non-zero status when a rule is violated.
- The script accepts a `CHECK_ROOT` override for deterministic fixture runs, which is exercised by `tests/e2e/v24/boundary_enforcement_test.go` so corruption of the gate is caught by unit tests without touching the real repo content.
- **Evidence commands:**
  - `bash scripts/ci/enforce-boundaries.sh`
  - `go test ./tests/e2e/v24 -run TestBoundaryEnforcement`
