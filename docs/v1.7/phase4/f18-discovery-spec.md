# Phase 4 — F18 Discovery Spec

## Directory Publication Model
- Relays publish signed `DirectoryEntry` records that contain the namespace, canonical alias, and versioned fingerprint.
- Each entry includes the signing key ID and a deterministic timestamp to avoid replay ambiguity.

## Indexer Response Contract
- Clients query indexers with anchored `DirectoryEntry` fingerprints.
- Responses are signed and include a freshness window plus the enforced trust flag for compliance-aware clients.

## Join Funnel Controls
- Join requests are categorized as `invite`, `request`, or `open`; each path carries rate-limit metadata and abuse labels.
- The server attaches deterministic `TrustWarning` strings whenever a join decision defers or rejects, enabling consistent UX flows.
