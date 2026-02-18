# Phase 4 F15 Proto Delta

Proto delta for F15 remains additive:
- Adds `ScreenShareFrame` message with deterministic version and checksum metadata.
- Introduces `ScreenshareAdaptation` enum for selection strategies.
- No existing field numbers are repurposed; all new enums/messages occupy new numbers with `reserved` sections left untouched.
