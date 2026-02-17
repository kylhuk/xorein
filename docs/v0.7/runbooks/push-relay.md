# Push Relay Runbook

Purpose-built to exercise the relay-authenticated push pipeline that now enforces payload size limits, honors idempotent delivery, and keeps relay blindness intact.

## Prerequisites

1. Go toolchain (`go1.22+`).
2. Configured GOPATH/GOMODCACHE so `go run` can resolve the module.

## Run the relay service

1. Start the push relay with the `default` profile (auth metadata and payload size safeguards are enabled automatically):
   ```sh
   go run ./cmd/push-relay --profile default
   ```
2. The CLI will log `push relay running profile=default port=5005` once registration/auth completes. The relay keeps a deterministic registration map so repeated restarts remain idempotent.

## Verify health

1. Trigger forward path manually (optionally from `cmd/aether` or a standalone Go snippet) by invoking `push.Service.Forward`. Look for `push.adapter.missing` or `push.payload.unauthorized` in logs to debug auth metadata issues.
2. When the adapter receives the ciphertext envelope, confirm the mock relay state in `cmd/push-relay` logs (the mock adapter prints nothing, which maintains ciphertext-only behavior).

## Recovery notes

- Payloads larger than 256 KiB are rejected with `push.payload.too_large`; trim payloads and retry.
- Duplicate envelope IDs are dropped early to keep the pipeline idempotent (`push.forward` returns success without forwarding a second time).
- Unknown recipients or missing auth metadata trigger `push.recipient.unknown` or `push.payload.unauthorized`; ensure registration metadata includes `push.auth_token`.
