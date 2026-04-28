package crypto

import (
	"crypto/rand"
	"fmt"
)

// RandomBytes fills a freshly allocated slice of length n with cryptographically
// random bytes. Panics if the OS entropy source fails — at that point the system
// is in an unrecoverable state and continuing would be insecure.
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("random: length must be positive, got %d", n)
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("random: entropy source failed: %w", err)
	}
	return b, nil
}

// RandomKey32 generates a 32-byte cryptographic key.
func RandomKey32() ([KeySize32]byte, error) {
	raw, err := RandomBytes(KeySize32)
	if err != nil {
		return [KeySize32]byte{}, err
	}
	var k [KeySize32]byte
	copy(k[:], raw)
	return k, nil
}

// RandomKey16 generates a 16-byte cryptographic key.
func RandomKey16() ([KeySize16]byte, error) {
	raw, err := RandomBytes(KeySize16)
	if err != nil {
		return [KeySize16]byte{}, err
	}
	var k [KeySize16]byte
	copy(k[:], raw)
	return k, nil
}

// RandomNonce12 generates a 12-byte AEAD nonce. Callers are responsible for
// ensuring nonce uniqueness per key; random nonces are safe with keys that are
// themselves ephemeral (key-per-message) or when message counts are well below
// 2^32 per key.
func RandomNonce12() ([NonceSize12]byte, error) {
	raw, err := RandomBytes(NonceSize12)
	if err != nil {
		return [NonceSize12]byte{}, err
	}
	var n [NonceSize12]byte
	copy(n[:], raw)
	return n, nil
}
