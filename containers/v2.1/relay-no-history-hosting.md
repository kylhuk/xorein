# Relay No-Long-History Hosting Contract

Relays participating in v2.1 must not retain plaintext history segments beyond metadata. Each Podman scenario should exercise:

1. Relay receives chat traffic metadata only and forwards to peers.
2. Relay keeps no local store of message bodies beyond in-flight buffers.
3. Relay exposes deterministic refusal reasons if asked to persist history.

This document is intentionally lightweight; the script and manifest represent the operational evidence for compliance.
