> **SUPERSEDED BY `docs/spec/v0.1/60-local-control-api.md`** — This document
> is a historical sketch of the local control API. The normative specification
> is `docs/spec/v0.1/60-local-control-api.md`. Do not use this document as an
> implementation reference.

# Xorein local control API v1

The Harmolyn frontend connects to a **local-only** Xorein runtime over a versioned control API.

## Transport

- Unix-like systems: Unix domain socket at `xorein-control.sock` in the configured data directory unless `--control` overrides it.
- Windows: loopback TCP endpoint only.
- Remote access is rejected. Authentication requires the bearer token stored in `control.token` in the configured data directory.

## Versioning

All stable endpoints are rooted at `/v1`.

## Endpoints

- `GET /v1/state` — full local snapshot, including identity, peers, servers, DMs, messages, voice session state, settings, and telemetry.
- `GET /v1/events` — Server-Sent Events stream for runtime, server, channel, message, and voice updates.
- `GET|POST /v1/identities` — read or create the local identity.
- `GET /v1/identities/backup` — export the local identity as JSON.
- `POST /v1/identities/restore` — restore a previously exported identity.
- `GET|POST /v1/servers` — list servers or create a server.
- `POST /v1/servers/join` — join a server from a signed deeplink invite.
- `POST /v1/servers/{serverID}/channels` — create a text or voice channel.
- `POST|DELETE /v1/peers/manual` — add or remove a manual peer address.
- `GET|POST /v1/dms` — list DMs or create one.
- `POST /v1/dms/{dmID}/messages` — send a DM message.
- `POST /v1/channels/{channelID}/messages` — send a channel message.
- `PATCH|DELETE /v1/messages/{messageID}` — edit or delete a message.
- `POST /v1/voice/{channelID}/join` — join voice.
- `POST /v1/voice/{channelID}/leave` — leave voice.
- `POST /v1/voice/{channelID}/mute` — update mute state.
- `POST /v1/voice/{channelID}/frames` — send opaque audio/media frames.

## Error model

Responses use JSON objects with:

- `code` — stable machine-readable error code
- `message` — human-readable explanation

Representative codes include `unauthorized`, `forbidden`, `invalid_request`, `invalid_signature`, `expired_invite`, `not_found`, and `method_not_allowed`.
