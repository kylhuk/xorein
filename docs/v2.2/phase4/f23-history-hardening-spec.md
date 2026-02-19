# F23 History Hardening Spec

The history-hardening expectations described here are now grounded in the Phase 5 evidence bundle. Gates G4 (history controls), G5 (adversarial/privacy abuse), and G6 (Podman replay scenarios) have closing artifacts, so the spec describes as-built controls rather than placeholders.

## Objective

- Reconcile history ingestion, archiving, and search surfaces with the v22 reliability goals.
- Surface deterministic hardening requirements so the handoff to Phase 5 evidence gathering is explicit.

## Scope & Constraints

1. Enumerate data integrity checks and mutability constraints for all historical segments touched since Phase 3.
2. Define remediation playbooks for `missing_backfill`, `stale-history`, and `replay-loop` observations.
3. Lock down UX messaging so coverage gaps and hardening status remain auditable.

## Execution summary

- **G4 – History control verification**: Pass (EV-v22-G4-001). The e2e/perf runs recorded at `artifacts/generated/v22-evidence/go-test-e2e-v22.txt` and `artifacts/generated/v22-evidence/go-test-perf-v22.txt` exercised the backfill protocol, coverage-gap labeling, and deterministic failure modes defined in this spec.
- **G5 – Adversarial/privacy-abuse matrix**: Pass (EV-v22-G5-002). `make check-full` output at `artifacts/generated/v22-evidence/make-check-full.txt` documents quota/retention abuse probes, replay/resilience checks, and the security scan suite; the `trivy filesystem` stage emitted a non-fatal warning that `--security-checks` is deprecated, but the run completed cleanly.
- **G6 – Podman replay scenarios**: Pass (EV-v22-G6-001). `artifacts/generated/v22-evidence/v22-history-scenarios.txt` logs capture offline catch-up, multi-archivist failover, quota-induced refusal, replica healing, and the relay no-long-history-hosting regression proof.

## Evidence index

- `EV-v22-G4-001`: `artifacts/generated/v22-evidence/go-test-e2e-v22.txt` + `artifacts/generated/v22-evidence/go-test-perf-v22.txt` – history backfill/perf suites covering integrity, retrieval, and coverage-gap UX controls.
- `EV-v22-G5-002`: `artifacts/generated/v22-evidence/make-check-full.txt` – adversarial/regression suite and govulncheck/gosec scans (non-fatal trivy warning about the deprecated `--security-checks` flag logged inline).
- `EV-v22-G6-001`: `artifacts/generated/v22-evidence/v22-history-scenarios.txt` – Podman scenario manifest with deterministic pass/fail/resilience results.
