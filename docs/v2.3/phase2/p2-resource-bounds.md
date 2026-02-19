# P2-T2 Resource Bounds (bounded resource controls)

## Contract
- Implemented in `pkg/v23/limits` so every history/search actor can instantiate deterministic budgets per request.
- Request guards are scoped to `ScopeBackfillVerification` and `ScopeIndexing`; `RequestBudget` enforces per-request CPU/IO ceilings and returns a typed `Refusal` with stable `ReasonCode`, `Scope`, and remediation text when a limit is reached.
- `DiskGuard` watches disk growth, surfaces an `AlarmState`, and refuses writes once `DiskHardLimitBytes` is reached.

## CPU + IO budgets
| Scope | CPU ceiling | IO ceiling |
| --- | --- | --- |
| Backfill verification | `250ms` per request (`BackfillVerificationBudget.CPULimit`) | `16 MiB` (`BackfillVerificationBudget.IOLimitBytes`) |
| Indexing maintenance | `200ms` per request (`IndexingBudget.CPULimit`) | `12 MiB` (`IndexingBudget.IOLimitBytes`) |

- Instantiate a budget with `NewRequestBudget(scope, BackfillVerificationBudget)` or `NewRequestBudget(scope, IndexingBudget)` for the corresponding flow.
- Budget accounting is request-scoped; `ConsumeCPU` and `ConsumeIO` return a `Refusal` once the respective ceiling is exceeded so callers can surface deterministic retry/error handling.

## Refusal taxonomy
| Reason code | Scope | Trigger | Remediation |
| --- | --- | --- | --- |
| `CPU_LIMIT_EXCEEDED` | `ScopeBackfillVerification` / `ScopeIndexing` | CPU nanoseconds recorded by `RequestBudget` surpass the configured ceiling | Reduce concurrency or request a larger verification budget before retrying. |
| `IO_LIMIT_EXCEEDED` | `ScopeBackfillVerification` / `ScopeIndexing` | IO bytes recorded by `RequestBudget` exceed the configured cap | Break the workload into smaller IO batches or re-run once IO pressure subsides. |
| `DISK_HARD_LIMIT` | `ScopeDiskGrowth` | `DiskGuard.AddUsage` pushes usage to `DiskHardLimitBytes` (220 GiB by default) | Clean up or archive data to return below the disk hard limit before writing more. |

## Disk growth guard
- Alarm threshold: `DiskAlarmThresholdBytes` (200 GiB). Any `AddUsage` call that pushes usage ≥ this threshold returns `AlarmState{Alarmed: true}` so operators can fire alarms before hitting the hard limit.
- Hard limit: `DiskHardLimitBytes` (220 GiB). `AddUsage` clamps usage at the hard limit and returns the deterministic `DISK_HARD_LIMIT` `Refusal`; every subsequent `AddUsage` call remains refused (fail-closed).  Guard creation enforces `alarmThreshold < hardLimit` and validates initial usage.

## Gate mapping
- **G3 (Reliability/performance SLO)**: Bounded CPU/IO + disk growth controls must be exercised and documented for archivist/backfill flows.
    - Evidence command: `go test ./pkg/v23/limits/...` (record the output in e.g. `artifacts/generated/v23-evidence/go-test-limits.txt` and attach `EV-v23-G3-001`).
    - The small test suite here asserts normal accounting, alarm triggering, and deterministic refusals for every `ReasonCode` so the gate can reference concrete command output.
