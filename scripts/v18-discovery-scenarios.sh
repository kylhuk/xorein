#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MANIFEST_DIR="$REPO_DIR/artifacts/generated/v18-discovery-scenarios"
MANIFEST_PATH="$MANIFEST_DIR/result-manifest.json"

mkdir -p "$MANIFEST_DIR"

declare -A status
scenario_order=(
  "pkg-v18"
  "e2e-v18"
  "perf-v18"
  "relay-boundary"
)

run_scenario() {
  local name=$1
  local command=$2
  local log_file=$3

  local exit_code=0
  set +e
  bash -lc "cd \"$REPO_DIR\" && $command" > "$log_file" 2>&1
  exit_code=$?
  set -e

  if [[ $exit_code -eq 0 ]]; then
    status["$name"]=pass
  else
    status["$name"]=fail
  fi

  statuses+=("$name|$command|$exit_code|${status[$name]}|$log_file")
  if [[ $exit_code -ne 0 ]]; then
    return 1
  fi
}

statuses=()
overall_status=0

run_scenario "pkg-v18" "go test ./pkg/v18/..." "$MANIFEST_DIR/pkg-v18.log" || overall_status=1
run_scenario "e2e-v18" "go test ./tests/e2e/v18/..." "$MANIFEST_DIR/e2e-v18.log" || overall_status=1
run_scenario "perf-v18" "go test ./tests/perf/v18/..." "$MANIFEST_DIR/perf-v18.log" || overall_status=1
run_scenario "relay-boundary" "go test ./tests/e2e/v18/... -run '^TestDiscoveryIntegrityEnforcesRelayBoundary$' -count=1" "$MANIFEST_DIR/relay-boundary.log" || overall_status=1

timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
cat <<JSON > "$MANIFEST_PATH"
{
  "version": "v18-discovery-scenarios",
  "timestamp": "$timestamp",
  "overall": "$(if [[ $overall_status -eq 0 ]]; then echo pass; else echo fail; fi)",
  "scenarios": [
JSON

for idx in "${!scenario_order[@]}"; do
  name="${scenario_order[$idx]}"
  entry=""
  # shellcheck disable=SC2155
  for s in "${statuses[@]}"; do
    if [[ $s == "$name|"* ]]; then
      entry="$s"
      break
    fi
  done

  IFS='|' read -r parsed_name parsed_command parsed_exit parsed_status parsed_log <<< "$entry"
  separator="," 
  if [[ $idx -eq 0 ]]; then
    separator=""
  fi

  cat <<JSON >> "$MANIFEST_PATH"
    ${separator}{
      "id": "$parsed_name",
      "command": "$parsed_command",
      "exit_code": $parsed_exit,
      "status": "$parsed_status",
      "log": "$parsed_log"
    }
JSON

done

cat <<'JSON' >> "$MANIFEST_PATH"
  ]
}
JSON

printf 'Result manifest written to %s\n' "$MANIFEST_PATH"
exit $overall_status
