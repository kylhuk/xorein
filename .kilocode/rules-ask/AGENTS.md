# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- Current repository is documentation-only: substantive content is in `aether-v3.md`; `README.md` is title-only.
- No runnable build/test/lint toolchain is present in this snapshot (`Makefile`, `Taskfile.yml`, Go module, CI workflows, and test/lint configs are absent).
- When users ask for commands from this repo, clarify that `make all` / `task all` in `aether-v3.md` describe intended future workflow, not executable commands here.
- Keep implemented-vs-planned distinction explicit in answers: roadmap checkboxes and pipeline sections are plans, not completed implementation.
- Preserve protocol-first framing from the doc: “the protocol is the product”; the specification is the contract and should be prioritized over client/UI specifics.
- Keep network model statements aligned with the plan: one binary, behavior selected by `--mode=client|relay|bootstrap`, and no special/privileged node class.
- Preserve compatibility constraints exactly: protobuf minor updates are additive-only; breaking changes require new multistream IDs plus downgrade negotiation.
- Keep governance/licensing wording aligned with document claims: AEP process + multi-implementation validation for breaking protocol changes, permissive code license, CC-BY-SA protocol spec.
- Open Decisions section is unresolved by design; do not present those items as decided facts.
