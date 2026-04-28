package protocol

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/aether/code_aether/pkg/v0_1/envelope"
)

// ErrSigMismatch is the sentinel returned when hybrid signature verification fails.
var ErrSigMismatch = errors.New("hybrid signature mismatch")

// VerifyHybridSig verifies a combined Ed25519 + ML-DSA-65 signature over canonicalBytes.
//
// sig is a base64url-no-pad string encoding Ed25519_sig (64 bytes) || ML-DSA-65_sig
// as produced by envelope.EncodeHybridSig. edPub must be 32 bytes; mldsaPub must be
// crypto.MLDSA65PublicKeySize (1952) bytes.
//
// Returns nil on success. Returns a *PeerStreamError with CodeSignatureMismatch on any
// verification failure, wrapping ErrSigMismatch as the cause.
func VerifyHybridSig(canonicalBytes []byte, sig string, edPub, mldsaPub []byte) error {
	edSig, mldsaSig, err := envelope.DecodeHybridSig(sig)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSigMismatch, err)
	}

	if !v0crypto.VerifyEd25519(ed25519.PublicKey(edPub), canonicalBytes, edSig) {
		return fmt.Errorf("%w: ed25519 signature invalid", ErrSigMismatch)
	}

	if err := v0crypto.VerifyMLDSA65(mldsaPub, canonicalBytes, mldsaSig); err != nil {
		return fmt.Errorf("%w: ml-dsa-65 signature invalid: %w", ErrSigMismatch, err)
	}

	return nil
}

// newSigMismatchError returns a *PeerStreamError with CodeSignatureMismatch.
// Callers that need a wire-level error (e.g. inside a HandleStream) should use this
// instead of returning ErrSigMismatch directly.
func newSigMismatchError(detail string) *PeerStreamError {
	return &PeerStreamError{
		Code:    CodeSignatureMismatch,
		Message: detail,
	}
}
