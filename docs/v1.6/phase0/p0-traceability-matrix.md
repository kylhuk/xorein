# Traceability Matrix

| Requirement | Artifact |
|------------|----------|
| Default role lifecycle | `pkg/v16/rbac/rbac.go`, `pkg/v16/rbac/rbac_test.go` |
| ACL allow/deny merge | `pkg/v16/acl/acl.go`, `pkg/v16/acl/acl_test.go` |
| Enforcement checks | `pkg/v16/enforcement/*`, `tests/e2e/v16/enforcement_test.go` |
| Admin UI helpers | `pkg/v16/ui/admin_ui.go` |
| Podman scenarios | `containers/v1.6/docker-compose.yml`, `scripts/v16-rbac-scenarios.sh` |
| Evidence & risk | `docs/v1.6/phase5/*` |
