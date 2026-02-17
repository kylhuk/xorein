# Roadmap Optional Polish (v11-v20)

## Purpose
- Standardize gate checklists, sign-off roles, evidence indexing, and deferral governance across `TODO_v11.md` to `TODO_v20.md`.

## Templates
- Gate checklist: `docs/templates/roadmap-gate-checklist.md`
- Sign-off and RACI: `docs/templates/roadmap-signoff-raci.md`
- Evidence index: `docs/templates/roadmap-evidence-index.md`
- Deferral register: `docs/templates/roadmap-deferral-register.md`

## Required per-version artifacts
- `docs/vX.Y/phase0/p0-gate-ownership.md`
- `docs/vX.Y/phase0/p0-traceability-matrix.md`
- `docs/vX.Y/phase5/p5-gate-signoff.md`
- `docs/vX.Y/phase5/p5-evidence-index.md`

## Verification
- Run `./scripts/verify-roadmap-docs.sh`.
- The script fails when required sections, template references, or cross-version handoff references are missing.
