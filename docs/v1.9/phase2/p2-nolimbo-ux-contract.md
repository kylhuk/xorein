# No-Limbo UX Contract

- Canonical states: `idle`, `connecting`, `in-call`, `recovering`.
- Recovery-first transitions ensure failed calls surface `recovering` before offering reconnect guidance.
- Clients must render `ActionShowRecover` before showing data entry prompts.
- State transitions are deterministic and mirrored in `pkg/v19/ui/nolimbo_ui.go` tests.
