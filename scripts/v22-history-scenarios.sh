#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
artifact_dir="$root/artifacts/generated/v22-history-scenarios"
manifest_file="$artifact_dir/manifest.txt"
run_log="$artifact_dir/run.log"
podman_image="docker.io/library/golang:1.24.12"
scenarios_file="$root/containers/v2.2/scenarios.conf"

mkdir -p "$artifact_dir"
: >"$run_log"

podman_available="false"
podman_blocker=""
declare -a scenario_records=()

env_status() {
    if [[ "$podman_available" == "true" ]]; then
        echo "pass"
    else
        echo "BLOCKED:ENV"
    fi
}

log_cmd() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" >> "$run_log"
}

run_probe() {
    local slug="$1"
    local label="$2"
    local pkg="$3"
    local pattern="$4"
    local required_output="${5:-PASS}"

    local log_path="$artifact_dir/$slug.log"
    local status
    local reason=""
    local actual_exit="n/a"
    local command_label

    status="$(env_status)"
    command_label="podman run --rm --userns=keep-id -v '$root:/workspace:Z' -w /workspace $podman_image bash -lc 'cd /workspace && /usr/local/go/bin/go test $pkg -run $pattern -count=1 -v'"

    if [[ "$podman_available" != "true" ]]; then
        reason="$podman_blocker"
        printf '%s\n' "blocked reason: $reason" >"$log_path"
    else
        log_cmd "RUN scenario=$label"
        log_cmd "command=$command_label"

        container_cmd="/usr/local/go/bin/go test $pkg -run $pattern -count=1 -v"
        set +e
        podman run --rm --userns=keep-id \
          -v "$root:/workspace:Z" \
          -w /workspace \
          "$podman_image" \
          bash -lc "cd /workspace && $container_cmd" \
          >"$log_path" 2>&1
        actual_exit=$?
        set -e

        output="$(cat "$log_path")"
        log_cmd "exit_code=$actual_exit"

        if [[ "$actual_exit" -ne 0 ]]; then
            status="fail"
            reason="exit_code:$actual_exit"
        elif [[ "$output" != *"$required_output"* ]]; then
            status="fail"
            reason="missing_required_output:$required_output"
        else
            status="pass"
        fi
    fi

    scenario_records+=("$slug|$status|$actual_exit|$reason|$pkg|$pattern|$required_output|$log_path|$command_label")
    log_cmd "status=$status reason=${reason:-none}"
}

if [[ ! -f "$scenarios_file" ]]; then
    echo "missing scenarios definition: $scenarios_file" >&2
    exit 1
fi

set +e
if ! command -v podman >/dev/null 2>&1; then
    podman_blocker="podman command not found"
    log_cmd "podman availability check failed: $podman_blocker"
elif ! podman_info=$(podman info 2>&1); then
    podman_info_single_line="${podman_info//$'\n'/ }"
    podman_blocker="podman runtime unavailable or unusable: $podman_info_single_line"
    log_cmd "podman availability check failed: $podman_blocker"
else
    podman_available="true"
    log_cmd "podman runtime check passed"
fi
set -e

while IFS='|' read -r slug label pkg pattern required; do
    [[ -z "$slug" || "$slug" == \#* ]] && continue
    run_probe "$slug" "$label" "$pkg" "$pattern" "$required"
done < "$scenarios_file"

{
    echo "suite: v22-history-scenarios"
    echo "generated_at: \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\""
    echo "podman_image: \"$podman_image\""
    echo "podman_available: $podman_available"
    echo "podman_blocker: \"$podman_blocker\""
    echo "artifact_dir: \"$artifact_dir\""
    echo "log: \"$run_log\""
    echo "scenarios:"
    for result in "${scenario_records[@]}"; do
        IFS='|' read -r result_slug result_status result_exit result_reason result_pkg result_pattern result_required result_log result_cmd <<<"$result"
        echo "  - slug: $result_slug"
        echo "    status: $result_status"
        echo "    package: $result_pkg"
        echo "    test_pattern: $result_pattern"
        echo "    expected_exit_code: 0"
        echo "    actual_exit_code: $result_exit"
        echo "    required_output: $result_required"
        echo "    failure_reason: \"$result_reason\""
        echo "    command: \"$result_cmd\""
        echo "    log: \"$result_log\""
    done
} >"$manifest_file"

overall_fail="false"
for result in "${scenario_records[@]}"; do
    IFS='|' read -r result_slug result_status _ <<<"$result"
    echo "[v22-history-scenarios] $result_slug -> $result_status"
    if [[ "$result_status" == "fail" ]]; then
        overall_fail="true"
    fi
done

if [[ "$overall_fail" == "true" ]]; then
    echo "[v22-history-scenarios] one or more scenario probes failed"
    exit 1
fi

echo "[v22-history-scenarios] wrote manifest: $manifest_file"
echo "[v22-history-scenarios] wrote log: $run_log"
exit 0
