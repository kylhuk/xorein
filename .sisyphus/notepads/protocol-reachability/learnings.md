# Append-only learnings for protocol-reachability


## 2026-04-21 static protocol/client audit
- Live runtime startup path remains `cmd/aether/main.go -> runRuntime -> network.NewP2PRuntime(network.Config{Mode: network.Mode(cfg.Role), ListenAddr: cfg.ListenAddr}) -> node.NewService(..., node.WithPeerRuntime(runtime)) -> runtime.SetHandler(service) -> service.Start(ctx)`. The local control endpoint is separate from this peer transport path.
- Repo Go client call path is `pkg/node/service.go:peerClient -> network.NewClient(1200*time.Millisecond) -> Client.Call -> NewRequest -> Client.Do -> NormalizePeerAddress -> peerInfoFromAddress -> h.NewStream(ctx, peerInfo.ID, peerTransportProtocols()...)`.
- The canonical repo peer multistream surface is derived from live code via `protocol.CanonicalProtocolStrings(protocol.FamilyPeer)` and currently resolves to exactly `/aether/peer/0.1.0` (`pkg/protocol/transport_test.go:TestPeerTransportCanonicalSurface`).
- The canonical repo peer capability surface is `protocol.DefaultPeerTransportFeatureFlags()`, which currently sorts to `cap.peer.bootstrap`, `cap.peer.delivery`, `cap.peer.join`, `cap.peer.manifest`, `cap.peer.metadata`, `cap.peer.relay`, and `cap.peer.transport`.
- Operation-specific requirements are additive over `cap.peer.transport`: `peer.info`/`peer.exchange` require metadata, bootstrap ops require bootstrap, manifest ops require manifest, `peer.join` requires join, `peer.deliver` requires delivery, and relay ops require relay (`pkg/network/transport.go:requiredCapabilities`).
- Executable positive evidence: `pkg/network/runtime_test.go:TestP2PRuntimeNegotiatesPeerTransport` proves the repo Go client negotiates `/aether/peer/0.1.0` successfully and exchanges a payload over the live runtime handler.
- Executable negative evidence: `pkg/protocol/transport_test.go` fail-closes unsupported protocol offers (`/aether/peer/1.0.0`) and unsupported required capabilities (`cap.peer.experimental`); `pkg/network/runtime_test.go` also proves the live runtime returns `unsupported-capability` with `MissingRequired=[cap.peer.experimental]`, and rejects opening a stream for `/aether/peer/1.0.0`.
- Scope note: this audit is intentionally limited to the repo's Go client/runtime path. README compatibility text is broader protocol documentation, but the current repo-reachable peer client surface is only the Go libp2p transport described above.
- Verification: `go test ./pkg/protocol ./pkg/network ./pkg/node` passed on 2026-04-21.

- Added `cmd/aether/main_smoke_test.go` to build the real `cmd/aether` binary, launch it as a child process, wait for the `xorein runtime ready ... listen=...` stdout line, and capture the live libp2p listen multiaddr before dialing.
- The smoke harness proves a real cross-process request/response by calling `network.NewClient(...).Call(..., network.OperationPeerInfo, ...)` against the spawned runtime and asserting the negotiated protocol is `/aether/peer/0.1.0` with a populated `node.PeerInfo` response.
- The same harness proves fail-closed negotiation by appending `cap.peer.experimental` to the request's required capabilities and asserting the live process returns `unsupported-capability` with the missing capability surfaced in `network.Error`.

## 2026-04-21 task 3 verdict packaging
- The packaged verdict stays intentionally narrow: repo Go client reachability is verified only for `/aether/peer/0.1.0`, with the live smoke harness covering both the positive peer-info exchange and the negative fail-closed capability check.
- The final evidence set now points at the dedicated task-3 logs (`.sisyphus/evidence/task-3-static-audit.log`, `.sisyphus/evidence/task-3-smoke-harness.log`, and `.sisyphus/evidence/task-3-verdict.md`) so later reviews do not need to infer scope from broader README language.
