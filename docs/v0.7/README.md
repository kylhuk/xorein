# v0.7 Demo Guide

This README keeps the runnable story for v0.7 in one place so operators and QA partners can exercise the new store-forward, history sync, push, and search helpers without hunting through phase docs.

## Demo checklist
1. Start infra (bootstrap, relay, push-relay; optional test harness profile is available):
   ```sh
   docker compose -f containers/v0.7/docker-compose.yml up
   ```
2. Build the binaries once for local client simulation:
   ```sh
   go build ./cmd/aether ./cmd/push-relay
   ```
3. Start two client profiles in separate shells:
   ```sh
   ./aether --mode=client --profile=testA
   ./aether --mode=client --profile=testB
   ```
4. Run relay and push-relay locally when you want direct command output in one terminal (instead of compose logs):
   ```sh
   ./aether --mode=relay --profile default
   ./push-relay --profile default
   ```
5. Validate deterministic behavior:
   - `E2E-SF-01` covers offline store-forward delivery after reconnect.
   - `E2E-HS-01` covers history sync and Merkle verification.
   - `E2E-SR-01` covers scoped search + required filters.
   - `E2E-PR-01` covers push relay mocked delivery + notification pipeline.
   - `E2E-EP-01` covers mode-epoch locked-history behavior.

## Single-command e2e run
The entire verification suite can be executed with one command that touches every v0.7 surface:
```sh
go test ./tests/e2e/v07 -run TestE2EScenarios
```
That command builds the demo binaries implicitly, exercises the push relay, history sync, search, and archivist layers, and emits repeatable success/failure state for release evidence.
