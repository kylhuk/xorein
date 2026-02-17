# Roadmap Gate Sign-off and RACI Template

Use this template to define gate ownership in `docs/vX.Y/phase0/p0-gate-ownership.md` and approvals in `docs/vX.Y/phase5/p5-gate-signoff.md`.

## Core roles
- `Plan Lead`: roadmap scope and dependency governance.
- `Protocol Lead`: wire/protobuf compatibility and protocol contracts.
- `Runtime Lead`: backend/relay implementation and reliability.
- `Client Lead`: Gio UX and client behavior.
- `QA Lead`: test coverage and evidence quality.
- `Security Lead`: security review and vulnerability closure.
- `Ops Lead`: Podman/container/runbook/rollback readiness.
- `Release Authority`: final promotion decision.

## RACI matrix

| Gate group | Responsible | Accountable | Consulted | Informed |
|---|---|---|---|---|
| Scope and dependencies (G0) | Plan Lead | Plan Lead | Protocol Lead, QA Lead | Release Authority |
| Compatibility and schema (G1) | Protocol Lead | Protocol Lead | Runtime Lead, QA Lead | Release Authority |
| Runtime/client implementation (G2-G3) | Runtime Lead, Client Lead | Runtime Lead | Protocol Lead, QA Lead | Release Authority |
| Validation and performance (G4-G5) | QA Lead | QA Lead | Runtime Lead, Client Lead, Ops Lead | Release Authority |
| Docs/evidence and conformance (G6-G9+) | Plan Lead, QA Lead | Plan Lead | Protocol Lead, Runtime Lead, Client Lead, Ops Lead, Security Lead | Release Authority |

## Required sign-off record
- Every gate row must include:
  - owner role,
  - approver role(s),
  - approver name/identifier,
  - timestamp (UTC),
  - linked evidence IDs.
