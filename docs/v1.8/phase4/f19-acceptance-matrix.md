# F19 Acceptance Matrix

| Scenario | Expected Outcome | Evidence |
|---|---|---|
| Fallback indexer path selected when primary is unreachable | Client returns a deterministic non-terminal blocker with reason `indexer-unreachable` | v19 spec implementation + automated acceptance tests |
| Join blocked by trust policy | Client shows stable blocker reason and preserves join intent state | v19 spec + user flow tests |
| Continuity restore after app suspend/restore | Incomplete join keeps last signed path and bounded retry metadata | v19 continuity tests |
| No-limbo UX across path transitions | UI never shows an undefined waiting state; one blocker state is always visible | v19 UX tests |
| QoL transition reason traceability | Transition reason includes path source, trust result, and next suggested action | v19 runtime + UI log assertions |
