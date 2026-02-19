# P2-T1 Provider Protocol

## ST1 → G3/G4
- Define PutBlobChunk/GetBlobChunk/PutManifest/GetManifest endpoint contracts that enforce deterministic chunk ordering, size checks, and manifest completion state. These contracts are gate evidence for G3 (blob upload/download) and G4 (provider behavior) because they capture the handshake between clients and storage providers.
- Optional quota/retention helpers expose hinty usage data without leaking enforcement logic, supporting G4’s quota and retention determinism.

## ST2 → G3
- Resumable transfers rely on the chunk presence query that is safe for private spaces (authorization gating) and deterministic resume tokens. Gate G3 flags this behavior because it tracks bounded transfer progress and retry/backoff tokens that are modeled deterministically.
- Concurrency and retry bounds are enforced by the ordered chunk model: only the next chunk index is accepted, so retries and backoff are predictable.

## ST3 → G4
- Refusal reasons are deterministic (invalid chunk order, chunk size mismatch, unsupported mime/profile), enabling G4’s provider refusal taxonomy. Every refusal path is documented so quota/retention enforcement and provider replication policies can rely on unambiguous failure semantics.
