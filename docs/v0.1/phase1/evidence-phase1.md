# Phase 1 Verification Evidence (Podman)

## Command

```bash
podman run --rm -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/alpine:3.20 sh -lc '
set -eu

echo "== Artifact existence checks =="
for f in \
  docs/v0.1/phase1/scope-contract.md \
  docs/v0.1/phase1/protocol-constraints.md \
  docs/v0.1/phase1/ownership-adr.md \
  docs/v0.1/phase1/acceptance-test-charter.md \
  docs/v0.1/phase1/review-template.md \
  docs/v0.1/phase1/evidence-phase1.md \
  TODO_v01.md; do
  if [ -f "$f" ]; then
    echo "FOUND $f"
  else
    echo "MISSING $f"
  fi
done

echo "== Acceptance content checks =="
grep -n "Success Criteria Traceability" docs/v0.1/phase1/scope-contract.md
grep -n "Deferred Items / Non-Goals" docs/v0.1/phase1/scope-contract.md
grep -n "Additive-only evolution" docs/v0.1/phase1/protocol-constraints.md
grep -n "Field reservation/no reuse" docs/v0.1/phase1/protocol-constraints.md
grep -n "Buf validation expectation" docs/v0.1/phase1/protocol-constraints.md
grep -n "Phase 1 P0 Task Owner Map" docs/v0.1/phase1/ownership-adr.md
grep -n "ADR Template" docs/v0.1/phase1/ownership-adr.md
grep -n "Repeatability Checklist" docs/v0.1/phase1/acceptance-test-charter.md

echo "== Phase 1 checkbox checks =="
grep -n "P1-T1 Freeze v0.1 scope contract" TODO_v01.md
grep -n "P1-T2 Build protocol constraints checklist" TODO_v01.md
grep -n "P1-T3 Define ownership and decision cadence" TODO_v01.md
grep -n "P1-T4 Define acceptance test charter" TODO_v01.md

echo "== Phase 1 sub-task checks =="
grep -n "Document v0.1 must-have user outcomes." TODO_v01.md
grep -n "Document v0.1 non-goals list." TODO_v01.md
grep -n "Link each success criterion to at least one execution task." TODO_v01.md
grep -n "Add single-binary mode invariant." TODO_v01.md
grep -n "Add protobuf additive-only minor-version rule." TODO_v01.md
grep -n "Add multistream major-version evolution rule." TODO_v01.md
grep -n "Add AEP and multi-implementation validation requirement for breaking changes." TODO_v01.md
grep -n "Assign role owners for protocol, networking, crypto, UI, ops, QA." TODO_v01.md
grep -n "Define escalation path for blocking decisions." TODO_v01.md
grep -n "Define merge and review policy for critical path changes." TODO_v01.md
grep -n "Define device and network prerequisites." TODO_v01.md
grep -n "Define data reset method between runs." TODO_v01.md
grep -n "Define evidence capture format." TODO_v01.md

echo "== Conditional Go checks =="
if [ -f go.mod ]; then
  echo "RUN go test ./..."
else
  echo "SKIP go test ./... (go.mod not present)"
fi

echo "== Conditional Buf checks =="
if find . -maxdepth 4 -type f \( -name "buf.yaml" -o -name "buf.work.yaml" \) | grep -q .; then
  echo "RUN buf lint"
  echo "RUN buf breaking"
  echo "RUN buf generate"
else
  echo "SKIP buf lint/breaking/generate (no buf.yaml or buf.work.yaml present)"
fi
'
```

## Output

```text
== Artifact existence checks ==
FOUND docs/v0.1/phase1/scope-contract.md
FOUND docs/v0.1/phase1/protocol-constraints.md
FOUND docs/v0.1/phase1/ownership-adr.md
FOUND docs/v0.1/phase1/acceptance-test-charter.md
FOUND docs/v0.1/phase1/review-template.md
FOUND docs/v0.1/phase1/evidence-phase1.md
FOUND TODO_v01.md
== Acceptance content checks ==
25:## Success Criteria Traceability (Phase 1)
17:## Deferred Items / Non-Goals
9:1. **Additive-only evolution:** minor-band changes are limited to adding fields/messages/enums.
10:2. **Field reservation/no reuse:** once a field number exists, it cannot be repurposed; removed numbers/names must be marked `reserved`.
14:6. **Buf validation expectation:** when `.proto`/`buf` assets are present, run `buf lint`, `buf breaking`, and `buf generate` (if generation is configured); if assets are absent, record explicit SKIP evidence.
13:## Phase 1 P0 Task Owner Map
1:# Ownership Matrix and ADR Template (Phase 1)
26:## ADR Template
34:## Repeatability Checklist (Run-Level)
== Phase 1 checkbox checks ==
163:- [x] `[Docs][P0][Effort:S][Owner:Tech Lead]` **P1-T1 Freeze v0.1 scope contract**
181:- [x] `[Docs][P0][Effort:S][Owner:Protocol Lead]` **P1-T2 Build protocol constraints checklist**
202:- [x] `[Ops][P0][Effort:S][Owner:Engineering Manager]` **P1-T3 Define ownership and decision cadence**
220:- [x] `[Validation][P0][Effort:S][Owner:QA Lead]` **P1-T4 Define acceptance test charter for five-minute first contact**
== Phase 1 sub-task checks ==
174:    - [x] Document v0.1 must-have user outcomes.
176:    - [x] Document v0.1 non-goals list.
178:    - [x] Link each success criterion to at least one execution task.
193:    - [x] Add single-binary mode invariant.
195:    - [x] Add protobuf additive-only minor-version rule.
197:    - [x] Add multistream major-version evolution rule.
199:    - [x] Add AEP and multi-implementation validation requirement for breaking changes.
213:    - [x] Assign role owners for protocol, networking, crypto, UI, ops, QA.
215:    - [x] Define escalation path for blocking decisions.
217:    - [x] Define merge and review policy for critical path changes.
231:    - [x] Define device and network prerequisites.
233:    - [x] Define data reset method between runs.
235:    - [x] Define evidence capture format.
== Conditional Go checks ==
SKIP go test ./... (go.mod not present)
== Conditional Buf checks ==
SKIP buf lint/breaking/generate (no buf.yaml or buf.work.yaml present)
```
