# Phase 4 - Reproducible build verification (P4-T1)

## Purpose and scope
- Build `cmd/xorein` and `cmd/harmolyn` deterministically for packaging readiness and `G6` evidence.
- Capture stable artifacts under `artifacts/generated/v26-evidence/repro-build`.
- Record hash outputs and support optional baseline replay comparison.

## Script contract
- `scripts/v26-repro-build-verify.sh`
- Command: `./scripts/v26-repro-build-verify.sh`
  - Builds `cmd/xorein` and `cmd/harmolyn` with deterministic flags (`-mod=readonly`, `-trimpath`, `-buildvcs=false`, fixed `-ldflags`).
  - Verifies the outputs are present and executable.
  - Writes
    - `artifacts/generated/v26-evidence/repro-build/binaries/xorein`
    - `artifacts/generated/v26-evidence/repro-build/binaries/harmolyn`
    - `artifacts/generated/v26-evidence/repro-build/checksums.txt`
    - `artifacts/generated/v26-evidence/repro-build/repro-build-manifest.txt`
    - `artifacts/generated/v26-evidence/repro-build/evidence-map.txt`
    - `artifacts/generated/v26-evidence/repro-build/run.log`
    - `artifacts/generated/v26-evidence/repro-build/baseline-comparison.txt`
- Baseline comparison:
  - Command: `./scripts/v26-repro-build-verify.sh --baseline <prior-run-dir>`
  - If provided, the script compares current hashes against `<prior-run-dir>/checksums.txt` and writes `baseline-comparison.txt`.
  - A mismatch terminates with non-zero status.

## Evidence mapping (EV-v26-G6-###)

| Evidence ID | Artifact | Command / trigger | Notes |
|---|---|---|---|
| EV-v26-G6-001 | `artifacts/generated/v26-evidence/repro-build/repro-build-manifest.txt` | `scripts/v26-repro-build-verify.sh` | Deterministic build command, package path, generated timestamp, and SHA-256 list for `xorein`/`harmolyn`. |
| EV-v26-G6-002 | `artifacts/generated/v26-evidence/repro-build/checksums.txt` | `scripts/v26-repro-build-verify.sh` | Stable hash inventory for binary outputs. |
| EV-v26-G6-003 | `artifacts/generated/v26-evidence/repro-build/baseline-comparison.txt` | `scripts/v26-repro-build-verify.sh --baseline ...` | Optional reproducibility comparison against a previous run directory. |

## Planned vs implemented
- Planned: full reproducibility check and evidence capture for all shipped binaries.
- Implemented: reproducible build script and deterministic hash/materialization under `artifacts/generated/v26-evidence/repro-build` are present; optional baseline comparison implemented via `--baseline`.
- Next: wire the generated artifacts into `p4-signing.md` signature workflow.
