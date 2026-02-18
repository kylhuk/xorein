# Phase 4 — F18 Proto Delta

Currently v1.7 does not change any protobuf schemas. The F18 plan reserves the following additions for v1.8 implementation:

1. Add a signed `DirectoryEntry` message with fields for namespace, alias, fingerprint, and signer metadata.
2. Introduce an `IndexerResponse` wrapper that bundles signed responses with freshness and enforcement flags.
3. Define `JoinPath` enums (`invite`, `request`, `open`) and overlay rate-limiting metadata with deterministic rejection reasons.

Future proto changes will reuse new field numbers next available after existing v17 fields, and all removed numbers are `reserved` to preserve wire compatibility.
