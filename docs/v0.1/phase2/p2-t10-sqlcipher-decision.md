# Phase 2 P2-T10: SQLCipher Integration Research Decision + Portability Risk Matrix

## Purpose
Capture the research decision record for SQLCipher integration, plus a portability risk matrix and linked mitigations. This is a Phase 2 research artifact with explicit planned vs. implemented separation.

## Context and references
- Task definition and deliverables: [`TODO_v01.md`](TODO_v01.md:369).
- Research exit rule (decision + rejected alternatives + residual risk): [`TODO_v01.md`](TODO_v01.md:1309).
- Risk register entry R5 (portability risk): [`TODO_v01.md`](TODO_v01.md:1325).
- Go ecosystem fit guidance (CGo only when required): [`aether-v3.md`](aether-v3.md:248).

## Decision record
- **Status:** Finalized research decision for planning scope (no implementation yet).
- **Decision statement:** Select **Option A** (SQLCipher via CGo-wrapped driver) as the v0.1 baseline path, with explicit fallback to non-default plaintext local storage mode only for unsupported environments until SQLCipher portability gates are satisfied.
- **Decision drivers:**
  - Cross-platform portability and build stability (R5). [`TODO_v01.md`](TODO_v01.md:1377)
  - Minimize CGo surface; isolate behind interfaces. [`aether-v3.md`](aether-v3.md:248)
  - Migration and key management expectations must be defined before implementation. [`TODO_v01.md`](TODO_v01.md:402)
  - Research Exit Rule requires explicit decision, rejected alternatives, residual risks, and downstream task updates. [`TODO_v01.md`](TODO_v01.md:1360)

### Candidate approaches (comparison scaffold)
| Option | Summary | Portability considerations | CGo boundary | Status |
|---|---|---|---|---|
| A | SQLCipher via CGo-wrapped driver | Requires native toolchain per target OS | Required | **Selected** |
| B | SQLCipher via prebuilt library bindings | Toolchain + prebuilt artifacts | Required | Rejected |
| C | Pure-Go encrypted storage alternative (non-SQLCipher) | Highest portability, but deviates from SQLCipher baseline | None | Rejected |

**Rejected alternatives:**
- **Option B rejected:** prebuilt-binding supply-chain and version-drift risks are higher than a constrained in-repo CGo boundary, and cross-platform reproducibility expectations would still require native packaging work.
- **Option C rejected:** out of scope for v0.1 SQLCipher requirement and would invalidate risk/mitigation linkage to R5 without reducing migration/key-management complexity.

## Portability risk matrix + mitigations
| Risk ID | Risk | Trigger/Signal | Impact | Mitigation link | Status |
|---|---|---|---|---|---|
| R5 | SQLCipher integration portability blocks builds | CI or local builds fail due to CGo or native library gaps | Blocks encrypted storage baseline | P2-T10 decision record (this doc) + P5-T4/P5-T5 constraints (defined below) | Active |
| R5-a | Platform toolchain drift | Build failures on specific OS/arch | Inconsistent developer/CI environments | Pin supported OS/arch + compiler toolchain matrix in storage build docs before merge | Planned |
| R5-b | Migration instability | Schema/crypto upgrades fail across versions | Data loss or rollback requirements | Enforce backup-before-migration and wrong-key negative tests in storage validation | Planned |
| R5-c | Key management ambiguity | Unclear key lifecycle or rotation | Security regressions | Key lifecycle runbook + explicit fallback mode boundaries before default enablement | Planned |

## Residual risks
- SQLCipher CGo dependency may still block certain targets until platform packaging/toolchain standardization is completed.
- Encrypted migration rollback safety remains contingent on backup/restore discipline and pre-migration validation coverage.
- Operator misconfiguration risk persists until key lifecycle runbook and verification checks are integrated into release gating.

## Planned mitigations (linked actions)
- **Define migration and key management expectations** (P2-T10 sub-task). [`TODO_v01.md`](TODO_v01.md:402)
- **Define CI validation for encrypted DB operations** (P2-T10 sub-task). [`TODO_v01.md`](TODO_v01.md:404)
- **Update downstream tasks with concrete decision outcomes** (Research exit rule). [`TODO_v01.md`](TODO_v01.md:1365)

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

## Downstream task constraint updates (Research Exit Rule item 4)
- **P5-T4 execution constraints:**
  - Migration plan must include encrypted-schema forward migration plus backup-before-migration requirement.
  - Migration fixtures must include wrong-key and unavailable-SQLCipher negative paths.
- **P5-T5 execution constraints:**
  - End-to-end lifecycle validation must include encrypted-open success, wrong-key failure semantics, and fallback-mode behavior checks.
  - Validation report must reference the selected Option A driver boundary and residual risk status for R5.
- **Risk R5 linkage update:**
  - R5 closure requires CI evidence for encrypted DB create/open/migrate/wrong-key paths on declared supported OS/arch matrix.

## Planned vs. implemented separation
No SQLCipher integration is implemented in this repository. This document records finalized research decisions and planning constraints only. All mitigations remain planned actions and must not be treated as active controls until corresponding build/validation tasks are completed.
