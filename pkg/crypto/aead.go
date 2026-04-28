package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// AEAD algorithm identifiers used in ciphertext_format fields.
const (
	AlgChaCha20Poly1305 = "chacha20-poly1305"
	AlgAES128GCM        = "aes-128-gcm"
)

const (
	KeySize32   = 32
	KeySize16   = 16
	NonceSize12 = 12
)

// ErrDecryptFailed is returned when AEAD authentication fails.
var ErrDecryptFailed = errors.New("decryption failed: authentication tag mismatch")

// SealChaCha20Poly1305 encrypts plaintext with the given 32-byte key, 12-byte nonce,
// and optional additional data. Returns authenticated ciphertext (appended MAC).
func SealChaCha20Poly1305(key [KeySize32]byte, nonce [NonceSize12]byte, plaintext, additionalData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, fmt.Errorf("chacha20poly1305: %w", err)
	}
	return aead.Seal(nil, nonce[:], plaintext, additionalData), nil
}

// OpenChaCha20Poly1305 decrypts and authenticates ciphertext. Returns ErrDecryptFailed
// if the authentication tag does not match.
func OpenChaCha20Poly1305(key [KeySize32]byte, nonce [NonceSize12]byte, ciphertext, additionalData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key[:])
	if err != nil {
		return nil, fmt.Errorf("chacha20poly1305: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce[:], ciphertext, additionalData)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return plaintext, nil
}

// SealAES128GCM encrypts plaintext with the given 16-byte key, 12-byte nonce,
// and optional additional data. Returns authenticated ciphertext.
func SealAES128GCM(key [KeySize16]byte, nonce [NonceSize12]byte, plaintext, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return aead.Seal(nil, nonce[:], plaintext, additionalData), nil
}

// OpenAES128GCM decrypts and authenticates ciphertext. Returns ErrDecryptFailed
// if the authentication tag does not match.
func OpenAES128GCM(key [KeySize16]byte, nonce [NonceSize12]byte, ciphertext, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce[:], ciphertext, additionalData)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return plaintext, nil
}
