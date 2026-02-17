# VA-E7 Threshold Registry

This registry lists deterministic thresholds that gate reviewers compare against when executing the perf runbook and deterministic perf suites.

| Metric | Threshold | Action | Notes |
|---|---|---|---|
| Server membership | 1000 members | Pass if deterministic campaign includes 1000 baseline | Verified by `TestIncrementalScaleCampaign` in `tests/perf/v09/deterministic_suite_test.go`. |
| Voice participants | 50 participants | Pass if tiered SFU plan covers 50 with deterministic forwarding checks | Verified by `TestDeterministicVoiceCascade` in `tests/perf/v09/deterministic_suite_test.go`. |
| Latency | <= 200 ms median | Flag regression if median latency crosses this bound | Verified by `TestLatencyThresholdHelper` in `tests/perf/v09/deterministic_suite_test.go`. |
| Incremental expansion | +50 members per iteration | Pass if hierarchy and sharding outputs remain deterministic across increments | Verified by `TestIncrementalScaleCampaign` in `tests/perf/v09/deterministic_suite_test.go`. |
| Security mode gating | Hard switch at 1200 load with deterministic hysteresis behavior | Pass if `NextSecurityMode` transitions match campaign assertions | Verified by `TestIncrementalScaleCampaign` and `pkg/v09/scale/plan.go`. |
