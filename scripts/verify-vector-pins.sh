#!/usr/bin/env bash
# Verify integrity of test vector files against pin.sha256.
# Exit 1 on any mismatch or missing file.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VECTOR_DIR="$SCRIPT_DIR/../docs/spec/v0.1/91-test-vectors"
PIN_FILE="$VECTOR_DIR/pin.sha256"

if [[ ! -f "$PIN_FILE" ]]; then
  echo "ERROR: pin.sha256 not found at $PIN_FILE" >&2
  exit 1
fi

FAILED=0
while IFS= read -r line; do
  [[ -z "$line" || "$line" == \#* ]] && continue
  expected_hash="${line%% *}"
  filename="${line##* }"
  filepath="$VECTOR_DIR/$filename"

  if [[ ! -f "$filepath" ]]; then
    echo "MISSING: $filename" >&2
    FAILED=1
    continue
  fi

  actual_hash="$(sha256sum "$filepath" | awk '{print $1}')"
  if [[ "$actual_hash" != "$expected_hash" ]]; then
    echo "MISMATCH: $filename" >&2
    echo "  want: $expected_hash" >&2
    echo "  got:  $actual_hash" >&2
    FAILED=1
  else
    echo "OK: $filename"
  fi
done < "$PIN_FILE"

if [[ $FAILED -ne 0 ]]; then
  echo "Vector pin verification FAILED." >&2
  exit 1
fi
echo "All vector pins verified OK."
