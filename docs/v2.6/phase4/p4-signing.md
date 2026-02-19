# Phase 4 - Signing and provenance proof (P4-T1, ST2)

## Purpose and scope
- Define a reproducible signature workflow for v26 build artifacts.
- Ensure the same payload is signed by each release run so verifier output is stable.

## Signing payload design
- Recommended payload file:
  - `artifacts/generated/v26-evidence/repro-build/checksums.txt`
- The payload should include all binary names and deterministic hashes so any drift is detectable.

## Planned execution
- Deterministic command example:
  ```bash
  ./scripts/v26-repro-build-verify.sh
  # then sign the checksums payload via your preferred detached-signature tool
  ```
- The resulting signature file and signature command output must be written beside the reproducibility artifacts for evidence.

## Evidence mapping (EV-v26-G6-###)

| Evidence ID | Artifact | Command / trigger | Notes |
|---|---|---|---|
| EV-v26-G6-002 | `artifacts/generated/v26-evidence/repro-build/checksums.txt` | `scripts/v26-repro-build-verify.sh` | Baseline and payload for signing workflow; reused by external verifier. |

## Planned vs implemented
- Planned: external verifier-friendly detached signature generation and signature verification script integration.
- Implemented: payload path is fixed and stable so signing can be chained deterministically after rebuild.
