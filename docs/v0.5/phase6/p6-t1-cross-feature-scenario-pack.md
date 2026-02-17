# Phase 6 · P6-T1 Cross-Feature Scenario Pack

## Purpose
Collect the integrated positive-path and recovery/adversarial scenarios that span all eight v0.5 scope bullets, as required by Section 6.24 of `TODO_v05.md`.

## Contract
- Every scope bullet appears in at least one end-to-end scenario (bot service → SDK → slash → shim → emoji/reaction → webhook) with deterministic expected outcomes, as required by P6-T1-ST1 and VA-X1.
- Failure, spoof, and recovery scenarios listed in P6-T1-ST2 (replayed webhooks, invalid credentials, emoji quota races, gateway resume bursts, concurrent reaction toggles) specify expected reason codes, retries, and fallback actions so evaluators can reproduce outcomes.
- Scenario documentation links each story back to the relevant VA artifacts and gate owners so V5-G6 reviewers can verify coverage and tie each expectation to a pass/fail status.

## Deterministic contract pack
- The integrated scenario pack combines positive, negative, and recovery stories for every VA artifact, including reason-class references and validation obligations captured in the Phase 0 verification matrix.
- Each story names the owning gate, outlines the reason codes emitted for success/rejection/recovery, and links to the applicable artifact documentation so V5-G6 reviewers can resolve any traceability gaps before V5-G7.
