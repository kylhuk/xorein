# Phase 0: Traceability Matrix

| Requirement | Artifact | Notes |
|------------|----------|-------|
| Spaces lifecycle | `pkg/v13/spaces` + tests | founder/admin, visibility, membership validation |
| Join policies | `pkg/v13/joinpolicy`, join policy tests | invite/request/open behaviors, relay policy regression |
| Text channels/chat baseline | `pkg/v13/channels`, `pkg/v13/chat`, tests, `pkg/v13/ui`, e2e/perf | deterministic message states and UI labels |
| Podman validation | `scripts/v13-e2e-podman.sh`, containers/v1.3/* | deterministic manifest and relay regression probe |
| `F14` spec | Phase 4 docs | voice baseline, proto delta, acceptance matrix |
