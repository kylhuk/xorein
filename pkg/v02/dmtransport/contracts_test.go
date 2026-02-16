package dmtransport

import "testing"

func TestSelectTransportPath(t *testing.T) {
	t.Run("direct", func(t *testing.T) {
		decision := SelectTransportPath(true, true, SecurityModeSeal)
		if decision.Path != DeliveryPathDirect {
			t.Fatalf("path=%q want %q", decision.Path, DeliveryPathDirect)
		}
		if decision.Reason != ReasonDirectAvailable {
			t.Fatalf("reason=%q want %q", decision.Reason, ReasonDirectAvailable)
		}
	})

	t.Run("offline", func(t *testing.T) {
		decision := SelectTransportPath(false, true, SecurityModeSeal)
		if decision.Path != DeliveryPathOffline {
			t.Fatalf("path=%q want %q", decision.Path, DeliveryPathOffline)
		}
		if decision.Reason != ReasonPeerOffline {
			t.Fatalf("reason=%q want %q", decision.Reason, ReasonPeerOffline)
		}
	})

	t.Run("unsupported mode", func(t *testing.T) {
		decision := SelectTransportPath(true, true, SecurityModeClear)
		if decision.Path != DeliveryPathReject {
			t.Fatalf("path=%q want %q", decision.Path, DeliveryPathReject)
		}
		if decision.Reason != ReasonUnsupportedSecurityMode {
			t.Fatalf("reason=%q want %q", decision.Reason, ReasonUnsupportedSecurityMode)
		}
	})
}

func TestTransitionSession(t *testing.T) {
	binding := SessionBinding{SessionID: "s1", ConversationID: "c1", PeerID: "peer"}

	open := TransitionSession(SessionStateClosed, SessionEventOpenRequested, binding)
	if open.State != SessionStateOpening || open.Reason != ReasonSessionOpenRequested {
		t.Fatalf("open request decision=%+v", open)
	}

	confirmed := TransitionSession(open.State, SessionEventOpenConfirmed, binding)
	if confirmed.State != SessionStateOpen || confirmed.Reason != ReasonSessionOpened {
		t.Fatalf("open confirm decision=%+v", confirmed)
	}

	timedOut := TransitionSession(confirmed.State, SessionEventTimeout, binding)
	if timedOut.State != SessionStateRetrying || timedOut.Reason != ReasonSessionTimedOut {
		t.Fatalf("timeout decision=%+v", timedOut)
	}

	retry := TransitionSession(timedOut.State, SessionEventRetryTick, binding)
	if retry.State != SessionStateOpening || retry.Reason != ReasonSessionRetryScheduled {
		t.Fatalf("retry decision=%+v", retry)
	}

	closed := TransitionSession(SessionStateOpen, SessionEventClose, binding)
	if closed.State != SessionStateClosed || closed.Reason != ReasonSessionClosed {
		t.Fatalf("close decision=%+v", closed)
	}
}

func TestTransitionSessionRequiresBinding(t *testing.T) {
	decision := TransitionSession(SessionStateClosed, SessionEventOpenRequested, SessionBinding{})
	if decision.State != SessionStateClosed || decision.Reason != ReasonInvalidSessionBinding {
		t.Fatalf("invalid binding decision=%+v", decision)
	}
}

func TestValidateEnvelopeMetadata(t *testing.T) {
	base := EnvelopeMetadata{
		MessageID:            "m1",
		ConversationID:       "c1",
		SenderID:             "alice",
		RecipientID:          "bob",
		SessionID:            "s1",
		SecurityMode:         SecurityModeSeal,
		ModeEpochID:          "epoch-1",
		IntegrityTag:         []byte{1},
		AdditionalFieldCount: 0,
	}

	tests := []struct {
		name   string
		meta   EnvelopeMetadata
		accept bool
		reason ReasonCode
	}{
		{name: "valid", meta: base, accept: true, reason: ReasonIntegrityVerified},
		{
			name: "missing integrity tag",
			meta: func() EnvelopeMetadata {
				v := base
				v.IntegrityTag = nil
				return v
			}(),
			accept: false,
			reason: ReasonMissingIntegrityTag,
		},
		{
			name: "metadata minimization violation",
			meta: func() EnvelopeMetadata {
				v := base
				v.AdditionalFieldCount = 1
				return v
			}(),
			accept: false,
			reason: ReasonMetadataMinimizationFailed,
		},
		{
			name: "unsupported mode",
			meta: func() EnvelopeMetadata {
				v := base
				v.SecurityMode = SecurityModeClear
				return v
			}(),
			accept: false,
			reason: ReasonUnsupportedSecurityMode,
		},
		{
			name: "invalid session binding",
			meta: func() EnvelopeMetadata {
				v := base
				v.SessionID = ""
				return v
			}(),
			accept: false,
			reason: ReasonInvalidSessionBinding,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateEnvelopeMetadata(tc.meta)
			if got.Accepted != tc.accept || got.Reason != tc.reason {
				t.Fatalf("ValidateEnvelopeMetadata()=%+v want accepted=%v reason=%q", got, tc.accept, tc.reason)
			}
		})
	}
}
