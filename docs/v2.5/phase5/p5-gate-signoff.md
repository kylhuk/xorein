# Phase 5 gate signoff checklist (P5-T1)

## Gate readiness

| Item | Gate | Evidence IDs | Status |
| --- | --- | --- | --- |
| Mandatory evidence commands executed and archived | G9 | EV-v25-G9-001, EV-v25-G9-002, EV-v25-G9-003, EV-v25-G9-004, EV-v25-G9-005, EV-v25-G10-004 | pass |
| Relay-boundary oriented evidence and scenario manifest coverage | G10 | EV-v25-G10-001, EV-v25-G10-002, EV-v25-G10-003, EV-v25-G9-003 | pass |
| Risk visibility and blocker tracking updated | G9/G10 | `p5-risk-register.md` | pass |

## Signoff notes
- `G9` has complete artifact capture for the mandatory command set.
- `G10` passes in this run with: `EV-v25-G10-001` (`buf breaking --against '.git#branch=v20'`), `EV-v25-G10-002` (`go test ./tests/perf/v25/...`), and `EV-v25-G10-003` scenario manifest coverage.
