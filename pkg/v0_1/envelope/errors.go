// Package envelope implements canonical payload construction, signing, and verification
// per docs/spec/v0.1/02-canonical-envelope.md.
package envelope

import (
	"errors"

	apb "github.com/aether/code_aether/gen/go/proto"
)

// VerificationError wraps an EnvelopeVerificationError code with detail.
type VerificationError struct {
	Code   apb.EnvelopeVerificationError
	Detail string
}

func (e *VerificationError) Error() string {
	return e.Code.String() + ": " + e.Detail
}

// Sentinel errors for each code — use errors.As to check the code field.
var (
	ErrSignatureMismatch       = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_SIGNATURE_MISMATCH, Detail: "one or both signatures did not verify"}
	ErrUnsignedPayload         = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_UNSIGNED_PAYLOAD, Detail: "envelope has no signature"}
	ErrUnsupportedPayloadType  = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_UNSUPPORTED_PAYLOAD_TYPE, Detail: "unrecognised payload type"}
	ErrCanonicalizationMismatch = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_CANONICALIZATION_MISMATCH, Detail: "canonical payload does not match recomputed value"}
	ErrExpiredSignature        = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_EXPIRED_SIGNATURE, Detail: "signed_at is outside the permitted freshness window"}
	ErrUntrustedSigner         = &VerificationError{Code: apb.EnvelopeVerificationError_ENVELOPE_VERIFICATION_ERROR_UNTRUSTED_SIGNER, Detail: "signer is not authorised for this payload type"}
)

// IsVerificationError returns true when err is an EnvelopeVerificationError
// with the given code.
func IsVerificationError(err error, code apb.EnvelopeVerificationError) bool {
	var ve *VerificationError
	return errors.As(err, &ve) && ve.Code == code
}
