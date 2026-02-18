# v1.2 Phase 2 - UX Contract for Identity and Recovery

## Status
UX state contract is implemented in `pkg/ui/shell.go` with a v1.2 adapter layer in `pkg/v12/ui/flow.go`.

## New user onboarding contract
1. Enter `identity_setup` route.
2. Display mandatory warning:
   - "No host or relay can reset your password. Keep a local backup (BackupID + BackupPassword) to recover."
3. Require explicit acknowledgement before identity creation.

## Recovery contract
- Restore flow records deterministic reason taxonomy fields:
  - identity: `identity-mismatch`, `identity-duplicate`
  - backup: `backup-password`, `backup-corrupt`, `backup-outdated`
- Successful restore marks state as completed and re-enables guarded navigation paths.

## Determinism requirements
- Unknown reason labels are rejected with deterministic validation errors.
- Blank identity setup remains blocked by `ErrIdentitySetupRequired`.

## Planned vs implemented
- Contract is implemented for shell state and v1.2 adapter state transitions, and both are tested.
- Full rendered Gio screens remain future UI rendering scope.
