# Phase 4 `F14` Proto Delta

- No new fields are added to existing v13 messages; voice-only payloads will be additive in v14.
- `voice_session.proto` will introduce `SessionFlavor`, `Route`, and `RecoveryHint` to cover direct/relay toggles.
- Proto changes are planned to remain optional so v13 clients ignore voice-specific frames safely.
