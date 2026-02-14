# Phase 2 P2-T9: Layered Test Strategy Matrix + KPI Scaffold

## Purpose
Provide a Phase 2 test-layer matrix and KPI scaffold that stays aligned with existing scaffolds and placeholders. This document captures planning intent and current implementation boundaries.

## Implemented (current state)
The repository currently exposes placeholder test scaffolds via Make targets and scripts, not production test suites. See:
- [`Makefile`](Makefile:1) targets: [`make check-fast`](Makefile:14), [`make check-full`](Makefile:15), [`make test`](Makefile:33).
- Reproducibility scaffolds: [`docs/v0.1/phase2/repro-tools.md`](docs/v0.1/phase2/repro-tools.md:1), [`docs/v0.1/phase2/repro-policy.md`](docs/v0.1/phase2/repro-policy.md:1).

## Layered test strategy matrix (planned)
| Layer | Scope | Ownership | Evidence target | Status |
|---|---|---|---|---|
| L0: Repo hygiene | Format/lint/compile placeholders | Engineering | `make check-fast` logs | Implemented (placeholder) |
| L1: Reproducibility | Dhall + checksum scripts | Engineering | `scripts/dhall-verify.sh`, `scripts/repro-checksums.sh` outputs | Implemented (placeholder) |
| L2: Protocol compatibility | Buf lint/breaking checks | Protocol | Buf outputs and reports | Planned |
| L3: Unit tests | Package-level correctness | Engineering | `go test ./...` evidence | Planned |
| L4: Integration | Multi-package seams and flows | Engineering/QA | Runbooks + evidence bundles | Planned |
| L5: Acceptance | First-contact journeys | QA | Acceptance test charter evidence | Planned |

## Strictness and scenario policy scaffold (P2-T9)

### Crypto and protocol strictness (planned)
- Crypto package tests: require deterministic known-answer vectors and explicit negative-path assertions for malformed inputs.
- Protocol package tests: require wire-compat checks for additive schema evolution and rejection tests for unsupported/invalid envelopes.
- Merge policy: any waived crypto/protocol failure requires explicit owner sign-off and a follow-up remediation item.

### Integration scenarios for LAN and relay fallback (planned)
- Scenario I1: same-LAN direct connectivity success path.
- Scenario I2: LAN discovery failure with relay fallback success path.
- Scenario I3: relay path unavailable with explicit user-visible failure semantics.
- Scenario I4: reconnect after transient disconnect with fallback-frequency KPI capture.

### Failure triage policy for flaky tests (planned)
- Tag flaky failures with scenario/layer labels and quarantine reason.
- Require issue link + owner + expiry date for every quarantine.
- Keep flaky tests non-blocking only for approved quarantine window; expired quarantines become merge-blocking.

References for acceptance criteria and evidence models:
- [`docs/v0.1/phase1/acceptance-test-charter.md`](docs/v0.1/phase1/acceptance-test-charter.md:1)
- [`docs/v0.1/phase1/scope-contract.md`](docs/v0.1/phase1/scope-contract.md:1)

## KPI scaffold (planned)
These KPIs are placeholders; thresholds must be set when the underlying tests exist.

| KPI | Definition | Evidence source | Target (planned) | Status |
|---|---|---|---|---|
| Build determinism | Reproducible artifacts unchanged across runs | `make pipeline` artifact hashes | TBD | Planned |
| Lint/format hygiene | Zero unwaived lint errors | `make check-fast` logs | TBD | Planned |
| Unit test coverage | Percent of statements covered | `go test ./...` coverage report | TBD | Planned |
| Integration stability | Pass rate across integration runs | Runbook evidence bundles | TBD | Planned |
| Acceptance pass rate | First-contact test success rate | Acceptance evidence logs | TBD | Planned |

## Planned vs. implemented separation
Only the placeholder scaffolds named in **Implemented (current state)** are present. All L2–L5 layers and KPI thresholds are forward-looking and must not be treated as active testing coverage.
