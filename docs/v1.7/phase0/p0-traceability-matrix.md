# Traceability Matrix (Phase 0)

| Requirement | Artifact |
|---|---|
| Signed moderation events (+ deterministic rejection reasons) | `pkg/v17/moderation/moderation.go` + tests |
| Replication ordering and duplicate handling | `pkg/v17/modsync` |
| Append-only audit visibility | `pkg/v17/audit` |
| Official-client enforcement status signals | `pkg/v17/ui/enforcement_ui.go` + `docs/v1.7/phase2/p2-enforcement-ux-contract.md` |
| Adversarial, partition, and perf coverage | `tests/e2e/v17`, `tests/perf/v17` |
| Podman moderation scenarios + relay boundary regression | `scripts/v17-moderation-scenarios.sh`, `containers/v1.7`, `docs/v1.7/phase3/p3-podman-scenarios.md` |
| v18 discovery spec package | `docs/v1.7/phase4/*` |
