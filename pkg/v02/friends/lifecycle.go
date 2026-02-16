package friends

import "strings"

type RequestState string

const (
	RequestStatePending  RequestState = "pending"
	RequestStateAccepted RequestState = "accepted"
	RequestStateDeclined RequestState = "declined"
	RequestStateCanceled RequestState = "canceled"
	RequestStateBlocked  RequestState = "blocked"
)

type RequestAction string

const (
	RequestActionAccept  RequestAction = "accept"
	RequestActionDecline RequestAction = "decline"
	RequestActionCancel  RequestAction = "cancel"
	RequestActionBlock   RequestAction = "block"
)

type TransitionReason string

const (
	ReasonStateTransition   TransitionReason = "state-transition"
	ReasonInvalidTransition TransitionReason = "invalid-transition"
)

func NextRequestState(current RequestState, action RequestAction) (RequestState, TransitionReason) {
	switch current {
	case RequestStatePending:
		switch action {
		case RequestActionAccept:
			return RequestStateAccepted, ReasonStateTransition
		case RequestActionDecline:
			return RequestStateDeclined, ReasonStateTransition
		case RequestActionCancel:
			return RequestStateCanceled, ReasonStateTransition
		case RequestActionBlock:
			return RequestStateBlocked, ReasonStateTransition
		}
	case RequestStateAccepted:
		if action == RequestActionBlock {
			return RequestStateBlocked, ReasonStateTransition
		}
	}
	return current, ReasonInvalidTransition
}

type Request struct {
	FromID string
	ToID   string
	State  RequestState
}

type ConcurrentResolution struct {
	State              RequestState
	CanonicalRequester string
	CanonicalRecipient string
}

// ResolveConcurrent applies deterministic resolution for simultaneous pending requests.
func ResolveConcurrent(a Request, b Request) ConcurrentResolution {
	requester := strings.TrimSpace(a.FromID)
	recipient := strings.TrimSpace(a.ToID)
	if requester == "" || recipient == "" {
		return ConcurrentResolution{}
	}
	if a.State == RequestStatePending && b.State == RequestStatePending && a.FromID == b.ToID && a.ToID == b.FromID {
		if requester > recipient {
			requester, recipient = recipient, requester
		}
		return ConcurrentResolution{
			State:              RequestStateAccepted,
			CanonicalRequester: requester,
			CanonicalRecipient: recipient,
		}
	}
	return ConcurrentResolution{
		State:              a.State,
		CanonicalRequester: requester,
		CanonicalRecipient: recipient,
	}
}
