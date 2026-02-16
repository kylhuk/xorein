# Phase 2 P2-T4: CI Trigger and Workflow Matrix

## Purpose
Document the current CI trigger surface and its mapped workflow targets, with explicit separation between implemented automation and planned extensions.

## Implemented (current state)
The repository defines four GitHub Actions workflows under [`.github/workflows/`](.github/workflows/ci.yml:1):

| Trigger | Workflow | Job(s) | Make target(s) | Notes |
|---|---|---|---|---|
| `push` to `main` | [`ci.yml`](.github/workflows/ci.yml:1) | `check` matrix (`fast`, `full`) | [`make check-fast`](Makefile:14) / [`make check-full`](Makefile:15) | Uses Podman install step and runs the matrix profiles. |
| `pull_request` to `main` | [`ci.yml`](.github/workflows/ci.yml:1) | `check` matrix (`fast`, `full`) | [`make check-fast`](Makefile:14) / [`make check-full`](Makefile:15) | Same as push; CI parity for PRs. |
| Scheduled nightly (`0 3 * * *`) | [`nightly.yml`](.github/workflows/nightly.yml:1) | `nightly-check` | [`make check-full`](Makefile:15) | Full checks on a nightly cadence. |
| Scheduled security audit (`0 1 * * 1`) | [`security-audit.yml`](.github/workflows/security-audit.yml:1) | `audit` | [`make scan`](Makefile:39) | Placeholder compliance scan surface. |
| Manual release dispatch | [`release.yml`](.github/workflows/release.yml:1) | `release` | [`make build`](Makefile:44) | Release preparation placeholder. |

### Supporting build/test scaffolds
- The workflow targets align with the placeholder build/test stages defined in [`Makefile`](Makefile:1), including the Podman-backed targets for generate/compile/lint/test/scan.
- Reproducibility scaffolds referenced by `make test` are documented in [`docs/v0.1/phase2/repro-tools.md`](docs/v0.1/phase2/repro-tools.md:1) and [`docs/v0.1/phase2/repro-policy.md`](docs/v0.1/phase2/repro-policy.md:1).

## Planned (not implemented yet)
- Expand workflow matrices to include platform portability checks once portability acceptance criteria are defined.
- Add release evidence capture steps and link them to the Phase 2 evidence log (pending artifact).
- Introduce explicit workflow annotations for protocol/buf checks once protocol automation is activated.

## Planned vs. implemented separation
Only the workflows and Make targets listed in **Implemented (current state)** exist today. The items in **Planned (not implemented yet)** are forward-looking and must not be treated as active CI behavior.
