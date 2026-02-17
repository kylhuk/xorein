# Phase 4 · P4 Emoji & Reaction Contracts

## Purpose
Describe the deterministic contracts covering custom emoji lifecycle, picker resolution, reaction state, and governance integration from Phase 4 of `TODO_v05.md`.

## Deterministic obligations
- Emoji upload validation, normalization, naming collision handling, and rejection reasons follow the schema in P4-T1 so that the max-50 quota is enforced uniformly across server operators and bots in VA-E1.
- The emoji picker data model, ordering keys, availability filtering, and shortcode tokenization/resolution rules in P4-T2 mandate deterministic selection and fallback text handling, documented in VA-E2.
- Reaction actor uniqueness, add/remove/toggle transitions, duplicate suppression, count aggregation, and recovery behavior from P4-T3 ensure single convergent state per actor/emoji pair even under retries; VA-E3 records the expected outcomes.
- Permission mapping, moderation guardrails, and audit-entry requirements for emoji and reaction actions in P4-T4 tie back to v0.4 governance controls so that every action class has deterministic permission and audit behavior described in VA-E4.

## Deterministic contract tables
| Input / condition | Outcome + reason codes | Artifact | Validation obligation |
|---|---|---|---|
| Emoji upload + quota enforcement | Positive: `emoji.upload.success`; Negative: `emoji.upload.invalid` on malformed asset or quota breach; Recovery: `emoji.upload.recover` via replacement/deletion. | VA-E1 | Anchor upload/quota schemas and rejection codes so reviewers can confirm reason-coded outcomes appear in the verification matrix and scenario pack. |
| Emoji picker + shortcode resolution | Positive: `emoji.picker.success` for deterministic selection; Negative: `emoji.picker.conflict` on collisions; Recovery: `emoji.picker.fallback` to text fallback. | VA-E2 | Document ordering/fallback rules and tie to scenario packs covering missing assets or ambiguous shortcodes. |
| Reaction lifecycle | Positive: `emoji.reaction.success` for converge add/remove; Negative: `emoji.reaction.invalid` for duplicate/stale input; Recovery: `emoji.reaction.recover` reconciliation path. | VA-E3 | Keep reaction toggles and aggregation reason codes linked to verification tests and integrated scenario events. |
| Governance for emoji/reaction actions | Positive: `emoji.gov.success` for authorized actions; Negative: `emoji.gov.denied` for permission failures; Recovery: `emoji.gov.recover` during audit/appeal. | VA-E4 | Map permission matrix and audit hooks to reason-class taxonomy for defendable governance coverage in release dossiers. |

These tables extend the verification matrix to cover emoji/reaction deterministic contracts and their validation obligations.
