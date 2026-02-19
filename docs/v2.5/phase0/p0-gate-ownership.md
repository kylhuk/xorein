# Gate Ownership – Phase 0 (G0)

| Gate | Owner | Responsibilities | Evidence | Notes |
| --- | --- | --- | --- | --- |
| `G0` (scope lock) | Product/Spec + Architecture Council | Confirm ST1–ST4 artifacts, approve go/no-go/non-go list, freeze data classes, MIME tiers, and supported size bands. | `EV-v25-G0-001` (scope table), `EV-v25-G0-002` (mime-size freeze), `EV-v25-G0-004` (traceability matrix) | Sign-offs recorded in `docs/templates/roadmap-gate-checklist.md`. |

## Evidence requirements
- Phase 0 cannot close without `EV-v25-G0-001` through `EV-v25-G0-004` placeholders resolved.
- Command hints: record scope verification commands (e.g., `make check-full`) before execution and log outputs for the gate checklist.

## Handoff notes
- Future phases must reference this ownership file when citing `G0` completion and use the frozen artifacts listed here for dependency tracing.
