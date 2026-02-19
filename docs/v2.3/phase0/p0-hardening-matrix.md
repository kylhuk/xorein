# Hardening matrix (Phase 0 P0-T1 ST1/ST2)

This matrix captures the `G0` and `G11` criteria that must be satisfied before v23 can claim scope lock. Each row references the evidence placeholder that will appear in the evidence index once the artifact is ready. Nothing here claims implementation; it records planning expectations.

| Gate | Criterion | Description / targets | Evidence placeholder |
| --- | --- | --- | --- |
| `G0` | Hardening invariant: privacy (no keyword leakage) | Documented coverage labels default, assisted keyword access off, default telemetry limited to metadata counts, no plaintext keyword retention. | `EV-v23-G0-001` (privacy) |
| `G0` | Hardening invariant: relay boundary | Explicit statement and review confirming relays have no long-term history storage/manifests; any temporary caches are flagged for `F24`. | `EV-v23-G0-002` (relay boundary) |
| `G0` | Hardening invariant: quotas/refusal reasons | Quota enforcement (Archivist + relay) maps to deterministic refusal codes and audit logs so operators can diagnose. | `EV-v23-G0-003` (quota) |
| `G0` | Hardening invariant: durability labeling | Replica shortages surface a durability degraded signal in API/ops, with documentation for both UI labels and alerting. | `EV-v23-G0-004` (durability) |
| `G11` | Architecture coverage audit completeness | Every persisted data class (see P0-T2 audit) has writer/storage/retrieval/confidentiality/retention entries; missing or unknown classes are tied to `F24` seeds or listed as deferred. | `EV-v23-G11-001` (audit base) |
| `G11` | Architecture coverage audit approval | Audit (Phase 0) reviewed and accepted by Chief Architect with trace into Phase 4 final audit report; unknown data classes flagged for follow-up. | `EV-v23-G11-002` (audit sign-off) |

Add additional `EV-v23-G0-###` or `EV-v23-G11-###` placeholders as subcriteria surface; reference this matrix when populating the evidence index. The review of this matrix is part of `G0` and `G11` gating discussions.
