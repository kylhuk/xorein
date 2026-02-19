# Phase 2 SLO scorecard (P2-T2)

This scorecard captures the deterministic performance/reliability probes required for `G4`. Each row corresponds to one of the ST1–ST4 categories defined in `TODO_v26.md` and maps to EV-v26-G4 evidence identifiers.

| Metric | Category | Baseline | Target | Measurement method | Deterministic failure reason | EV-v26-G4-### |
| --- | --- | --- | --- | --- | --- | --- |
| `startup_reconnect` | ST1 | 3.0s | ≤3.0s | Synthetic stopwatch derived from pre-recorded startup/reconnect telemetry (fixed trace) | Startup/reconnect exceeds 3.0s stability budget | EV-v26-G4-001 |
| `message_send_latency_p95` | ST2 | 120ms | ≤120ms | Deterministic relay-local message path with bounded payload | P95 latency above 120ms | EV-v26-G4-002 |
| `backfill_throughput` | ST3 | 400 ev/sec | ≥400 ev/sec | Precomputed backfill batch throughput simulation with fixed chunk sizes | Throughput drops below 400 events/sec | EV-v26-G4-003 |
| `blob_transfer_stability` | ST4 | 80 MB/sec | ≥80 MB/sec | Deterministic blob transfer harness with fixed 1MiB payloads and resource bounds | Blob transfer throughput or resource bounds violated | EV-v26-G4-004 |

## Test harness
- Each metric is exercised by `tests/perf/v26/perf_v26_scorecard_test.go`; the harness uses fixed values so the test is deterministic and fast, mimicking the measurement methods above.

## Deterministic failure guidance
- When an SLO fails, the test prints the metric name, category, and the reason from the table above, making debugging reproducible without wall-clock variability.
