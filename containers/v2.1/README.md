# v2.1 Podman Scenarios

These assets describe persistence and relay-boundary checks that run inside lightweight Podman containers.

`scripts/v21-persistence-scenarios.sh` executes podman-backed probes, writes per-scenario logs and a run manifest, and records each scenario as `pass`, `fail`, or `BLOCKED:ENV` based on command results.
