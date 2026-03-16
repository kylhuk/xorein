# Xorein release naming

The first release candidate uses the following canonical runtime-facing names:

- runtime name: `xorein`
- local control socket name: `xorein-control.sock`
- control token file: `control.token`
- protocol namespace: `/aether/...`

The protocol namespace remains `/aether` for wire compatibility, while the user-facing runtime and local control surface use `xorein` naming.
