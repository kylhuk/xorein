# P10-T6 First-Launch Shell Flow Validation

Date: 2026-02-15

## Dependency feasibility

`P10-T2`, `P10-T3`, `P10-T4`, and `P10-T5` are marked complete in [`TODO_v01.md`](TODO_v01.md).

## Validation scope

- Guided first-launch shell flow from blank state
- Interrupted-flow recovery behavior (navigation interruption + server switch)
- Actionable error-state guards across shell and voice flows

## Evidence anchors

- Guided flow: [`TestFirstLaunchGuidedFlowFromBlankState()`](pkg/ui/shell_test.go:1319)
- Interrupted recovery: [`TestFirstLaunchInterruptedFlowRecovery()`](pkg/ui/shell_test.go:1382)
- Error-state/actionable guard coverage: [`TestShellGuardsAndStateErrors()`](pkg/ui/shell_test.go:95), [`TestVoiceSessionGuards()`](pkg/ui/shell_test.go:1121), [`TestVoiceControlAndDeviceValidationGuards()`](pkg/ui/shell_test.go:1267)

## Podman verification command

`podman run --rm --userns=keep-id -v "$PWD":/workspace:Z -w /workspace docker.io/library/golang:1.23.4 bash -lc 'export PATH=/usr/local/go/bin:$PATH; gofmt -w pkg/ui/shell_test.go; go test ./pkg/ui; go test ./...; go build ./...'`

Observed output:

```text
ok  	github.com/aether/code_aether/pkg/ui	0.002s
go: downloading google.golang.org/protobuf v1.36.1
?   	github.com/aether/code_aether/cmd/aether	[no test files]
?   	github.com/aether/code_aether/gen/go/proto	[no test files]
?   	github.com/aether/code_aether/pkg/app	[no test files]
?   	github.com/aether/code_aether/pkg/crypto	[no test files]
?   	github.com/aether/code_aether/pkg/network	[no test files]
ok  	github.com/aether/code_aether/pkg/phase4	(cached)
ok  	github.com/aether/code_aether/pkg/phase5	(cached)
ok  	github.com/aether/code_aether/pkg/phase6	(cached)
ok  	github.com/aether/code_aether/pkg/phase7	(cached)
?   	github.com/aether/code_aether/pkg/storage	[no test files]
ok  	github.com/aether/code_aether/pkg/phase8	(cached)
ok  	github.com/aether/code_aether/pkg/phase9	(cached)
ok  	github.com/aether/code_aether/pkg/protocol	(cached)
ok  	github.com/aether/code_aether/pkg/ui	(cached)
```
