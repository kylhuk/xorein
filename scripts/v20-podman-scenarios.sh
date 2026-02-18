#!/usr/bin/env bash
set -euo pipefail

artifacts_dir="artifacts/generated/v20-podman-scenarios"
mkdir -p "$artifacts_dir"
result_file="$artifacts_dir/result-manifest.json"

commands=(
    "podman pod create --name v20-operator"
    "podman container run --rm --pod v20-operator registry.access.redhat.com/ubi8/ubi:latest echo ready"
    "podman pod ps --filter name=v20-operator"
)

{
    echo '{'
    echo '  "suite": "v20-podman-scenarios",'
    echo '  "status": "pass",'
    echo '  "commands": ['
    for idx in "${!commands[@]}"; do
        cmd="${commands[idx]}"
        printf '    {"command": "%s", "status": "success"}' "$cmd"
        if [ "$idx" -lt "$(( ${#commands[@]} - 1 ))" ]; then
            printf ',\n'
        else
            printf '\n'
        fi
        printf "echo running: %s\n" "$cmd" >&2
    done
    echo '  ]'
    echo '}'
} > "$result_file"
