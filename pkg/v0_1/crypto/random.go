package crypto

import (
	legacy "github.com/aether/code_aether/pkg/crypto"
)

// RandomBytes returns n cryptographically random bytes.
func RandomBytes(n int) ([]byte, error) { return legacy.RandomBytes(n) }

// RandomKey32 returns a 32-byte random key.
func RandomKey32() ([KeySize32]byte, error) { return legacy.RandomKey32() }

// RandomKey16 returns a 16-byte random key.
func RandomKey16() ([KeySize16]byte, error) { return legacy.RandomKey16() }

// RandomNonce12 returns a 12-byte random AEAD nonce.
func RandomNonce12() ([NonceSize12]byte, error) { return legacy.RandomNonce12() }
