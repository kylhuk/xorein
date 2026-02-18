#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACT_DIR="$REPO_DIR/artifacts/generated/v11-relay-smoke"
MANIFEST_PATH="$ARTIFACT_DIR/result-manifest.json"
SMOKE_BINARY="$REPO_DIR/bin/aether"
SMOKE_BINARY_CONTAINER=""
PODMAN_IMAGE="docker.io/library/golang:1.24.8"
RELAY_LISTEN="127.0.0.1:4001"
RELAY_STORE_BASE="/workspace/artifacts/generated/v11-relay-smoke/stores"

ALLOWED_MODE="session-metadata"
FORBIDDEN_MODE="durable-message-body"

ALLOWED_LOG="$ARTIFACT_DIR/allowed-relay-session-metadata.log"
FORBIDDEN_LOG="$ARTIFACT_DIR/forbidden-relay-durable-message-body.log"

mkdir -p "$ARTIFACT_DIR"

has_elf_binary() {
	local candidate="$1"

	if [[ ! -x "$candidate" ]]; then
		return 1
	fi

	local file_type
	if ! command -v file >/dev/null 2>&1; then
		return 1
	fi
	file_type="$(file "$candidate" 2>/dev/null || true)"
	[[ "$file_type" == *"ELF"* ]]
}

require_command() {
	local command_name="$1"
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "[relay-smoke] required command not found: $command_name" >&2
		exit 2
	fi
}

build_or_reuse_binary() {
	if has_elf_binary "$SMOKE_BINARY"; then
		echo "[relay-smoke] using existing local relay binary: $SMOKE_BINARY" >&2
		echo "true"
		return
	fi

	require_command go
	local smoke_goarch
	smoke_goarch="$(go env GOARCH)"
	if [[ -z "$smoke_goarch" ]]; then
		echo "[relay-smoke] failed to resolve GOARCH via go env" >&2
		exit 2
	fi

	SMOKE_BINARY="$ARTIFACT_DIR/relay-smoke"
	echo "[relay-smoke] building relay binary at $SMOKE_BINARY" >&2
	rm -f "$SMOKE_BINARY"
	(cd "$REPO_DIR" && GOOS=linux GOARCH="$smoke_goarch" CGO_ENABLED=0 go build -o "$SMOKE_BINARY" ./cmd/aether)
	echo "false"
}

run_probe() {
	local mode="$1"
	local expected_exit="$2"
	local required_text="$3"
	local log_file="$4"
	local output
	local exit_code
	local status="pass"
	local reason=""

	set +e
	podman run --rm --userns=keep-id \
		-v "$REPO_DIR:/workspace:Z" \
		-w /workspace \
		"$PODMAN_IMAGE" \
		"$SMOKE_BINARY_CONTAINER" \
		--mode relay \
		--relay-listen "$RELAY_LISTEN" \
		--relay-store "$RELAY_STORE_BASE/$mode" \
		--relay-persistence-mode "$mode" \
		--relay-health-interval 10ms >"$log_file" 2>&1
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

	echo "$mode|$expected_exit|$exit_code|$status|$reason|$log_file|$(printf '%s' "$required_text")"
}

existing_binary=$(build_or_reuse_binary)
require_command podman
SMOKE_BINARY_CONTAINER="/workspace${SMOKE_BINARY#$REPO_DIR}"

allowed_result="$(run_probe "$ALLOWED_MODE" 0 "Relay runtime active mode=relay" "$ALLOWED_LOG")"
forbidden_result="$(run_probe "$FORBIDDEN_MODE" 16 "policy violation" "$FORBIDDEN_LOG")"

IFS='|' read -r allowed_mode allowed_expected_exit allowed_exit allowed_status allowed_reason allowed_log allowed_required <<<"$allowed_result"
IFS='|' read -r forbidden_mode forbidden_expected_exit forbidden_exit forbidden_status forbidden_reason forbidden_log forbidden_required <<<"$forbidden_result"

cat > "$MANIFEST_PATH" <<EOF
{
  "suite": "v11-relay-smoke",
  "binary": "$SMOKE_BINARY",
  "binary_reused": $existing_binary,
  "podman_image": "$PODMAN_IMAGE",
  "artifact_dir": "$ARTIFACT_DIR",
  "probes": [
    {
      "name": "allowed-persistence-mode",
      "mode": "$allowed_mode",
      "expected_exit_code": $allowed_expected_exit,
      "actual_exit_code": $allowed_exit,
      "status": "$allowed_status",
      "failure_reason": "$allowed_reason",
      "required_output": "$allowed_required",
      "log": "$allowed_log"
    },
    {
      "name": "forbidden-persistence-mode",
      "mode": "$forbidden_mode",
      "expected_exit_code": $forbidden_expected_exit,
      "actual_exit_code": $forbidden_exit,
      "status": "$forbidden_status",
      "failure_reason": "$forbidden_reason",
      "required_output": "$forbidden_required",
      "log": "$forbidden_log"
    }
  ]
}
EOF

echo "[relay-smoke] wrote manifest: $MANIFEST_PATH"

if [[ "$allowed_status" != "pass" || "$forbidden_status" != "pass" ]]; then
	echo "[relay-smoke] one or more checks failed"
	exit 1
fi

echo "[relay-smoke] all checks passed"
exit 0
