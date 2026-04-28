// Package crypto implements the locked v0.1 cryptographic profile.
// See docs/spec/v0.1/01-cryptographic-primitives.md for the normative definition.
package crypto

import (
	legacy "github.com/aether/code_aether/pkg/crypto"
)

// Re-exported constants from the legacy package.
const (
	AlgChaCha20Poly1305 = legacy.AlgChaCha20Poly1305
	AlgAES128GCM        = legacy.AlgAES128GCM
	KeySize32           = legacy.KeySize32
	KeySize16           = legacy.KeySize16
	NonceSize12         = legacy.NonceSize12
)

// ErrDecryptFailed is returned when AEAD authentication fails.
var ErrDecryptFailed = legacy.ErrDecryptFailed

// SealChaCha20Poly1305 encrypts plaintext with a 32-byte key and 12-byte nonce.
// Used for all message-level encryption in Seal, Crowd, and Channel modes (spec §2).
func SealChaCha20Poly1305(key [KeySize32]byte, nonce [NonceSize12]byte, plaintext, ad []byte) ([]byte, error) {
	return legacy.SealChaCha20Poly1305(key, nonce, plaintext, ad)
}

// OpenChaCha20Poly1305 decrypts and authenticates. Returns ErrDecryptFailed on failure.
func OpenChaCha20Poly1305(key [KeySize32]byte, nonce [NonceSize12]byte, ciphertext, ad []byte) ([]byte, error) {
	return legacy.OpenChaCha20Poly1305(key, nonce, ciphertext, ad)
}

// SealAES128GCM encrypts with AES-128-GCM. Used for MediaShield (SFrame, spec §2).
func SealAES128GCM(key [KeySize16]byte, nonce [NonceSize12]byte, plaintext, ad []byte) ([]byte, error) {
	return legacy.SealAES128GCM(key, nonce, plaintext, ad)
}

// OpenAES128GCM decrypts and authenticates AES-128-GCM. Returns ErrDecryptFailed on failure.
func OpenAES128GCM(key [KeySize16]byte, nonce [NonceSize12]byte, ciphertext, ad []byte) ([]byte, error) {
	return legacy.OpenAES128GCM(key, nonce, ciphertext, ad)
}
