# P9-T5 Relay Container Build and Publication Workflow

## Scope

This document defines the implemented relay container workflow for v0.1 execution scope: deterministic build inputs, digest capture, signing handoff, SBOM handling, publish gates, and rollback guidance.

## Image tag and digest policy

1. Relay image coordinates are Makefile-driven:
   - `RELAY_IMAGE_REPO` default: `localhost/aether-relay`
   - `RELAY_IMAGE_TAG` default: `v0.1.0`
2. Build uses digest-pinned bases in [`containers/relay/Containerfile`](containers/relay/Containerfile).
3. Build output records:
   - image ID: `artifacts/generated/relay-container/image-id.txt`
   - repo digest: `artifacts/generated/relay-container/image-digest.txt`
4. Publication is digest-first. Deploy and rollback references must use immutable digest form (`repo@sha256:...`), not mutable tags.

## Publication gating requirements

Workflow entrypoint: `make relay-container-workflow`.

Stages:
1. `relay-container-build`
   - Builds image via Podman from pinned Containerfile.
   - Captures image ID + repo digest as artifacts.
2. `relay-container-sign`
   - Produces digest checksum sidecar (`image-digest.txt.sha256`).
   - Emits signing command template for operator key-backed signing.
3. `relay-container-sbom`
   - Emits SPDX JSON artifact at `artifacts/generated/relay-container/sbom/sbom.spdx.json`.
   - Emits SBOM checksum sidecar.
4. `relay-container-publish-check`
   - Verifies digest/checksum/SBOM prerequisites exist.
   - Writes publication checklist and rollback steps.

## Rollout and rollback guidance

Rollout gate:
- Push/publish only after digest capture, digest checksum, SBOM + SBOM checksum, and signing command generation all succeed.

Rollback:
1. Re-deploy prior known-good digest from release history.
2. Re-point deployment manifests from candidate digest to previous digest.
3. Verify relay health on rollback digest before unpausing rollout.

## Implemented artifacts

- [`Makefile`](Makefile) targets:
  - `relay-container-workflow`
  - `relay-container-build`
  - `relay-container-sign`
  - `relay-container-sbom`
  - `relay-container-publish-check`
- [`containers/relay/Containerfile`](containers/relay/Containerfile) with digest-pinned base images.
- Generated outputs under `artifacts/generated/relay-container/` from Podman execution.
