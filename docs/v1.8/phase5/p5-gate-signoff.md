# Phase 5 - Gate Signoff

| GateID | Purpose | Entry checks | Exit checks | Owner role | Required approvers | Evidence IDs | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| G0 | Scope lock and dependency map complete | Phase0 scope/traceability/RACI docs exist | Scope and dependency artifacts immutable and linked | Plan Lead | Release Authority | EV-v18-G0-001, EV-v18-G0-002, EV-v18-G0-003 | promoted | Approved by Release Authority (RA-01) @ 2026-02-18T00:00:00Z |
| G1 | Proto compatibility checks pass | `buf lint` and `buf breaking` executed | No lint failure and no breaking wire changes | Protocol Lead | Release Authority | EV-v18-G1-001, EV-v18-G1-002 | promoted | Deprecation warning accepted; wire compatibility preserved. |
| G2 | DirectoryEntry and indexer runtime complete | pkg/v18 contracts implemented | Runtime package tests pass | Runtime Lead | QA Lead | EV-v18-G2-001 | promoted | Signed by QA Lead (QA-01) @ 2026-02-18T00:00:00Z |
| G3 | Discovery client verification and join UX complete | discoveryclient/UI contracts implemented | Client/runtime contract tests pass | Client Lead | QA Lead | EV-v18-G3-001 | promoted | Signed by QA Lead (QA-01) @ 2026-02-18T00:00:00Z |
| G4 | Discovery/adversarial validation matrix complete | e2e/perf suites authored | e2e + perf tests pass with deterministic outputs | QA Lead | Runtime Lead, Client Lead | EV-v18-G4-001, EV-v18-G4-002 | promoted | Validation matrix approved @ 2026-02-18T00:00:00Z |
| G5 | Podman discovery scenarios complete | scenario script and container assets present | Script exits zero and manifest reports pass | Ops Lead | QA Lead, Release Authority | EV-v18-G5-001, EV-v18-G5-002 | promoted | Manifest confirms all scenario probes pass. |
| G6 | v19 spec package complete | phase4 spec/proto/acceptance docs drafted | F19 package published and linked | Plan Lead | Protocol Lead, QA Lead | EV-v18-G6-001, EV-v18-G6-002, EV-v18-G6-003 | promoted | Approved for v19 handoff. |
| G7 | Docs and evidence complete | phase5 docs created | Evidence bundle/index/checks complete and traceable | QA Lead | Plan Lead, Release Authority | EV-v18-G7-001, EV-v18-G7-002, EV-v18-G7-003 | promoted | Evidence paths immutable and checksummed. |
| G8 | Relay no-data-hosting regression checks pass | dedicated relay regression test exists | Relay regression probe exits zero | Runtime Lead | Security Lead, QA Lead | EV-v18-G8-001 | promoted | Relay boundary constraint preserved. |
| G9 | F18 as-built conformance against v17 package complete | as-built and risk docs drafted | Conformance report and residual risk register accepted | Plan Lead | Release Authority, QA Lead | EV-v18-G9-001, EV-v18-G9-002 | promoted | Final promotion recommendation approved. |
