#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")"/.. && pwd)"
artifact_dir="$root/artifacts/generated/v13-e2e-podman"
mkdir -p "$artifact_dir"
status="pass"
if ! go test "${root}/tests/e2e/v13/..."; then
    status="fail"
fi
if ! go test "${root}/tests/perf/v13/..."; then
    status="fail"
fi
cat <<JSON > "$artifact_dir/result-manifest.json"
{
  "status": "$status",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "checks": ["e2e", "perf"],
  "invocation": "scripts/v13-e2e-podman.sh"
}
JSON
if [ "$status" = "fail" ]; then
    exit 1
fi
