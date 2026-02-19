# Phase 0 Binary Boundary (v2.1)

- **Xorein** is the runtime/daemon binary that hosts the protocol stack, storage engine, crypto, DHT/pubsub, and persistence components. It must never import or depend on Gio or harmolyn-only packages.
- **harmolyn** is the Gio-based frontend that depends on Xorein through explicit interfaces or narrow internal APIs, enabling a headless build of Xorein.
- Build systems enforce the separation by keeping `cmd/xorein` and `cmd/harmolyn` in distinct modules or by running dependency checks as part of `G0`. Any violation must be documented in this artifact before Phase 1 work proceeds.
