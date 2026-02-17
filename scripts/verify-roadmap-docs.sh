#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

versions=(11 12 13 14 15 16 17 18 19 20)
required_headings=(
  "## Version Isolation Contract (mandatory)"
  "## Entry criteria (must be true before implementation starts)"
  "## Promotion gates (must all pass)"
  "## Mandatory command evidence (attach exact outputs in Phase 5)"
  "## Roadmap conformance templates (mandatory)"
  "## Phase plan"
  "## Risk register"
  "## Decision log"
)
required_template_refs=(
  "docs/templates/roadmap-gate-checklist.md"
  "docs/templates/roadmap-signoff-raci.md"
  "docs/templates/roadmap-evidence-index.md"
  "docs/templates/roadmap-deferral-register.md"
)

fail_count=0

require_in_file() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$file"; then
    printf 'ERROR: %s missing "%s"\n' "$file" "$needle"
    fail_count=$((fail_count + 1))
  fi
}

for v in "${versions[@]}"; do
  file="TODO_v${v}.md"

  if [[ ! -f "$file" ]]; then
    printf 'ERROR: missing %s\n' "$file"
    fail_count=$((fail_count + 1))
    continue
  fi

  for heading in "${required_headings[@]}"; do
    require_in_file "$file" "$heading"
  done

  for ref in "${required_template_refs[@]}"; do
    require_in_file "$file" "$ref"
  done

  require_in_file "$file" "Relay no-data-hosting"
  require_in_file "$file" "as-built conformance"
  require_in_file "$file" "p0-gate-ownership.md"
  require_in_file "$file" "p0-traceability-matrix.md"
  require_in_file "$file" "p5-gate-signoff.md"
  require_in_file "$file" "p5-evidence-index.md"
  require_in_file "$file" "EV-v${v}-GX-###"

done

for v in 11 12 13 14 15 16 17 18 19; do
  file="TODO_v${v}.md"
  next=$((v + 1))
  require_in_file "$file" "f${next}-acceptance-matrix.md"
done

require_in_file "TODO_v20.md" "f21-acceptance-matrix.md"
require_in_file "TODO_v20.md" "f21-deferral-register.md"

if [[ $fail_count -ne 0 ]]; then
  printf 'Roadmap verification failed with %d issue(s).\n' "$fail_count"
  exit 1
fi

printf 'Roadmap verification passed for TODO_v11.md..TODO_v20.md\n'
