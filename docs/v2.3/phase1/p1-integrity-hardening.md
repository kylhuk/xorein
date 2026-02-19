# P1-T3: Integrity Hardening Log

## ST1 – Harden manifest/head verification
- `pkg/v23/integrity` validates every manifest segment (ID, hash, time order) and recomputes the canonical hash so malformed inputs fail fast.
- Heads require metadata alignment and deterministic SHA-256 signatures derived from membership keys before they are accepted.

## ST2 – Deterministic error mapping
- Each invalid manifest/head scenario returns a typed validation error with a stable code (e.g., `MANIFEST_SEGMENT_TIME_ORDER_INVALID`, `HEAD_SIGNATURE_MISMATCH`).
- Tests (unit and e2e) assert the exact code so operators and automation can rely on repeatable signals.

## Gate mapping (G2)
- G2 (Security hardening) is satisfied by ST1–ST2 because manifest/head validation is hardened against malformed data and exposes deterministic failure codes for automation, fulfilling the integrity hardening requirement.

## Evidence commands
- `go test ./pkg/v23/integrity`
- `go test ./tests/e2e/v23/integrity_manifest_test.go`
- `go test ./tests/e2e/v23/integrity_head_test.go`
