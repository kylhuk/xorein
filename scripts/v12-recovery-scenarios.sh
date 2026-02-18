#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACT_DIR="$REPO_DIR/artifacts/generated/v12-recovery-scenarios"
MANIFEST_PATH="$ARTIFACT_DIR/result-manifest.json"
PODMAN_IMAGE="docker.io/library/golang:1.24.12"

NEW_DEVICE_LOG="$ARTIFACT_DIR/new-device-restore.log"
LOST_PASSWORD_LOG="$ARTIFACT_DIR/lost-password-without-backup.log"
TAMPER_LOG="$ARTIFACT_DIR/backup-tamper-detection.log"
RELAY_BOUNDARY_LOG="$ARTIFACT_DIR/relay-boundary-regression.log"

mkdir -p "$ARTIFACT_DIR"

require_command() {
	local command_name="$1"
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "[v12-recovery] required command not found: $command_name" >&2
		exit 2
	fi
}

run_probe() {
	local name="$1"
	local expected_exit="$2"
	local required_text="$3"
	local log_file="$4"
	local command="$5"

	local output
	local exit_code
	local status="pass"
	local reason=""

	set +e
	podman run --rm --userns=keep-id \
		-v "$REPO_DIR:/workspace:Z" \
		-w /workspace \
		"$PODMAN_IMAGE" \
		bash -lc "$command" >"$log_file" 2>&1
	exit_code=$?
	set -e

	output="$(cat "$log_file")"

	if [[ "$exit_code" -ne "$expected_exit" ]]; then
		status="fail"
		reason="unexpected_exit_code:$exit_code"
	fi

	if [[ "$status" == "pass" && -n "$required_text" && "$output" != *"$required_text"* ]]; then
		status="fail"
		reason="missing_required_output"
	fi

	echo "$name|$expected_exit|$exit_code|$status|$reason|$log_file|$(printf '%s' "$required_text")|$command"
}

require_command podman

new_device_result="$(run_probe "new-device-restore" 0 "PASS" "$NEW_DEVICE_LOG" "/usr/local/go/bin/go test ./tests/e2e/v12 -run TestNewDeviceRestoreScenario -count=1 -v")"
lost_password_result="$(run_probe "lost-password-without-backup" 0 "PASS" "$LOST_PASSWORD_LOG" "/usr/local/go/bin/go test ./tests/e2e/v12 -run TestLostPasswordWithoutBackupScenario -count=1 -v")"
tamper_result="$(run_probe "backup-tamper-detection" 0 "PASS" "$TAMPER_LOG" "/usr/local/go/bin/go test ./pkg/v12/backup -run TestRestoreDetectsTamperedCiphertext -count=1 -v")"
relay_boundary_result="$(run_probe "relay-boundary-regression" 0 "PASS" "$RELAY_BOUNDARY_LOG" "/usr/local/go/bin/go test ./tests/e2e/v12 -run TestRelayBoundaryRegressionScenario -count=1 -v")"

IFS='|' read -r new_name new_expected new_actual new_status new_reason new_log new_required new_command <<<"$new_device_result"
IFS='|' read -r lost_name lost_expected lost_actual lost_status lost_reason lost_log lost_required lost_command <<<"$lost_password_result"
IFS='|' read -r tamper_name tamper_expected tamper_actual tamper_status tamper_reason tamper_log tamper_required tamper_command <<<"$tamper_result"
IFS='|' read -r relay_name relay_expected relay_actual relay_status relay_reason relay_log relay_required relay_command <<<"$relay_boundary_result"

cat > "$MANIFEST_PATH" <<EOF
{
  "suite": "v12-recovery-scenarios",
  "podman_image": "$PODMAN_IMAGE",
  "artifact_dir": "$ARTIFACT_DIR",
  "probes": [
    {
      "name": "$new_name",
      "command": "$new_command",
      "expected_exit_code": $new_expected,
      "actual_exit_code": $new_actual,
      "status": "$new_status",
      "failure_reason": "$new_reason",
      "required_output": "$new_required",
      "log": "$new_log"
    },
    {
      "name": "$lost_name",
      "command": "$lost_command",
      "expected_exit_code": $lost_expected,
      "actual_exit_code": $lost_actual,
      "status": "$lost_status",
      "failure_reason": "$lost_reason",
      "required_output": "$lost_required",
      "log": "$lost_log"
    },
    {
      "name": "$tamper_name",
      "command": "$tamper_command",
      "expected_exit_code": $tamper_expected,
      "actual_exit_code": $tamper_actual,
      "status": "$tamper_status",
      "failure_reason": "$tamper_reason",
      "required_output": "$tamper_required",
      "log": "$tamper_log"
    },
    {
      "name": "$relay_name",
      "command": "$relay_command",
      "expected_exit_code": $relay_expected,
      "actual_exit_code": $relay_actual,
      "status": "$relay_status",
      "failure_reason": "$relay_reason",
      "required_output": "$relay_required",
      "log": "$relay_log"
    }
  ]
}
EOF

echo "[v12-recovery] wrote manifest: $MANIFEST_PATH"

if [[ "$new_status" != "pass" || "$lost_status" != "pass" || "$tamper_status" != "pass" || "$relay_status" != "pass" ]]; then
	echo "[v12-recovery] one or more checks failed"
	exit 1
fi

echo "[v12-recovery] all checks passed"
exit 0
