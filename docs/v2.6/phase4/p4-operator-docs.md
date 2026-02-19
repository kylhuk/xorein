# Phase 4 operator documentation (as-built truth)

This artifact tracks the operator documentation portion of the `G7` gate. It states what is planned for the final handoff runbooks and what can already be considered part of the as-built truth by referencing the phase 3 operator deliverables, keeping planning and implementation commentary side by side.

## Planning vs implemented boundary
| Operator topic | Planning note | As-built note/reference | Artifact | Evidence placeholder |
| --- | --- | --- | --- | --- |
| Relay runbook + rollback drills | Plan to publish the consolidated relay operator runbook and drill narrative for `G5` readiness. | The relay planning output already leans on `docs/v2.6/phase3/p3-relay-runbook.md` and `p3-relay-rollback-drill.md`, so plan text simply references those as the just-in-time as-built statements. | `docs/v2.6/phase3/p3-relay-runbook.md` | `EV-v26-G7-301` |
| Archivist/blob operator guidance | Plan to document storage alarms, quota tuning, and recovery notes. | As-built statements cite `docs/v2.6/phase3/p3-archivist-runbook.md` and related drill notes, framing them as the deliverables yet to be packaged in a single G7 doc. | `docs/v2.6/phase3/p3-archivist-runbook.md` | `EV-v26-G7-302` |
| Indexer + push relay + TURN readiness | Plan to combine all aux service readiness guidance into the operator handbook. | The operator doc will point to the aux services runbook currently captured in `docs/v2.6/phase3/p3-aux-services-runbook.md` and note whether TURN is shipped or deferred. | `docs/v2.6/phase3/p3-aux-services-runbook.md` | `EV-v26-G7-303` |
| Evidence index + gate checklist | Plan to include a checklist referencing the gate templates so evidence curators can certify `G7`. | We already know the templates will be used: `docs/templates/roadmap-evidence-index.md` for evidence numbering plus `docs/templates/roadmap-gate-checklist.md` for gating. | `docs/templates/roadmap-evidence-index.md` | `EV-v26-G7-304` |

## Evidence & checklist tooling
- Include the `docs/templates/roadmap-gate-checklist.md` entry for `G7` with operator-specific owners and the `docs/templates/roadmap-signoff-raci.md` RACI detail.
- Evidence entries will follow `EV-v26-G7-###` numbering; placeholders above are intended to be replaced by the real drill outputs once the docs collect them.
- The operator doc build command should record the script invocation (e.g., `make docs-v2.6`) as `EV-v26-G7-305` for traceability.

## Next steps (planning)
- Package the relay, archivist, and aux service runbooks into this phase 4 narrative without claiming further implementation work.
- Confirm all operator scenarios match the drill reports from phase 3 (relay rollback, archivist corruption recovery, aux service failures) so the gate can cite concrete evidence entries.
- Capture the final doc generation output with the evidence index template so this artifact can flip from planning to implemented status.
