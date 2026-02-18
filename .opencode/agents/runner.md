---
description: Run commands and capture outputs (no file edits).
mode: subagent
model: openai/gpt-5.1-codex-mini
options:
  reasoningEffort: low
permission:
  task: deny
  edit: deny
  bash:
    "*": allow
    "git status*": allow
    "git diff*": allow
    "git log*": allow
    "git show*": allow
    "go version*": allow
    "go env*": allow
    "go list*": allow
    "go test*": allow
    "go test": allow
    "go vet*": allow
    "golangci-lint*": allow
    "buf*": allow
    "gofmt*": allow
    "goimports*": allow
    "git commit*": deny
    "git push*": deny
    "rm*": deny
    "sudo*": deny
---

You execute commands to gather evidence.

Rules
- Never edit files.
- Return: command(s) run + trimmed outputs + interpretation (pass/fail).
