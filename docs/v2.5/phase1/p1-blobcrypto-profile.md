# Blobcrypto Profile

This artifact captures the v25 Phase 1 encryption envelope expectations and references the helper APIs that implement the scope-constrained, deterministic handling discussed in the TODO. Generating a per-blob AEAD key, wrapping it under the right scope key, and deriving chunk nonces are the primitive operations.

## Envelope semantics
- `GenerateDEK` produces a fresh 32-byte DEK for each blob and is paired with `WrapSpaceDEK` or `WrapDMSessionDEK` depending on the asset scope.
- `Envelope` stores `scope`, `keyId`, the wrapped ciphertext, and the wrap nonce; `UnwrapDEK` verifies that the registry still recognizes the wrapping key before returning the DEK.
- Each registered scope has a dedicated wrapping key (`ScopeSpaceAsset` for spaces, `ScopeDMSession` for DMs) so that the envelope carries the correct authorization domain.

## Deterministic refusal taxonomy
- `missing_key_material`: there is no registered key for the requested scope.
- `invalid_envelope`: envelope structure or key IDs do not match the registry.
- `key_revoked`: the rotation hook reported a revoked key ID.
- `auth_failure`: AEAD authentication failed while unwrapping the DEK.

## Key registry and rotation hook
- `KeyRegistry` holds the scope → key material map and accepts an optional `RotationHook` that can return an error whenever the caller sees a retired key ID. No rotation is forced in v25, but the hook provides the future rotation/rollback check interface.
- Wrap helpers (`WrapSpaceDEK`, `WrapDMSessionDEK`) and `UnwrapDEK` share the same registry so client code can short-circuit once a scope key is missing or marked revoked.

## Chunk encryption helpers
- `DeriveChunkNonce` deterministically combines the blob hash with the chunk index, guaranteeing a repeatable 12-byte nonce for AES-GCM given the `aead.NonceSize()`.
- `EncryptChunk` / `DecryptChunk` plug into the derived nonce to seal per-chunk ciphertext under the blob’s DEK, and callers may pass additional associated data as needed.

## Evidence commands
Use the helper command below to exercise this profile and capture the evidence noted in the TODO when running G2-phase proofs:

- `go test ./pkg/v25/blobcrypto`
- `go test ./tests/e2e/v25/blobcrypto_envelope_test.go`
