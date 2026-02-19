# F24 deferral register (Phase4 P4-T2 ST3)

This register captures the work intentionally deferred beyond the planning scope so reviewers can trace the `G10` seed catalog to future releases. Each row identifies the responsible role, the rationale, and the next version where the item will be reevaluated.

| ID | Deferred item | Rationale | Owner role | Revisit target version | Gate impact |
| --- | --- | --- | --- | --- | --- |
| `D24-01` Desktop background mode daemon for `harmolyn` ↔ `xorein` | Optional daemon introduces local persistence (service metadata + refresh tokens) that risks violating the relay boundary in `G8`. Delay until operators can prove the service does not persist beyond encrypted OS service metadata and revisit after v24 baseline workloads. | Architecture lead | `v25` | `G8`, `G10`, `G11` |
| `D24-02` Assisted search runtime hints | The research seed (`F24-C`) only provides a ledger/spec; the actual opt-in assisted search persistence and privacy monitoring will be implemented once upstream research (encryption, SSE/PIR evaluation) is complete. Deferring prevents claiming any unknown persistence is covered before it exists. | Privacy product manager | `v25` | `G10`, `G11` |
| `D24-03` Multi-provider encrypted blobstore pinning | Planning for `F24-A` currently lists the capability, but bundling provider-specific pinning (IPFS-style) introduces new operational runbooks that should land after the `F24` spec proves out. Revisit when `v25` drive-level automation for multi-provider alignment is ready. | Storage reliability lead | `v25` | `G10`, `G11` |
