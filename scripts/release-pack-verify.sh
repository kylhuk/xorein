#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
export PATH="$ROOT_DIR/scripts:$ROOT_DIR/.cache/xorein-tools/bin:$PATH"
cd "$ROOT_DIR"

GENERATED_DIR="artifacts/generated"
RELEASE_PACK_DIR="$GENERATED_DIR/release-pack"
SIGN_DIR="$RELEASE_PACK_DIR/signing"
BUILD_BIN="bin/aether"

DEFAULT_RELEASE_SIGNING_SEED_HEX="000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
RELEASE_SIGNING_SEED_HEX="${RELEASE_SIGNING_SEED_HEX:-$DEFAULT_RELEASE_SIGNING_SEED_HEX}"
if [[ ! "$RELEASE_SIGNING_SEED_HEX" =~ ^[0-9a-fA-F]{64}$ ]]; then
	echo "release-pack verify: RELEASE_SIGNING_SEED_HEX must be 64 hex characters" >&2
	exit 1
fi
if [[ "$RELEASE_SIGNING_SEED_HEX" != "$DEFAULT_RELEASE_SIGNING_SEED_HEX" && "${ALLOW_NONFIXTURE_RELEASE_SIGNING_SEED:-0}" != "1" ]]; then
	echo "release-pack verify: non-fixture signing seed requires ALLOW_NONFIXTURE_RELEASE_SIGNING_SEED=1" >&2
	exit 1
fi

mkdir -p "$RELEASE_PACK_DIR/sbom" "$SIGN_DIR"
STAMP_COMMIT="$(git rev-parse HEAD 2>/dev/null || true)"
printf 'release-pack-stamp\ncommit=%s\n' "${STAMP_COMMIT:-unknown}" > "$GENERATED_DIR/stamp.txt"
sha256sum "$GENERATED_DIR/stamp.txt" "$BUILD_BIN" > "$GENERATED_DIR/checksums.txt"

test -f "$GENERATED_DIR/stamp.txt"
test -x "$BUILD_BIN"

"$BUILD_BIN" --preflight --repo-root "$ROOT_DIR" > "$RELEASE_PACK_DIR/preflight.txt"

sha256sum "$GENERATED_DIR/stamp.txt" "$BUILD_BIN" > "$RELEASE_PACK_DIR/checksums.txt"
sha256sum "$GENERATED_DIR/checksums.txt" "$GENERATED_DIR/stamp.txt" > "$RELEASE_PACK_DIR/checksums-stage.txt"

BIN_SHA="$(sha256sum "$BUILD_BIN" | awk '{print $1}')"
STAMP_SHA="$(sha256sum "$GENERATED_DIR/stamp.txt" | awk '{print $1}')"
cat > "$RELEASE_PACK_DIR/sbom/sbom.spdx.json" <<EOF
{
  "spdxVersion": "SPDX-2.3",
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "aether-v0.1-release-pack",
  "documentNamespace": "https://aether.invalid/spdx/release-pack/v0.1",
  "creationInfo": {
    "created": "1970-01-01T00:00:00Z",
    "creators": ["Tool: scripts/release-pack-verify.sh"]
  },
  "files": [
    {
      "SPDXID": "SPDXRef-File-BinAether",
      "fileName": "bin/aether",
      "checksums": [
        {"algorithm": "SHA256", "checksumValue": "$BIN_SHA"}
      ]
    },
    {
      "SPDXID": "SPDXRef-File-Stamp",
      "fileName": "artifacts/generated/stamp.txt",
      "checksums": [
        {"algorithm": "SHA256", "checksumValue": "$STAMP_SHA"}
      ]
    }
  ]
}
EOF
sha256sum "$RELEASE_PACK_DIR/sbom/sbom.spdx.json" > "$RELEASE_PACK_DIR/sbom/sbom.spdx.json.sha256"

cat > "$SIGN_DIR/release-pack-sign.go" <<'EOF'
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "sign", "sign or verify")
	payloadPath := flag.String("payload", "", "payload path")
	seedHex := flag.String("seed-hex", "", "ed25519 seed hex for sign mode")
	pubHex := flag.String("pub-hex", "", "ed25519 public key hex for verify mode")
	sigB64 := flag.String("sig-b64", "", "signature base64")
	flag.Parse()

	payload, err := os.ReadFile(*payloadPath)
	if err != nil {
		panic(err)
	}

	switch *mode {
	case "sign":
		seed, err := hex.DecodeString(*seedHex)
		if err != nil {
			panic(err)
		}
		if len(seed) != ed25519.SeedSize {
			panic("invalid seed size")
		}
		priv := ed25519.NewKeyFromSeed(seed)
		pub := priv.Public().(ed25519.PublicKey)
		digest := sha256.Sum256(payload)
		sig := ed25519.Sign(priv, payload)
		fmt.Printf("payload_sha256=%s\n", hex.EncodeToString(digest[:]))
		fmt.Printf("public_key_hex=%s\n", hex.EncodeToString(pub))
		fmt.Printf("signature_base64=%s\n", base64.StdEncoding.EncodeToString(sig))
	case "verify":
		pub, err := hex.DecodeString(*pubHex)
		if err != nil {
			panic(err)
		}
		sig, err := base64.StdEncoding.DecodeString(*sigB64)
		if err != nil {
			panic(err)
		}
		digest := sha256.Sum256(payload)
		if !ed25519.Verify(ed25519.PublicKey(pub), payload, sig) {
			fmt.Printf("payload_sha256=%s\n", hex.EncodeToString(digest[:]))
			fmt.Println("verification=FAILED")
			os.Exit(1)
		}
		fmt.Printf("payload_sha256=%s\n", hex.EncodeToString(digest[:]))
		fmt.Println("verification=OK")
	default:
		panic("unsupported mode")
	}
}
EOF

SBOM_SHA="$(awk '{print $1}' "$RELEASE_PACK_DIR/sbom/sbom.spdx.json.sha256")"
CHECKSUMS_SHA="$(sha256sum "$RELEASE_PACK_DIR/checksums.txt" | awk '{print $1}')"
cat > "$SIGN_DIR/payload.txt" <<EOF
artifact=artifacts/generated/release-pack/checksums.txt
artifact=artifacts/generated/release-pack/sbom/sbom.spdx.json
artifact=artifacts/generated/release-pack/sbom/sbom.spdx.json.sha256
sbom_sha256=$SBOM_SHA
checksums_sha256=$CHECKSUMS_SHA
EOF

go run artifacts/generated/release-pack/signing/release-pack-sign.go --mode sign --payload artifacts/generated/release-pack/signing/payload.txt --seed-hex "$RELEASE_SIGNING_SEED_HEX" > "$SIGN_DIR/sign-output.txt"

pub_hex="$(grep '^public_key_hex=' "$SIGN_DIR/sign-output.txt" | cut -d= -f2- | tr -d '\r\n')"
sig_b64="$(grep '^signature_base64=' "$SIGN_DIR/sign-output.txt" | cut -d= -f2- | tr -d '\r\n')"
go run artifacts/generated/release-pack/signing/release-pack-sign.go --mode verify --payload artifacts/generated/release-pack/signing/payload.txt --pub-hex "$pub_hex" --sig-b64 "$sig_b64" > "$SIGN_DIR/verify-output.txt"

cat > "$RELEASE_PACK_DIR/signature-verification.txt" <<EOF
signature-workflow: ed25519-detached
signing_method: go-ed25519-host
payload: artifacts/generated/release-pack/signing/payload.txt
sign_output:
$(cat "$SIGN_DIR/sign-output.txt")
verify_output:
$(cat "$SIGN_DIR/verify-output.txt")
status: verified
EOF
