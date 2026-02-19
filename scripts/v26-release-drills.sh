#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACT_DIR="$ROOT_DIR/artifacts/generated/v26-release-drills"
MANIFEST_FILE="$ARTIFACT_DIR/manifest.txt"
RUN_LOG="$ARTIFACT_DIR/run.log"

mkdir -p "$ARTIFACT_DIR"
: > "$RUN_LOG"

GO_OK=true
if ! command -v go >/dev/null 2>&1; then
    GO_OK=false
fi

REPRO_OK=true
if [[ ! -x "$ROOT_DIR/scripts/v26-repro-build-verify.sh" ]]; then
    REPRO_OK=false
fi

log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" | tee -a "$RUN_LOG"
}

run_drill() {
    local drill_id="$1"
    local label="$2"
    local command_to_run="$3"
    local require_go="${4:-false}"
    local require_script="${5:-false}"
    local status="pass"
    local reason=""
    local actual_exit="n/a"
    local log_path="$ARTIFACT_DIR/$drill_id.log"

    : > "$log_path"

    if [[ "$require_go" == "true" && "$GO_OK" != "true" ]]; then
        status="BLOCKED:ENV"
        reason="go command not found"
    fi

    if [[ "$status" == "pass" && "$require_script" == "true" && "$REPRO_OK" != "true" ]]; then
        status="BLOCKED:ENV"
        reason="scripts/v26-repro-build-verify.sh is missing or not executable"
    fi

    if [[ "$status" == "pass" ]]; then
        log "RUN drill=$drill_id label=$label"
        log "command=cd \"$ROOT_DIR\" && $command_to_run"
        set +e
        bash -lc "cd \"$ROOT_DIR\" && $command_to_run" >"$log_path" 2>&1
        actual_exit=$?
        set -e

        if [[ "$actual_exit" -ne 0 ]]; then
            status="fail"
            reason="exit_code:$actual_exit"
        fi
    else
        echo "blocked reason: $reason" >"$log_path"
    fi

    DRILL_SUMMARY+=("$drill_id|$status|$actual_exit|$reason|$command_to_run|$log_path")
    log "status=$drill_id result=$status reason=${reason:-none} exit=$actual_exit"
}

usage() {
  cat <<'USAGE'
Usage: ./scripts/v26-release-drills.sh

Runs the v2.6 release-readiness drills and writes deterministic artifacts to:
  artifacts/generated/v26-release-drills/manifest.txt
  artifacts/generated/v26-release-drills/run.log
  artifacts/generated/v26-release-drills/<drill>.log
USAGE
}

if [[ ${1:-} == "--help" ]]; then
    usage
    exit 0
fi

if [[ $# -gt 0 ]]; then
    usage
    exit 2
fi

log "release-drills start"
log "root_dir=$ROOT_DIR"

DRILL_SUMMARY=()

run_drill "release-e2e-v26" "go test regression surface" "go test ./tests/e2e/v26/..." true false
run_drill "release-perf-v26" "go test perf surface" "go test ./tests/perf/v26/..." true false
run_drill "release-repro-build-v26" "repro build verification" "./scripts/v26-repro-build-verify.sh" true true
run_drill "release-runbook-readiness" "runbook presence audit" "test -f docs/v2.6/phase3/p3-relay-runbook.md && test -f docs/v2.6/phase3/p3-archivist-runbook.md && test -f docs/v2.6/phase3/p3-aux-services-runbook.md" false false

overall="pass"
{
    echo "suite: v26-release-drills"
    echo "generated_at: \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\""
    echo "artifact_dir: \"$ARTIFACT_DIR\""
    echo "root_dir: \"$ROOT_DIR\""
    echo "drills:"
    for result in "${DRILL_SUMMARY[@]}"; do
        IFS='|' read -r drill_id status actual_exit reason command log_path <<<"$result"
        echo "  - drill: $drill_id"
        echo "    status: $status"
        echo "    command: $command"
        echo "    exit_code: $actual_exit"
        echo "    reason: \"$reason\""
        echo "    log: \"$log_path\""

        if [[ "$status" == "fail" ]]; then
            overall="fail"
        elif [[ "$status" == "BLOCKED:ENV" && "$overall" == "pass" ]]; then
            overall="BLOCKED:ENV"
        fi
    done
    echo "overall_status: $overall"
} > "$MANIFEST_FILE"

log "release-drills finished"
log "overall_status=$overall"
log "manifest=$MANIFEST_FILE"

if [[ "$overall" == "fail" || "$overall" == "BLOCKED:ENV" ]]; then
    exit 1
fi

exit 0
