# Phase 2 P2-T10: SQLCipher Integration Research Decision + Portability Risk Matrix

## Purpose
Capture the research decision record for SQLCipher integration, plus a portability risk matrix and linked mitigations. This is a Phase 2 research artifact with explicit planned vs. implemented separation.

## Context and references
- Task definition and deliverables: [`TODO_v01.md`](TODO_v01.md:369).
- Research exit rule (decision + rejected alternatives + residual risk): [`TODO_v01.md`](TODO_v01.md:1309).
- Risk register entry R5 (portability risk): [`TODO_v01.md`](TODO_v01.md:1325).
- Go ecosystem fit guidance (CGo only when required): [`aether-v3.md`](aether-v3.md:248).

## Decision record (draft)
- **Status:** Draft (research in progress; no implementation).
- **Decision statement:** Pending selection of SQLCipher integration approach. Final choice must document portability constraints and fallback options, per P2-T10 acceptance criteria in [`TODO_v01.md`](TODO_v01.md:375).
- **Decision drivers:**
  - Cross-platform portability and build stability (R5). [`TODO_v01.md`](TODO_v01.md:1325)
  - Minimize CGo surface; isolate behind interfaces. [`aether-v3.md`](aether-v3.md:248)
  - Migration and key management expectations must be defined before implementation. [`TODO_v01.md`](TODO_v01.md:379)

### Candidate approaches (comparison scaffold)
| Option | Summary | Portability considerations | CGo boundary | Status |
|---|---|---|---|---|
| A | SQLCipher via CGo-wrapped driver | Requires native toolchain per target OS | Required | Under review |
| B | SQLCipher via prebuilt library bindings | Toolchain + prebuilt artifacts | Required | Under review |
| C | Pure-Go encrypted storage alternative (non-SQLCipher) | Highest portability, but deviates from SQLCipher baseline | None | Rejected (out of scope unless SQLCipher fails) |

**Rejected alternatives:** None formally recorded yet; Option C is marked out-of-scope unless SQLCipher feasibility fails.

## Portability risk matrix + mitigations
| Risk ID | Risk | Trigger/Signal | Impact | Mitigation link | Status |
|---|---|---|---|---|---|
| R5 | SQLCipher integration portability blocks builds | CI or local builds fail due to CGo or native library gaps | Blocks encrypted storage baseline | P2-T10 decision record (this doc) + portability validation plan (planned) | Active |
| R5-a | Platform toolchain drift | Build failures on specific OS/arch | Inconsistent developer/CI environments | Define supported platform matrix + toolchain doc (planned) | Planned |
| R5-b | Migration instability | Schema/crypto upgrades fail across versions | Data loss or rollback requirements | Define migration expectations and test plan (planned) | Planned |
| R5-c | Key management ambiguity | Unclear key lifecycle or rotation | Security regressions | Document key management expectations (planned) | Planned |

## Planned mitigations (linked actions)
- **Define migration and key management expectations** (P2-T10 sub-task). [`TODO_v01.md`](TODO_v01.md:379)
- **Define CI validation for encrypted DB operations** (P2-T10 sub-task). [`TODO_v01.md`](TODO_v01.md:381)
- **Update downstream tasks with concrete decision outcomes** (Research exit rule). [`TODO_v01.md`](TODO_v01.md:1311)

## Migration and key management expectations (policy scaffold)

- Migration expectations (planned):
  - Migrations are forward-only and version-stamped; rollback path requires documented backup/restore procedure.
  - Encrypted schema changes must be validated against at least one prior schema version before merge.
- Key management expectations (planned):
  - Key material is never persisted in plaintext in repo artifacts or logs.
  - Key lifecycle actions (provision, rotate, revoke) require explicit operator runbook steps and audit evidence.
  - Recovery/fallback mode must be defined before enabling encrypted storage by default.

## CI validation policy for encrypted DB operations (planned)

- Required CI checks once implementation exists:
  1. create/open encrypted DB,
  2. migration from previous encrypted schema,
  3. negative test for wrong-key open,
  4. portability smoke on supported OS/arch matrix.
- Evidence requirement: CI logs must show each check result and the selected SQLCipher driver/toolchain path.

## Planned vs. implemented separation
No SQLCipher integration is implemented in this repository. This document records research intent and scaffolding only. All mitigations are planned actions and must not be treated as active controls until corresponding tasks are completed.
