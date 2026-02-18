# Phase 2 RBAC UX Contract

- UI surfaces expose role summaries, permission hints, and upgrade guidance.
- Admin helpers must rely on deterministic role metadata from `pkg/v16/ui/admin_ui.go`.
- Permission hints respect allowed/denied states and remain textual (no runtime rendering assumptions).
