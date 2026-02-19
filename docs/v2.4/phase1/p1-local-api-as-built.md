# Phase 1 local API as-built narrative

| ST | Summary | Gate | Evidence command |
| --- | --- | --- | --- |
| ST1 | Proto surface and deterministic refusal taxonomy discovered in `proto/xorein_local/v1/local.proto` and supporting docs. | G1 | `buf lint` & `buf breaking` (planned as per TODO_v24 command hints). |
| ST2 | Listener scaffolding enforces Unix domain socket/named pipe only bindings; `ListenerConfig.ValidateLocalBind` returns `RefusalReasonNonLocalBind` on remote transports. | G2 | `go test ./pkg/v24/localapi` (local API scaffolding unit suite). |
| ST3 | Handshake/session middleware issues `SessionToken`, enforces version equality, and rejects invalid tokens with `RefusalReasonInvalidToken`. | G2 | `go test ./pkg/v24/localapi` (session tests). |
| ST4 | Audit records, cursor helpers, and event envelope semantics minimize payload exposure in logs while providing deterministic cursor advancement/state. | G2 | `go test ./tests/e2e/v24/localapi_security_test.go` (security invariants). |

## Deterministic evidence mapping
- Recorded mappings live under `docs/v2.4/phase1/p1-local-api-as-built.md` until the EV rows are published (e.g., `EV-v24-G2-001` for the local API unit tests, `EV-v24-G2-002` for the e2e security invariants).
- Command outputs will be captured in `artifacts/generated/v24-evidence/` with names matching `localapi-tests.txt` once executed.
