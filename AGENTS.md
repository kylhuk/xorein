# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- Repository state is documentation-only right now: `aether-v3.md` is the substantive source and `README.md` only contains the project title.
- No runnable project tooling/config is present in this repo snapshot (`Makefile`, `Taskfile.yml`, Go module, CI workflows, linters, and test configs are absent).
- Treat commands in `aether-v3.md` (for example `make all` / `task all`) as planned architecture, not executable instructions for this checkout.
- Keep a strict implemented-vs-planned distinction: roadmap checklists and pipeline sections in `aether-v3.md` describe intended future work.
- Protocol-first constraint: documentation states “the protocol is the product” and the specification is the contract; UI/client details are secondary.
- Network model assumption in docs: no special nodes, single binary with mode flags (`--mode=client|relay|bootstrap`).
- Compatibility rules documented here are strong constraints: protobuf minor changes are additive-only; major protocol changes require new multistream IDs and downgrade negotiation.
- Governance rules in the plan are explicit: breaking changes require an AEP flow and multi-implementation validation before finalization.
- Keep licensing language aligned with the document: code permissive (MIT-like) and protocol specification CC-BY-SA.
- Open Decisions are unresolved by design; do not present them as settled facts in new docs or generated guidance.
