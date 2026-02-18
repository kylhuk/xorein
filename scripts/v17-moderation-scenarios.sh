#!/usr/bin/env bash
set -uo pipefail

manifest_dir="artifacts/generated/v17-moderation-scenarios"
manifest_path="$manifest_dir/result-manifest.json"
mkdir -p "$manifest_dir"

declare -A results
scenario_order=(
  "pkg-v17"
  "e2e-v17"
  "perf-v17"
  "relay-policy"
)

run_scenario() {
  local name=$1
  shift
  if "$@"; then
    results["$name"]="pass"
  else
    results["$name"]="fail"
    exit_code=1
  fi
}

exit_code=0
run_scenario pkg-v17 go test ./pkg/v17/...
run_scenario e2e-v17 go test ./tests/e2e/v17/...
run_scenario perf-v17 go test ./tests/perf/v17/...
run_scenario relay-policy go test ./pkg/v11/relaypolicy

time_stamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
{
  printf '{\n'
  printf '  "version": "v17-moderation-scenarios",\n'
  printf '  "timestamp": "%s",\n' "$time_stamp"
  printf '  "scenarios": [\n'
  first=1
  for name in "${scenario_order[@]}"; do
    if [[ $first -eq 0 ]]; then
      printf ',\n'
    else
      first=0
    fi
    printf '    {"name": "%s", "status": "%s"}' "$name" "${results[$name]:-fail}"
  done
  printf '\n  ]\n'
  printf '}\n'
} > "$manifest_path"

printf 'Result manifest written to %s\n' "$manifest_path"
exit "$exit_code"
