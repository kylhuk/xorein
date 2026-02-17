# Phase 4 · P4-T1 Push Relay & Desktop Notification Contract

## Objective
Lock the encrypted push relay envelope, provider mapping, token lifecycle, retry semantics, and desktop-native notification coherence rules so V7-G4 can freeze the notification contracts before any runtime implementation claims completion.

## Contract
### Encrypted push envelope & relay binding
- The relay-visible envelope metadata and integrity tags live in `pkg/v07/pushrelay/contracts.go`. They must preserve ciphertext opacity, minimize metadata, and fit every provider bridge (FCM/APNs) without introducing new plaintext leak vectors.

### Provider routing & token lifecycle
- Provider-bridge mapping, retry behavior, token registration/rotation/revocation, and dedupe semantics are described in `pkg/v07/pushrelay/contracts.go`. The contract bounds failure classification and ensures deterministic retry/backoff sequences plus explicit dead-letter signaling.

### Desktop notification coherence
- Desktop trigger, dedupe, suppression, action handling, and degraded fallback semantics (including unread/attention convergence) are captured in `pkg/v07/notification/contracts.go`. The contract ensures parity across event streams and includes fallback paths when platform APIs are unavailable.

## Evidence anchors
| Artifact | Description | Evidence anchor |
|---|---|---|
| `VA-P1` | Relay envelope metadata minimization | Section "Encrypted push envelope & relay binding" |
| `VA-P2` | Provider forwarding rules | Same section |
| `VA-P3` | Token lifecycle states | Section "Provider routing & token lifecycle" |
| `VA-P4` | Retry/dedupe/failure taxonomy | Same section |
| `VA-P5` | Desktop trigger/dedupe coherence | Section "Desktop notification coherence" |
| `VA-P6` | Action handling & degraded fallback | Same section |

This doc lets V7-G4 reviewers evaluate the encrypted push relay and desktop notification constraints while tying each artifact to the planned `pkg/v07` seams.
