#!/bin/bash
set -euo pipefail
ARTIFACTS_DIR="artifacts/generated"
mkdir -p "$ARTIFACTS_DIR"
find "$ARTIFACTS_DIR" -type f ! -path "$ARTIFACTS_DIR/release-pack/signing/*" -print0 | sort -z | xargs -0 sha256sum > "$ARTIFACTS_DIR/checksums.txt"
