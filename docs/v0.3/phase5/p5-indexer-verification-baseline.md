# v0.3 Phase 5 - P5 Indexer and Verification Baseline

> Status: Execution artifact. Optional indexer posture and signed response verification now reference `pkg/v03/indexer` deliverables.

## Purpose

Describe how optional community-run indexers integrate, remain non-authoritative, and how clients verify signed responses before accepting discovery results.

## Scope Summary

- Optional indexer reference contract with interchangeability and revocability requirements.
- Signed directory/search response verification plus invalid-signature handling.
- Multi-indexer query/merge/de-dup guidance preserving privacy and trust posture.

## Acceptance Anchors

1. Indexer usage is always optional; clients may skip indexers without blocking exploration.
2. Signed responses include deterministic fields (issuer, timestamp, signature) and invalid-signature handling paths that reject payloads before processing.
3. Multi-indexer merge/de-dup preserves non-authoritative trust model and respects privacy-preserving query posture (no forced multi-query by default unless explicitly configured).

## Evidence Mapping

| Contract | Doc | Code/Test Evidence |
|---|---|---|
| Optional indexer contract | `docs/v0.3/phase5/p5-indexer-verification-baseline.md` | `pkg/v03/indexer/contracts.go`, `pkg/v03/indexer/contracts_test.go` |
| Signed response verification | `docs/v0.3/phase5/p5-indexer-verification-baseline.md` | `pkg/v03/indexer/contracts_test.go` |
| Multi-indexer query guidance | `docs/v0.3/phase5/p5-indexer-verification-baseline.md` | `pkg/v03/indexer/contracts.go`, `pkg/v03/indexer/contracts_test.go` |

## Trust Summary

- Any signed response failure route must log the reason and expose the rejection state via release-phase governance documents without implying removal of optionality.
