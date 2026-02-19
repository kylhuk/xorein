# F24 proto delta (Phase4 P4-T2 ST1)

The `F24` seed package closes Phase 4 while keeping all wire contracts additive-only. No new protobuf files or fields are required for the seeds described in `docs/v2.3/phase4/f24-backlog-and-spec-seeds.md`.

## Delta summary

- **Proto delta:** _empty_ (no new fields/messages). The current schema already captures blob metadata, device metadata, and hint channels with sufficient extensibility for the required seeds.
- **Rationale:** adding fields would delay `G10` and risk violating the additive-only contract. Seed behaviors stay within existing messages, coupled with operational scripts, spec prose, or future v24 code.
- **Evidence linkage:** completeness of the audit (P0-T2) and acceptance matrix (Phase 4 ST2) is what unlocks `G10` and `G11`. The absence of proto changes is explicitly noted here so reviewers know we considered wire compatibility.
