# Phase 2 Foundation: Scaffold & Ownership Boundaries (v0.1)

## Purpose

This artifact records the concrete scaffold for Phase 2 Task 1 (P2-T1). It
clarifies which packages own which responsibilities, how imports should stay
restricted, and what seams exist for future implementation. This is foundation
level work only — there is intentionally no production logic yet.

## Ownership Boundaries
- `cmd/aether` owns the single binary entry point (wiring, CLI flags, top-level
  orchestration). It may import only public package APIs (no internal/private
 helpers) and should defer implementation details to the `pkg/` tree.
- `pkg/protocol`, `pkg/network`, `pkg/crypto`, `pkg/storage`, and `pkg/ui` each
  represent a distinct ownership area. For P2-T1 they are doc-only placeholders;
  concrete code will be added in later deliverables.
- `pkg/app` (and its seams definitions) acts as the bridge between `cmd/aether` and
  the owned layers. The app package is the only component that can reference the
  seam interfaces and should not reach into package implementations directly.

## Import Conventions
- `cmd/aether` imports only the top-level seam definitions from `pkg/app` plus
  standard library packages. Any additional dependency must be justified by a P2
  requirement.
- Owner packages (`pkg/protocol`, etc.) currently contain only documentation and
  must not introduce implementation imports until those packages are ready for
  Phase 2 engineering work.
- Circular dependencies are discouraged; seams exist precisely to keep `cmd/
  ` and feature packages decoupled.

## Interface Seams
- See `pkg/app/seams.go` for the minimal interfaces each package provides during
  Phase 2 scaffolding. The intent is to keep app-level wiring and unit tests
  focused on behaviour-driven contracts and to avoid premature coupling with
  actual implementations.
- Each seam interface should be owned by its respective package (e.g.,
  `NetworkSeam` by `pkg/network`) even if the current implementation is still
  deferred.

## Phase Statement

This document is Phase 2 foundation scaffolding only. The packages exist to
capture ownership and interface expectations; the actual networking, cryptography,
storage, and UI logic will arrive in future tasks once the requirements are fully
specified.
