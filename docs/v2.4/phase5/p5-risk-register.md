# Phase 5 risk register (P5-T1 ST2–ST4)

| ID | Risk | Mitigation | Evidence / Exit criterion |
| --- | --- | --- | --- |
| R24-1 | Local API regression or multi-client crash recovery gaps delay G5/G6 closure. | Podman harness `scripts/v24-daemon-scenarios.sh` + scenario manifest/log (`artifacts/generated/v24-daemon-scenarios/manifest.txt`) exercise daemon reuse, crash recovery, and stale socket handling. | EV-v24-G8-009 plus narrative in `p5-evidence-bundle.md` establish multi-client/regression coverage. |
| R24-2 | Build or lint warnings hide protocol violations in `cmd/xorein` or `cmd/harmolyn`. | `buf lint`, `buf breaking`, `go build` for both binaries, and `make check-full` run; advisory warnings noted and tracked in this register. | EV-v24-G8-001…EV-v24-G8-008 (warnings are deprecated `DEFAULT` name + Trivy flag; gate-ready because no failures). |
| R24-3 | UI dependency drift (G9) introduces Gio imports into backend binaries. | `scripts/ci/enforce-boundaries.sh` confirms `cmd/xorein`/`pkg/xorein` remain Gio-free and ensures `cmd/harmolyn` only references the local API client. | EV-v24-G9-001 with logs in `artifacts/generated/v24-evidence/enforce-boundaries.txt`. |
| R24-4 | F25 spec package lacks clarity for downstream teams. | Release the F25 blob-store spec, proto delta, and acceptance matrix under `docs/v2.4/phase4/` and tie them to the as-built conformance narrative. | Documented in `p5-as-built-conformance.md` plus the F25 doc set; reviewers close the risk when the spec package is published and referred in G7/G8 artifacts. |

## Residual risk disposition
- R24-1 remains monitored by collecting scenario manifests/logs for future regression cycles; nothing new beyond EV-v24-G8-009 was required for this closure.
- R24-2 and R24-3 have advisory but acceptable warning levels; no further action is planned unless the tooling surface escalates to a failure.
- R24-4 is closed with the published F25 docs but remains a communication risk for v25 teams; include the spec links in release notes to keep downstream consumers aligned.
