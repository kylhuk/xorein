# Phase 6 - Task 7: Cross-Platform Deeplink Registration Strategy

## Objective
Document the platform-specific deeplink registration requirements for Gio-based v0.1 clients so that join flows remain deterministic on desktop, mobile, and sandboxed environments.

## Target Platforms
1. **macOS / Windows Desktop (Gio native):**
   - Register `aether://` custom scheme via platform installers (plist on macOS, registry entries on Windows).
   - Fallback: copy/paste deeplink prompt when registration missing.
2. **Linux Desktop:**
   - `.desktop` file with `MimeType=x-scheme-handler/aether` and `gio mime x-scheme-handler/aether aether.desktop` update.
   - Flatpak/Snap: register via portal (`xdg-desktop-portal --register-app`) and ship `.desktop` override inside sandbox; expose manual `join.aether.chat/<code>` entry if host denies portal requests.
3. **Android (Gio + gomobile):**
   - `AndroidManifest.xml` intent-filter for `aether://join/*` + App Link handshake.
   - Mitigation for OEMs that block implicit intents: show manual code entry plus QR scanning fallback.
4. **iOS (gio-ios / gomobile):**
   - `CFBundleURLTypes` entry for `aether` scheme.
   - Universal Links deferred until v0.2; fallback is shareable join code entry (also rendered as numeric invite code and `https://join.aether.chat/<code>`).

## Sandbox / Installer Guidance
- **macOS pkg**: installer writes `~/Library/Preferences/com.aether.chat.plist` with `CFBundleURLSchemes=["aether"]`. Provide remediation script (`/usr/libexec/PlistBuddy`) to re-register when LaunchServices cache breaks.
- **Windows MSI**: registry path `HKCU\Software\Classes\aether` with default `URL:aether Protocol` and `shell\open\command="%LOCALAPPDATA%\Aether\aether.exe" --deeplink "%1"`. Document repair step using `reg import` snippet in support guide.
- **Snap/Flatpak**: include portal manifest entry `session_bus` permission `org.freedesktop.portal.OpenURI`; failure path triggers inline copy prompt with CLI fallback.

## Residual Risks & Mitigations
| ID | Risk | Impact | Mitigation |
|----|------|--------|------------|
| RR-P6-T7-1 | Sandboxed environments may block custom scheme even after portal request. | Users cannot auto-open join links. | Present join code UI with clipboard detection + instructions to run `aether --deeplink <link>`; document in operator FAQ. |
| RR-P6-T7-2 | Windows registry corruption leaves stale handler. | Deeplinks silently fail to launch client. | Include repair tool (`aether --repair-deeplink`) that rewrites HKCU keys; add to troubleshooting guide. |
| RR-P6-T7-3 | Mobile OEMs disable implicit intents (notably on downstream Android skins). | Tap on invite fails. | Detect via `PackageManager.resolveActivity`; if nil, surface manual entry instructions and QR copy. |

Residual risk owners logged in `TODO_v01.md` and linked to Phase 11 operator docs.

## Compatibility Notes
- Desktop CLI path `cmd/aether/main.go` already accepts `--deeplink` arguments ensuring test coverage without GUI registration.
- Sandboxed environments (Flatpak, Snap) need portal prompts; instructions added to Phase 11 operator docs.
- Mobile browsers may downgrade to `https://join.aether.chat/<code>` which the shell converts to deeplink via `pkg/phase6/deeplink.go` normalization.

## Evidence
- Parser/router implementation: [`pkg/phase6/deeplink.go`](../../../pkg/phase6/deeplink.go)
- CLI validation: [`cmd/aether/main.go`](../../../cmd/aether/main.go)
- Tests covering normalization and error paths: [`pkg/phase6/deeplink_test.go`](../../../pkg/phase6/deeplink_test.go)
- Sandbox repair CLI fallback: `aether --deeplink <link>` (documented in `cmd/aether/main.go`).
