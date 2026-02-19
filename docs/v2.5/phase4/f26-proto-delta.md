# F26 Proto Delta

Planning-only additive delta that defines the protobuffer surface `F26` will publish in order to enable the final closure workflow (gate `G8`). This file does not carry live wire changes; it merely records the future grammar so older versions can safely ignore the new definitions.

## Additive intent

- Surface a `FinalClosureStatus` message that lists gates (`G0`–`G10`), their status, and optional evidence references. The message remains optional until v26 consumes it.
- Introduce `ReleaseArtifactCatalog` and `ReleaseArtifactEntry` metadata so `F26` can advertise the reproducible-build catalog and SHA-256 fingerprints without renaming existing services.
- Extend the telemetry envelope with `EvidenceIndexRef` fields so clients and QA systems can fetch the `EV-v25-G8-###` records in a standardized way.
- Publish `GateSignature` with signer identity, gate ID, timestamp, and digest (optional) to seal the final conformance check while remaining additive.

## Compatibility guarantees

All fields enumerated above are additive: new messages, optional fields, and repeated entries only. No existing field numbers are reused, no services are renamed, and the new messages stay optional so legacy clients and relays continue to operate without being aware of the closure metadata.

Future versions may extend `ReleaseArtifactEntry` with additional metadata (e.g., `artifact_type`, `release_candidate`), but v25 freezes only the fields listed above to keep `G8` additive and backward compatible.
