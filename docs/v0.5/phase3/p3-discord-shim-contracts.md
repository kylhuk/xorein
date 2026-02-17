# Phase 3 · P3 Discord Shim Contracts

## Purpose
Surface the deterministic REST subset, Gateway translation, pattern coverage, limitation catalog, and migration guidance defined in Phase 3 of `TODO_v05.md` for `aether-discord-shim`.

## Deterministic obligations
- Each supported REST endpoint (messages, channels, guilds, members, roles) includes an endpoint-to-native mapping, explicit unsupported fields, and error translation semantics so every request/response is predictable as described in P3-T1 and VA-D1.
- Gateway event translation, intent boundaries, heartbeat, reconnect, resume flows, and failure handling are spelled out in P3-T2 and VA-D2 so translated events converge deterministically even when sessions drop or resume.
- The 80 percent coverage corpus pulls from common `discord.py` and `discord.js` patterns defined in P3-T3; the scoring method and unsupported-feature taxonomy make coverage claims auditable per VA-D3.
- The shim limitations matrix, native-equivalent mapping, and migration scenario template in P3-T4 depict when bots should switch to the Native Bot API and how to recover from unsupported behavior, supplying deterministic rollback/playbook guidance in VA-D4.

## Deterministic contract tables
| Input / condition | Outcome + reason codes | Artifact | Validation obligation |
|---|---|---|---|
| Shim REST endpoint translation | Positive: `shim.rest.success` for matched endpoint + fields; Negative: `shim.rest.unsupported` for unmapped fields; Recovery: `shim.rest.migrate` directing native API fallback. | VA-D1 | Trace translation matrix against canonical field map and unsupported catalog so reviewers can review reason-coded failures per `VA-X1`. |
| Gateway event translation + session semantics | Positive: `shim.gateway.success` for supported event intents; Negative: `shim.gateway.session-error` for heartbeat or resume problems; Recovery: `shim.gateway.resume` path after reconnect. | VA-D2 | Link gateway matrices and heartbeat flows to scenario pack steps to ensure failure/recovery reason classes are consistent. |
| 80% pattern coverage plus unsupported taxonomy | Positive: `shim.coverage.success` when canon patterns meet target; Negative: `shim.coverage.unsupported` when gaps exist; Recovery: `shim.coverage.rollback` to identify migration debt. | VA-D3 | Capture coverage scoring evidence and unsupported declarations so audit reviewers can reproduce coverage/rollback reason classes. |
| Migration guidance and limitation catalog | Positive: `shim.migration.success` when limitations map to native alternatives; Negative: `shim.migration.blocked` when no shim path exists; Recovery: `shim.migration.recover` via native preference playbook. | VA-D4 | Keep migrations scenario template and limitation matrix linked to release-deferral stories for transparent fallback reason codes. |

These tables provide concrete input/outcome expectations for the V5-G3 review and feed the Phase 0 verification matrix.
