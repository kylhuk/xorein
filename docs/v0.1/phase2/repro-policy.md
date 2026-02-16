# Reproducible Build & Provenance Policy (Phase 2)

1. The build pipeline (`make pipeline`) produces deterministic timestamps by writing placeholder artifacts to `artifacts/generated/stamp.txt` and controlling order of stages.
2. Checksums: [`scripts/repro-checksums.sh`](scripts/repro-checksums.sh:1) records SHA256 digests for generated outputs before publishing.
3. Signing policy (planned gate): release artifacts must include signature material and signer identity metadata before publication; unsigned release artifacts are policy violations.
4. SBOM policy (planned gate): SBOM artifacts are emitted to `artifacts/generated/sbom/` with digest sidecars and are referenced from release evidence.
5. Any future automation must log Podman image digests used for tooling to support provenance audits.
