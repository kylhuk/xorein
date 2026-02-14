# Reproducibility Tooling Notes (Phase 2)

1. [`scripts/repro-checksums.sh`](scripts/repro-checksums.sh:1) generates SHA256 digests and lives in the source tree so downstream automation can capture checksum manifests.
2. [`scripts/dhall-verify.sh`](scripts/dhall-verify.sh:1) validates presence of Dhall sources and emits deterministic placeholder verification output inside Podman.
3. Signature workflow (planned): publish detached signatures plus signer metadata next to release artifacts.
4. SBOM workflow (planned): emit SBOMs under `artifacts/generated/sbom/` and verify checksums before publication.
5. Tooling provenance: record Podman image digest and command line in evidence for checksum/signature/SBOM-producing steps.
