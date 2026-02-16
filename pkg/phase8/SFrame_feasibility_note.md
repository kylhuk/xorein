SFrame feasibility note (Phase 8)

Standards maturity

* SFrame is standardized as IETF Proposed Standard RFC 9605 (published August 2024). It’s explicitly designed for multiparty calls where an SFU can read forwarding metadata without access to media content, and it’s defined independently of RTP. ([IETF Datatracker][1])
* The RTP packetization + SDP signaling layer for SFrame is still an Internet-Draft (“RTP Payload Format for SFrame”), updated January 9, 2026. If you need true RTP-level interop outside “WebRTC encoded transforms”, expect spec churn here. ([AVTCORE][2])

Implementation maturity and practical options

Option 1 (recommended for web): WebRTC Encoded Transforms + SFrame in a Worker

* Approach: Use `RTCRtpSender/RTCRtpReceiver.transform` (worker-based) to encrypt/decrypt encoded audio frames using SFrame framing (RFC 9605). The W3C Encoded Transform spec explicitly defines “SFrame transforms” and cipher suites tied to RFC 9605. ([W3C][3])
* Compatibility: RTCRtpScriptTransform support is now broad: Safari 15.4+, Firefox 117+, Chrome 141+ (plus modern Edge/Opera equivalents), with ~78% global usage as of Jan 2026. ([Can I Use][4])
* Maintenance: Moderate. SFrame (RFC) is stable, but WebRTC transform APIs are still evolving at the spec level. Interop is actively being driven (RTCRtpScriptTransform is a 2025 WebRTC Interop focus area). ([WebKit][5])
* Back-compat: Historically, Chromium shipped a different, older Insertable Streams API; Mozilla’s shipping note calls this out and also notes WebKit shipped since 15.4. If you must support older Chromium, keep a shim path for `createEncodedStreams`. ([Google Groups][6])

Option 2 (future-looking): Built-in browser SFrame transforms (RTCSFrameSender/ReceiverTransform)

* Approach: Use the W3C-defined `RTCSFrameSenderTransform` / `RTCSFrameReceiverTransform` objects (instead of custom crypto code). ([W3C][3])
* Maturity risk: Spec-defined, but shipping/interop status is less clear than RTCRtpScriptTransform. Treat as “opt-in upgrade” once confirmed in your target browsers.

Option 3 (native clients): Embed an RFC 9605 library

* Approach: Use a native SFrame implementation in mobile/desktop clients and integrate at the encoded-frame boundary.
* Evidence of implementable libraries:

  * Rust `sframe` crate explicitly implements “SFrame (RFC 9605)” and supports different crypto backends, including a `ring` backend that can compile to Wasm32 (useful if you want one core for native + web via wasm). ([Docs.rs][7])
  * Cisco has a C++ implementation, but its README still references draft-era alignment caveats (“spec is still in progress… might not match exactly”), so it needs compliance validation against RFC 9605 test vectors before adoption. ([GitHub][8])

Option 4 (vendor SDK): Use a provider’s E2EE feature

* Current state looks “available but often beta / limited”:

  * Cloudflare RealtimeKit: “True end-to-end encryption… is in beta”; explicitly notes cloud recording / AI / transcription generally unavailable while E2EE is enabled. ([RealtimeKit Documentation][9])
  * Dyte: E2EE announced as beta with similar server-feature limitations. ([Dyte][10])
  * Agora: lists “End-to-end encryption (Beta)” but the guide is not yet available (documentation maturity gap). ([docs-staging.agora.io][11])
* If you need predictable SLAs and feature completeness, treat “beta E2EE” vendor offerings as higher operational risk.

Constraints and unresolved risks (to document in the ticket)

* Key management remains the dominant risk. Early SFrame implementations explicitly call out that key management/exchange is the missing deployment piece (“Bring your own KMS”), with MLS commonly referenced as the long-term direction. ([Medium][12])
* Server-side features that require media access (recording, transcription, AI processing) will not work when media stays end-to-end encrypted; vendors document this as a hard limitation today. ([RealtimeKit Documentation][9])
* API fragmentation/interop: Historically two incompatible browser APIs existed (legacy Chromium vs worker-first standard); while this is converging now, you still need capability detection and a “no-E2EE” fallback. ([webrtcHacks][13])
* RTP-level interop outside the WebRTC transform context is not fully standardized yet (packetization/SDP signaling still draft). ([AVTCORE][2])
* Security model nuance: RTCRtpScriptTransform improves privacy against middleboxes (e.g., SFU) but does not inherently protect against the application/provider if they control the JS and keys; Mozilla explicitly discusses this limitation. ([Google Groups][6])

Voice encryption fallback decision criteria (explicit)

Decision gates (if any gate fails, fall back per next section)

1. Client capability gate

* Enable SFrame E2EE only if the client supports either:

  * `RTCRtpScriptTransform` / encoded transforms (preferred), or
  * legacy Chromium insertable streams API (only if you choose to maintain it).
* If not supported: fall back to SRTP-only (DTLS-SRTP), or block “E2EE-required” rooms. ([Can I Use][4])

2. Product feature gate

* If the call requires any server-side media feature (cloud recording, transcription/AI, server-side mixing/processing), do not enable E2EE for that session (or require explicit user opt-in acknowledging those features will be disabled). ([RealtimeKit Documentation][9])

3. Operational reliability gate (suggested measurable thresholds)

* If E2EE increases call setup failure rate by >0.2% absolute vs baseline, or causes decrypt/authentication errors in >0.5% of sessions, treat SFrame as “high risk” and revert to SRTP-only by default until fixed (keep feature-flagged rollout).

4. Performance gate (voice-specific)

* If E2EE increases median audio end-to-end latency by >30 ms or pushes sustained client CPU >15% above baseline on your lowest supported mobile device class, revert to SRTP-only for that device class (capability-based policy).

5. Vendor maturity gate (only if using a vendor SDK)

* If the vendor labels E2EE as beta, lacks published limitations/SLAs, or support can’t commit to timelines for critical issues, treat as “high risk” and default to SRTP-only (or switch vendors). ([RealtimeKit Documentation][9])

Fallback modes (what you do when gates fail)
A) Default fallback: DTLS-SRTP transport encryption only

* WebRTC already uses DTLS + SRTP to encrypt media in transit; this protects against network attackers but not against the SFU/service decrypting. ([RealtimeKit Documentation][9])

B) “E2EE-required” room policy (optional)

* If E2EE is a hard requirement: block join for clients that fail the capability gate (instead of silently downgrading).

C) Voice-only pragmatic fallback for small calls (optional)

* For 1:1 or very small groups, prefer P2P where feasible (SRTP becomes effectively end-to-end because there is no SFU middlebox). RFC 9605 frames this as removing the SFU from the trust boundary. ([IETF Datatracker][1])

## Implementation recommendation for v0.1 (planning-only)

- **Planning-only signal:** This recommendation records a target path; it does **not** claim SFrame is implemented or enabled in v0.1. Execution evidence must be captured before any doc asserts readiness.
- **Web clients (preferred target):** adopt Option 1 once client work exists. Chromium, Firefox, and Safari ship `RTCRtpScriptTransform`, enabling a worker-based SFrame transform with an optional shim for legacy Chromium `createEncodedStreams` when older versions are in scope. This keeps crypto isolated from libwebrtc internals and lets MLS-backed key distribution evolve independently.
- **Native clients:** embed an RFC 9605 implementation (Option 3) at the encoded-frame boundary. Treat the Rust `sframe` crate as the reference core and compile it to Wasm for browsers (if library parity is desired) and to a static library for native builds so cipher selections (`AES_128_GCM_SHA256_128`) remain aligned.
- **Operational guardrails:** ship SFrame behind an experimentation flag. The feature is enabled only when the capability gate (#1), product feature gate (#2), and reliability/performance gate (#3/#4) are green. Downgrade to SRTP-only is a single configuration toggle.

## Acceptance check and downstream actions

- **Compatibility + maintenance assessment:** Options 1–4 document browser/library maturity and vendor limitations; the recommendation above records the selected approach per client, satisfying the “confirm practical options” deliverable.
- **Fallback + decision criteria:** Gates 1–5 plus fallback modes A–C provide explicit downgrade behavior and room-blocking policies, covering the fallback deliverable. Reliability/performance guardrails (<0.2% setup regression, <30 ms added latency, <15% sustained CPU delta) give measurable rollback triggers.
- **Dependencies:** MLS-backed key management remains the gating dependency for enabling SFrame by default. Carry the “no server-side media features while E2EE is active” limitation into operator/user quickstarts so release communications stay accurate. Residual risks (R2/R3) inherit this decision record in `TODO_v01.md`.
- **Doc propagation:** When updating quickstarts, add a “Security and media-scope note” that references this research doc and reiterates that SFrame is not enabled in the v0.1 runtime.

[1]: https://datatracker.ietf.org/doc/html/rfc9605 "
            RFC 9605 - Secure Frame (SFrame): Lightweight Authenticated Encryption for Real-Time Media
        "
[2]: https://ietf-wg-avtcore.github.io/draft-ietf-avtcore-rtp-sframe/draft-ietf-avtcore-rtp-sframe.html?utm_source=chatgpt.com "RTP Payload Format for SFrame"
[3]: https://www.w3.org/TR/webrtc-encoded-transform/ "WebRTC Encoded Transform"
[4]: https://caniuse.com/mdn-api_rtcrtpscripttransform "RTCRtpScriptTransform API | Can I use... Support tables for HTML5, CSS3, etc"
[5]: https://webkit.org/blog/16458/announcing-interop-2025/ "  Announcing Interop 2025 | WebKit"
[6]: https://groups.google.com/a/mozilla.org/g/dev-platform/c/Gowr5Fx5jng "Intent to ship RTCRtpScriptTransform"
[7]: https://docs.rs/sframe "sframe - Rust"
[8]: https://github.com/cisco/sframe "GitHub - cisco/sframe: Implementation of RFC 9605"
[9]: https://docs.realtime.cloudflare.com/guides/capabilities/misc/end-to-end-encryption "End to End encryption for audio and video calls"
[10]: https://dyte.io/blog/end-to-end-encryption/ "End-to-End Encryption"
[11]: https://docs-staging.agora.io/en/interactive-live-streaming/advanced-features/end-to-end-encryption?platform=web "Interactive Live Streaming End-to-end encryption (Beta) | Agora Docs"
[12]: https://medooze.medium.com/sframe-js-end-to-end-encryption-for-webrtc-f9a83a997d6d "SFrame.js: end to end encryption for WebRTC | by Medooze | Medium"
[13]: https://webrtchacks.com/end-to-end-encryption-in-webrtc-4-years-later/ "End-to-End Encryption in WebRTC… 4 Years Later - webrtcHacks"
