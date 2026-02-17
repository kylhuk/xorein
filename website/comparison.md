# Launch Surface Comparison

This comparison surface ensures the claims differentiate Harmolyn/xorein from legacy vendors using deterministic evidence anchors (VA-W2).

| Dimension | Claims | Evidence |
|---|---|---|
| Naming governance | Harmolyn (client) vs xorein (protocol) with legacy Aether traceability | `docs/v1.0/phase3/p3-spec-publication.md` naming section + `pkg/v10/governance/naming.go` alias map |
| Launch readiness | Bootstrap 10+ nodes + relay program | `docs/v1.0/phase7/p7-bootstrap-infra.md`, `docs/v1.0/phase8/p8-relay-program.md`, `containers/v1.0/docker-compose.yml` |
| Reproducibility | Deterministic build + CLI witness | `pkg/v10/repro/verification.go`, `pkg/v10/scenario` loop |
| Security audit readiness | Threat model + decision-closure governance | `docs/v1.0/phase2/p2-security-audit.md`, `open_decisions.md` closure register |
