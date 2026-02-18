# Phase 2 UX Contract

Gio client surfaces must follow deterministic contracts:
- Source map: display/window only.
- Controls: start enabled only when idle, stop enabled when streaming.
- Indicators: show bitrate `q=<kbps>` and recovery hint text from `pkg/v15/ui`.
- Recovery cues must reference fallback/error reason and advise to pause traffic or restart.
