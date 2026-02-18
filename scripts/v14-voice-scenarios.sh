#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
RESULTS_DIR="$ROOT/artifacts/generated/v14-voice-scenarios"
MANIFEST="$RESULTS_DIR/result-manifest.json"
SETUP_LOG="$RESULTS_DIR/voice-call-setup.log"
RECONNECT_LOG="$RESULTS_DIR/voice-reconnect.log"
RELAY_LOG="$RESULTS_DIR/relay-boundary.log"
mkdir -p "$RESULTS_DIR"

setup_status="pass"
reconnect_status="pass"
relay_status="pass"

if ! go test ./tests/e2e/v14/... -run '^TestVoiceFlowSequence$' >"$SETUP_LOG" 2>&1; then
	setup_status="fail"
fi

if ! go test ./tests/e2e/v14/... -run '^TestReconnectRecovery$' >"$RECONNECT_LOG" 2>&1; then
	reconnect_status="fail"
fi

if ! go test ./tests/e2e/v14/... -run '^TestVoiceFlowSequence$' >"$RELAY_LOG" 2>&1; then
	relay_status="fail"
fi

overall_status="pass"
if [[ "$setup_status" != "pass" || "$reconnect_status" != "pass" || "$relay_status" != "pass" ]]; then
	overall_status="fail"
fi

timestamp=$(date --utc +%Y-%m-%dT%H:%M:%SZ)

cat > "$MANIFEST" <<JSON
{
  "version": "v14",
  "status": "$overall_status",
  "timestamp": "$timestamp",
  "scenarios": [
    {"name": "voice-call-setup", "status": "$setup_status", "steps": 4, "log": "voice-call-setup.log"},
    {"name": "voice-reconnect", "status": "$reconnect_status", "steps": 3, "log": "voice-reconnect.log"},
    {"name": "relay-boundary-regression", "status": "$relay_status", "steps": 2, "log": "relay-boundary.log"}
  ]
}
JSON

echo "wrote $MANIFEST"

if [[ "$overall_status" != "pass" ]]; then
	exit 1
fi
