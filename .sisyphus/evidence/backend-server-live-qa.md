## Backend server live QA

- Runtime start: `bin/aether --mode client --data-dir /tmp/xorein-qa.5yPAPd/data --control /tmp/xorein-qa.5yPAPd/control.sock --listen 127.0.0.1:0`
- Startup log: `xorein runtime ready role=client peer_id=212a8d54d34b9fb94ea95edb8c8fc002 listen=127.0.0.1:36633 control=/tmp/xorein-qa.5yPAPd/control.sock`
- `/v1/state` with valid token returned `200` and an empty server list before mutation.
- Wrong bearer token on `/v1/state` returned `401 unauthorized`.
- `/v1/events` emitted `event: ready` / `data: {"version":"v1"}` and timed out with curl exit `28` as expected.
- `POST /v1/servers` created `qa-server` with ID `server-78037b6214f22f6e`.
- Restart from the same data dir preserved the server and `peer_id`.
- Invalid mode check returned exit code `2` with the expected deterministic error.
