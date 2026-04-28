package envelope

import (
	"bytes"
	"crypto/ed25519"
	"time"

	apb "github.com/aether/code_aether/gen/go/proto"
	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

// TrustFn is an optional callback that reports whether the signer embedded in
// env is trusted to carry its payload type. A nil TrustFn skips step 6.
type TrustFn func(env *apb.SignedEnvelope) bool

// VerifyOption is a functional option for VerifyEnvelopeFull.
type VerifyOption func(*verifyConfig)

type verifyConfig struct {
	trustFn TrustFn
}

// WithTrustFn sets the signer-authorisation callback (spec 02 §2.3 step 6).
// If not provided, step 6 is skipped.
func WithTrustFn(fn TrustFn) VerifyOption {
	return func(c *verifyConfig) { c.trustFn = fn }
}

// VerifyEnvelopeFull verifies env per spec 02 §2.3 and returns:
//   - an ordered list of all VerificationError entries found
//   - the recomputed canonical bytes (nil if step 3 failed)
//   - a fatal error if verification could not be completed (e.g., internal failure)
//
// Returns (nil, canonicalBytes, nil) on success.
// All 7 steps are executed in order; errors are collected rather than stopping
// at the first failure, except where a step's output is required by a later step.
func VerifyEnvelopeFull(env *apb.SignedEnvelope, now time.Time, stored bool, opts ...VerifyOption) (errs []error, canonical []byte, err error) {
	cfg := &verifyConfig{}
	for _, o := range opts {
		o(cfg)
	}

	// Step 1: shape check.
	if env == nil || len(env.Signature) == 0 {
		return []error{ErrUnsignedPayload}, nil, nil
	}
	// Also check for the payload field — canonical_payload and signer are
	// required fields per §2.3 step 1.
	if env.Signer == nil || env.CanonicalPayload == nil {
		return []error{ErrUnsignedPayload}, nil, nil
	}

	// Step 2: algorithm check.
	if env.SignatureAlgorithm != apb.SignatureAlgorithm_SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65 {
		errs = append(errs, ErrUnsupportedPayloadType)
	}

	// Step 3: recompute canonical payload; compare byte-for-byte.
	msg, typeErr := payloadForType(env)
	var recomputed []byte
	if typeErr == nil {
		recomputed, err = BuildCanonicalPayload(uint32(env.PayloadType), msg, int64(env.SignedAt))
		if err != nil {
			errs = append(errs, ErrCanonicalizationMismatch)
		} else if !bytes.Equal(recomputed, env.CanonicalPayload) {
			errs = append(errs, ErrCanonicalizationMismatch)
		}
	} else {
		// Cannot recompute canonical bytes without knowing the type.
		errs = append(errs, ErrCanonicalizationMismatch)
	}

	// Step 4: verify both signatures.
	signer := env.Signer
	if len(signer.SigningPublicKey) == 0 || len(signer.MlDsa65PublicKey) == 0 {
		errs = append(errs, ErrSignatureMismatch)
	} else {
		edPub := ed25519.PublicKey(signer.SigningPublicKey)
		edOK := v0crypto.VerifyEd25519(edPub, env.CanonicalPayload, env.Signature)
		mlErr := v0crypto.VerifyMLDSA65(signer.MlDsa65PublicKey, env.CanonicalPayload, env.MlDsa65Signature)
		if !edOK || mlErr != nil {
			errs = append(errs, ErrSignatureMismatch)
		}
	}

	// Step 5: freshness window.
	window := LiveFreshnessWindow
	if stored {
		window = StoredFreshnessWindow
	}
	signedAt := time.UnixMilli(int64(env.SignedAt))
	age := now.Sub(signedAt)
	if age > window || age < -window {
		errs = append(errs, ErrExpiredSignature)
	}

	// Step 6: signer authorisation (optional).
	if cfg.trustFn != nil && !cfg.trustFn(env) {
		errs = append(errs, ErrUntrustedSigner)
	}

	// Step 7: known payload type check.
	if typeErr != nil {
		errs = append(errs, ErrUnsupportedPayloadType)
	}

	if len(errs) > 0 {
		return errs, recomputed, nil
	}
	return nil, recomputed, nil
}
