# v0.3 Phase 4 - P4 Discovery and Admission Baseline

> Status: Execution artifact. Directory entry publication, explore/preview, and invite/request contracts reference `pkg/v03/discovery` deliverables.

## Purpose

Document deterministic contracts for `DirectoryEntry` publication/retrieval, explore/discover browsing plus preview, and invite/request-to-join workflows pulled forward into v0.3.

## Scope Summary

- `DirectoryEntry` create/update/withdraw flows with deterministic TTL and freshness guidance.
- Explore/discover browsing, server preview rendering, and stale-data failure handling.
- Invite plus request-to-join lifecycle with policy matrix and deterministic outcomes.

## Acceptance Anchors

1. Publish/retrieve operations include reason codes for stale entries, intermittent publisher availability, and evict signals tied to OD3 decisions.
2. Explore/preview flows specify deterministic staleness handling, partial data fallbacks, and preview security disclosures.
3. Invite/request lifecycle enumerates positive (accepted), neutral (pending), negative (denied/canceled) outcomes with policy references.

## Evidence Mapping

| Contract | Doc | Code/Test Evidence |
|---|---|---|
| DirectoryEntry contract | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |
| Explore + preview | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |
| Invite/request | `docs/v0.3/phase4/p4-discovery-admission-baseline.md` | `pkg/v03/discovery/contracts.go`, `pkg/v03/discovery/contracts_test.go` |

## Decision Linkage

- OD3-01 through OD3-04 remain open; this doc references them for TTL, ranking, upgrade path, and query posture without resolving them.
