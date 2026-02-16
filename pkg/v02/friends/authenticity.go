package friends

import "time"

type AuthenticityReason string

const (
	AuthenticityReasonValid            AuthenticityReason = "authenticity-valid"
	AuthenticityReasonInvalidRequest   AuthenticityReason = "authenticity-invalid-request"
	AuthenticityReasonMissingSignature AuthenticityReason = "authenticity-missing-signature"
	AuthenticityReasonInvalidSigner    AuthenticityReason = "authenticity-invalid-signer"
	AuthenticityReasonInvalidSignature AuthenticityReason = "authenticity-invalid-signature"
	AuthenticityReasonReplayDetected   AuthenticityReason = "authenticity-replay-detected"
	AuthenticityReasonExpired          AuthenticityReason = "authenticity-expired"
)

type SignedRequest struct {
	RequestID          string
	Nonce              string
	FromID             string
	ToID               string
	CreatedAtUnix      int64
	ExpiresAtUnix      int64
	SignerID           string
	SignatureAlgorithm RequestSignatureAlgorithm
	Signature          []byte
}

type RequestReplayState struct {
	Seen map[string]struct{}
}

type AuthenticityDecision struct {
	Allowed   bool
	Reason    AuthenticityReason
	ReplayKey string
	NextState RequestReplayState
}

type RequestSignatureVerifier func(request SignedRequest) bool

func ValidateRequestAuthenticity(request SignedRequest, state RequestReplayState, now time.Time, verify RequestSignatureVerifier) AuthenticityDecision {
	next := copyReplayState(state)
	replayKey := makeReplayKey(request)

	if replayKey == "" || request.FromID == "" || request.ToID == "" || request.CreatedAtUnix <= 0 || request.ExpiresAtUnix <= request.CreatedAtUnix {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonInvalidRequest, ReplayKey: replayKey, NextState: next}
	}
	if request.SignatureAlgorithm == RequestSignatureAlgorithmUnspecified || len(request.Signature) == 0 || request.SignerID == "" {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonMissingSignature, ReplayKey: replayKey, NextState: next}
	}
	if request.SignerID != request.FromID {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonInvalidSigner, ReplayKey: replayKey, NextState: next}
	}
	if now.Unix() > request.ExpiresAtUnix {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonExpired, ReplayKey: replayKey, NextState: next}
	}
	if _, ok := next.Seen[replayKey]; ok {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonReplayDetected, ReplayKey: replayKey, NextState: next}
	}
	if verify != nil && !verify(request) {
		return AuthenticityDecision{Allowed: false, Reason: AuthenticityReasonInvalidSignature, ReplayKey: replayKey, NextState: next}
	}
	next.Seen[replayKey] = struct{}{}
	return AuthenticityDecision{Allowed: true, Reason: AuthenticityReasonValid, ReplayKey: replayKey, NextState: next}
}

func makeReplayKey(request SignedRequest) string {
	if request.RequestID == "" || request.Nonce == "" {
		return ""
	}
	return request.RequestID + ":" + request.Nonce
}

func copyReplayState(state RequestReplayState) RequestReplayState {
	seen := make(map[string]struct{}, len(state.Seen))
	for replayKey := range state.Seen {
		seen[replayKey] = struct{}{}
	}
	return RequestReplayState{Seen: seen}
}
