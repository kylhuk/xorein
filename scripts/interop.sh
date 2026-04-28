#!/usr/bin/env bash
# interop.sh — Seal-DM round-trip conformance driver (spec 90 §7)
#
# Runs the two-implementation Seal-DM interop harness. Currently both
# "implementations" are the same binary (same process, isolated key material),
# exercising the full X3DH + Double Ratchet path.
#
# Exit code: 0 = pass, 1 = failure

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "=== interop.sh: Seal-DM interop harness ==="
echo ""

cd "$REPO_ROOT"

echo "[interop] Running Seal-DM round-trip tests..."
go test -race -v -run 'TestSeal' ./pkg/v0_1/spectest/interop/...

echo ""
echo "=== interop.sh: PASS ==="
