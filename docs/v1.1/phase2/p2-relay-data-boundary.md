# Phase 2 - Relay Data Boundary

Relay v11 is strictly a control-plane and transport fabric. The relay does not host user-owned content beyond transient in-flight frames; every durable datum is either kept on clients/peers or stored in downstream v12+ services that explicitly claim that responsibility. `pkg/v11/relaypolicy/policy.go` encodes this intent, enumerating the allowed storage classes and persistence modes, while `pkg/v11/relaypolicy/policy_test.go` guarantees the lists are deterministic and that validation rejects forbidden configurations. Together with `tests/e2e/v11/relay_boundary_test.go`, these artifacts form the executable realization of EV-v11-G3-002.

## Intent and rationale
- Relay is **control-plane + transport only**: it relays routing directives, ACL decisions, heartbeat acknowledgments, and short-lived QoS telemetry between peers.
- Relay does **not** provide durable hosting for user content: message bodies, large blobs, or long-term state live elsewhere.
- This keeps the relay boundary additive and auditable; every data class either maps to a documented policy (in `pkg/v11/relaypolicy/policy.go` and `pkg/v11/relaypolicy/policy_test.go`) or triggers the boundary test harness in `tests/e2e/v11/relay_boundary_test.go`.

## Allowed vs forbidden data classes
| Data class | Allowed | Example | Justification |
|---|---|---|---|
| Session metadata | ✅ | `pkg/v11/relaypolicy/policy.go` defines `StorageClassSessionMetadata` guarded by `PersistenceModeSessionMetadata` | Needed for connection/session state without storing payloads; policy permits the metadata while validation forbids extending storage duration. |
| Transient metadata | ✅ | `pkg/v11/relaypolicy/policy.go` adds `StorageClassTransientMetadata` with `PersistenceModeTransientMetadata` | Operational signals such as heartbeats and QoS counters are short-lived and replayable without creating a durable hosting surface. |
| Durable message body | ❌ | `pkg/v11/relaypolicy/policy.go` maps `StorageClassDurableMessageBody` to `PersistenceModeDurableMessageBody` | Keeping message bodies would create a secondary hosting surface; validation rejects the associated persistence mode. |
| Attachment payload | ❌ | `pkg/v11/relaypolicy/policy.go` maps `StorageClassAttachmentPayload` to `PersistenceModeAttachmentPayload` | Files and attachments are owned by peers/clients; the relay must not cache them beyond transient transport windows. |
| Media frame archive | ❌ | `pkg/v11/relaypolicy/policy.go` maps `StorageClassMediaFrameArchive` to `PersistenceModeMediaFrameArchive` | Historical media archives belong to owning peers; policy validation turns on the forbidden mode to keep relays stateless. |

## Evidence anchor table
| Anchor | Description | Evidence |
|---|---|---|
| EV-v11-G3-001 | Relay policy packages enumerate allowed/forbidden storage classes and persistence modes | `pkg/v11/relaypolicy/policy.go` defines the classes/modes and `pkg/v11/relaypolicy/policy_test.go` confirms the deterministic lists. |
| EV-v11-G3-002 | Boundary tests exercise allowed/forbidden persistence modes | `tests/e2e/v11/relay_boundary_test.go` validates that session metadata modes pass while durable message bodies fail. |

## Planned vs Implemented
- **Planned:** Define the relay data boundary with evidence anchors and close the gap between policy intents and enforcement tests before handing gate EV-v11-G3 to QA.
- **Implemented:** Policies live under `pkg/v11/relaypolicy/policy.go`, deterministic checks run in `pkg/v11/relaypolicy/policy_test.go`, and the boundary scenario is scaffolded in `tests/e2e/v11/relay_boundary_test.go`; documentation now captures the data-class expectations without claiming completion of the evidence artifacts.
