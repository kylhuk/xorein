package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	mldsa65 "github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

const (
	// MLDSA65PublicKeySize is the size of an ML-DSA-65 public key.
	MLDSA65PublicKeySize = mldsa65.PublicKeySize // 1952
	// MLDSA65PrivateKeySize is the size of an ML-DSA-65 private key.
	MLDSA65PrivateKeySize = mldsa65.PrivateKeySize // 4032
	// MLDSA65SignatureSize is the size of an ML-DSA-65 signature.
	MLDSA65SignatureSize = mldsa65.SignatureSize // 3309
)

// ErrMLDSAVerify is returned when ML-DSA-65 signature verification fails.
var ErrMLDSAVerify = errors.New("mldsa65: signature verification failed")

// GenerateMLDSA65Keypair generates a fresh ML-DSA-65 signing keypair.
// Returns (publicKey bytes, privateKey bytes, error).
func GenerateMLDSA65Keypair() ([]byte, []byte, error) {
	pk, sk, err := mldsa65.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("mldsa65 keygen: %w", err)
	}
	pkBytes := make([]byte, MLDSA65PublicKeySize)
	skBytes := make([]byte, MLDSA65PrivateKeySize)
	var pkArr [mldsa65.PublicKeySize]byte
	var skArr [mldsa65.PrivateKeySize]byte
	pk.Pack(&pkArr)
	sk.Pack(&skArr)
	copy(pkBytes, pkArr[:])
	copy(skBytes, skArr[:])
	return pkBytes, skBytes, nil
}

// SignMLDSA65 signs msg with the private key. Returns MLDSA65SignatureSize (3309) bytes.
// Uses deterministic (non-randomized) signing per spec 01 §6.
func SignMLDSA65(skBytes, msg []byte) ([]byte, error) {
	if len(skBytes) != MLDSA65PrivateKeySize {
		return nil, fmt.Errorf("mldsa65 sign: private key must be %d bytes, got %d", MLDSA65PrivateKeySize, len(skBytes))
	}
	var skArr [mldsa65.PrivateKeySize]byte
	copy(skArr[:], skBytes)
	var sk mldsa65.PrivateKey
	sk.Unpack(&skArr)
	sig := make([]byte, MLDSA65SignatureSize)
	if err := mldsa65.SignTo(&sk, msg, nil, false, sig); err != nil {
		return nil, fmt.Errorf("mldsa65 sign: %w", err)
	}
	return sig, nil
}

// MLDSA65PublicKeyFromPrivate extracts the public key from an ML-DSA-65 private key.
func MLDSA65PublicKeyFromPrivate(skBytes []byte) ([]byte, error) {
	if len(skBytes) != MLDSA65PrivateKeySize {
		return nil, fmt.Errorf("mldsa65: private key must be %d bytes, got %d", MLDSA65PrivateKeySize, len(skBytes))
	}
	var skArr [mldsa65.PrivateKeySize]byte
	copy(skArr[:], skBytes)
	var sk mldsa65.PrivateKey
	sk.Unpack(&skArr)
	pk := sk.Public().(*mldsa65.PublicKey)
	pkBytes := make([]byte, MLDSA65PublicKeySize)
	var pkArr [mldsa65.PublicKeySize]byte
	pk.Pack(&pkArr)
	copy(pkBytes, pkArr[:])
	return pkBytes, nil
}

// VerifyMLDSA65 verifies a ML-DSA-65 signature. Returns ErrMLDSAVerify on failure.
func VerifyMLDSA65(pkBytes, msg, sig []byte) error {
	if len(pkBytes) != MLDSA65PublicKeySize {
		return fmt.Errorf("mldsa65 verify: public key must be %d bytes, got %d", MLDSA65PublicKeySize, len(pkBytes))
	}
	if len(sig) != MLDSA65SignatureSize {
		return fmt.Errorf("mldsa65 verify: signature must be %d bytes, got %d", MLDSA65SignatureSize, len(sig))
	}
	var pkArr [mldsa65.PublicKeySize]byte
	copy(pkArr[:], pkBytes)
	var pk mldsa65.PublicKey
	pk.Unpack(&pkArr)
	if !mldsa65.Verify(&pk, msg, nil, sig) {
		return ErrMLDSAVerify
	}
	return nil
}
