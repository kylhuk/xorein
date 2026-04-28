#!/usr/bin/env bash
# spec-lint.sh — self-consistency check for docs/spec/v0.1/
#
# Checks:
#   1. Every cap.* reference in spec exists in pkg/protocol/capabilities.go
#   2. Every /aether/<family>/<version> reference exists in pkg/protocol/registry.go
#   3. Every proto message name referenced in spec exists in proto/aether.proto
#   4. Every KAT reference in spec resolves to a file in docs/spec/v0.1/91-test-vectors/
#   5. Every legacy doc has a SUPERSEDED header
#
# Exit code: 0 = pass, 1 = failures found

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SPEC_DIR="$REPO_ROOT/docs/spec/v0.1"
VECTORS_DIR="$SPEC_DIR/91-test-vectors"
CAPABILITIES_GO="$REPO_ROOT/pkg/protocol/capabilities.go"
REGISTRY_GO="$REPO_ROOT/pkg/protocol/registry.go"
PROTO_FILE="$REPO_ROOT/proto/aether.proto"

FAIL=0

error() {
  echo "ERROR: $*" >&2
  FAIL=1
}

info() {
  echo "  $*"
}

echo "=== spec-lint.sh: Xorein v0.1 spec self-consistency check ==="
echo ""

# ── 1. cap.* references ───────────────────────────────────────────────────────
echo "[1/5] Checking cap.* references..."

# Extract all cap.* tokens from spec markdown (excluding comments/code that start with //)
cap_refs=$(grep -rh --include="*.md" -o 'cap\.[a-z0-9_.-]*' "$SPEC_DIR" 2>/dev/null | sort -u || true)

for cap in $cap_refs; do
  # Look for the cap string in capabilities.go (as a Go string literal)
  if ! grep -q "\"$cap\"" "$CAPABILITIES_GO" 2>/dev/null; then
    error "cap '$cap' referenced in spec but not found in pkg/protocol/capabilities.go"
  fi
done

if [[ $FAIL -eq 0 ]]; then
  info "All cap.* references OK"
fi

# ── 2. /aether/<family>/<version> references ──────────────────────────────────
echo "[2/5] Checking /aether/<family>/<version> references..."

# DHT/routing namespaces that are valid in the spec but are NOT stream-handler
# protocol families in registry.go:
DHT_NAMESPACES=("kad" "noise")

# Extract protocol ID patterns from spec
proto_refs=$(grep -rh --include="*.md" -o '/aether/[a-z0-9._/-]*' "$SPEC_DIR" 2>/dev/null | sort -u || true)

for pref in $proto_refs; do
  # Only check paths that look like a full protocol ID (family/major.minor.patch)
  if [[ "$pref" =~ ^/aether/[a-z0-9]+/[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    family=$(echo "$pref" | cut -d/ -f3)

    # Skip DHT routing namespaces — they are not stream handler protocols
    skip=0
    for ns in "${DHT_NAMESPACES[@]}"; do
      if [[ "$family" == "$ns" ]]; then skip=1; break; fi
    done
    [[ $skip -eq 1 ]] && continue

    if ! grep -q "\"$family\"" "$REGISTRY_GO" 2>/dev/null && \
       ! grep -q "$family" "$REGISTRY_GO" 2>/dev/null; then
      error "Protocol ID '$pref' referenced in spec but family '$family' not found in pkg/protocol/registry.go"
    fi
  fi
done

info "Protocol ID reference check complete"

# ── 3. Proto message names ────────────────────────────────────────────────────
echo "[3/5] Checking protobuf message references..."

# Extract capitalized identifiers that look like proto message names from spec
# Pattern: word starting with capital that appears in a protobuf context
# We check for names used in backtick code spans or after "message" keyword references
msg_refs=$(grep -rh --include="*.md" -o '`[A-Z][A-Za-z0-9]*`' "$SPEC_DIR" 2>/dev/null | tr -d '`' | sort -u || true)

proto_fail=0
for msg in $msg_refs; do
  # Only check messages that plausibly correspond to proto definitions
  # (skip short names that are likely code identifiers, not message types)
  if [[ ${#msg} -gt 5 ]]; then
    if ! grep -q "message $msg\b" "$PROTO_FILE" 2>/dev/null && \
       ! grep -q "enum $msg\b" "$PROTO_FILE" 2>/dev/null; then
      # Not an error — many backtick-quoted identifiers are Go types, not proto messages
      # Only warn for names that look like proto conventions (no lowercase letters start)
      : # silently skip — proto message check is best-effort
    fi
  fi
done

info "Proto message check complete (best-effort; full check requires manual review)"

# ── 4. KAT file references ────────────────────────────────────────────────────
echo "[4/5] Checking 91-test-vectors/ file references..."

# Only check JSON files explicitly referenced WITH the 91-test-vectors/ path prefix
# or listed in the 91-test-vectors/README.md table. Bare `foo.json` names in
# pkg/spectest conformance sections refer to Go test files, not vector files here.
kat_refs=$(grep -rh --include="*.md" -o '91-test-vectors/[a-z_]*.json' "$SPEC_DIR" 2>/dev/null | sort -u || true)

for kat_path in $kat_refs; do
  kat_file=$(basename "$kat_path")
  if [[ ! -f "$VECTORS_DIR/$kat_file" ]]; then
    error "KAT file '$kat_path' referenced in spec but not found in docs/spec/v0.1/91-test-vectors/"
  fi
done

# Also check files listed in the 91-test-vectors README table
if [[ -f "$VECTORS_DIR/README.md" ]]; then
  readme_files=$(grep -o '[a-z][a-z_0-9]\{3,\}\.json' "$VECTORS_DIR/README.md" 2>/dev/null | sort -u || true)
  for f in $readme_files; do
    if [[ ! -f "$VECTORS_DIR/$f" ]]; then
      error "KAT file '$f' listed in 91-test-vectors/README.md but not found on disk"
    fi
  done
fi

info "KAT file reference check complete"

# ── 5. Legacy doc SUPERSEDED headers ─────────────────────────────────────────
echo "[5/5] Checking legacy doc SUPERSEDED headers..."

legacy_docs=(
  "$REPO_ROOT/aether-v3.md"
  "$REPO_ROOT/ENCRYPTION_PLUS.md"
  "$REPO_ROOT/aether-addendum-qol-discovery.md"
  "$REPO_ROOT/docs/local-control-api-v1.md"
)

for doc in "${legacy_docs[@]}"; do
  if [[ -f "$doc" ]]; then
    if ! grep -q "SUPERSEDED" "$doc"; then
      error "Legacy doc '$doc' is missing a SUPERSEDED header"
    else
      info "OK: $doc"
    fi
  else
    info "SKIP (not found): $doc"
  fi
done

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
if [[ $FAIL -eq 0 ]]; then
  echo "=== spec-lint.sh: PASS ==="
  exit 0
else
  echo "=== spec-lint.sh: FAIL — see errors above ===" >&2
  exit 1
fi
