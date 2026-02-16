#!/bin/bash
set -euo pipefail

REQUIRED_SOURCES=(
	config/dhall/types.dhall
	config/dhall/default.dhall
	config/dhall/env.dhall
)

for f in "${REQUIRED_SOURCES[@]}"; do
	if [[ ! -f "$f" ]]; then
		echo "missing required dhall source: $f" >&2
		exit 1
	fi
done

ARTIFACT_DIR="artifacts/generated/bootstrap-plan"
PLAN_JSON="$ARTIFACT_DIR/bootstrap-plan.json"
PLAN_TEXT="$ARTIFACT_DIR/bootstrap-plan.txt"
DHALL_EVAL_IMAGE="${DHALL_EVAL_IMAGE:-docker.io/dhallhaskell/dhall:latest}"
DHALL_JSON_IMAGE="${DHALL_JSON_IMAGE:-docker.io/dhallhaskell/dhall-json:latest}"
PYTHON_IMAGE="${PYTHON_IMAGE:-docker.io/library/python:3.11-slim}"

mkdir -p "$ARTIFACT_DIR"

podman run --rm --userns=keep-id -v "$PWD":/work:Z -w /work "$DHALL_EVAL_IMAGE" \
	dhall --file config/dhall/env.dhall >/dev/null

podman run --rm --userns=keep-id -v "$PWD":/work:Z -w /work "$DHALL_JSON_IMAGE" \
	dhall-to-json --file config/dhall/env.dhall >"$PLAN_JSON"

podman run --rm --userns=keep-id -v "$PWD":/work:Z -w /work "$PYTHON_IMAGE" \
	python - <<'PY'
import json
import pathlib

plan_path = pathlib.Path("artifacts/generated/bootstrap-plan/bootstrap-plan.json")
summary_path = pathlib.Path("artifacts/generated/bootstrap-plan/bootstrap-plan.txt")

data = json.loads(plan_path.read_text())
lines = []
for env in data.get("bootstrapEnvironments", []):
    lines.append(f"[{env.get('environment', 'unknown')}]")
    for node in env.get("nodes", []):
        node_cfg = node.get("node", {})
        metrics = node.get("metrics", {})
        health = node.get("health", {})
        listen = ", ".join(node_cfg.get("listen", []))
        announce = ", ".join(node_cfg.get("announce", []))
        lines.append(
            f"- {node_cfg.get('name', 'unknown')} ({node_cfg.get('region', 'n/a')}): "
            f"listen={listen} announce={announce} contact={node_cfg.get('contact', '')}"
        )
        lines.append(
            f"  metrics={metrics.get('listenAddr', '')} "
            f"expectPeers={health.get('expectPeers', '?')} interval={health.get('interval', '')}"
        )
    lines.append("")

summary = "\n".join(line for line in lines if line.strip()) + "\n"
summary_path.write_text(summary)
PY

echo "Bootstrap plan artifacts written to $PLAN_JSON and $PLAN_TEXT"
