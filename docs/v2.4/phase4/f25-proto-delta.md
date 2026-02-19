# F25 Proto Delta

Planning-only additive delta to describe what proto assets `F25` will publish. No actual wire changes occur in v24; this document simply records the future additive steps so implementers understand the target surface without breaking existing clients.

## Additive intent

- Introduce `BlobRef` message with fields for `hash`, `size`, `mime`, chunk metadata, encryption envelope, and access control.
- Define `BlobUploadRequest`/`BlobDownloadRequest` RPC payloads that remain optional in the local API until v25 implementation catches up.
- Add `StorageCapability` descriptors that can annotate `Archivist`-like providers, including optional chunking and quota hints.
- Extend the local API service with new RPCs `DescribeBlob`, `ListBlobRefs`, and `RetrieveBlobReference` guarded by capability bits (added as new optional enums).

## Compatibility guarantees

All fields referenced above are additive-only: no existing message field numbers are reused, no field types are altered, and no services are renamed. Future clients can add these definitions while older clients ignore them because they appear behind optional `oneof`s, `repeated`s, or new RPCs with optional request/response fields.

No renumbering, no breaking changes, and no `required` declarations appear in the proposed `F25` delta so both v24 and v25 clients can evolve safely.
