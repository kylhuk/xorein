# v2.0 Deployment Runbook

1. Build and tag artifacts consistent with release policy.
2. Deploy via `podman-compose -f containers/v2.0/docker-compose.yml up -d`.
3. Confirm relay health and continuity via `pkg/v20/hardening` expectations.
4. Run `scripts/v20-podman-scenarios.sh` after deployment to produce deterministic manifest.
5. For rollback, stop the pod (`podman pod rm -f v20-operator`) and redeploy.
