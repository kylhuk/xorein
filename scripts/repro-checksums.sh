#!/bin/bash
set -euo pipefail
ARTIFACTS_DIR="artifacts/generated"
mkdir -p "$ARTIFACTS_DIR"
find "$ARTIFACTS_DIR" -type f -print0 | sort -z | xargs -0 sha256sum > "$ARTIFACTS_DIR/checksums.txt"
