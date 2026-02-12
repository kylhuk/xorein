# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- This repository snapshot is documentation-only: `aether-v3.md` is the substantive artifact; `README.md` is title-only.
- No runnable architecture/tooling files exist here (`Makefile`, `Taskfile.yml`, Go module, CI workflows, lint/test configs are absent).
- Architectural commands in `aether-v3.md` (for example `make all` / `task all`) must be treated as target-state design, not executable repo instructions.
- Keep implemented-vs-planned boundaries explicit in design output: roadmap checklists and pipeline sections describe future work.
- Preserve the document’s central architectural axiom: protocol/spec is primary (“the protocol is the product”); UI clients are implementations.
- Keep network architecture statements aligned with docs: one binary with mode flags (`--mode=client|relay|bootstrap`), capability differences not privileged node classes.
- Preserve compatibility architecture constraints: protobuf minor evolution is additive-only; major protocol shifts require new multistream IDs and downgrade negotiation.
- Preserve governance architecture constraints: breaking changes follow AEP process and require validation by at least two independent implementations.
- Keep licensing architecture language aligned: permissive code licensing, CC-BY-SA protocol spec.
- Open Decisions in `aether-v3.md` are unresolved and must stay presented as open choices, not finalized architecture.
