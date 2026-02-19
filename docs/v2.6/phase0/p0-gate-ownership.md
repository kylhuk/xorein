# Phase 0 Gate Ownership (Planning)

This planning-only note defines who owns the `G0` Scope Lock and terminal criteria plus the surrounding approvals. It follows the RACI schema from `docs/templates/roadmap-signoff-raci.md` so that downstream phases can update their tables consistently.

## `G0` scope lock and terminal freeze
- **Gate name:** `G0` Scope Lock ➜ Terminalization readiness.
- **Objective:** Align `F26` acceptance inputs, traceability matrix, deferral policy, and gate sign-off artifacts before advancing to the security and regression phases.
- **Acceptance:** See `docs/v2.6/phase0/p0-scope-lock.md`; evidence placeholder `EV-v26-G0-014` records the assembled RACI package.

## RACI table (planning placeholder)
| Activity | Responsible | Accountable | Consulted | Informed | Evidence placeholder |
| --- | --- | --- | --- | --- | --- |
| Scope freeze & acceptance import (`p0-scope-lock`) | Release architect | Program manager | Proto review board, Docs lead | Operations, QA | `EV-v26-G0-015` |
| Traceability matrix maintenance (`p0-traceability-matrix`) | Docs lead | Program manager | Release architect, QA | All phase owners | `EV-v26-G0-016` |
| Terminal deferral policy validation (`p0-terminal-deferral-policy`) | Risk owner | Security lead | Program manager, QA | Stakeholders | `EV-v26-G0-017` |
| Gate ownership + sign-off planning (`p0-gate-ownership`) | Program manager | Director of release | Ops readiness, Governance | Release council | `EV-v26-G0-018` |

## Gate approvals
- `Responsible` individuals prepare the artifacts above.
- `Accountable` parties (usually at least one gate approver) sign the `G0` readiness memo once evidence placeholders are populated.
- `Consulted` parties respond with input into the sign-off comments before `G0` passes; their review is captured in the final sign-off sheet produced in Phase 5.
- `Informed` parties are updated once `G0` closes so they can retarget downstream work.

Each row above is a planning placeholder; the actual names and evidence entries will be filled in during execution before `G0` is called Ready.
