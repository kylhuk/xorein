# Core rules (applies in all modes)

## Speed with proof
- Never claim “done” without showing command output (build/test/lint) or clearly stating what could not be verified.
- Prefer the smallest correct change; avoid drive-by refactors.
- If the repository is currently documentation/planning-only, do not imply implementation completion.

## Scope discipline
- Break work into small, verifiable steps. Prefer “one package / one concern” changes.
- Keep context lean: only open/mention the files needed to do the current step.

## Safety & security
- Never print or persist secrets or key material in logs, tests, fixtures, or examples.
- Security-sensitive code (crypto, auth, network parsing) requires extra caution and explicit negative-path tests.

## Protocol compatibility
- Treat wire formats and protobuf schemas as compatibility contracts.
- Additive changes only unless explicitly authorized.
