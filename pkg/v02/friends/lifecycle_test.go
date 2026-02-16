package friends

import "testing"

func TestFriendRequestTransitions(t *testing.T) {
	state, reason := NextRequestState(RequestStatePending, RequestActionAccept)
	if state != RequestStateAccepted || reason != ReasonStateTransition {
		t.Fatalf("unexpected transition: state=%q reason=%q", state, reason)
	}
	rejected, reason := NextRequestState(RequestStateBlocked, RequestActionAccept)
	if rejected != RequestStateBlocked || reason != ReasonInvalidTransition {
		t.Fatalf("blocked transition mismatch: state=%q reason=%q", rejected, reason)
	}
}

func TestResolveConcurrentRequestsDeterministic(t *testing.T) {
	a := Request{FromID: "z", ToID: "a", State: RequestStatePending}
	b := Request{FromID: "a", ToID: "z", State: RequestStatePending}
	resolved := ResolveConcurrent(a, b)
	if resolved.State != RequestStateAccepted {
		t.Fatalf("state=%q want %q", resolved.State, RequestStateAccepted)
	}
	if resolved.CanonicalRequester != "a" {
		t.Fatalf("requester=%q want a", resolved.CanonicalRequester)
	}
}
