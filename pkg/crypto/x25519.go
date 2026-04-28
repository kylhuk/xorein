package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// X25519KeySize is the size of an X25519 public or private key in bytes.
const X25519KeySize = 32

// ErrLowOrderPoint is returned when the DH result is the all-zero point, which
// indicates a low-order input key and must be rejected per RFC 7748 §6.1.
var ErrLowOrderPoint = errors.New("x25519: low-order input point; key exchange rejected")

// GenerateX25519Keypair generates a fresh X25519 private key and its corresponding
// public key. The private key is clamped per RFC 7748 §5.
func GenerateX25519Keypair() (private, public [X25519KeySize]byte, err error) {
	privBytes, err := RandomBytes(X25519KeySize)
	if err != nil {
		return [X25519KeySize]byte{}, [X25519KeySize]byte{}, fmt.Errorf("generate x25519: %w", err)
	}
	copy(private[:], privBytes)
	// curve25519.X25519 applies RFC 7748 clamping internally.
	pubBytes, err := curve25519.X25519(private[:], curve25519.Basepoint)
	if err != nil {
		return [X25519KeySize]byte{}, [X25519KeySize]byte{}, fmt.Errorf("derive public key: %w", err)
	}
	copy(public[:], pubBytes)
	return private, public, nil
}

// X25519DH performs Diffie-Hellman with the local private key and remote public key.
// Returns ErrLowOrderPoint if the resulting shared secret is all zeros.
func X25519DH(private, public [X25519KeySize]byte) ([X25519KeySize]byte, error) {
	shared, err := curve25519.X25519(private[:], public[:])
	if err != nil {
		return [X25519KeySize]byte{}, fmt.Errorf("x25519 dh: %w", err)
	}
	// Reject the all-zero output (low-order point attack).
	allZero := true
	for _, b := range shared {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return [X25519KeySize]byte{}, ErrLowOrderPoint
	}
	var out [X25519KeySize]byte
	copy(out[:], shared)
	return out, nil
}
