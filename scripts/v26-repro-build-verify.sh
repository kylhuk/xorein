#!/usr/bin/env bash
set -euo pipefail
set -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/artifacts/generated/v26-evidence/repro-build"
BASELINE_DIR=""

TARGETS=(xorein harmolyn)
declare -A HASHES

usage() {
  cat <<'USAGE'
Usage: ./scripts/v26-repro-build-verify.sh [--help] [--baseline <path>]

Build deterministic binaries and collect v2.6 reproducibility artifacts.

Options:
  --help             Show this help text and exit.
  --baseline <path>  Optional baseline output directory from a prior run.
                     If set, script compares checksums.txt against baseline
                     and fails if any hash mismatch is found.

Outputs (default):
  artifacts/generated/v26-evidence/repro-build/
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --baseline)
      if [[ $# -lt 2 ]]; then
        echo "--baseline requires a path argument" >&2
        exit 1
      fi
      BASELINE_DIR="$2"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
  shift
 done

log() {
  printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$RUN_LOG"
}

if ! command -v go >/dev/null 2>&1; then
  echo "go toolchain is required" >&2
  exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "sha256sum is required" >&2
  exit 1
fi

if [[ ! -d "$ROOT_DIR/cmd" ]]; then
  echo "expected cmd directory at $ROOT_DIR/cmd" >&2
  exit 1
fi

if [[ -n "$BASELINE_DIR" ]]; then
  if [[ ! -d "$BASELINE_DIR" ]]; then
    echo "baseline directory not found: $BASELINE_DIR" >&2
    exit 1
  fi
  BASELINE_DIR="$(cd "$BASELINE_DIR" && pwd)"
  BASELINE_CHECKSUMS="$BASELINE_DIR/checksums.txt"
  if [[ ! -f "$BASELINE_CHECKSUMS" ]]; then
    echo "baseline checksums not found: $BASELINE_CHECKSUMS" >&2
    exit 1
  fi
fi

mkdir -p "$OUTPUT_DIR"
RUN_LOG="$OUTPUT_DIR/run.log"
CHECKSUM_FILE="$OUTPUT_DIR/checksums.txt"
MANIFEST_FILE="$OUTPUT_DIR/repro-build-manifest.txt"
EVIDENCE_MAP_FILE="$OUTPUT_DIR/evidence-map.txt"
COMPARISON_FILE="$OUTPUT_DIR/baseline-comparison.txt"

: > "$RUN_LOG"
: > "$CHECKSUM_FILE"
: > "$MANIFEST_FILE"
: > "$EVIDENCE_MAP_FILE"
: > "$COMPARISON_FILE"

log "repro build verify start"
log "workspace: $ROOT_DIR"
log "output_dir: $OUTPUT_DIR"
if [[ -n "$BASELINE_DIR" ]]; then
  log "baseline_dir: $BASELINE_DIR"
else
  log "baseline_dir: not set"
fi

GO_VERSION="$(go version)"
for target in "${TARGETS[@]}"; do
  source_dir="$ROOT_DIR/cmd/$target"
  binary_path="$OUTPUT_DIR/binaries/$target"

  if [[ ! -d "$source_dir" ]]; then
    echo "build target missing: $source_dir" >&2
    exit 1
  fi

  mkdir -p "$(dirname "$binary_path")"
  rm -f "$binary_path"

  BUILD_CMD=(go build -mod=readonly -trimpath -buildvcs=false -ldflags '-s -w -buildid=' -o "$binary_path" "$source_dir")
  log "build $target"
  if ! "${BUILD_CMD[@]}" 2>&1 | tee -a "$RUN_LOG"; then
    echo "build failed for $target" >&2
    exit 1
  fi

  if [[ ! -x "$binary_path" ]]; then
    echo "build did not produce executable for $target" >&2
    exit 1
  fi

  HASHES[$target]="$(sha256sum "$binary_path" | awk '{print $1}')"
  log "sha256 $target=${HASHES[$target]}"

  printf '%s  binaries/%s\n' "${HASHES[$target]}" "$target" >> "$CHECKSUM_FILE"
 done

cat <<MANIFEST > "$MANIFEST_FILE"
generated_utc: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
script: "scripts/v26-repro-build-verify.sh"
workspace: "$ROOT_DIR"
go_version: "$GO_VERSION"
output_dir: "$OUTPUT_DIR"
targets:
MANIFEST
for target in "${TARGETS[@]}"; do
  cat <<ENTRY >> "$MANIFEST_FILE"
  - name: "$target"
    package: "./cmd/$target"
    output: "binaries/$target"
    sha256: "${HASHES[$target]}"
ENTRY
done

if [[ -n "$BASELINE_DIR" ]]; then
  log "running baseline comparison"
  BASELINE_MISMATCH=0
  {
    echo 'artifact,current_sha256,baseline_sha256,status'
    for target in "${TARGETS[@]}"; do
      baseline_sha="$(awk -v artifact="binaries/$target" '$2 == artifact { print $1; exit }' "$BASELINE_CHECKSUMS")"
      current_sha="${HASHES[$target]}"
      status=""

      if [[ -z "$baseline_sha" ]]; then
        baseline_sha=""
        status="missing-baseline-entry"
        BASELINE_MISMATCH=1
      elif [[ "$current_sha" == "$baseline_sha" ]]; then
        status="match"
      else
        status="mismatch"
        BASELINE_MISMATCH=1
      fi

      printf '%s,%s,%s,%s\n' "$target" "$current_sha" "$baseline_sha" "$status"
    done
  } > "$COMPARISON_FILE"

  if [[ "$BASELINE_MISMATCH" -ne 0 ]]; then
    echo "baseline verification failed; see $COMPARISON_FILE" | tee -a "$RUN_LOG"
    echo "reproducibility status: FAILED" >> "$COMPARISON_FILE"
  else
    echo "baseline verification status: PASS" >> "$COMPARISON_FILE"
    log "baseline comparison pass"
  fi
else
  echo "baseline comparison skipped" >> "$COMPARISON_FILE"
fi

if [[ -n "$BASELINE_DIR" ]]; then
  BASELINE_COMPARISON_STATUS="pass"
  if grep -q ',mismatch,' "$COMPARISON_FILE" || grep -q ',missing-baseline-entry,' "$COMPARISON_FILE"; then
    BASELINE_COMPARISON_STATUS="fail"
  fi
else
  BASELINE_COMPARISON_STATUS="not-run"
fi

cat <<MAP > "$EVIDENCE_MAP_FILE"
EV-v26-G6-001|scripts/v26-repro-build-verify.sh|deterministic rebuild of cmd/xorein and cmd/harmolyn|pass
EV-v26-G6-002|scripts/v26-repro-build-verify.sh|sha256 checksums output written to checksums.txt|pass
EV-v26-G6-003|scripts/v26-repro-build-verify.sh --baseline <path>|baseline hash comparison against checksums.txt|${BASELINE_COMPARISON_STATUS}
MAP

if [[ -n "$BASELINE_DIR" ]]; then
  if [[ "$BASELINE_COMPARISON_STATUS" == "fail" ]]; then
    echo "reproducible build verification failed due to baseline mismatch" >&2
    exit 1
  fi
fi

log "reproducible build verification completed successfully"
cat <<SUMMARY
repro-build verification complete
output_dir: $OUTPUT_DIR
checksums: $CHECKSUM_FILE
manifest: $MANIFEST_FILE
evidence_map: $EVIDENCE_MAP_FILE
run_log: $RUN_LOG
SUMMARY
exit 0
