# SLO Scorecard

| Metric | Target (%) | Observed (%) | Status |
| --- | --- | --- | --- |
| Login success rate | >= 99.5 | 99.72 | Pass
| Call connectivity score | >= 92.0 | 94.5 | Pass
| Relay stability | >= 99.9 | 99.95 | Pass

Evidence: `tests/perf/v20/slo-scorecard_test.go` captures deterministic pass/fail logic for these ratios.
