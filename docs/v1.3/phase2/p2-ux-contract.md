# Phase 2 UX Contract

- Spaces list shows founder controls, channel previews, and join buttons with deterministic labels (`pkg/v13/ui`).
- Channel timelines surface delivery/read state via `chat.DeliveryState` and `ui.DeliveryLabel` to avoid ambiguous status.
- Composer shows `HasDraft` to prevent unintentional sends; empty drafts keep peripheral state idle.
- Join panel copies `JoinPolicy` state from runtime and reflects open/invite-only semantics without guessing.
- No-limbo requirement: UI labels must never show `pending` once ack/ failure is resolved.
