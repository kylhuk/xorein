# Phase 5 As-Built Conformance

v15 implements the F15 screenshare contract baseline from `docs/v1.4/phase4/f15-screenshare-spec.md` with deterministic capture and adaptation contracts in `pkg/v15/capture` and `pkg/v15/screenshare`, and control hint wiring in `pkg/v15/ui`.

| Acceptance criterion (`docs/v1.4/phase4/f15-acceptance-matrix.md`) | v15 evidence | Result |
|---|---|---|
| Frame contract defined | `docs/v1.4/phase4/f15-screenshare-spec.md`, `pkg/v15/capture/capture.go`, `tests/e2e/v15/first_frame_test.go` | pass |
| Proto delta additive | `docs/v1.4/phase4/f15-proto-delta.md`, `artifacts/generated/v15-evidence/buf-breaking.txt` | deferred to v16 spec handoff |
| Adaptation behaviors described | `docs/v1.4/phase4/f15-screenshare-spec.md`, `pkg/v15/screenshare/screenshare.go`, `tests/perf/v15/adaptation_steps_test.go` | pass |
