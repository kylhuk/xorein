# Five-Minute First Contact Acceptance Test Charter

## Preconditions and Repeatability
- Environment: Two clean Linux hosts (local VMs) connected to same LAN, Podman installed for deterministic builds and runs.
- Device requirements: CPU with virtualization, TLS-enabled network stack, time-synced nodes.
- Network prerequisites: LAN broadcast/mDNS enabled for discovery, relay fallback over UDP.
- Data reset: before each run, clear local caches/db/logs for both peers using a documented reset command set in the runbook. If script automation is later added (for example `scripts/reset-first-contact.sh`), update this charter with exact path and command evidence.

## Test Flow
1. Launch `aether --mode=client` on Host A; record identity creation time and success tones.
2. Launch `aether --mode=client` on Host B; link to Host A via server manifest deeplink. Capture log segments for handshake (`protocol-log:identity, join`).
3. Exchange MLS-protected chat message from A to B; confirm ack and chat history entry.
4. Initiate voice mesh session with both peers; verify connection stats (packet loss <1%, latency <100ms). Capture `webrtc-state` output.
5. Engage relay client in `--mode=relay` on Host A while Host B switches to the relay path; record connectivity reason codes and failover timestamp.

## Expected Outcomes
- Identity created and verified within 60s.
- Server join completed and chat message delivered in sequence with MLS encryption flag.
- Voice session stable for at least 2 minutes, no fatal dropouts.
- Relay fallback engages when direct path is delayed beyond 30s (traceable via reason code logs).

## Evidence Capture Format
- Capture `journalctl` snippets for identity/server/voice subsystems with timestamps and reason codes (format: `[component][timestamp][reason] message`).
- Store artifacts in `evidence/phase1/` with metadata `run-id`, `identity-hash`, `network-profile`.
- Repeatability checklist: run data reset script, confirm network prerequisites, start clients in order, archive logs.

## Acceptance Criteria Mapping
| Criterion | Artifact | Notes |
|---|---|---|
| First-launch usability | Charter steps + evidence format | P1-T4 ensures deterministic flows. |
| Device/network prerequisites | Preconditions section | Required before each execution batch. |
| Reset and evidence method | Repeatability checklist + data reset note | Required for clean-state repeatability. |

## Repeatability Checklist (Run-Level)
- [ ] Confirm two clean hosts, synchronized time, and LAN prerequisites.
- [ ] Execute documented reset commands on both hosts before each run.
- [ ] Start Host A and Host B in the same order for every run.
- [ ] Record timestamps for identity create, join, first message, voice join, relay fallback.
- [ ] Capture logs in the defined evidence format and store under `evidence/phase1/`.
