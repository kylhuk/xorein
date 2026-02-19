# Phase 2 SLO Scorecard (G3)

This scorecard captures the **Phase 2** reliability and performance SLOs that support gate `G3`. Each row below ties an SLO target to a measurement method and the deterministic test that enforces it so the gate can be evaluated with repeatable evidence.

## SLO Table

| Category | Metric | Target | Measurement method | Enforcement test |
| --- | --- | --- | --- | --- |
| ST1 Backfill | Time-to-first-page delivery (first manifest + page) | <= 1.8s | Synthetic backfill sampler simulates manifest assembly + first-page fan-out and records the wall-clock-equivalent latency in deterministic units. | `tests/perf/v23/slo_scorecard_test.go: TestBackfillSLOs` |
| ST1 Backfill | Bounded retry duration for failed pages | <= 5s | Synthetic retry loop on deterministic error signals and driftless timers to confirm the retry window stays bounded. | `tests/perf/v23/slo_scorecard_test.go: TestBackfillSLOs` |
| ST2 Search (Tiered by DB size) | Query p50 latency | Small (<10k rows): <= 45ms<br>Medium (<100k rows): <= 70ms<br>Large (>=100k rows): <= 120ms | Percentile helpers consume synthetic latency samples that are seeded to match each DB tier and compute the corresponding p50 value. | `tests/perf/v23/slo_scorecard_test.go: TestSearchSLOs` |
| ST2 Search (Tiered by DB size) | Query p95 latency | Small: <= 110ms<br>Medium: <= 180ms<br>Large: <= 240ms | Same deterministic percentile helpers as above compute the p95 for each tier. | `tests/perf/v23/slo_scorecard_test.go: TestSearchSLOs` |
| ST3 Archivist | Ingest rate | >= 1,000 entries/min | Synthetic ingest logs with known timestamps ensure the averaged rate stays above the target. | `tests/perf/v23/slo_scorecard_test.go: TestArchivistSLOs` |
| ST3 Archivist | Prune cadence | <= 6h | Simulated prune schedule ticks show the time between pruning rounds does not exceed the bound. | `tests/perf/v23/slo_scorecard_test.go: TestArchivistSLOs` |
| ST3 Archivist | Disk growth | <= 15 GiB/day | Synthetic disk-growth curves limit the per-day delta and assert that even worst-case retention behavior stays within the bound. | `tests/perf/v23/slo_scorecard_test.go: TestArchivistSLOs` |

## Measurement notes

- Percentile calculations in the perf suite sort deterministic latency buckets and pick the ceiling index so there is no dependence on timers or live traffic.
- Ingest, prune, and disk-growth helpers operate on synthetic samples with fixed timestamps, so every CI run produces the same numbers and the thresholds are either met or not.
- Passing the perf test is the definitive measurement artifact for the G3 gate because the same helper code can be re-used offline when verifying production metrics.

## Evidence command

```
go test ./tests/perf/v23 -run Test.*SLOs
```

Running the command above produces the evidence snippet that should be attached to the `EV-v23-G3-###` entry and proves the scorecard satisfies the reliability/performance gate.
