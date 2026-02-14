SHELL := /bin/bash
BIN_DIR := bin
BUILD_BIN := $(BIN_DIR)/aether
GENERATED_DIR := artifacts/generated

.PHONY: all pipeline check-fast check-full generate compile lint test scan build clean

STAGE_ORDER := generate compile lint test scan

all: check-full build

pipeline: generate compile lint test scan build

check-fast: generate compile lint
check-full: generate compile lint test scan

generate:
	@echo "[generate] placeholder for proto/Dhall generation"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/busybox:1.36.1 \
		sh -c 'mkdir -p "$(GENERATED_DIR)" && printf "phase2-scaffold-generated\n" > "$(GENERATED_DIR)/stamp.txt"'

compile:
	@echo "[compile] validating workspace readiness"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/work":Z -w "/work" docker.io/library/busybox:1.36.1 sh -c "test -f cmd/aether/main.go && echo compile placeholder"

lint:
	@echo "[lint] running baseline checks"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/work":Z -w "/work" docker.io/library/busybox:1.36.1 sh -c "test -f .golangci.yml && echo golangci-lint placeholder"

test:
	@echo "[test] running deterministic repro scaffolds"
	@set -euo pipefail
	@./scripts/dhall-verify.sh
	@./scripts/repro-checksums.sh

scan:
	@echo "[scan] compliance scan placeholder"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/busybox:1.36.1 sh -c "echo scan placeholder && test -f docs/v0.1/phase2/repro-policy.md"

build:
	@echo "[build] packaging binaries into $(BUILD_BIN)"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/busybox:1.36.1 \
		sh -c 'mkdir -p "$(BIN_DIR)" && printf "aether-phase2-placeholder-binary\n" > "$(BUILD_BIN)"'

clean:
	@echo "[clean] removing artifacts"
	@rm -rf "$(BIN_DIR)" artifacts
