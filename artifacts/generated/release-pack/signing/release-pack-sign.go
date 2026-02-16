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
