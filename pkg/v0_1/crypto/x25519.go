package crypto

import (
	legacy "github.com/aether/code_aether/pkg/crypto"
)

// X25519KeySize is the size of an X25519 key in bytes.
const X25519KeySize = legacy.X25519KeySize

// ErrLowOrderPoint is returned when the DH output is the all-zero point (RFC 7748 §6.1).
var ErrLowOrderPoint = legacy.ErrLowOrderPoint

// GenerateX25519Keypair generates a fresh clamped X25519 keypair.
func GenerateX25519Keypair() (private, public [X25519KeySize]byte, err error) {
	return legacy.GenerateX25519Keypair()
}

// X25519DH computes the Diffie-Hellman shared secret. Returns ErrLowOrderPoint
// if the result is the all-zero point.
func X25519DH(private, public [X25519KeySize]byte) ([X25519KeySize]byte, error) {
	return legacy.X25519DH(private, public)
}
