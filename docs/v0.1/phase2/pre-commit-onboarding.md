# Pre-commit Onboarding (Phase 2)

1. Install [`pre-commit`](https://pre-commit.com/) via the platform package manager (pip, Homebrew, etc.) or the provided bootstrap script.
2. Run `pre-commit install` to wire the hooks into your git repo.
3. The repo currently uses the [`pre-commit-hooks`](.pre-commit-config.yaml:1) collection; baseline hooks validate YAML/JSON syntax, merge-conflict markers, protected branch commits, and whitespace hygiene.
4. CI triggers the same hooks via `make check-fast`, so local execution keeps PR hygiene aligned with automation.
5. To run all hooks locally without committing, use `pre-commit run --all-files`.

## Bypass policy and review requirements (P2-T5)

- Default policy: bypass is disallowed for normal development flow. Commits to `main` are blocked locally by [`no-commit-to-branch`](.pre-commit-config.yaml:1).
- Emergency bypass (`SKIP=<hook-id> git commit ...`) is allowed only when:
  1. the failure is a confirmed false positive or tool outage, and
  2. a follow-up remediation task is created before merge.
- Any bypassed PR must include:
  - the exact skipped hook IDs,
  - rationale and impact assessment,
  - reviewer acknowledgment from the owning domain (Platform/DevEx for repo hooks).
- Merge rule: unresolved bypasses are merge-blocking unless an explicit approver note is recorded in PR review.
