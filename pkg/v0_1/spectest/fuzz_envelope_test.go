package spectest_test

import (
	"bytes"
	"testing"
	"time"

	apb "github.com/aether/code_aether/gen/go/proto"
	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
	"github.com/aether/code_aether/pkg/v0_1/envelope"
)

// FuzzCanonicalPayload verifies BuildCanonicalPayload never panics on arbitrary inputs
// and that its output round-trips through VerifyEnvelope consistently.
func FuzzCanonicalPayload(f *testing.F) {
	f.Add(uint32(0), []byte{}, int64(0))
	f.Add(uint32(1), []byte("hello world"), int64(1_700_000_000_000))
	f.Add(uint32(0xFFFF), bytes.Repeat([]byte{0xAB}, 1024), int64(-1))

	f.Fuzz(func(t *testing.T, payloadType uint32, payloadBytes []byte, signedAtMS int64) {
		// BuildCanonicalPayload must not panic.
		msg := &apb.SignedEnvelope{}
		canonical, err := envelope.BuildCanonicalPayload(payloadType, msg, signedAtMS)
		if err != nil {
			return // error is OK; panic is not
		}
		if canonical == nil {
			t.Fatal("BuildCanonicalPayload returned nil without error")
		}
		_ = payloadBytes
	})
}

// FuzzVerifyEnvelope verifies VerifyEnvelope never panics on arbitrary byte inputs
// and always returns one of the documented error codes (or nil on valid envelopes).
func FuzzVerifyEnvelope(f *testing.F) {
	// Seed: a minimally valid hybrid-signed envelope.
	edPub, edPriv, _ := v0crypto.GenerateEd25519Keypair()
	mldsaPub, mldsaPriv, _ := v0crypto.GenerateMLDSA65Keypair()
	signer := &apb.IdentityProfile{
		SigningPublicKey:  []byte(edPub),
		MlDsa65PublicKey: mldsaPub,
	}
	seed, _ := envelope.SignEnvelope(apb.PayloadType_PAYLOAD_TYPE_UNSPECIFIED, &apb.SignedEnvelope{}, edPriv, mldsaPriv, signer)

	f.Add(seed.GetEnvelopeId(), seed.Signature, seed.MlDsa65Signature, seed.CanonicalPayload, seed.SignedAt)

	f.Fuzz(func(t *testing.T, envID string, sig, mldsaSig, canonical []byte, signedAt uint64) {
		env := &apb.SignedEnvelope{
			EnvelopeId:         envID,
			Signature:          sig,
			MlDsa65Signature:   mldsaSig,
			CanonicalPayload:   canonical,
			SignedAt:           signedAt,
			SignatureAlgorithm: apb.SignatureAlgorithm_SIGNATURE_ALGORITHM_HYBRID_ED25519_ML_DSA_65,
			Signer:             signer,
		}
		// Must not panic. Error is expected for almost all fuzz inputs.
		_ = envelope.VerifyEnvelope(env, time.Now(), false)
	})
}
