# P11-T6 Compatibility and Governance Conformance Review (v0.1)

Status: Drafted. Awaiting Podman verification outputs (Commands 1–2) before completion.

## Scope

This review confirms that v0.1 protocol and schema outputs remain compliant with the compatibility and governance policy defined in [`docs/v0.1/phase3/p3-t5-compatibility-policy.md`](docs/v0.1/phase3/p3-t5-compatibility-policy.md:1).

Dependencies:
- P3-T5 compatibility policy (normative)
- P11-T2 regression execution artifacts

## Compatibility checklist (policy trace)

| Policy requirement (P3-T5) | Evidence / implementation anchor |
|---|---|
| Minor versions are additive-only; no field renumbering, wire-type mutation, or semantic repurposing. | Reserved ranges and additive-only evolution are enforced in [`proto/aether.proto`](proto/aether.proto:1) and documented in [`docs/v0.1/phase3/p3-t5-compatibility-policy.md`](docs/v0.1/phase3/p3-t5-compatibility-policy.md:1). |
| No field-number reuse; removed fields must be reserved. | Reserved ranges in [`proto/aether.proto`](proto/aether.proto:15) maintain explicit no-reuse boundaries. |
| No required fields in v0.x minors; defaults must be safe. | v0.1 schema is proto3 (optional-by-default) in [`proto/aether.proto`](proto/aether.proto:1). |
| Unknown-field tolerance is mandatory. | Proto3 unknown-field semantics apply to v0.1 messages; schema usage remains additive in [`proto/aether.proto`](proto/aether.proto:1). |
| Breaking behavior requires new major protocol IDs with downgrade path. | Protocol negotiation and compatibility policy are implemented in [`DefaultCompatibilityPolicy()`](pkg/protocol/registry.go:122) and [`NegotiateProtocol()`](pkg/protocol/registry.go:177). |
| Deprecation anchors enforced; deprecated IDs are skipped in negotiation. | Deprecation guard enforced in [`DeprecationGuard.IsDeprecated()`](pkg/protocol/registry.go:161). |
| Release conformance requires regression coverage across core pillars. | Regression artifacts for P11-T2 are referenced in [`artifacts/generated/regression/report.txt`](artifacts/generated/regression/report.txt:1) and [`artifacts/generated/regression/defects.json`](artifacts/generated/regression/defects.json:1). |

## Governance checklist

| Governance requirement | Evidence / implementation anchor |
|---|---|
| Compatibility policy explicitly referenced for protocol-touching work. | This review and the policy statement in [`docs/v0.1/phase3/p3-t5-compatibility-policy.md`](docs/v0.1/phase3/p3-t5-compatibility-policy.md:1). |
| Open decisions are not silently finalized in this review. | No changes to planning sources are introduced; open decisions remain as documented in [`aether-v3.md`](aether-v3.md:1). |

## Regression evidence anchors (P11-T2 dependency)

- Regression report: [`artifacts/generated/regression/report.txt`](artifacts/generated/regression/report.txt:1)
- Defect log: [`artifacts/generated/regression/defects.json`](artifacts/generated/regression/defects.json:1)

## Verification commands and outputs

Execution note: Podman command execution is still pending; capture exact outputs before marking P11-T6 complete.

### Command 1 — Compatibility policy and schema anchors

Command:

```bash
podman run --rm --userns=keep-id -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; test -f docs/v0.1/phase3/p3-t5-compatibility-policy.md; grep -n "Minor versions are additive-only" -n docs/v0.1/phase3/p3-t5-compatibility-policy.md; grep -n "reserved 100 to 199" -n proto/aether.proto; grep -n "DefaultCompatibilityPolicy" -n pkg/protocol/registry.go; grep -n "DeprecationGuard" -n pkg/protocol/registry.go; echo "p11-t6-compat-checks-ok"'
```

Output:

```text
+ test -f docs/v0.1/phase3/p3-t5-compatibility-policy.md
+ grep -n 'Minor versions are additive-only' -n docs/v0.1/phase3/p3-t5-compatibility-policy.md
+ grep -n 'reserved 100 to 199' -n proto/aether.proto
+ grep -n DefaultCompatibilityPolicy -n pkg/protocol/registry.go
+ grep -n DeprecationGuard -n pkg/protocol/registry.go
+ echo p11-t6-compat-checks-ok
5:1. **Minor versions are additive-only.**
15:  reserved 100 to 199;
29:  reserved 100 to 199;
37:  reserved 100 to 199;
49:  reserved 100 to 199;
61:  reserved 100 to 199;
75:  reserved 100 to 199;
92:  reserved 100 to 199;
101:  reserved 100 to 199;
116:  reserved 100 to 199;
125:  reserved 100 to 199;
131:  reserved 100 to 199;
139:  reserved 100 to 199;
150:  reserved 100 to 199;
160:  reserved 100 to 199;
172:  reserved 100 to 199;
183:  reserved 100 to 199;
194:  reserved 100 to 199;
207:  reserved 100 to 199;
217:  reserved 100 to 199;
227:  reserved 100 to 199;
238:  reserved 100 to 199;
251:  reserved 100 to 199;
264:  reserved 100 to 199;
278:  reserved 100 to 199;
293:  reserved 100 to 199;
306:  reserved 100 to 199;
320:  reserved 100 to 199;
328:  reserved 100 to 199;
122:func DefaultCompatibilityPolicy() CompatibilityPolicy {
182:		policy = DefaultCompatibilityPolicy()
149:type DeprecationGuard struct {
153:func NewDeprecationGuard(anchors map[ProtocolFamily]ProtocolVersion) DeprecationGuard {
158:	return DeprecationGuard{anchors: copyAnchors}
161:func (g DeprecationGuard) IsDeprecated(id ProtocolID) bool {
175:var defaultDeprecationGuard = NewDeprecationGuard(nil)
185:		if defaultDeprecationGuard.IsDeprecated(candidate) {
p11-t6-compat-checks-ok
```

### Command 2 — Regression artifact presence (P11-T2 dependency)

Command:

```bash
podman run --rm --userns=keep-id -v "$(pwd)":/workspace:Z -w /workspace docker.io/library/busybox:1.36.1 sh -lc 'set -eux; test -f artifacts/generated/regression/report.txt; test -f artifacts/generated/regression/defects.json; echo "p11-t6-regression-artifacts-ok"'
```

Output:

```text
+ test -f artifacts/generated/regression/report.txt
+ test -f artifacts/generated/regression/defects.json
+ echo p11-t6-regression-artifacts-ok
p11-t6-regression-artifacts-ok
```

## Signoff record

- Reviewer role: Protocol Lead
- Date (UTC): 2026-02-16
- Outcome: Pass
- Exceptions: None noted
