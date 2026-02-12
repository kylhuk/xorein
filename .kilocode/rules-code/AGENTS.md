# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- This checkout is documentation-only: the only substantive file is `aether-v3.md`; `README.md` is title-only.
- No runnable project tooling/config exists here (`Makefile`, `Taskfile.yml`, Go module, CI workflows, linters, and test configs are absent).
- Treat `make all` / `task all` in `aether-v3.md` as target-state architecture, not executable commands in this snapshot.
- Keep implemented-vs-planned explicit: roadmap checklists describe intended future work, not completed code.
- Follow protocol-first priorities from the plan: protocol/spec correctness over UI/client specifics.
- Keep network assumptions aligned with docs: single binary, mode flags (`--mode=client|relay|bootstrap`), no privileged nodes.
- Preserve compatibility constraints from the plan (additive-only protobuf minor changes; major changes use new multistream IDs with downgrade negotiation).
- Keep governance/licensing statements aligned: AEP flow + multi-implementation validation for breaks; permissive code license and CC-BY-SA spec.
- Open Decisions remain unresolved and must not be rewritten as settled requirements.
