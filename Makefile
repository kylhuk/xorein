SHELL := /bin/bash
BIN_DIR := bin
BUILD_BIN := $(BIN_DIR)/aether
GENERATED_DIR := artifacts/generated
TOOLS_CACHE_DIR := .cache/xorein-tools
TOOLS_BIN_DIR := $(TOOLS_CACHE_DIR)/bin
RELAY_ARTIFACTS_DIR := $(GENERATED_DIR)/relay-container
RELAY_CONTAINERFILE := containers/relay/Containerfile
RELAY_IMAGE_REPO ?= localhost/aether-relay
RELAY_IMAGE_TAG ?= v0.1.0
RELAY_IMAGE := $(RELAY_IMAGE_REPO):$(RELAY_IMAGE_TAG)
RELEASE_PACK_DIR := $(GENERATED_DIR)/release-pack
RELEASE_PACK_SIGN_DIR := $(RELEASE_PACK_DIR)/signing
RELEASE_SIGNING_IMAGE ?= docker.io/library/golang:1.24.8
export PATH := $(CURDIR)/scripts:$(CURDIR)/$(TOOLS_BIN_DIR):$(PATH)

.PHONY: all pipeline conformance check-fast check-full generate compile lint spec-lint test race scan build clean relay-container-workflow relay-container-build relay-container-sign relay-container-sbom relay-container-publish-check release-pack-verify interop

STAGE_ORDER := generate compile lint test race scan build

all: check-full build

pipeline: generate compile lint test race scan build

# conformance: full pipeline + Seal-DM interop check (spec 90)
conformance: pipeline interop

interop:
	@echo "[interop] running Seal-DM interop harness"
	@bash scripts/interop.sh

check-fast: generate compile lint spec-lint
check-full: generate compile lint test race scan

spec-lint:
	@echo "[spec-lint] running spec self-consistency check"
	@bash scripts/spec-lint.sh

generate:
	@echo "[generate] running protobuf compatibility checks"
	@set -euo pipefail; \
	if [[ -f buf.yaml ]]; then \
		buf lint; \
		buf breaking --against '.git#branch=main'; \
	else \
		echo "[generate] skipping buf checks (buf.yaml missing)"; \
	fi

compile:
	@echo "[compile] compiling all Go packages"
	@set -euo pipefail
	@go build ./...

lint:
	@echo "[lint] running hygiene and Go lint checks"
	@set -euo pipefail
	@pre-commit run --all-files
	@golangci-lint run ./...

test:
	@echo "[test] verifying test vector pins"
	@bash scripts/verify-vector-pins.sh
	@echo "[test] running Go test suite"
	@set -euo pipefail
	@go test ./...

race:
	@echo "[race] running Go race test suite"
	@set -euo pipefail
	@go test -race ./...

scan:
	@echo "[scan] running security suite"
	@set -euo pipefail; \
	ART_DIR="$(GENERATED_DIR)/security"; \
	mkdir -p "$$ART_DIR"; \
	echo "[scan] govulncheck ./..."; \
	govulncheck ./... | tee "$$ART_DIR/govulncheck.txt"; \
	echo "[scan] gosec ./..."; \
	go test ./... >/tmp/gotest.log; \
	gosec -fmt text -severity medium -confidence medium ./... | tee "$$ART_DIR/gosec.txt"; \
	echo "[scan] trivy filesystem (vuln+secret)"; \
	TRIVY_CACHE_DIR="$(CURDIR)/.cache/trivy" trivy fs --scanners vuln,secret --security-checks vuln,secret --skip-dirs .git,.cache,artifacts/generated,bin --no-progress --exit-code 1 --severity HIGH,CRITICAL --format json . | tee "$$ART_DIR/trivy-fs.json"

build:
	@echo "[build] building runnable binary into $(BUILD_BIN)"
	@set -euo pipefail
	@mkdir -p "$(BIN_DIR)"
	@rm -f "$(BUILD_BIN)"
	@go build -o "$(BUILD_BIN)" ./cmd/aether

clean:
	@echo "[clean] removing artifacts"
	@rm -rf "$(BIN_DIR)" artifacts

relay-container-workflow: relay-container-build relay-container-sign relay-container-sbom relay-container-publish-check

relay-container-build:
	@echo "[relay-container] building relay image $(RELAY_IMAGE)"
	@set -euo pipefail
	@mkdir -p "$(RELAY_ARTIFACTS_DIR)"
	@podman build --pull=always --file "$(RELAY_CONTAINERFILE)" --tag "$(RELAY_IMAGE)" .
	@podman image inspect "$(RELAY_IMAGE)" --format '{{.Id}}' > "$(RELAY_ARTIFACTS_DIR)/image-id.txt"
	@podman image inspect "$(RELAY_IMAGE)" --format '{{index .RepoDigests 0}}' > "$(RELAY_ARTIFACTS_DIR)/image-digest.txt"

relay-container-sign:
	@echo "[relay-container] emitting signing material"
	@set -euo pipefail
	@test -f "$(RELAY_ARTIFACTS_DIR)/image-digest.txt"
	@sha256sum "$(RELAY_ARTIFACTS_DIR)/image-digest.txt" > "$(RELAY_ARTIFACTS_DIR)/image-digest.txt.sha256"
	@printf "cosign sign --key <key-ref> %s\n" "$(RELAY_IMAGE)@$(shell cat $(RELAY_ARTIFACTS_DIR)/image-digest.txt | sed 's|^.*@||')" > "$(RELAY_ARTIFACTS_DIR)/signing-command.txt"

relay-container-sbom:
	@echo "[relay-container] generating deterministic SBOM placeholder"
	@set -euo pipefail
	@test -f "$(RELAY_ARTIFACTS_DIR)/image-digest.txt"
	@mkdir -p "$(RELAY_ARTIFACTS_DIR)/sbom"
	@printf '{\n  "spdxVersion": "SPDX-2.3",\n  "name": "aether-relay",\n  "image": "%s",\n  "imageDigest": "%s",\n  "documentNamespace": "https://aether.invalid/spdx/relay/%s"\n}\n' "$(RELAY_IMAGE)" "$(shell cat $(RELAY_ARTIFACTS_DIR)/image-digest.txt)" "$(RELAY_IMAGE_TAG)" > "$(RELAY_ARTIFACTS_DIR)/sbom/sbom.spdx.json"
	@sha256sum "$(RELAY_ARTIFACTS_DIR)/sbom/sbom.spdx.json" > "$(RELAY_ARTIFACTS_DIR)/sbom/sbom.spdx.json.sha256"

relay-container-publish-check:
	@echo "[relay-container] writing publication and rollback checklist"
	@set -euo pipefail
	@test -f "$(RELAY_ARTIFACTS_DIR)/image-digest.txt"
	@test -f "$(RELAY_ARTIFACTS_DIR)/image-digest.txt.sha256"
	@test -f "$(RELAY_ARTIFACTS_DIR)/sbom/sbom.spdx.json"
	@printf '%s\n' \
		'Publication gates:' \
		'1. Build uses pinned source commit and records image digest.' \
		'2. Digest checksum is present (image-digest.txt.sha256).' \
		'3. SBOM + SBOM checksum are present.' \
		'4. Signing command is recorded and must be executed before push.' \
		'' \
		'Rollback:' \
		'- Re-deploy previous known-good digest from release history.' \
		'- Re-point deployment manifests from candidate digest to prior digest.' \
		'- Confirm rollback digest health before unpausing rollout.' \
		> "$(RELAY_ARTIFACTS_DIR)/publication-checklist.txt"

release-pack-verify: build
	@echo "[release-pack] generating reproducible verification bundle"
	@./scripts/release-pack-verify.sh
