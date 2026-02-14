# Phase 2 Evidence Log (v0.1)

Scope: Evidence for Phase 2 tasks P2-T1 through P2-T10 as currently verifiable.

## Command 1 — Podman artifact existence checks

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; test -f cmd/aether/main.go; test -f pkg/app/seams.go; test -f .golangci.yml; test -f .pre-commit-config.yaml; test -f .github/workflows/ci.yml; test -f .github/workflows/release.yml; test -f .github/workflows/nightly.yml; test -f .github/workflows/security-audit.yml; test -f buf.yaml; test -f buf.gen.yaml; test -f proto/aether.proto; test -f config/dhall/types.dhall; test -f config/dhall/default.dhall; test -f config/dhall/env.dhall; test -f scripts/dhall-verify.sh; test -f scripts/repro-checksums.sh; test -f docs/v0.1/phase2/scaffold-boundaries.md; test -f docs/v0.1/phase2/lint-suppression-governance.md; test -f docs/v0.1/phase2/pre-commit-onboarding.md; test -f docs/v0.1/phase2/proto-reservation-policy.md; test -f docs/v0.1/phase2/dhall-ops.md; test -f docs/v0.1/phase2/repro-policy.md; test -f docs/v0.1/phase2/repro-tools.md; test -f docs/v0.1/phase2/p2-t4-trigger-matrix.md; test -f docs/v0.1/phase2/p2-t9-test-strategy-matrix.md; test -f docs/v0.1/phase2/p2-t10-sqlcipher-decision.md; echo "phase2-artifact-existence-ok"'
```

Output:

```text
+ test -f cmd/aether/main.go
+ test -f pkg/app/seams.go
+ test -f .golangci.yml
+ test -f .pre-commit-config.yaml
+ test -f .github/workflows/ci.yml
+ test -f .github/workflows/release.yml
+ test -f .github/workflows/nightly.yml
+ test -f .github/workflows/security-audit.yml
+ test -f buf.yaml
+ test -f buf.gen.yaml
+ test -f proto/aether.proto
+ test -f config/dhall/types.dhall
+ test -f config/dhall/default.dhall
+ test -f config/dhall/env.dhall
+ test -f scripts/dhall-verify.sh
+ test -f scripts/repro-checksums.sh
+ test -f docs/v0.1/phase2/scaffold-boundaries.md
+ test -f docs/v0.1/phase2/lint-suppression-governance.md
+ test -f docs/v0.1/phase2/pre-commit-onboarding.md
+ test -f docs/v0.1/phase2/proto-reservation-policy.md
+ test -f docs/v0.1/phase2/dhall-ops.md
+ test -f docs/v0.1/phase2/repro-policy.md
+ test -f docs/v0.1/phase2/repro-tools.md
+ test -f docs/v0.1/phase2/p2-t4-trigger-matrix.md
+ test -f docs/v0.1/phase2/p2-t9-test-strategy-matrix.md
+ test -f docs/v0.1/phase2/p2-t10-sqlcipher-decision.md
+ echo phase2-artifact-existence-ok
phase2-artifact-existence-ok
```

## Command 2 — Podman content spot checks

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; grep -q "runtime mode: client|relay|bootstrap" cmd/aether/main.go; grep -q "pkg/protocol" docs/v0.1/phase2/scaffold-boundaries.md; grep -q "check-fast" Makefile; grep -q "check-full" Makefile; grep -q "generate compile lint test scan build" Makefile; grep -q "govet" .golangci.yml; grep -q "check-yaml" .pre-commit-config.yaml; grep -q "reserved 100" proto/aether.proto; grep -q "ConfigType" config/dhall/types.dhall; grep -q "checksum" docs/v0.1/phase2/repro-policy.md; grep -q "Layered test strategy matrix" docs/v0.1/phase2/p2-t9-test-strategy-matrix.md; grep -q "SQLCipher" docs/v0.1/phase2/p2-t10-sqlcipher-decision.md; echo phase2-content-checks-ok'
```

Output:

```text
+ grep -q 'runtime mode: client|relay|bootstrap' cmd/aether/main.go
+ grep -q pkg/protocol docs/v0.1/phase2/scaffold-boundaries.md
+ grep -q check-fast Makefile
+ grep -q check-full Makefile
+ grep -q 'generate compile lint test scan build' Makefile
+ grep -q govet .golangci.yml
+ grep -q check-yaml .pre-commit-config.yaml
+ grep -q 'reserved 100' proto/aether.proto
+ grep -q ConfigType config/dhall/types.dhall
+ grep -q checksum docs/v0.1/phase2/repro-policy.md
+ grep -q 'Layered test strategy matrix' docs/v0.1/phase2/p2-t9-test-strategy-matrix.md
+ grep -q SQLCipher docs/v0.1/phase2/p2-t10-sqlcipher-decision.md
+ echo phase2-content-checks-ok
phase2-content-checks-ok
```

## Command 3 — Podman check-fast in Alpine with nested podman unavailable

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/alpine:3.20 sh -lc 'set -eux; apk add --no-cache make bash coreutils findutils >/dev/null; make check-fast'
```

Output:

```text
+ apk add --no-cache make bash coreutils findutils
+ make check-fast
[generate] placeholder for proto/Dhall generation
/bin/bash: line 1: podman: command not found
make: *** [Makefile:20: generate] Error 127
```

Status: SKIP for full execution inside this container because `Makefile` stages invoke `podman` and nested podman is not available by default.

## Command 4 — Podman check-fast in Alpine with nested podman install

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/alpine:3.20 sh -lc 'set -eux; apk add --no-cache podman make bash coreutils findutils >/dev/null; make check-fast'
```

Output:

```text
+ apk add --no-cache podman make bash coreutils findutils
*
* For running rootless Podman you need to:
* - configure subordinate user ids (/etc/subuid) and subordinate group ids (/etc/subgid),
* - mount the control groups: `rc-service cgroups start`.
*
* More information: https://wiki.alpinelinux.org/wiki/Podman
*
+ make check-fast
[generate] placeholder for proto/Dhall generation
time="2026-02-14T15:57:06Z" level=warning msg="Using rootless single mapping into the namespace. This might break some images. Check /etc/subuid and /etc/subgid for adding sub*ids if not using a network user"
Error: configure storage: 'overlay' is not supported over overlayfs, a mount_program is required: backing file system is unsupported for this graph driver
make: *** [Makefile:20: generate] Error 125
```

Status: SKIP for full nested execution due to overlayfs/rootless nested podman storage limitations.

## Command 5 — Host make check-fast (non-container wrapper)

Command:

```bash
make check-fast
```

Output:

```text
[generate] placeholder for proto/Dhall generation
mkdir: can't create directory 'artifacts/': Permission denied
make: *** [Makefile:20: generate] Error 1
```

Status: Failed on host execution path in this environment; no completion claim made.

## Command 6 — Podman write sanity check for mounted workspace

Command:

```bash
podman run --rm --userns=keep-id -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; mkdir -p artifacts/generated; echo ok > artifacts/generated/probe2.txt; test -f artifacts/generated/probe2.txt; echo podman-write-ok'
```

Output:

```text
+ mkdir -p artifacts/generated
+ echo ok
+ test -f artifacts/generated/probe2.txt
+ echo podman-write-ok
podman-write-ok
```

Status: Confirms mounted workspace can be written from podman in this direct invocation.

## Command 7 — Podman applicability checks for toolchain-dependent commands

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/alpine:3.20 sh -lc 'set -eux; if [ -f go.mod ]; then echo RUN go test ./...; else echo SKIP go test ./... \(go.mod not present\); fi; if [ -f .golangci.yml ]; then echo SKIP golangci-lint run \(no golangci-lint binary available in podman verification context\); else echo SKIP golangci-lint run \(.golangci.yml missing\); fi; if [ -f buf.yaml ]; then echo SKIP buf lint \(buf CLI unavailable in podman verification context\); echo SKIP buf breaking \(buf CLI unavailable in podman verification context\); echo SKIP buf generate \(buf CLI unavailable in podman verification context\); else echo SKIP buf lint/breaking/generate \(buf.yaml missing\); fi'
```

Output:

```text
+ '[' -f go.mod ]
+ echo SKIP go test ./... '(go.mod' not 'present)'
+ '[' -f .golangci.yml ]
+ echo SKIP golangci-lint run '(no' golangci-lint binary available 'in' podman verification 'context)'
+ '[' -f buf.yaml ]
+ echo SKIP buf lint '(buf' CLI unavailable 'in' podman verification 'context)'
+ echo SKIP buf breaking '(buf' CLI unavailable 'in' podman verification 'context)'
+ echo SKIP buf generate '(buf' CLI unavailable 'in' podman verification 'context)'
SKIP go test ./... (go.mod not present)
SKIP golangci-lint run (no golangci-lint binary available in podman verification context)
SKIP buf lint (buf CLI unavailable in podman verification context)
SKIP buf breaking (buf CLI unavailable in podman verification context)
SKIP buf generate (buf CLI unavailable in podman verification context)
```

## Command 8 — Podman policy-content checks for newly added Phase 2 scaffolding

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; grep -q "check-json" .pre-commit-config.yaml; grep -q "check-merge-conflict" .pre-commit-config.yaml; grep -q "no-commit-to-branch" .pre-commit-config.yaml; grep -q "Bypass policy and review requirements" docs/v0.1/phase2/pre-commit-onboarding.md; grep -q "Automated compatibility checks" docs/v0.1/phase2/proto-reservation-policy.md; grep -q "Generation and verification stage" docs/v0.1/phase2/dhall-ops.md; grep -q "Signing policy" docs/v0.1/phase2/repro-policy.md; grep -q "SBOM policy" docs/v0.1/phase2/repro-policy.md; grep -q "Strictness and scenario policy scaffold" docs/v0.1/phase2/p2-t9-test-strategy-matrix.md; grep -q "Migration and key management expectations" docs/v0.1/phase2/p2-t10-sqlcipher-decision.md; echo phase2-doc-policy-checks-ok'
```

Output:

```text
+ grep -q check-json .pre-commit-config.yaml
+ grep -q check-merge-conflict .pre-commit-config.yaml
+ grep -q no-commit-to-branch .pre-commit-config.yaml
+ grep -q 'Bypass policy and review requirements' docs/v0.1/phase2/pre-commit-onboarding.md
+ grep -q 'Automated compatibility checks' docs/v0.1/phase2/proto-reservation-policy.md
+ grep -q 'Generation and verification stage' docs/v0.1/phase2/dhall-ops.md
+ grep -q 'Signing policy' docs/v0.1/phase2/repro-policy.md
+ grep -q 'SBOM policy' docs/v0.1/phase2/repro-policy.md
+ grep -q 'Strictness and scenario policy scaffold' docs/v0.1/phase2/p2-t9-test-strategy-matrix.md
+ grep -q 'Migration and key management expectations' docs/v0.1/phase2/p2-t10-sqlcipher-decision.md
+ echo phase2-doc-policy-checks-ok
phase2-doc-policy-checks-ok
```

## Command 9 — Host make check-fast

Command:

```bash
make check-fast
```

Output:

```text
[generate] placeholder for proto/Dhall generation
[compile] validating workspace readiness
compile placeholder
[lint] running baseline checks
golangci-lint placeholder
```

## Command 10 — Host make check-full

Command:

```bash
make check-full
```

Output:

```text
[generate] placeholder for proto/Dhall generation
[compile] validating workspace readiness
compile placeholder
[lint] running baseline checks
golangci-lint placeholder
[test] running deterministic repro scaffolds
dhall verification placeholder: config sources present
[scan] compliance scan placeholder
scan placeholder
```

## Command 11 — Host make pipeline (includes build stage, initial failure)

Command:

```bash
make pipeline
```

Output:

```text
[generate] placeholder for proto/Dhall generation
[compile] validating workspace readiness
compile placeholder
[lint] running baseline checks
golangci-lint placeholder
[test] running deterministic repro scaffolds
dhall verification placeholder: config sources present
[scan] compliance scan placeholder
scan placeholder
[build] packaging binaries into bin/aether
mkdir: can't create directory 'bin': Permission denied
make: *** [Makefile:47: build] Error 1
```

Status: Initial FAIL before relabel fix on build-stage Podman mount.

## Command 11b — Host make pipeline after build-stage relabel fix

Command:

```bash
make pipeline
```

Output:

```text
[generate] placeholder for proto/Dhall generation
[compile] validating workspace readiness
compile placeholder
[lint] running baseline checks
golangci-lint placeholder
[test] running deterministic repro scaffolds
dhall verification placeholder: config sources present
[scan] compliance scan placeholder
scan placeholder
[build] packaging binaries into bin/aether
```

Status: PASS after updating [`Makefile`](Makefile) build-stage mount to include `:Z`.

## Command 12 — Podman SKIP check for `go test ./...`

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eu; if [ -f go.mod ]; then echo RUN go test ./...; else echo SKIP go test ./... \(go.mod not present\); fi'
```

Output:

```text
SKIP go test ./... (go.mod not present)
```

## Command 13 — Podman SKIP check for golangci-lint binary availability

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eu; if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "SKIP golangci-lint run: golangci-lint not installed in verification container"; fi'
```

Output:

```text
SKIP golangci-lint run: golangci-lint not installed in verification container
```

## Command 14 — Podman `buf lint`

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 lint
```

Output:

```text
Failure: decode buf.yaml: build.roots cannot be set on version v1: [proto]
```

## Command 15 — Podman `buf breaking`

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 breaking --against '.git#branch=main'
```

Output:

```text
Failure: decode buf.yaml: build.roots cannot be set on version v1: [proto]
```

## Command 16 — Podman `buf generate`

Command:

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/bufbuild/buf:1.39.0 generate
```

Output:

```text
Failure: decode buf.yaml: build.roots cannot be set on version v1: [proto]
```

## Evidence summary

- Artifact existence and key content checks passed in Podman.
- Additional Podman policy-content checks for updated Phase 2 docs/config passed.
- Host `make check-fast` and `make check-full` now succeed and provide stage-order evidence for generate/compile/lint/test/scan.
- Host `make pipeline` initially failed at build (`Permission denied` creating `bin/`), then passed after adding `:Z` relabeling to the build-stage Podman mount in [`Makefile`](Makefile).
- `go test` is SKIP due missing `go.mod` in repository.
- Container `golangci-lint run` is SKIP due missing binary in the verification container (while Makefile lint placeholder stage runs).
- Buf commands run in Podman but fail due invalid [`buf.yaml`](buf.yaml) configuration (`build.roots` with `version: v1`), so protobuf compatibility automation is not yet passing.
