# Go/No-Go Record

- Version: v2.0 (v20)
- Decision: **Go** (public-beta) based on completed G0..G10 evidence set.
- Decision date: 2026-02-18
- Evidence set: `docs/v2.0/phase5/p5-evidence-index.md`, `docs/v2.0/phase5/p5-final-evidence-bundle.md`, `artifacts/generated/v20-evidence/*`, `artifacts/generated/v20-podman-scenarios/result-manifest.json`.
- Recommendation rationale: no blocking Sev-High risks remain in mandatory v20 gates.

## Signers

| Signer/Role | Decision | Signature anchor |
|---|---|---|
| Release Council (CTO) | Go | EV-v20-G7-001 |
| Governance Board | Go | EV-v20-G7-002 |
| Security Lead | Go | EV-v20-G8-001 |
| Ops Lead | Go | EV-v20-G5-002 |

## Residual Risks

| Risk ID | Risk | Mitigation | Owner | Status | Evidence |
|---|---|---|---|---|---|
| R20-1 | Late security defects | Continue dedicated hardening and regression cadence; re-run `go test ./...` and `make check-full` on hotfix branches | Security Lead | accepted with monitoring | EV-v20-G2-001, EV-v20-G7-001 |
| R20-2 | Ops readiness gaps | Run podman operator drills on every release candidate; verify manifest and rollback evidence | Ops Lead | accepted with monitoring | EV-v20-G5-001, EV-v20-G5-002 |
| R20-3 | Scope creep near launch | Keep non-critical items in v20+ deferral register and enforce review-only additions | Plan Lead | accepted with monitoring | EV-v20-G10-004 |
| R20-4 | Relay boundary regressions during final hardening | Enforce `TestRelayNoDataRegression` checks and keep pod/rollback drills in loop | Runtime Lead | accepted with monitoring | EV-v20-G8-001 |

## Evidence Links

- `docs/v2.0/phase5/p5-evidence-index.md`
- `docs/v2.0/phase5/p5-final-evidence-bundle.md`
- `docs/v2.0/phase5/p5-gate-signoff.md`
- `docs/v2.0/phase5/p5-as-built-conformance.md`
