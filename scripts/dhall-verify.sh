#!/bin/bash
set -euo pipefail

for f in config/dhall/types.dhall config/dhall/default.dhall config/dhall/env.dhall; do
  if [[ ! -f "$f" ]]; then
    echo "missing required dhall source: $f" >&2
    exit 1
  fi
done

podman run --rm --userns=keep-id -v "$PWD":"/work":Z -w "/work" docker.io/library/busybox:1.36.1 \
  sh -lc 'set -eu; echo "dhall verification placeholder: config sources present"'
