# Phase 3 - Podman relay smoke

## Purpose and scope
- Exercise the v11 relay policy within a Podman container so the runtime artifact is verified before the Podman smoke gate (EV-v11-G4) is marked complete.
- Focus on the persistence-mode probes that must pass or fail deterministically: relays running in Podman must accept allowed persistence modes and reject forbidden ones before the gate runner is satisfied.
- Capture deterministic evidence for Podman smoke in the form of the manifest/log files produced by `scripts/v11-relay-smoke.sh` so they can be added to the gate evidence index.

## Command usage
- The smoke script lives at `scripts/v11-relay-smoke.sh`. Run it from the workspace root after Podman is available and the containers/v1.1 assets are built.
- Example usage:
  ```bash
  ./scripts/v11-relay-smoke.sh
  ```
- The script runs both probes (allowed and forbidden persistence mode), captures logs, and writes a deterministic pass/fail manifest.

## Expected pass/fail probes
| Probe | Persistence mode | Expected result | Rationale |
|---|---|---|---|
| Allowed metadata path | `session-metadata` or `transient-metadata` | Pass | Podman smoke runs the relay with an allowed persistence-mode flag defined in `pkg/v11/relaypolicy/policy.go`; the command must exit zero and emit the runtime-active marker. |
| Forbidden payload path | `durable-message-body`, `attachment-payload`, or `media-frame-archive` | Fail | The script requests a forbidden persistence mode, the relay exits with `ValidationError`, and the manifest records the rejection so we can prove the boundary enforcement in Podman. |

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| EV-v11-G4-001 | Podman smoke passes when allowed persistence modes are used | `scripts/v11-relay-smoke.sh` run logs and manifest for the `session-metadata` probe (capture STDOUT/stderr and manifest file in phase artifacts). |
| EV-v11-G4-002 | Podman smoke fails when forbidden persistence modes are requested | Same script run with the forbidden persistence modes; manifest/logs document the explicit `ValidationError` exit. |

## Planned vs implemented
- **Planned:** Integrate smoke manifest/log outputs into Phase 5 evidence publishing for EV-v11-G4 promotion review.
- **Implemented:** `scripts/v11-relay-smoke.sh` and the v1.1 container baseline provide deterministic allowed/forbidden probes; evidence publication remains pending Phase 5 indexing/sign-off.
