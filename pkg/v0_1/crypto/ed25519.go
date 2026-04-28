package crypto

import (
	"crypto/ed25519"
	"crypto/sha512"
	"errors"
	"fmt"

	"filippo.io/edwards25519"
)

const (
	// Ed25519PublicKeySize is the size of an Ed25519 public key.
	Ed25519PublicKeySize = ed25519.PublicKeySize // 32
	// Ed25519PrivateKeySize is the size of an Ed25519 private key (seed + public).
	Ed25519PrivateKeySize = ed25519.PrivateKeySize // 64
	// Ed25519SignatureSize is the size of an Ed25519 signature.
	Ed25519SignatureSize = ed25519.SignatureSize // 64
)

// ErrInvalidEd25519Key is returned for malformed key material.
var ErrInvalidEd25519Key = errors.New("ed25519: invalid key material")

// GenerateEd25519Keypair generates a fresh Ed25519 keypair using crypto/rand.
// Returns (publicKey, privateKey, error).
func GenerateEd25519Keypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("ed25519 generate: %w", err)
	}
	return pub, priv, nil
}

// SignEd25519 signs msg with the given Ed25519 private key. Returns 64 bytes.
func SignEd25519(priv ed25519.PrivateKey, msg []byte) []byte {
	return ed25519.Sign(priv, msg)
}

// VerifyEd25519 verifies an Ed25519 signature. Returns true if valid.
func VerifyEd25519(pub ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(pub, msg, sig)
}

// Ed25519PublicToX25519 converts an Ed25519 public key to an X25519 public key
// using the Bernstein-Hamburg birational map (spec 01 §1.1).
// u = (1 + y) / (1 - y) mod p, where y is the Ed25519 y-coordinate.
func Ed25519PublicToX25519(edPub ed25519.PublicKey) ([X25519KeySize]byte, error) {
	if len(edPub) != Ed25519PublicKeySize {
		return [X25519KeySize]byte{}, ErrInvalidEd25519Key
	}
	p, err := new(edwards25519.Point).SetBytes(edPub)
	if err != nil {
		return [X25519KeySize]byte{}, fmt.Errorf("ed25519 → x25519: %w", err)
	}
	// edwards25519 Point → X25519 (Montgomery u-coordinate).
	// filippo.io/edwards25519 exposes BytesMontgomery() which performs the map.
	mont := p.BytesMontgomery()
	var out [X25519KeySize]byte
	copy(out[:], mont)
	return out, nil
}

// Ed25519PrivateToX25519 converts an Ed25519 private key seed to an X25519 scalar
// via clamp(SHA-512(seed)[0:32]) per spec 01 §1.1 and RFC 7748 §4.1.
func Ed25519PrivateToX25519(edPriv ed25519.PrivateKey) ([X25519KeySize]byte, error) {
	if len(edPriv) != Ed25519PrivateKeySize {
		return [X25519KeySize]byte{}, ErrInvalidEd25519Key
	}
	// The Go ed25519.PrivateKey is 64 bytes: seed (32) || public key (32).
	seed := edPriv[:32]
	h := sha512.Sum512(seed)
	// Clamp per RFC 7748 §5.
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	var out [X25519KeySize]byte
	copy(out[:], h[:32])
	return out, nil
}
