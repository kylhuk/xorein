# Protocol Constraints Checklist (Phase 1)

## Protocol-first Checklist
1. Single-binary mode (`--mode=client|relay|bootstrap`) is immutable for v0.1 planning and implementation. [`TODO_v01.md:7`](TODO_v01.md:7)
2. Protocol/spec contract remains primary over UI behavior; no UI-driven protocol overrides. [`AGENTS.md:3`](AGENTS.md:3)
3. Control-plane defaults (Noise + QUIC) and media security defaults (ICE/SRTP/SFrame) may only change through formal governance. [`TODO_v01.md:20-27`](TODO_v01.md:20-27)

## Compatibility Checklist (Required)
1. **Additive-only evolution:** minor-band changes are limited to adding fields/messages/enums.
2. **Field reservation/no reuse:** once a field number exists, it cannot be repurposed; removed numbers/names must be marked `reserved`.
3. **Backward compatibility defaults:** no required fields; safe defaults for absent values.
4. **Forward compatibility defaults:** unknown fields are ignored safely; capability fallback is explicit.
5. **Multistream major-version rule:** breaking protocol behavior requires new major IDs plus downgrade pathway documentation. [`TODO_v01.md:128-150`](TODO_v01.md:128-150)
6. **Buf validation expectation:** when `.proto`/`buf` assets are present, run `buf lint`, `buf breaking`, and `buf generate` (if generation is configured); if assets are absent, record explicit SKIP evidence.

## Governance Checklist
1. AEP-style governance process is mandatory for any breaking-change candidate.
2. Breaking-change candidates require validation across at least two independent implementations before closure.
3. Critical-path review artifacts must include this checklist by reference. [`docs/v0.1/phase1/review-template.md:1`](docs/v0.1/phase1/review-template.md:1)
4. Constraints must remain traceable to scope and acceptance artifacts. [`docs/v0.1/phase1/scope-contract.md:1`](docs/v0.1/phase1/scope-contract.md:1)

## Prohibited Change Quick List
- Renumbering existing fields.
- Reusing removed field numbers/names.
- Changing wire types in-place for existing fields.
- Introducing required fields in v0.1 minor evolution.
