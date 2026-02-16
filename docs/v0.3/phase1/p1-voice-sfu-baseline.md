# v0.3 Phase 1 - P1 Voice & SFU Baseline

> Status: Execution artifact. Voice quality, peer SFU election, and relay SFU mode contracts now map to `pkg/v03` deliverables and scenario evidence.

## Purpose

Define deterministic contracts for RNNoise/Opus ABR/jitter/FEC+DTX, peer-SFU election for 9+ participants, and relay-SFU mode while keeping security semantics explicit.

## Scope Summary

- Signal quality adaptation: RNNoise + Opus ABR + adaptive jitter + FEC/DTX fallback.
- Peer-SFU election and tie-break for overloaded voice sessions (9+ participants).
- Relay SFU mode trigger (`--sfu-enabled=true`) with compatibility boundaries.

## Acceptance Anchors

1. RNNoise and ABR parameters produce deterministic quality levels under noisy and low-bandwidth conditions and document fallback reasons during mode transitions.
2. Peer-SFU election includes deterministic triggers, tie-break rules, and state machine-based handoff signals.<br>
3. Relay SFU mode exposes explicit enable/disable states and fallback plans in single-binary builds.

## Evidence Mapping

| Element | Doc | Code/Test Evidence |
|---|---|---|
| RNNoise + ABR + jitter + FEC/DTX contracts | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| Peer-SFU election | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |
| Relay SFU mode | `docs/v0.3/phase1/p1-voice-sfu-baseline.md` | `pkg/v03/voice/contracts.go`, `pkg/v03/voice/contracts_test.go` |

## Dependencies

- Phase 0 scope/compatibility guardrails (`docs/v0.3/phase0/p0-t1-scope-contract.md`).
- Proto compatibility review (`docs/v0.3/phase0/p0-t2-compatibility-governance-checklist.md`).
