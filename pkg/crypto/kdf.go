package crypto

import (
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Extract derives a pseudorandom key (PRK) from a secret and optional salt using
// HKDF-SHA-256. If salt is nil, HKDF uses the zero-filled hash-length salt per RFC 5869.
func Extract(secret, salt []byte) []byte {
	return hkdf.Extract(sha256.New, secret, salt)
}

// Expand derives keying material of exactly length bytes from a PRK and a label.
// The label must match one of the constants in labels.go; callers MUST NOT pass
// unvalidated external strings as labels.
func Expand(prk []byte, label string, length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("expand: length must be positive, got %d", length)
	}
	r := hkdf.Expand(sha256.New, prk, []byte(label))
	out := make([]byte, length)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, fmt.Errorf("expand: %w", err)
	}
	return out, nil
}

// DeriveKey is a convenience wrapper that performs Extract then Expand in one call.
// Equivalent to HKDF(Hash=SHA-256, IKM=secret, salt=salt, info=label, L=length).
func DeriveKey(secret, salt []byte, label string, length int) ([]byte, error) {
	prk := Extract(secret, salt)
	return Expand(prk, label, length)
}

// DeriveKey32 derives exactly 32 bytes using DeriveKey; the most common case.
func DeriveKey32(secret, salt []byte, label string) ([KeySize32]byte, error) {
	raw, err := DeriveKey(secret, salt, label, KeySize32)
	if err != nil {
		return [KeySize32]byte{}, err
	}
	var out [KeySize32]byte
	copy(out[:], raw)
	return out, nil
}
