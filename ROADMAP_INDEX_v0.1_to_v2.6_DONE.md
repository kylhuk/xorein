# Roadmap Index (v0.1 → v2.6 “DONE”)

This index is an operator/agent-friendly map of the TODO files and what each version is responsible for implementing/specifying. The intent is that executing the TODO chain end-to-end yields a complete, deployable system with **Xorein** (backend) and **harmolyn** (frontend).

## Binaries (expected by the end)
Primary:
- `cmd/xorein` — backend node/runtime (relay/bootstrap/client/archivist/blob provider as configured)
- `cmd/harmolyn` — frontend UI (Gio) attaching to Xorein via local API (v24+)

Common auxiliary binaries/services referenced in earlier TODOs (if present in your repo roadmap):
- `cmd/indexer` — public discovery indexer (v18)
- `cmd/aether-bridge` (or renamed bridge runtime) — bridge bot/runtime (v11)
- push relay service (v11)
- TURN/STUN deployment artifacts (v11)

## TODO files in order

### v0.x → v1.0
- `TODO_v01.md` (v0.1)
- `TODO_v02.md` (v0.2)
- `TODO_v03.md` (v0.3)
- `TODO_v04.md` (v0.4)
- `TODO_v05.md` (v0.5)
- `TODO_v06.md` (v0.6)
- `TODO_v07.md` (v0.7)
- `TODO_v08.md` (v0.8)
- `TODO_v09.md` (v0.9)
- `TODO_v10.md` (v1.0)

### v1.1 → v2.0
- `TODO_v11.md` (v1.1) — bridges + “missing-to-be-real” gap closures (TURN/STUN, mobile wake, multi-device)
- `TODO_v12.md` (v1.2) — identity + backup recovery; specifies Spaces/chat
- `TODO_v13.md` (v1.3) — Spaces + chat baseline; specifies voice
- `TODO_v14.md` (v1.4) — voice baseline; specifies screen share
- `TODO_v15.md` (v1.5) — screen share; specifies RBAC/ACL
- `TODO_v16.md` (v1.6) — RBAC/ACL; specifies moderation/audit
- `TODO_v17.md` (v1.7) — moderation/audit; specifies discovery/indexer
- `TODO_v18.md` (v1.8) — discovery/indexer; specifies connectivity/QoL
- `TODO_v19.md` (v1.9) — connectivity orchestrator + QoL; specifies release hardening
- `TODO_v20.md` (v2.0) — production/public-beta hardening + release conformance; produces v20+ seed package

### v2.1 → v2.6 (v20+ architecture completion track)
These files close the “history/search/persistence + binary split + durable data plane” work that was explicitly deferred to v20+ in the earlier roadmap.

- `TODO_v21.md` (v2.1) — encrypted local timeline persistence + local search baseline (Xorein + harmolyn)
- `TODO_v22.md` (v2.2) — distributed history/backfill + Archivist capability (ciphertext-only) + coverage semantics
- `TODO_v23.md` (v2.3) — history/search hardening + operator readiness + architecture coverage audit; produces F24 seeds
- `TODO_v24.md` (v2.4) — Xorein daemon + harmolyn attach via local API; multi-client attach; produces F25 spec
- `TODO_v25.md` (v2.5) — ciphertext blob store plane for attachments/assets + replication + anti-enumeration; produces F26 spec
- `TODO_v26.md` (v2.6) — terminal full-stack closure (“DONE”): end-to-end regression, ops drills, reproducible builds, final evidence

## Definition of “DONE” (as enforced by v26)
A v2.6 promotion requires:
- all critical journeys pass end-to-end across both binaries
- operator runbooks/drills exist with evidence
- reproducible build + signing + SBOM exist for all shipped binaries
- the terminal deferral register contains **zero core deferrals**
