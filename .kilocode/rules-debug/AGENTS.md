# AGENTS.md

This file provides guidance to agents when working with code in this repository.

- Repository state is documentation-only: `aether-v3.md` is the only substantive source; `README.md` is title-only.
- There is no runnable debug/tooling stack in this snapshot (`Makefile`, `Taskfile.yml`, Go module, CI workflows, lint/test configs are absent).
- Treat pipeline commands documented in `aether-v3.md` (for example `make all` / `task all`) as target-state architecture, not commands you can run here.
- Keep implemented-vs-planned explicit while debugging claims: roadmap checklists are future intent, not evidence of existing behavior.
- Protocol constraints in docs are authoritative: protocol/spec correctness is primary; UI/client behavior is secondary.
- Network assumptions to preserve in analysis: same binary with mode flags (`--mode=client|relay|bootstrap`), no privileged node class.
- Compatibility constraints are strict: protobuf minor versions are additive-only; major changes require new multistream IDs with downgrade negotiation.
- Governance/licensing constraints are part of correctness context: AEP process + multi-implementation validation for breakage, permissive code license, CC-BY-SA spec.
- Open Decisions are intentionally unresolved; do not debug or document them as finalized requirements.
