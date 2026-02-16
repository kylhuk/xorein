SHELL := /bin/bash
BIN_DIR := bin
BUILD_BIN := $(BIN_DIR)/aether
GENERATED_DIR := artifacts/generated
RELAY_ARTIFACTS_DIR := $(GENERATED_DIR)/relay-container
RELAY_CONTAINERFILE := containers/relay/Containerfile
RELAY_IMAGE_REPO ?= localhost/aether-relay
RELAY_IMAGE_TAG ?= v0.1.0
RELAY_IMAGE := $(RELAY_IMAGE_REPO):$(RELAY_IMAGE_TAG)
RELEASE_PACK_DIR := $(GENERATED_DIR)/release-pack
RELEASE_PACK_SIGN_DIR := $(RELEASE_PACK_DIR)/signing
RELEASE_SIGNING_IMAGE ?= docker.io/library/golang:1.24.8

.PHONY: all pipeline check-fast check-full generate compile lint test scan build clean relay-container-workflow relay-container-build relay-container-sign relay-container-sbom relay-container-publish-check release-pack-verify

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
	@echo "[scan] running security suite"
	@set -euo pipefail; \
	ART_DIR="$(GENERATED_DIR)/security"; \
	mkdir -p "$$ART_DIR"; \
	echo "[scan] govulncheck ./..."; \
	podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/golang:1.24.8 bash -lc 'set -euo pipefail; export PATH="/usr/local/go/bin:/go/bin:$$PATH"; export GOMODCACHE=/tmp/gomodcache; go install golang.org/x/vuln/cmd/govulncheck@v1.1.4; govulncheck ./...' | tee "$$ART_DIR/govulncheck.txt"; \
	echo "[scan] gosec ./..."; \
	podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/golang:1.24.8 bash -lc 'set -euo pipefail; export PATH="/usr/local/go/bin:/go/bin:$$PATH"; export GOMODCACHE=/tmp/gomodcache; go install github.com/securego/gosec/v2/cmd/gosec@v2.22.2; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./... >/tmp/gotest.log; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 gosec -fmt text -severity medium -confidence medium ./...' | tee "$$ART_DIR/gosec.txt"; \
	echo "[scan] trivy filesystem (vuln+secret)"; \
	podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/aquasec/trivy:0.50.0 fs --scanners vuln,secret --security-checks vuln,secret --skip-dirs /workspace/.git --no-progress --exit-code 1 --severity HIGH,CRITICAL --format json /workspace | tee "$$ART_DIR/trivy-fs.json"

build:
	@echo "[build] packaging binaries into $(BUILD_BIN)"
	@set -euo pipefail
	@podman run --rm --userns=keep-id -v "$(PWD)":"/workspace":Z -w "/workspace" docker.io/library/busybox:1.36.1 \
		sh -c 'mkdir -p "$(BIN_DIR)" && printf "aether-phase2-placeholder-binary\n" > "$(BUILD_BIN)"'

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

release-pack-verify:
	@echo "[release-pack] generating reproducible verification bundle"
	@./scripts/release-pack-verify.sh
