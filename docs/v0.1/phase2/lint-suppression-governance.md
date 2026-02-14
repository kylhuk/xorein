# Lint Suppression Governance (Phase 2)

This document explains how the Phase 2 foundation discipline uses [`golangci-lint`](.golangci.yml:1).

## Severity Policy

- Rules configured in [`.golangci.yml`](.golangci.yml:1) default to error severity for `errcheck` and `staticcheck` while `unused` remains at warning level to avoid noisy blockers while still calling them out.
- Any deviation from the severity map requires explicit justification in a pull request, referencing the failing check and describing why it is acceptable for the given surface.

## Suppressions

- Suppressing a linter must happen through well-justified inline comments (e.g., `//nolint:errcheck`) only when:
  1. The team has assessed forward-compatibility risks and documented them.
  2. The comment explicitly names the linter and the reason (e.g., `//nolint:staticcheck // TODO: add interface once proto fields ship`).
- Prefer suppressing at the narrowest scope (variable or statement) rather than disabling entire files.

## Enforcement

- CI runs `make check-full` (see [`Makefile`](Makefile:1)) which includes `lint` so any suppression must not hide new regressions.
- Reviewers should request justification or alternative solutions before merging suppressed lints.
