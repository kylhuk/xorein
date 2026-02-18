# Build Attestation

- Build artifacts signed and hashed per release policy.
- Podman-friendly layout described in `containers/v2.0/docker-compose.yml` ensures reproducible registry images.
- Attestation records stored under `artifacts/generated/v20-podman-scenarios/result-manifest.json` after running `scripts/v20-podman-scenarios.sh`.
