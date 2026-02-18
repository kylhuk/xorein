# Phase 4 `F14` Voice Specification

- Defines baseline voice session lifecycle: dial, join, negotiate media route, leave.
- Identifies deterministic states (pending, in-call, recovering) to align with client state machine.
- Captures fallback topology decisions (direct, mesh, relay) with clearly documented recovery transitions.
- Notes that current v13 runtime only surfaces messaging; voice remains spec-only for v14 implementation.
