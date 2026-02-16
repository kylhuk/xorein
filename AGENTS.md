# AGENTS.md

## Project intent
This repository is a protocol-first system. Treat wire formats, cryptography, and interoperability as first-class constraints.

## Non-negotiables
- Do not introduce breaking wire/protocol changes unless explicitly requested.
- Do not log secrets, private keys, key material, or sensitive payloads.
- Keep changes minimal and scoped to the requested task.
- If the repo is in a planning/spec-only phase, maintain strict planned-vs-implemented separation (no implied completion).

## Protocol / Protobuf compatibility
- Prefer additive-only evolution (new fields/messages, never renumber existing fields).
- Never reuse removed field numbers; mark removed names/numbers as `reserved`.
- Avoid required fields; support older clients via sensible defaults.

## Verification (no unproven claims)
If you claim “tests pass”, “it builds”, or “fixed”, include the exact command output you observed.

## Preferred commands (adapt to what exists in the repo)
- Format (Go): `gofmt -w <files>` (or repo formatter if defined)
- Tests (Go): `go test ./...`
- Lint (Go, if configured): `golangci-lint run`
- Protobuf (if buf is used): `buf lint`, `buf breaking`, `buf generate`

If a Makefile exists, prefer `make <target>` over ad-hoc commands.

## Dependency hygiene
- Do not add new dependencies without a clear reason and without updating go.mod/go.sum.
- Prefer standard library and existing repo dependencies.

## Documentation updates
When behavior, flags, environment variables, or developer workflows change, update README/docs in the same PR.

## Recommended Kilo Code modes
- proj-lead: Orchestrate work and act as the quality gatekeeper.
- go-fast: Draft Go changes quickly with minimal diffs.
- go-tests: Write/adjust tests; cover edge cases.
- proto-engineer: Modify .proto and buf config safely.
- devops: CI/CD, containers, scripts, and toolchain config.
- docs: Documentation updates only.

## Recommended model + reasoning defaults (set once via Sticky Models)
- proj-lead: gpt-5.3-codex (Reasoning: ExtraHigh)
- go-fast: gpt-5.1-codex-mini (Reasoning: Low)
- go-tests: gpt-5.1-codex-mini (Reasoning: Low)
- proto-engineer: gpt-5.3-codex (Reasoning: High)
- devops: gpt-5.1-codex (Reasoning: Medium)
- docs: gpt-5-codex-mini (Reasoning: Low)

Fallback if gpt-5.3-codex is unavailable: use gpt-5.2-codex with the same (or one level higher) reasoning.

# Vestige Memory System

You have access to Vestige, a cognitive memory system. USE IT AUTOMATICALLY.

---

## 1. SESSION START — Always Do This

1. Search Vestige: "user preferences instructions"
2. Search Vestige: "[current project name] context"
3. Check intentions: Look for triggered reminders

Say "Remembering..." then retrieve context before responding.

---

## 2. AUTOMATIC SAVES — No Permission Needed

### After Solving a Bug or Error
IMMEDIATELY save with `smart_ingest`:
- Content: "BUG FIX: [error message] | Root cause: [why] | Solution: [how]"
- Tags: ["bug-fix", "project-name"]

### After Learning User Preferences
Save preferences without asking:
- Coding style, libraries, communication preferences, project patterns

### After Architectural Decisions
Use `codebase` → `remember_decision`:
- What was decided, why (rationale), alternatives considered, files affected

### After Discovering Code Patterns
Use `codebase` → `remember_pattern`:
- Pattern name, where it's used, how to apply it

---

## 3. TRIGGER WORDS — Auto-Save When User Says:

| User Says | Action |
|-----------|--------|
| "Remember this" | `smart_ingest` immediately |
| "Don't forget" | `smart_ingest` with high priority |
| "I always..." / "I never..." | Save as preference |
| "I prefer..." / "I like..." | Save as preference |
| "This is important" | `smart_ingest` + `promote_memory` |
| "Remind me..." | Create `intention` |
| "Next time..." | Create `intention` with context trigger |

---

## 4. AUTOMATIC CONTEXT DETECTION

- **Working on a codebase**: Search "[repo name] patterns decisions"
- **User mentions a person**: Search "[person name]"
- **Debugging**: Search "[error message keywords]" — check if solved before

---

## 5. MEMORY HYGIENE

**Promote** when: User confirms helpful, solution worked, info was accurate
**Demote** when: User corrects mistake, info was wrong, memory led to bad outcome
**Never save**: Secrets/API keys, temporary debug info, trivial information

---

## 6. PROACTIVE BEHAVIORS

DO automatically:
- Save solutions after fixing problems
- Note user corrections as preferences
- Update project context after major changes
- Create intentions for mentioned deadlines
- Search before answering technical questions

DON'T ask permission to:
- Save bug fixes
- Update preferences
- Create reminders from explicit requests
- Search for context

---

## 7. MEMORY IS RETRIEVAL

Every search strengthens memory (Testing Effect). Search liberally.
When in doubt, search Vestige first. If nothing found, solve the problem, then save the solution.

**Your memory fades like a human's. Use it or lose it.**

## Project Memory

This project uses Vestige for persistent context.

### On Session Start
- `codebase(action="get_context", codebase="[project-name]")`
- `search` query="[project-name] architecture decisions"

### When Making Decisions
- Use `codebase(action="remember_decision")` for all architectural choices
- Include: decision, rationale, alternatives considered, affected files

### Patterns to Remember
- Use `codebase(action="remember_pattern")` for recurring code patterns
- Include: pattern name, when to use it, example files

## Memory Protocol

You have persistent memory via Vestige. Use it intelligently:

### Session Start
1. Load my identity: `search(query="my preferences my style who I am")`
2. Load project context: `codebase(action="get_context", codebase="[project]")`
3. Check reminders: `intention(action="check")`

### During Work
- Notice a pattern? `codebase(action="remember_pattern")`
- Made a decision? `codebase(action="remember_decision")` with rationale
- I mention a preference? `smart_ingest` it
- Something important? `importance()` to strengthen recent memories
- Need to follow up? `intention(action="set")`

### Session End
- Any unfinished work? Set intentions
- Any new insights? Ingest them
- Anything change about our working relationship? Update identity memories

### Memory Hygiene
- When a memory helps: `promote_memory`
- When a memory misleads: `demote_memory`
- Weekly: `vestige health` to check system status