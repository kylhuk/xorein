# Risk Register

> As-built artifact that tracks residual risks now that the mandatory command evidence and manifests have been recorded.

| ID | Risk | Mitigation | Exit criterion | Status |
|---|---|---|---|---|
| R21-1 | Local store corruption causes data loss or crash loops. | Deterministic corruption detection, safe fallback stories, restart tests, and logs that surface `STORE_CORRUPT`. | Corruption regressions show deterministic recovery in `tests/e2e/v21` or `go test ./pkg/v21`. | monitoring (store regression suite passed; continue to watch corruption probes) |
| R21-2 | Search index might surface redacted content or index tombstones incorrectly. | Tombstone enforcement, index rebuild workflows, and dedicated search/privacy tests. | Redaction/search tests pass (`tests/e2e/v21/search_*`, `tests/perf/v21`). | monitoring (search/perf suites validated coverage; keep tombstone flows under observation) |
| R21-3 | Retention/prune cycles break pagination or cause data drift. | Stable cursor contracts, property tests, and `pkg/v21/store/retention` invariants. | Retention regression tests complete without pagination regressions. | monitoring (retention/pagination invariants exercised; watch for future drift) |
| R21-4 | UX surfaces expose ambiguous history availability states. | Explicit timeline banners, search coverage labels, and reason taxonomy documentation. | UX walkthrough/rerun ensures no ambiguous states. | monitoring (timeline/search scenarios passed; keep UX banners under verification) |
| R21-5 | Relay accidentally stores durable history segments, violating the no-long-history boundary. | Persistence Podman scenarios, targeted regression scripts, and relay manifest proofs. | `scripts/v21-persistence-scenarios.sh` reports relay passes; `EV-v21-G8-001` recorded. | closed (manifest confirms relay no-long-history-hosting probe pass) |
| R21-6 | Backend frontend boundary leaks UI dependencies into Xorein. | Header-level boundary doc (`docs/v2.1/phase0/p0-binary-boundary.md`) plus `go build ./cmd/xorein` without Gio imports. | `go build ./cmd/xorein` completes cleanly (`EV-v21-G7-001`). | closed (Xorein/harmolyn builds pass; UI deps remain isolated) |
| R21-7 | Evidence capture lags command outputs, delaying gate sign-off. | Evidence index placeholders, command-to-EV mapping, and `P5-T1` governance doc updates. | All mandatory EV rows (G1–G9) reference real artifact files. | closed (EV rows now point at recorded artifacts; bundle annotated) |

The register now serves as the as-built record for residual risk once the corresponding evidence paths materialized.
