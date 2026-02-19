#!/usr/bin/env bash
set -euo pipefail

CHECK_ROOT="${CHECK_ROOT:-$(cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)" && pwd)}"

violations=()

report_matches() {
  local label="$1"
  local matches
  matches="$2"
  if [ -z "$matches" ]; then
    return
  fi
  while IFS= read -r match; do
    if [ -z "$match" ]; then
      continue
    fi
    violations+=("$label: $match")
  done <<< "$matches"
}

scan_dir() {
  local directory="$1"
  local pattern="$2"
  local label="$3"
  local matches
  if [ ! -d "$directory" ]; then
    return
  fi
  while IFS= read -r -d $'\0' file; do
    matches="$(grep -nE "$pattern" "$file" || true)"
    report_matches "$label" "$matches"
  done < <(find "$directory" -name '*.go' -print0)
}

check_st1() {
  local label="ST1 (Gio import)"
  local pattern='"gioui\.org[^" ]*"'
  local path
  for path in "$CHECK_ROOT/cmd/xorein" "$CHECK_ROOT/pkg/xorein"; do
    scan_dir "$path" "$pattern" "$label"
  done
}

check_st2() {
  local label="ST2 (protocol runtime import)"
  local path="$CHECK_ROOT/cmd/harmolyn"
  if [ ! -d "$path" ]; then
    return
  fi
  local forbidden=(
    "github.com/aether/code_aether/pkg/xorein"
    "github.com/aether/code_aether/pkg/v24/daemon"
    "github.com/aether/code_aether/pkg/v23"
    "github.com/aether/code_aether/pkg/v22"
    "github.com/aether/code_aether/pkg/v21"
    "github.com/aether/code_aether/pkg/v20"
    "github.com/aether/code_aether/pkg/v19"
    "github.com/aether/code_aether/pkg/v18"
    "github.com/aether/code_aether/pkg/phase5"
    "github.com/aether/code_aether/pkg/protocol"
  )
  local pattern
  local matches
  while IFS= read -r -d $'\0' file; do
    for pattern in "${forbidden[@]}"; do
      matches="$(grep -n "$pattern" "$file" || true)"
      report_matches "$label" "$matches"
    done
  done < <(find "$path" -name '*.go' -print0)
}

main() {
  check_st1
  check_st2

  if [ ${#violations[@]} -eq 0 ]; then
    return 0
  fi

  printf 'Boundary enforcement violations (%d)\n' "${#violations[@]}" >&2
  local violation
  for violation in "${violations[@]}"; do
    printf '  %s\n' "$violation" >&2
  done
  exit 1
}

main
