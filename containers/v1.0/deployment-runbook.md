# v1.0 Deployment Runbook

## Purpose
Document the deterministic deployment references for VA-N1, VA-N2, VA-N4, VA-B7, and VA-R* evidence.

## Steps
1. **Provision bootstrap nodes** (VA-N1): follow topology map in `docs/v1.0/phase7/p7-bootstrap-infra.md` and adopt `containers/v1.0/docker-compose.yml` replica counts.
2. **Monitor and operate** (VA-N2): instrument the `relay-node` health metrics and record events inside `containers/v1.0/docker-compose.yml` via `com.release.anchor` labels.
3. **Perform relay handover drills** (VA-N4): refer to `docs/v1.0/phase8/p8-relay-program.md` operator handover section, capture logs in `containers/v1.0/deployment-runbook.md`, and confirm reliability scoring from `pkg/v10/relay/policy.go`.
4. **Evidence gating** (VA-B7/VA-R*): append normalized artifacts from `pkg/v10/repro/verification.go` and the release manifest `releases/VA-B1-release-manifest.md` when archiving run results.

## Continuity
- Keep the `relay-node` service running while the `bootstrap-node` service performs rolling updates; `containers/v1.0/deployment-runbook.md` notes stuck nodes and provides deterministic restart instructions.
