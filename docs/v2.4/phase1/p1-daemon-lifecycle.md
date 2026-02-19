# Daemon lifecycle hardening (P1-T2)

## ST1 – deterministic state + lock surface (G2)
- Added an explicit `State` graph plus `ValidateTransition` to refuse any unexpected lifecycle move.
- Introduced `LockManager`/`RefusalError` so only one owner may acquire the daemon lock and invalid transitions surface deterministic reason strings.
- **Evidence commands:** `go test ./pkg/v24/daemon -run TestLockManager`.

## ST2 – crash-safe bounded restart (G2)
- Supervisor executes start attempts with exponential backoff and caps retries, returning `SupervisorRefusal` with a `retry-after` when the cap is hit.
- Ensures crash loops become deterministic refusals instead of unbounded restarts.
- **Evidence commands:** `go test ./pkg/v24/daemon -run TestSupervisorRetries|TestSupervisorRefusal`.

## ST3 – doctor checks for operators (G3)
- Doctor reports socket permissions, lock ownership, stale socket status, daemon version, health probe summary/state, last crash marker, and next recommended action so operators can repair broken daemon artifacts and correlate incidents with version/health metadata.
- NextAction logic now combines those probes with the socket/lock picture so `xorein doctor` deterministically instructs operators when to remove stale sockets, restart the daemon, or keep monitoring a healthy owner.
- Doctor report is structured and pluggable so `xorein doctor` can feed consistent outputs to logging/garage tooling.
- **Evidence commands:** `go test ./pkg/v24/daemon/doctor`.
