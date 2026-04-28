# Append-only issues for protocol-reachability


## 2026-04-21 static protocol/client audit
- No repo-Go-client scope mismatch found in the audited runtime and transport code. The only caution is documentation scope: README describes the broader protocol family, while current reachability claims should stay limited to the repo Go client using `/aether/peer/0.1.0` plus the `cap.peer.*` transport capabilities above.

- No implementation blockers encountered while adding the separate-process smoke harness; the only issue found during development was a compile-time missing return in the new test helper, which was fixed before verification.

## 2026-04-21 task 3 verdict packaging
- Keep future wording narrow: the evidence only verifies repo Go client reachability, not external interoperability or third-party client compatibility.
