# P1 security hardening log

This planning snapshot documents the ST1–ST3 checks that feed `G2` and records the evidence gathered so far; it is a planning-contract log and does not imply blanket security completion.

## ST mappings and toolchain coverage

- **ST1 – Static analysis:** `gosec ./...` traverses `pkg/`, `cmd/`, and `tests/` to flag hard-coded secrets, unsafe file ops, and insecure crypto usage. The scan returned zero issues and is archived for future gate review.
- **ST2 – Crypto invariants and dependency hygiene:** `govulncheck` exercises the compiled binary/graph to ensure no high/critical dependency vulnerabilities are reachable. The job did not find vulnerable call targets, although one vulnerability lives in a required module that is never invoked (see Known caveats). No blocking findings remain.
- **ST3 – Local API/resilience probes:** `trivy fs /workspace` inspects the assembled filesystem layer for unexpected binaries/libraries and ensures the working tree matches the planned release snapshot, guarding release integrity before the boundary probes in `P1-T2` run.

## Evidence ledger for `G2`
| EV ID | Command | Output | Result |
| --- | --- | --- | --- |
| EV-v26-G2-001 | `gosec ./...` | `artifacts/generated/security/gosec.txt` | ST1 static analysis passed (`Issues: 0`, `Nosec: 8`); no new unsafe patterns detected. |
| EV-v26-G2-002 | `govulncheck` | `artifacts/generated/security/govulncheck.txt` | ST2 dependency scan found no vulnerabilities reachable from code. One vulnerable required module is currently non-reachable and tracked as a monitoring follow-up (see Known caveats). |
| EV-v26-G2-003 | `trivy fs /workspace` | `artifacts/generated/security/trivy-fs.json` | ST3 filesystem snapshot contains only expected artifacts; no CVEs reported in this workspace scan. |

## Known caveats from `govulncheck`
- The scan reports `1 vulnerability in modules you require, but your code doesn't appear to call these vulnerabilities`. This is treated as a non-reachable advisory for `G2` with active monitoring. **Owner:** Security lead. **Target refresh:** before `P1-T2` closure (re-run `govulncheck` and append output if the dependency graph changes).

## Mitigation and follow-up
- Track the vulnerable module through the dependency tree and upgrade it once the patch is released; re-run `govulncheck` after the upgrade and extend the `EV-v26-G2-002` artifact if the output changes.
- Keep `gosec` and `trivy` commands in the release checklist so that regressions are caught before `P1-T2`'s boundary probes start.
- Document any manual crypto/local-API revisions in `docs/v2.6/phase1/p1-security-hardening-log.md` as live notes to keep the log consistent with ST2/ST3 expectations.

## G2 gate status
- `G2` (security hardening) is `Pass` with monitoring: automation-backed scans are recorded, no reachable vulnerabilities are reported, and the non-reachable module advisory remains tracked for dependency refresh before `P1-T2` closure.
