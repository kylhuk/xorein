# Phase 3 · P3-T1 Scoped Search Filter Contract

## Objective
Define the SQLCipher FTS5 indexing, required filters, and scoped-response semantics so V7-G3 can gate the search contract and keep results confined to channel/server/DM boundaries before any implementation touches `pkg/v07/search`.

## Contract
### Scoped indexing model
- The index document model, tokenization profile, normalization rules, and scope partition keys (including `mode_epoch_id`) live in `pkg/v07/search/contracts.go` and guarantee that identical corpora produce canonical entries for channel, server, and DM scopes.

### Search lifecycle & required filters
- Filter combinatorics (from user, date range, has file, has link) plus query normalization reside in `pkg/v07/search/contracts.go`. Invalid combinations are rejected deterministically (including malformed date ranges), and valid queries map to canonical tokens regardless of input ordering or casing.

### Response envelope & authorization guardrails
- Response ordering, pagination tokens, partial failure signaling, and authorization/redaction checks are spelled out in the same filter contract, ensuring scoped results respect privacy boundaries and provide explicit degraded-mode signals.

### Privacy & evidence requirements
- SQLCipher-at-rest assumptions and evidence requirements for scoped search operations (including auditing the required filters) are captured in this doc so V7-G3 can track compliance before gating.

## Evidence anchors
| Artifact | Description | Evidence anchor |
|---|---|---|
| `VA-S1` | Scoped index entries | Section "Scoped indexing model" |
| `VA-S2` | Index lifecycle tied to retention | Same section |
| `VA-S3` | Filter normalization & invalid combinations | Section "Search lifecycle & required filters" |
| `VA-S4` | Response pagination and partial failures | Section "Response envelope & authorization guardrails" |
| `VA-S5` | Authorization and scope isolation | Same section |
| `VA-S6` | Privacy/evidence assumptions | Section "Privacy & evidence requirements" |

This narrative keeps the scoped-search contracts and required filter semantics auditable for V7-G3 while referencing the future `pkg/v07/search` seams.
