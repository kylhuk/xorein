# Phase 0 · P0-T3 Verification Evidence Matrix

## Purpose
Describe the planned evidence schema for every deterministic contract that `TODO_v07.md` demands, so V7-G0 can certify traceability before any code-freeze claims.

## Contract
- Map each `VA-*` artifact from the execution plan to a preferred evidence type (doc narrative, deterministic helper, scenario pack, audit record) and annotate the gate owner plus reason-class triad (positive, negative, recovery).
- Include a sweep matrix that ties every evidence row back to the release-gate dossier (`docs/v0.7/phase5/p5-t3-release-gate-handoff.md`) so reviewers can trace the chain from V7-G0 all the way to V7-G6.

### Artifact matrix by scope

#### Store-forward and retention/artchivist (`VA-D1`..`VA-D6`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-D1 | P1-T1 | TTL-preserved dht entry (`storeforward.ttl.success`) | Pruned entry surfaced (`storeforward.ttl.prune`) | Retry/purge fallback (`storeforward.ttl.retry`) | Freshness | V7-G1 | `docs/v0.7/phase1/p1-t1-store-forward-retention-archivist.md`
| VA-D2 | P1-T1 | k=20 replication success (`storeforward.replica.success`) | Target shortfall flagged (`storeforward.replica.shortfall`) | Degraded restore path (`storeforward.replica.repair`) | Durability | V7-G1 | Same doc
| VA-D3 | P1-T2 | Policy override honored (`retention.policy.success`) | Conflict rejection (`retention.policy.conflict`) | Audit trail recorded (`retention.policy.audit`) | Retention policy | V7-G1 | Same doc
| VA-D4 | P1-T2 | Purge executed deterministically (`retention.purge.success`) | Premature purge warning (`retention.purge.reject`) | Restore from archivist (`retention.purge.recover`) | Retention transition | V7-G1 | Same doc
| VA-D5 | P1-T3 | Archivist enrolls (`archivist.enroll.success`) | Enrollment rejected (`archivist.enroll.reject`) | Controlled withdrawal (`archivist.withdraw.recover`) | Role lifecycle | V7-G1 | Same doc
| VA-D6 | P1-T3 | Integrity checks pass (`archivist.integrity.success`) | Coverage gap flagged (`archivist.integrity.alert`) | Fallback to peer store (`archivist.integrity.recover`) | Archivist obligations | V7-G1 | Same doc

#### History sync and Merkle (`VA-H1`..`VA-H7`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-H1 | P2-T1 | Sync negotiation succeeds (`history.sync.negotiate.success`) | Version mismatch (`history.sync.negotiate.reject`) | Downgrade backoff (`history.sync.negotiate.recover`) | Negotiation | V7-G2 | `docs/v0.7/phase2/p2-t1-history-sync-merkle.md`
| VA-H2 | P2-T1 | Request/response lifecycle closes (`history.sync.lifecycle.complete`) | Request rejection (`history.sync.lifecycle.failure`) | Resume on checkpoint (`history.sync.lifecycle.resume`) | Lifecycle | V7-G2 | Same doc
| VA-H3 | P2-T2 | Merkle root matches canonical (`history.merkle.success`) | Chunk divergence (`history.merkle.mismatch`) | Proof remediation (`history.merkle.remediate`) | Merkle integrity | V7-G2 | Same doc
| VA-H4 | P2-T2 | Proof exchange validated (`history.proof.verify.success`) | Proof failure (`history.proof.verify.failure`) | Re-request path (`history.proof.verify.retry`) | Proof handling | V7-G2 | Same doc
| VA-H5 | P2-T3 | Retention-aware sync window honored (`history.sync.window.good`) | Window violation flagged (`history.sync.window.blocked`) | Alternate range recovery (`history.sync.window.recover`) | Retention-aware sync | V7-G2 | Same doc
| VA-H6 | P2-T3 | Archivist source accepted (`history.sync.archivist.success`) | Source mismatch flagged (`history.sync.archivist.reject`) | Fallback to canonical (`history.sync.archivist.recover`) | Archivist source | V7-G2 | Same doc
| VA-H7 | P2-T1 | Mode epoch boundaries signaled (`history.mode.epoch.success`) | Locked-history confusion (`history.mode.epoch.blocked`) | Explicit re-sharing action (`history.mode.epoch.recover`) | Mode epoch | V7-G2 | Same doc

#### Scoped search filters (`VA-S1`..`VA-S6`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-S1 | P3-T1 | Scoped index entries generated (`search.index.success`) | Tokenization error reported (`search.index.failure`) | Rebuild from backup (`search.index.recover`) | Indexing | V7-G3 | `docs/v0.7/phase3/p3-t1-scoped-search-filters.md`
| VA-S2 | P3-T1 | Lifecycle respects retention (`search.index.lifecycle.success`) | Retention mismatch (`search.index.lifecycle.failure`) | Repair from archived payload (`search.index.lifecycle.recover`) | Index lifecycle | V7-G3 | Same doc
| VA-S3 | P3-T2 | Search filter query accepted (`search.query.filters.success`) | Invalid combination (`search.query.filters.invalid`) | Normalized fallback (`search.query.filters.recover`) | Filter normalization | V7-G3 | Same doc
| VA-S4 | P3-T2 | Response envelope complete (`search.response.success`) | Partial failure (`search.response.partial`) | Pagination retry (`search.response.recover`) | Response handling | V7-G3 | Same doc
| VA-S5 | P3-T3 | Authorization/redaction honor scope (`search.auth.success`) | Scope leak blocked (`search.auth.failure`) | Scoped fallback (`search.auth.recover`) | Authorization | V7-G3 | Same doc
| VA-S6 | P3-T3 | At-rest constraints preserved (`search.privacy.success`) | Privacy violation detected (`search.privacy.failure`) | Evidence audit (`search.privacy.recover`) | Privacy | V7-G3 | Same doc

#### Push relay & desktop notifications (`VA-P1`..`VA-P6`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-P1 | P4-T1 | Relay envelope metadata minimal (`push.envelope.success`) | Metadata leak flagged (`push.envelope.failure`) | Retry with masked fields (`push.envelope.recover`) | Relay blindness | V7-G4 | `docs/v0.7/phase4/p4-t1-push-relay-desktop-notifications.md`
| VA-P2 | P4-T1 | Provider forwarding succeeds (`push.forward.success`) | Provider error channel (`push.forward.failure`) | Dead-letter fallback (`push.forward.recover`) | Provider mapping | V7-G4 | Same doc
| VA-P3 | P4-T2 | Token lifecycle renews (`push.token.success`) | Token stale error (`push.token.failure`) | Rotation recovery (`push.token.recover`) | Token lifecycle | V7-G4 | Same doc
| VA-P4 | P4-T2 | Retry/dedupe enforces policy (`push.retry.success`) | Dead-letter triggered (`push.retry.failure`) | Backoff reset (`push.retry.recover`) | Retry policy | V7-G4 | Same doc
| VA-P5 | P4-T3 | Desktop triggers fire (`push.desktop.success`) | Notification suppression (`push.desktop.failure`) | Fallback state (`push.desktop.recover`) | Trigger coherence | V7-G4 | Same doc
| VA-P6 | P4-T3 | Action handlers resolve (`push.desktop.action.success`) | API missing feedback (`push.desktop.action.failure`) | Degraded fallback (`push.desktop.action.recover`) | Action handling | V7-G4 | Same doc

#### Integrated validation & governance (`VA-I1`..`VA-I3`, `VA-R1`)
| Artifact | Task | Positive | Negative | Recovery | Reason-class | Gate owner | Evidence anchor |
|---|---|---|---|---|---|---|---|
| VA-I1 | P5-T1 | Scenario pack covers positive flows (`validation.scenario.positive`) | Gap flagged (`validation.scenario.gap`) | Recovery scenarios drafted (`validation.scenario.recover`) | Coverage | V7-G5 | `docs/v0.7/phase5/p5-t1-integrated-validation.md`
| VA-I2 | P5-T2 | Governance review clarifies controls (`governance.review.pass`) | Non-compliance reported (`governance.review.alert`) | Escalation path logged (`governance.review.recover`) | Governance readiness | V7-G5 | `docs/v0.7/phase5/p5-t2-governance-readiness-audit.md`
| VA-I3 | P5-T2 | Compatibility checklist satisfied (`governance.checklist.pass`) | Triggered governance meatball (`governance.checklist.alert`) | Additional evidence collected (`governance.checklist.recover`) | Compatibility | V7-G5 | Same doc
| VA-R1 | P5-T3 | Release gate handoff complete (`release.handoff.success`) | Missing evidence noted (`release.handoff.alert`) | Deferred item docketed (`release.handoff.recover`) | Release conformance | V7-G6 | `docs/v0.7/phase5/p5-t3-release-gate-handoff.md`

This matrix keeps every `VA-*` artifact tied to the new docs, gate owners, and implementation hopefuls so V7-G0 can certify the compliance story before any runtime commitments occur.
