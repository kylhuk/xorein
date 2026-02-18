# v1.2 Phase 2 - Proto Changelog

## Status
Implemented additive schema updates in `proto/aether.proto`.

## Added enum
- `BackupReason`
  - `BACKUP_REASON_UNSPECIFIED`
  - `BACKUP_REASON_PASSWORD`
  - `BACKUP_REASON_CORRUPT`
  - `BACKUP_REASON_IDENTITY_MISMATCH`
  - `BACKUP_REASON_IDENTITY_DUPLICATE`
  - `BACKUP_REASON_OUTDATED`

## Added messages
- `IdentityState`
- `BackupManifest`
- `BackupRestoreResult`

## Compatibility notes
- Additive-only changes; no renumbering or type mutation of existing fields.
- Existing wire contracts remain stable.
- Go bindings were regenerated via `buf generate` at `gen/go/proto/aether.pb.go`.

## Planned vs implemented
- v1.2 proto deltas are implemented.
- v1.3 proto items remain planning artifacts in phase4.
