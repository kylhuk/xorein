# v23 Phase 4 Release Notes — History/Search Plane

## Summary
- Documents the Archivist operator contract (roles, quotas, refusal labels, durability states) to satisfy gate `G6` and prep future runbook/monitoring handoffs.
- Communicates history availability, coverage labels, and backfill behaviors to users so search responses remain interpretable and honest.
- Captures the security/privacy metadata model, default guarantees, and explicit non-guarantees, linking back to the architecture coverage audit (`G11`) and the pending `F24` seeds that resolve any residual persistence gaps (`G10`).

## Gate mapping
- **G6**: The release docs referenced here complete the P4-T1 artifact set required for gate sign-off.
- **G10**: References to non-guaranteed persistence classes are flagged for `F24` seeds or deferrals so the gate remains additive-only.
- **G11**: Aligns with the architecture coverage audit by cross-referencing every metadata class mentioned in the operator/user/security narratives.

## Known constraints
- These notes describe the planned documentation surface. Behavioral implementation and runbook automation remain gated until G6 certification is earned.
- No assisted-search leakage budgets are proposed; any such feature must be opt-in and scoped to a future rollout.

## Verification
- Docs-only verification (tests N/A; writing/peer review is the verification step for these artifacts).
