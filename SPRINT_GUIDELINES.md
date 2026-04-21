# SPRINT_GUIDELINES.md

> Status: Governance and planning policy for sprint execution.
>
> Repository posture: implementation in progress. This guideline defines required planning and delivery behavior and does not claim implemented completion.

---

## 1. Vision and Operating Principles

### 1.1 Vision
Run disciplined, version-scoped sprints that improve real user outcomes while preserving protocol integrity, governance safety, and explicit planning-vs-implementation status.

### 1.2 Operating principles
1. **Protocol-first:** protocol and specification contracts are primary; client and UI behavior are downstream implementations.
2. **Planning discipline first:** no coding begins without explicit scope, acceptance criteria, risk handling, and QA strategy.
3. **No silent scope expansion:** out-of-scope work is deferred, not absorbed.
4. **Governance-safe evolution:** additive minor evolution only; incompatible behavior follows major-path governance rules.
5. **Evidence over narrative:** gate exits require traceable pass/fail evidence, not prose-only confidence.
6. **Status clarity:** every artifact must preserve planned-vs-implemented separation.
7. **Open decisions remain open:** unresolved protocol decisions are tracked, not written as finalized architecture.

---

## 2. Priority Stack

Priority order for sprint decisions and trade-offs:
1. **Performance**
2. **Security**
3. **Reliability**
4. **QoL**
5. **Maintainability**
6. **Accessibility**
7. **Documentation**

If priorities conflict, escalate with explicit rationale in the sprint risk log and record the chosen trade-off.

---

## 3. Sprint Model: One Sprint per Minor Version

1. One sprint maps to one minor-version roadmap band.
2. Sprint scope must be traceable to that version band only.
3. Cross-version prerequisites are handled as carry-back dependencies, not silent re-scoping.
4. Each sprint has a single release-conformance exit gate.
5. The same gate discipline applies to release-band planning artifacts (including v1.0 release planning).

---

## 4. Required Sprint Artifacts

Each sprint must maintain all artifacts below:
1. **Sprint planning charter** with scope boundaries and explicit non-goals.
2. **Prioritized backlog** with task ordering and dependency mapping.
3. **Acceptance criteria matrix** mapping scope items to measurable outcomes.
4. **Risk log** with owner, status, mitigation, and escalation path.
5. **Release notes draft** for planned changes, known limits, and deferrals.

Artifacts must include evidence links or evidence placeholders for gate review.

---

## 5. QoL Guideline Set

### 5.1 Mandatory 10 percent effort-reduction objective
Each sprint must include at least one QoL objective that delivers **10% less user effort** on a priority journey.

### 5.2 Required QoL framing
1. Define the target journey and baseline effort metric before sprint work begins.
2. Define the post-change target proving the 10 percent effort reduction.
3. Preserve no-limbo behavior: user-visible degraded paths must include state, reason class, and next action.
4. Preserve recovery-first behavior for interruption-prone flows.
5. Keep deterministic reason taxonomy aligned across user-visible messaging and diagnostics.

### 5.3 QoL evidence requirement
Sprint closure must include an evidence-backed QoL scorecard with pass/fail outcomes and trace links.

---

## 6. Additional Measurable Sprint Guidelines

### 6.1 Quality
- 100 percent of in-scope items are mapped to tasks, acceptance criteria, and evidence anchors.
- 100 percent of gate checklist rows include pass/fail status and evidence links.

### 6.2 Reliability
- Critical journeys include positive, adverse, degraded, and recovery scenarios.
- No ambiguous terminal user state is accepted in closure evidence.

### 6.3 Performance
- For impacted journeys, baseline and target performance expectations are documented before execution.
- Regressions require explicit risk-log entry and governance sign-off before closure.

### 6.4 Process
- Scope changes require explicit approval and trace update.
- Risk log is updated at each gate transition.
- Deferrals include rationale, owner role, and next-version target.

---

## 7. QA Requirements

QA expectations apply even during docs/planning phases.

1. Every acceptance criterion must map to a planned validation approach.
2. Test strategy must include positive, negative, degraded, and recovery coverage.
3. Integration and regression strategy must be declared for cross-feature changes.
4. Test evidence model must be defined before implementation handoff.
5. In docs-only stages, QA artifacts are planning contracts and must be clearly labeled as planned.

---

## 8. Code Review Policy and Merge/Release Gates

### 8.1 Code review policy
1. No change merges without peer review.
2. Protocol-touching changes require compatibility and governance checklist review.
3. Breaking-change candidates require AEP-path evidence and validation by at least two independent implementations before finalization.

### 8.2 Merge gates
All must be satisfied before merge:
- Scope/acceptance trace is up to date.
- QA/test strategy and evidence links are present.
- Documentation updates are complete.
- Risk log entries and mitigations are current.

### 8.3 Release gates
All must be satisfied before release handoff:
- Release-conformance checklist complete.
- Open decisions status is accurate and unresolved items remain explicitly open.
- Release notes draft and deferral register are complete.

---

## 9. Definition of Ready

Work is ready to enter a sprint only when:
1. Scope is version-bounded and source-traceable.
2. Acceptance criteria are explicit and testable.
3. Dependencies and prerequisites are identified.
4. Risks are logged with owner and mitigation intent.
5. QA strategy is defined.
6. Planned-vs-implemented status labeling is explicit.

---

## 10. Definition of Done

### 10.1 Planning artifact done
1. Scope, exclusions, and gate criteria are complete and internally consistent.
2. Acceptance matrix and evidence model are complete.
3. Risk log and deferral register are updated.
4. Review sign-off is recorded.
5. Artifact language stays planning-only where implementation is not complete.

### 10.2 Implementation done
1. Acceptance criteria pass with evidence.
2. QA strategy is executed and results are recorded.
3. Code review and merge gates are satisfied.
4. Documentation and release notes are updated.
5. Observability, rollback, and operational readiness checks are complete.

---

## 11. End-of-Sprint Closure Checklist

### 11.1 Planning closure
- [ ] Scope trace matrix is complete and version-bounded.
- [ ] Backlog status is current and unresolved work is explicitly deferred.
- [ ] Acceptance criteria and evidence links are complete.

### 11.2 Implementation closure
- [ ] Implemented work is explicitly identified; unimplemented work remains planned.
- [ ] Merge/review gates are complete for all merged changes.
- [ ] Carry-back dependencies and deferrals are recorded.

### 11.3 Testing and quality closure
- [ ] QA coverage includes positive, negative, degraded, and recovery paths.
- [ ] Regression checks are complete or explicitly deferred with risk sign-off.
- [ ] QoL scorecard includes 10 percent effort-reduction objective results.

### 11.4 Documentation and communication closure
- [ ] Documentation updates are complete and consistent with scope.
- [ ] Release notes draft is complete and traceable to sprint outputs.
- [ ] Planned-vs-implemented wording is explicit across all artifacts.

### 11.5 Observability and operations closure
- [ ] Observability expectations and diagnostics impact are documented.
- [ ] Rollback/recovery expectations are documented for release handoff.

### 11.6 Risk and governance closure
- [ ] Risk log is up to date with final statuses and residual risks.
- [ ] Compatibility/governance checks are complete.
- [ ] Open decisions remain unresolved unless authoritative source docs explicitly resolve them.
