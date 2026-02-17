# Phase 8 - Community Relay Program Launch Readiness

Phase 8 documents the community relay program policies, reliability scoring, onboarding, and incremental governance required for V10-G8.

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| VA-N4 | Relay policy and onboarding controls | `pkg/v10/relay/policy.go` scoring functions + onboarding steps listed here |
| VA-N5 | Abuse response and non-token incentive posture | `pkg/v10/relay/policy.go` abuse classification map |
| VA-N6 | Operator handover continuity evidence | `containers/v1.0/deployment-runbook.md` handover checklist + `docs/v1.0/phase8/p8-relay-program.md` notes |

## Planned vs Implemented
- **Planned:** Document the relay policy, onboarding checklist, reliability scoring, and non-token incentive approach before launching the program.
- **Implemented:** The `pkg/v10/relay` helpers codify scoring and abuse handling while this doc covers the onboarding + continuity runbook. Release and container docs cite the same anchors in the go/no-go dossier.

## Notes
- Running `pkg/v10/scenario` validates these relay controls as part of the `v1.0 genesis` summary.
