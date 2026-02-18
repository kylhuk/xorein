# Phase 5 Risk Register

| Risk ID | Description | Mitigation | Status |
|---------|-------------|------------|--------|
| R13-1 | Space policy ambiguity | Deterministic joinpolicy states + invite token checks | Active |
| R13-2 | Chat state drift | Explicit chat.DeliveryState transitions and tests | Active |
| R13-3 | Relay boundary regression | Reuse pkg/v11/relaypolicy checks in e2e + Podman manifest | Monitored |
| R13-4 | Podman scenario flakes | Deterministic script manifest + docker-compose scaffold | Active |
