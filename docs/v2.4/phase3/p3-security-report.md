# v2.4 Phase 3 Security Report

## ST1 - Unauthorized Access
- **Focus:** Listener ownership checks and token validation block any actor with insufficient local privileges.
- **Tests:** `tests/e2e/v24/localapi_security_auth_test.go`, `tests/e2e/v24/localapi_security_test.go`
- **Evidence commands:** `go test ./pkg/v24/localapi`, `go test ./tests/e2e/v24/localapi_security_auth_test.go`

## ST2 - Replay and Downgrade
- **Focus:** Handshake middleware now tracks handshake nonces and enforces a configurable minimum version so downgrades or reuse attempts are deterministically refused.
- **Tests:** `tests/e2e/v24/localapi_security_replay_test.go`
- **Evidence commands:** `go test ./pkg/v24/localapi`, `go test ./tests/e2e/v24/localapi_security_replay_test.go`

## ST3 - Injection and Fuzzing
- **Focus:** Frame validation enforces strict header length and payload bounds, rejecting malformed or oversized frames before processing.
- **Tests:** `tests/e2e/v24/localapi_security_injection_test.go`
- **Fuzz target:** `tests/fuzz/v24/validate_frame_fuzz_test.go` (no panics, catches bounds issues)
- **Evidence commands:** `go test ./tests/e2e/v24/localapi_security_injection_test.go`, `go test ./tests/fuzz/v24`
