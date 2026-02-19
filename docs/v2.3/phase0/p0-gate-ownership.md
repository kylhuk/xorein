# Gate ownership and approvers (Phase 0 P0-T1 ST4)

This RACI-style table identifies the primary owner and the gate approver for each gate (`G0`..`G11`). Ownership is required before we mark the corresponding gate as satisfied or submit evidence entries (e.g., `EV-v23-G0-###`).

| Gate | Primary owner (tasks) | Approver | Typical evidence placeholder(s) | Notes |
| --- | --- | --- | --- | --- |
| `G0` (Scope lock + hardening invariants) | Security/Architecture writer | Project lead | `EV-v23-G0-###` | Owner drafts scope lock/hardening matrix; approver confirms invariants and relay boundary defaults. |
| `G1` (Compatibility/schema checks) | Protocol compatibility lead | API owner | `EV-v23-G1-###` | Schema docs/tests referenced in Phase 1 `pkg/v23` modules. |
| `G2` (Security hardening) | Security engineering | Chief security officer | `EV-v23-G2-###` | Abuse, privacy, integrity, durability docs/test suites cite this gate. |
| `G3` (Reliability/SLO/perf) | Reliability engineer | Ops engineering lead | `EV-v23-G3-###` | SLO scorecards and perf test outputs referenced. |
| `G4` (Archivist operator readiness) | Operator docs owner | Site reliability lead | `EV-v23-G4-###` | Runbooks and drills documented in Phase 3. |
| `G5` (Regression matrix) | QA/regression lead | QA director | `EV-v23-G5-###` | Podman scenario reports feed this gate. |
| `G6` (Release docs and evidence) | Release writer | Release manager | `EV-v23-G6-###` | Release notes + evidence bundle in Phase 4. |
| `G7` (Go/no-go sign-off) | Program manager | Executive sponsor | `EV-v23-G7-###` | Captures final decision and risk sign-offs. |
| `G8` (Relay boundary regression check) | Relay boundary maintainer | Relay network architect | `EV-v23-G8-###` | Regression scenario verifying relay no-history constraint. |
| `G9` (`F23` as-built conformance) | Conformance reporter | Protocol owner | `EV-v23-G9-###` | Report vs v22 spec. |
| `G10` (`F24` seed publication) | Roadmap steward | Governance council | `EV-v23-G10-###` | Seed package + deferral register approvals. |
| `G11` (Architecture coverage audit) | Architecture coverage lead | Chief architect | `EV-v23-G11-###` | Audit artifact plus final approved report. |

The above owners/approvers should be tallied in the evidence index when marking gates; any missing approver must be noted as `BLOCKED:GATE` until assigned. When gates span multiple artifacts, the owner coordinates the combined submission.
