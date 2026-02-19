# P3 Podman Scenarios

This stage documents the deterministic container scenarios that prove persistence, search, and relay-boundary contracts.

1. **multi-peer restart/resume probe (podman-backed)**: execute the scenario probe in `scripts/v21-persistence-scenarios.sh` and verify scenario success from command outcomes.
2. **local history clear/recovery probe (podman-backed)**: execute the scenario probe and treat runtime-backed command results as definitive for clear/recovery behavior.
3. **relay no-long-history-hosting boundary probe (podman-backed)**: execute the scenario probe and record whether the relay boundary probe command passes.

Each run rewrites `artifacts/generated/v21-persistence-scenarios/manifest.txt` from live command outcomes and writes execution logs under `artifacts/generated/v21-persistence-scenarios`.

The manifest records scenario status values as:
- `pass`: probe command completed successfully
- `fail`: probe command returned non-zero or missing expected output
- `BLOCKED:ENV`: required podman runtime was unavailable or unusable
