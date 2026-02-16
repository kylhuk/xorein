package friends

import (
	"testing"
	"time"
)

func TestValidateRequestAuthenticity(t *testing.T) {
	now := time.Unix(1000, 0)
	request := SignedRequest{
		RequestID:          "req-1",
		Nonce:              "nonce-1",
		FromID:             "alice",
		ToID:               "bob",
		CreatedAtUnix:      now.Add(-time.Minute).Unix(),
		ExpiresAtUnix:      now.Add(time.Minute).Unix(),
		SignerID:           "alice",
		SignatureAlgorithm: RequestSignatureAlgorithmEd25519,
		Signature:          []byte{1, 2, 3},
	}

	accepted := ValidateRequestAuthenticity(request, RequestReplayState{Seen: map[string]struct{}{}}, now, func(SignedRequest) bool { return true })
	if !accepted.Allowed || accepted.Reason != AuthenticityReasonValid {
		t.Fatalf("accepted decision mismatch: %+v", accepted)
	}
	if _, ok := accepted.NextState.Seen[accepted.ReplayKey]; !ok {
		t.Fatalf("expected replay key to be recorded")
	}

	duplicate := ValidateRequestAuthenticity(request, accepted.NextState, now, func(SignedRequest) bool { return true })
	if duplicate.Allowed || duplicate.Reason != AuthenticityReasonReplayDetected {
		t.Fatalf("duplicate decision mismatch: %+v", duplicate)
	}

	expired := request
	expired.ExpiresAtUnix = now.Add(-time.Second).Unix()
	denied := ValidateRequestAuthenticity(expired, RequestReplayState{Seen: map[string]struct{}{}}, now, func(SignedRequest) bool { return true })
	if denied.Allowed || denied.Reason != AuthenticityReasonExpired {
		t.Fatalf("expired decision mismatch: %+v", denied)
	}

	signerMismatch := request
	signerMismatch.SignerID = "mallory"
	denied = ValidateRequestAuthenticity(signerMismatch, RequestReplayState{Seen: map[string]struct{}{}}, now, func(SignedRequest) bool { return true })
	if denied.Allowed || denied.Reason != AuthenticityReasonInvalidSigner {
		t.Fatalf("invalid signer decision mismatch: %+v", denied)
	}

	invalidSignature := ValidateRequestAuthenticity(request, RequestReplayState{Seen: map[string]struct{}{}}, now, func(SignedRequest) bool { return false })
	if invalidSignature.Allowed || invalidSignature.Reason != AuthenticityReasonInvalidSignature {
		t.Fatalf("invalid signature decision mismatch: %+v", invalidSignature)
	}

	missingSignature := request
	missingSignature.Signature = nil
	denied = ValidateRequestAuthenticity(missingSignature, RequestReplayState{Seen: map[string]struct{}{}}, now, nil)
	if denied.Allowed || denied.Reason != AuthenticityReasonMissingSignature {
		t.Fatalf("missing signature decision mismatch: %+v", denied)
	}
}
