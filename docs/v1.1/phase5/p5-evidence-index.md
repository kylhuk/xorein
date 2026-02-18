# Phase 5 · Evidence index (Planning in progress)

See `docs/templates/roadmap-evidence-index.md` for the canonical format. The table below captures current evidence snapshots and remaining TODO-mandated validations.

| EvidenceID | Gate | Command | OutputPath | TimestampUTC | Owner | Result | Checksum | Notes |
|---|---|---|---|---|---|---|---|---|
| EV-v11-G1-001 | G1 | `buf lint` | `pending/EV-v11-G1-001-buf-lint.log` | TBD | Protocol Lead | blocked (not run) | TBD | Required by TODO_v11 mandatory command evidence list. |
| EV-v11-G1-002 | G1 | `buf breaking` | `pending/EV-v11-G1-002-buf-breaking.log` | TBD | Protocol Lead | blocked (not run) | TBD | Required by TODO_v11 mandatory command evidence list. |
| EV-v11-G4-001 | G4 | `go test ./tests/e2e/v11/...` | `artifacts/v11/evidence/EV-v11-G4-001-go-test-e2e-v11.log` | 2026-02-18T07:40:17Z | QA Lead | pass | c31b8a70cd76034526d4193b2dbd1af46a5a64b464032f5ceabfa15a2561c688 | v11 relay boundary e2e suite passed. |
| EV-v11-G4-002 | G4 | `scripts/v11-relay-smoke.sh` | `artifacts/v11/evidence/EV-v11-G4-002-relay-smoke.log` | 2026-02-18T07:40:17Z | QA Lead | pass | 18a194b37150140db6a2ad5ab48197a7f9697bdf8d272f949bd0e13542cf9e3a | Script pass log; see manifest + probe logs below. |
| EV-v11-G4-003 | G4 | `scripts/v11-relay-smoke.sh (manifest)` | `artifacts/generated/v11-relay-smoke/result-manifest.json` | 2026-02-18T07:40:17Z | QA Lead | pass | da3c6dafdb307fd40b2f9e8b0a08101cf779fc2868ea518e21444d85ec1a7196 | Deterministic smoke manifest from latest run. |
| EV-v11-G4-004 | G4 | `scripts/v11-relay-smoke.sh (allowed probe log)` | `artifacts/generated/v11-relay-smoke/allowed-relay-session-metadata.log` | 2026-02-18T07:40:17Z | QA Lead | pass | fa95bdcd524a2faee46981e18dd938d65c89adbfd2196734e734cece66ac8661 | Allowed persistence mode accepted. |
| EV-v11-G4-005 | G4 | `scripts/v11-relay-smoke.sh (forbidden probe log)` | `artifacts/generated/v11-relay-smoke/forbidden-relay-durable-message-body.log` | 2026-02-18T07:40:17Z | QA Lead | pass | 4ba25953b3f9041e5d61a00018ebe20f65fee9226228df6d1a06dc0209c841a3 | Forbidden persistence mode rejected with policy violation. |
| EV-v11-G5-003 | G5 | `./scripts/verify-roadmap-docs.sh` | `artifacts/v11/evidence/EV-v11-G5-003-roadmap-verify.log` | 2026-02-18T07:40:17Z | QA Lead | pass | 5ceece84b9f1b833892a0970a857665cd3140fb189d96ea0ea80e08ab1970fe0 | Roadmap TODO policy verification passed. |
| EV-v11-G5-004 | G5 | `go test ./...` | `pending/EV-v11-G5-004-go-test-all.log` | TBD | QA Lead | blocked (not run) | TBD | Required by TODO_v11 mandatory command evidence list. |
| EV-v11-G5-005 | G5 | `make check-full` | `pending/EV-v11-G5-005-make-check-full.log` | TBD | QA Lead | blocked (not run) | TBD | Required by TODO_v11 mandatory command evidence list. |

## Notes
- When the commands run, update the `TimestampUTC`, `Result`, `Checksum`, and `Notes` columns plus the `OutputPath` to the actual log location, then propagate the EV IDs into this file and `p5-gate-signoff.md`.
- The TODO list for v11 mandatory commands still has open rows (`buf lint`, `buf breaking`, `go test ./...`, `make check-full`); keep those entries blocked until outputs are recorded.
