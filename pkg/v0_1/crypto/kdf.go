package crypto

import (
	legacy "github.com/aether/code_aether/pkg/crypto"
)

// Extract derives a PRK from a secret and optional salt using HKDF-SHA-256 (RFC 5869).
func Extract(secret, salt []byte) []byte {
	return legacy.Extract(secret, salt)
}

// Expand derives keying material from a PRK and label. Length must be positive.
func Expand(prk []byte, label string, length int) ([]byte, error) {
	return legacy.Expand(prk, label, length)
}

// DeriveKey runs HKDF extract+expand in one call.
func DeriveKey(secret, salt []byte, label string, length int) ([]byte, error) {
	return legacy.DeriveKey(secret, salt, label, length)
}

// DeriveKey32 derives exactly 32 bytes.
func DeriveKey32(secret, salt []byte, label string) ([KeySize32]byte, error) {
	return legacy.DeriveKey32(secret, salt, label)
}
