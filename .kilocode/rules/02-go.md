# Go code rules

- Run `gofmt` on any modified Go file.
- Prefer small, composable functions; avoid unnecessary cleverness.
- Error handling:
  - Wrap errors with context.
  - Avoid swallowing errors.
- Concurrency:
  - Avoid data races; prefer structured concurrency (context cancellation, explicit goroutine lifetimes).
- Tests:
  - New behavior should come with tests (unit first, integration if needed).
