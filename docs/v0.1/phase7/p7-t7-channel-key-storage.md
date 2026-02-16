# Phase 7 - Task 7: Channel Key Storage Hardening

## Objective
Document the memory, logging, and crash-handling constraints for channel MLS/Sender Keys artifacts so Phase 7 pipelines persist only what is required and surface residual risks explicitly.

## Scope
- MLS sender secrets and per-participant signer keys managed by [`Bootstrapper`](../../pkg/phase7/bootstrap.go).
- Pipeline sender secrets used in [`Pipeline.Send`](../../pkg/phase7/pipeline.go) / [`Pipeline.Receive`](../../pkg/phase7/pipeline.go).
- SFrame key derivation inputs shared with Phase 8 (`SenderKey` clones only).

## In-Memory Handling Constraints
1. **Clone on handoff** – All key material passed across goroutines must be cloned to prevent aliasing. Existing helpers (`cloneKey()` in [`bootstrap.go`](../../pkg/phase7/bootstrap.go)) remain the canonical primitive; new call sites must avoid `append` reuse.
2. **Zeroization policy** – Callers that finish using `[32]byte` secrets must zero the byte slice before release. Go runtime does not guarantee zeroing; add `for i := range secret { secret[i] = 0 }` before dropping references. TODO hook: add helper in Phase 7 once schedule allows (tracked under Residual Risk RR-P7-T7-1).
3. **Mutex-protected state** – `Bootstrapper` and `SFrameKeyDistributor` already gate key maps via mutexes. Any in-memory cache additions must follow the same locking to avoid races that prolong the life of stale keys.
4. **Rotation snapshots** – `Bootstrapper.Rotate` keeps at most two legacy sender keys (current + previous) to satisfy compatibility bridge semantics. Never increase retention depth without explicit risk review; tests exist in [`bootstrap_test.go`](../../pkg/phase7/bootstrap_test.go).

## At-Rest and Serialization Constraints
1. **No disk persistence** – Channel-level secrets remain process memory only. Persisted artifacts (profiles, manifests, SQLCipher stores) must never include MLS/Sender Keys; lint/test hooks should fail if new fields appear in protobuf structs.
2. **Diagnostics boundaries** – Structured logs and diagnostic exports must redact key material and derived identifiers. [`pkg/ui/shell.go`](../../pkg/ui/shell.go) redaction policy already enforces diagnostic anonymization; reuse its helpers for any Phase 7 debug output.
3. **Crash reports** – Fatal crash handlers must redact stacks to avoid dumping byte slices. Until a structured crash reporter exists, disable panic stack uploads for release builds that include Phase 7.

## Logging and Telemetry Constraints
1. **No logging of byte slices** – Functions touching `[]byte` secrets cannot include `%x` / `%v` logs. Instead, emit rotation counters or epoch numbers.
2. **Reason codes only** – When errors bubble up (e.g., `ErrInvalidSignature`, `ErrDuplicateMessage`), propagate typed errors without embedding participant IDs or secrets.
3. **Telemetry opt-in** – Any future metrics for MLS operations must be opt-in and aggregate-only (counts, durations), never raw identifiers.

## Residual Risks & Follow-Up Tasks
| ID | Risk | Impact | Mitigation / Next Step |
|----|------|--------|------------------------|
| RR-P7-T7-1 | No centralized zeroization helper; callers may forget to wipe buffers. | Secrets remain in heap pages longer than necessary. | Add `Zeroize(b []byte)` helper in Phase 7 utilities and update pipeline/phase8 call sites (new TODO item). |
| RR-P7-T7-2 | Go runtime copies of slices may survive beyond intended scope. | Secrets could linger in stack copies after goroutine exit. | Keep secrets scoped to shortest-lived goroutine; avoid logging/panicking with slices. Review again during Go 1.24 upgrade. |
| RR-P7-T7-3 | Crash dumps may include heap snapshots if panic reporting is enabled. | Potential leakage during support escalations. | Block crash upload features until redaction tooling lands (tie to Phase 11 diagnostics tasks). |

## Downstream Updates
- Update `TODO_v01.md` Track R-A entry to reflect constraints + residual risks.
- File TODO for implementing zeroization helper (dependency for future hardening pass).
- Coordinate with Phase 10 shell team so diagnostics exporters never surface MLS state (already satisfied via `DiagnosticRedactionPolicy`, but reaffirmed here).

## Evidence
- Code references: [`pkg/phase7/bootstrap.go`](../../pkg/phase7/bootstrap.go), [`pkg/phase7/pipeline.go`](../../pkg/phase7/pipeline.go), [`pkg/phase8/sframe.go`](../../pkg/phase8/sframe.go).
- Tests guarding rotation depth and pipeline behavior: [`pkg/phase7/bootstrap_test.go`](../../pkg/phase7/bootstrap_test.go), [`pkg/phase7/pipeline_test.go`](../../pkg/phase7/pipeline_test.go).
