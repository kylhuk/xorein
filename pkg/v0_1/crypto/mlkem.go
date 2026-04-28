package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	mlkem768 "github.com/cloudflare/circl/kem/mlkem/mlkem768"
)

const (
	// MLKEM768PublicKeySize is the size of an ML-KEM-768 encapsulation public key.
	MLKEM768PublicKeySize = mlkem768.PublicKeySize // 1184
	// MLKEM768PrivateKeySize is the size of an ML-KEM-768 decapsulation private key.
	MLKEM768PrivateKeySize = mlkem768.PrivateKeySize // 2400
	// MLKEM768CiphertextSize is the size of an ML-KEM-768 encapsulation ciphertext.
	MLKEM768CiphertextSize = mlkem768.CiphertextSize // 1088
	// MLKEM768SharedKeySize is the size of the shared secret (32 bytes).
	MLKEM768SharedKeySize = mlkem768.SharedKeySize // 32
)

// ErrMLKEMDecapsulate is returned when ML-KEM-768 decapsulation fails.
var ErrMLKEMDecapsulate = errors.New("mlkem768: decapsulation failed")

// GenerateMLKEM768Keypair generates a fresh ML-KEM-768 keypair.
// Returns (encapsulationPK bytes, decapsulationSK bytes, error).
func GenerateMLKEM768Keypair() ([]byte, []byte, error) {
	pk, sk, err := mlkem768.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("mlkem768 keygen: %w", err)
	}
	pkBytes := make([]byte, MLKEM768PublicKeySize)
	skBytes := make([]byte, MLKEM768PrivateKeySize)
	pk.Pack(pkBytes)
	sk.Pack(skBytes)
	return pkBytes, skBytes, nil
}

// MLKEM768Encapsulate encapsulates a fresh shared secret to the given public key.
// Returns (ciphertext []byte, sharedSecret []byte, error).
// ciphertext is MLKEM768CiphertextSize (1088) bytes; sharedSecret is 32 bytes.
func MLKEM768Encapsulate(pkBytes []byte) (ct, ss []byte, err error) {
	if len(pkBytes) != MLKEM768PublicKeySize {
		return nil, nil, fmt.Errorf("mlkem768 encapsulate: public key must be %d bytes, got %d", MLKEM768PublicKeySize, len(pkBytes))
	}
	var pk mlkem768.PublicKey
	if err := pk.Unpack(pkBytes); err != nil {
		return nil, nil, fmt.Errorf("mlkem768 encapsulate: unpack pk: %w", err)
	}
	ct = make([]byte, MLKEM768CiphertextSize)
	ss = make([]byte, MLKEM768SharedKeySize)
	pk.EncapsulateTo(ct, ss, nil)
	return ct, ss, nil
}

// MLKEM768Decapsulate decapsulates the ciphertext using the private key.
// Returns the 32-byte shared secret.
func MLKEM768Decapsulate(skBytes, ctBytes []byte) ([]byte, error) {
	if len(skBytes) != MLKEM768PrivateKeySize {
		return nil, fmt.Errorf("mlkem768 decapsulate: private key must be %d bytes, got %d", MLKEM768PrivateKeySize, len(skBytes))
	}
	if len(ctBytes) != MLKEM768CiphertextSize {
		return nil, fmt.Errorf("mlkem768 decapsulate: ciphertext must be %d bytes, got %d", MLKEM768CiphertextSize, len(ctBytes))
	}
	var sk mlkem768.PrivateKey
	if err := sk.Unpack(skBytes); err != nil {
		return nil, fmt.Errorf("mlkem768 decapsulate: unpack sk: %w", err)
	}
	ss := make([]byte, MLKEM768SharedKeySize)
	sk.DecapsulateTo(ss, ctBytes)
	return ss, nil
}
