# v0.1 Phase 1 Scope Contract and Traceability

## Scope Contract Summary
- **In-scope outcomes** mirror the Phase 1 mandate: identity creation, server manifest management, LAN first-contact journeys (identity → server → chat → voice), and headless relay bootstrapping consistent with [`TODO_v01.md:161-220`](TODO_v01.md:161-220).
- **Strict exclusion guardrail** keeps deferred work out of v0.1: DM protocol with X3DH, presence/friends, SFU, advanced media, bot frameworks, public discovery, and push relays remain explicit non-goals as flagged in [`TODO_v01.md:74-84`](TODO_v01.md:74-84).

## In-Scope Outcomes
1. Identity lifecycle from generation through profile signing and persistence. [`TODO_v01.md:32-45`](TODO_v01.md:32-45)
2. Server creation, manifest signing, and deeplink join flow. [`TODO_v01.md:32-45`](TODO_v01.md:32-45)
3. MLS-powered text chat with Sender Keys compatibility for migration needs. [`TODO_v01.md:32-45`](TODO_v01.md:32-45)
4. Full mesh voice for small groups plus relay fallback. [`TODO_v01.md:32-45`](TODO_v01.md:32-45)
5. Headless relay with DHT bootstrap and Circuit Relay v2. [`TODO_v01.md:32-45`](TODO_v01.md:32-45)
6. Baseline engineering foundation (CI, proto, builds). [`TODO_v01.md:46-47`](TODO_v01.md:46-47)
7. Diagnostics groundwork (reason codes, ring buffers). [`TODO_v01.md:47-48`](TODO_v01.md:47-48)
8. QoL contract enforcing deterministic first-contact clarity. [`TODO_v01.md:52-59`](TODO_v01.md:52-59)

## Deferred Items / Non-Goals
- DM protocol with X3DH & Double Ratchet [`TODO_v01.md:76-84`](TODO_v01.md:76-84)
- Presence, friends, notifications
- SFU/advanced voice tuning
- Screen share, file transfer, moderation, bots
- Public discovery, push notification relays
- Compliance hardening beyond baseline engineering controls

## Success Criteria Traceability (Phase 1)
| Success criterion | Linked Phase 1 artifact | Phase 1 Tasks | Evidence placeholder |
|---|---|---|---|
| First-launch usability (<5 min) | Acceptance test charter [`docs/v0.1/phase1/acceptance-test-charter.md:1`](docs/v0.1/phase1/acceptance-test-charter.md:1) | P1-T4 | Pending test runs |
| Identity lifecycle stability | Scope contract outcomes section | P1-T1 | Pending identity audit |
| Server bootstrap manifest flow | Scope contract outcomes section | P1-T1 | Pending manifest spec |
| Text security with MLS | Protocol constraints checklist [`docs/v0.1/phase1/protocol-constraints.md:1`](docs/v0.1/phase1/protocol-constraints.md:1) | P1-T2 | Pending proto convergence |
| Voice baseline mesh | Acceptance test charter | P1-T4 | Pending voice runbook |
| Relay baseline | Scope contract outcomes + constraints | P1-T1/P1-T2 | Pending relay evidence |
| Engineering foundation | Ownership + ADR artifact [`docs/v0.1/phase1/ownership-adr.md:1`](docs/v0.1/phase1/ownership-adr.md:1) | P1-T3 | Pending pipeline evidence |
| Diagnostics groundwork | Protocol constraints checklist | P1-T2 | Pending diagnostics spec |
| Scope integrity (deferred features) | Scope contract deferred section | P1-T1 | Evidence placeholder |

## Evidence Location Placeholders
- Scope contract draft: `docs/v0.1/phase1/scope-contract.md` (review pending)
- Traceability references: each success row above will link to execution evidence as it becomes available.
