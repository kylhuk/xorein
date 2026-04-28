package crypto

import (
	"crypto/ed25519"
	"errors"
	"fmt"
)

const (
	// HybridSignatureSize is Ed25519 (64 B) || ML-DSA-65 (3309 B) = 3373 bytes.
	// Source: spec 01 §6.
	HybridSignatureSize = Ed25519SignatureSize + MLDSA65SignatureSize // 64 + 3309 = 3373
)

// ErrHybridSignatureMismatch is returned when either component of a hybrid signature fails.
var ErrHybridSignatureMismatch = errors.New("hybrid signature: verification failed")

// HybridSign signs msg with both Ed25519 and ML-DSA-65 keys and returns the
// concatenated hybrid signature (3373 bytes): Ed25519_sig || ML-DSA-65_sig.
// Source: spec 01 §6.
func HybridSign(edPriv ed25519.PrivateKey, mldsaPriv, msg []byte) ([]byte, error) {
	edSig := SignEd25519(edPriv, msg)
	mldsaSig, err := SignMLDSA65(mldsaPriv, msg)
	if err != nil {
		return nil, fmt.Errorf("hybrid sign: %w", err)
	}
	out := make([]byte, HybridSignatureSize)
	copy(out[:Ed25519SignatureSize], edSig)
	copy(out[Ed25519SignatureSize:], mldsaSig)
	return out, nil
}

// HybridVerify verifies a hybrid signature produced by HybridSign.
// Both Ed25519 and ML-DSA-65 MUST verify; failure of either returns ErrHybridSignatureMismatch.
func HybridVerify(edPub ed25519.PublicKey, mldsaPub, msg, sig []byte) error {
	if len(sig) != HybridSignatureSize {
		return fmt.Errorf("hybrid verify: signature must be %d bytes, got %d", HybridSignatureSize, len(sig))
	}
	edSig := sig[:Ed25519SignatureSize]
	mldsaSig := sig[Ed25519SignatureSize:]
	if !VerifyEd25519(edPub, msg, edSig) {
		return ErrHybridSignatureMismatch
	}
	if err := VerifyMLDSA65(mldsaPub, msg, mldsaSig); err != nil {
		return ErrHybridSignatureMismatch
	}
	return nil
}

// CombineKEM produces the hybrid master secret for Seal mode X3DH.
// Implements spec 01 §5.1 / spec 11 §2.1:
//
//	hybrid_master = HKDF-SHA-256(IKM = x3dhSecret || ssMLKEM, salt = "", info = LabelSealHybridMaster, L = 32)
func CombineKEM(x3dhSecret, ssMLKEM []byte) ([KeySize32]byte, error) {
	ikm := make([]byte, len(x3dhSecret)+len(ssMLKEM))
	copy(ikm, x3dhSecret)
	copy(ikm[len(x3dhSecret):], ssMLKEM)
	return DeriveKey32(ikm, nil, LabelSealHybridMaster)
}
