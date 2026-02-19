# Phase 4 - SBOM and provenance (P4-T1, ST2)

## Purpose and scope
- Provide a deterministic software-bill-of-materials input for v26 packaging artifacts.
- Keep reproducible build evidence readable and repo-local without external signing tools.

## SBOM source for v2.6
- The reproducible build manifest is the canonical local SBOM input:
  - `artifacts/generated/v26-evidence/repro-build/repro-build-manifest.txt`
- The manifest records:
  - build timestamp
  - package paths
  - output paths
  - SHA-256 for each built binary

## Planned command
- `./scripts/v26-repro-build-verify.sh`
  - Runs with deterministic flags.
  - Rebuilds and writes the manifest and hash set.

## Evidence mapping (EV-v26-G6-###)

| Evidence ID | Artifact | Command / trigger | Notes |
|---|---|---|---|
| EV-v26-G6-001 | `artifacts/generated/v26-evidence/repro-build/repro-build-manifest.txt` | `scripts/v26-repro-build-verify.sh` | Manifest used as SBOM input for deterministic checksum audit. |

## Planned vs implemented
- Planned: full SPDX generation and package-manager transitive component capture.
- Implemented: deterministic manifest + checksum inventory are produced by the reproducible build script and can be treated as local SBOM input.
