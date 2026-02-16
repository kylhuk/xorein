# Phase 10 - Task 7: Gio Input & IME Edge Cases

## Objective
Identify v0.1-critical Gio text input and IME edge cases across desktop and mobile targets and document mitigations plus backlog follow-ups.

## Desktop (Linux/macOS/Windows)
1. **Dead keys / Compose sequences**
   - Gio’s text driver collapses intermediate key events; dead keys may not emit composed runes until `Editor.Insert` completes.
   - Mitigation: ensure `ValidateComposerMessage()` (see [`pkg/ui/shell.go`](../../pkg/ui/shell.go)) trims whitespace only after focus loss; do not debounce key events within 50ms.
2. **International IMEs (IBus/Fcitx/TSF)**
   - Gio requires `InputMethod.Editor` implementation to relay surrounding text. Shell already keeps last draft state; enforce `maxLength` guard after commit only.
3. **Accessibility / screen readers**
   - AT-SPI/VoiceOver may pause event delivery when focus tooltip surfaces. Avoid modal popups in composer region.

## Mobile (Android/iOS via gomobile)
1. **On-screen keyboard focus churn**
   - Voice push-to-talk button may steal focus; store drafts via `DraftScopeKey()` before toggling voice UI.
2. **CJK IMEs**
   - Composition window overlaps message list; reserve 200px bottom padding when IME visible.
3. **Hardware keyboards on tablets**
   - Enter key vs Shift+Enter needs same behavior as desktop (`HandleComposerEnter`). Covered in tests but document for QA.

## Fallbacks & User Guidance
- Provide “Paste join code” dialog for users whose IME cannot paste into main composer (rare but observed on restricted kiosk builds).
- Add settings toggle to disable markdown shortcuts, preventing accidental formatting on layouts where `*` requires AltGr.

## Residual Risks
| ID | Risk | Impact | Mitigation |
|----|------|--------|------------|
| RR-P10-T7-1 | Gio lacks per-platform IME diagnostics; regressions hard to detect. | Users may see dropped characters without logs. | Add debug flag to log IME composition events (without contents) for QA builds. |
| RR-P10-T7-2 | Android OEM keyboards kill background audio focus. | Voice bar loses mic permission mid-compose. | Persist voice-session state and surface toast instructing re-enable; backlog item created. |

## References
- Composer helpers: [`pkg/ui/shell.go`](../../pkg/ui/shell.go)
- Voice/diagnostic UI hooks: [`pkg/ui/shell.go`](../../pkg/ui/shell.go)
- Tests covering drafts/enter handling: [`pkg/ui/shell_test.go`](../../pkg/ui/shell_test.go)
