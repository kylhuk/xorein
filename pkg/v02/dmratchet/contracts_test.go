package dmratchet

import "testing"

func TestValidatePersistedStateCases(t *testing.T) {
	cases := []struct {
		name        string
		state       RatchetState
		expectation BindingExpectation
		config      RatchetConfig
		wantReason  ValidationReason
		wantValid   bool
		wantResync  bool
	}{
		{
			name: "valid state",
			state: RatchetState{
				SessionID:        "session-x",
				BindingID:        "binding-1",
				ModeEpoch:        "epoch-1",
				RootKey:          []byte("root"),
				SkippedPerChain:  map[ChainRole]uint32{ChainReceiving: 1},
				TotalSkippedKeys: 1,
				ReplayCache:      map[string]struct{}{"msg-001": {}},
			},
			expectation: BindingExpectation{SessionID: "session-x", BindingID: "binding-1", ModeEpoch: "epoch-1"},
			config:      DefaultConfig(),
			wantReason:  ReasonStateValid,
			wantValid:   true,
		},
		{
			name: "missing root key",
			state: RatchetState{
				SessionID: "session",
			},
			wantReason: ReasonMissingRootKey,
			wantResync: true,
		},
		{
			name: "missing session id",
			state: RatchetState{
				RootKey: []byte("root"),
			},
			wantReason: ReasonMissingSessionID,
			wantResync: true,
		},
		{
			name: "binding mismatch",
			state: RatchetState{
				SessionID: "session",
				BindingID: "binding",
				RootKey:   []byte("root"),
			},
			expectation: BindingExpectation{BindingID: "other"},
			wantReason:  ReasonSessionBindingMismatch,
			wantResync:  true,
		},
		{
			name: "mode epoch mismatch",
			state: RatchetState{
				SessionID: "session",
				BindingID: "binding",
				ModeEpoch: "epoch-a",
				RootKey:   []byte("root"),
			},
			expectation: BindingExpectation{ModeEpoch: "epoch-b"},
			wantReason:  ReasonModeEpochMismatch,
			wantResync:  true,
		},
		{
			name: "skipped per chain exceeded",
			state: RatchetState{
				SessionID:        "session",
				RootKey:          []byte("root"),
				TotalSkippedKeys: 1,
				SkippedPerChain: map[ChainRole]uint32{
					ChainReceiving: 3,
				},
			},
			config:     RatchetConfig{MaxSkippedKeysPerChain: 2},
			wantReason: ReasonSkippedKeysExceeded,
			wantResync: true,
		},
		{
			name: "total skipped keys exceeded",
			state: RatchetState{
				SessionID:        "session",
				RootKey:          []byte("root"),
				TotalSkippedKeys: 5,
			},
			config:     RatchetConfig{MaxTotalSkippedKeys: 3},
			wantReason: ReasonSkippedKeysExceeded,
			wantResync: true,
		},
		{
			name: "replay window exceeded",
			state: RatchetState{
				SessionID: "session",
				RootKey:   []byte("root"),
				ReplayCache: map[string]struct{}{
					"a": {},
					"b": {},
					"c": {},
				},
			},
			config:     RatchetConfig{ReplayWindow: 2},
			wantReason: ReasonValidationReplayWindowExceeded,
			wantResync: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := ValidatePersistedState(tc.state, tc.expectation, tc.config)
			if decision.Reason != tc.wantReason {
				t.Fatalf("expected %s reason, got %+v", tc.wantReason, decision)
			}
			if decision.Valid != tc.wantValid {
				t.Fatalf("expected valid=%t, got %+v", tc.wantValid, decision)
			}
			if decision.ResyncRequired != tc.wantResync {
				t.Fatalf("expected resync=%t, got %+v", tc.wantResync, decision)
			}
		})
	}
}

func TestEvaluateIncomingMessageCases(t *testing.T) {
	cases := []struct {
		name          string
		state         RatchetState
		message       IncomingMessage
		config        RatchetConfig
		wantAccept    bool
		wantReplay    bool
		wantResync    bool
		wantReason    DecisionReason
		wantGap       uint32
		wantSkipped   uint32
		wantSkipChain ChainRole
	}{
		{
			name:       "missing message id",
			message:    IncomingMessage{MessageID: ""},
			wantReason: ReasonMessageIDMissing,
			wantAccept: false,
			wantResync: false,
			wantReplay: false,
		},
		{
			name:       "duplicate payload",
			state:      RatchetState{ReplayCache: map[string]struct{}{"dup": {}}},
			message:    IncomingMessage{MessageID: "dup", Chain: ChainReceiving, Counter: 1},
			wantReason: ReasonReplayDetected,
			wantReplay: true,
		},
		{
			name:       "replay window resync",
			state:      RatchetState{ReplayCache: map[string]struct{}{"a": {}, "b": {}}},
			message:    IncomingMessage{MessageID: "c", Chain: ChainReceiving, Counter: 1},
			config:     RatchetConfig{ReplayWindow: 1},
			wantReason: ReasonDecisionReplayWindowExceeded,
			wantResync: true,
		},
		{
			name:       "counter behind",
			state:      RatchetState{ReceivingCounter: 5},
			message:    IncomingMessage{MessageID: "msg", Chain: ChainReceiving, Counter: 3},
			wantReason: ReasonCounterBehind,
			wantResync: false,
		},
		{
			name:       "in order",
			state:      RatchetState{ReceivingCounter: 5},
			message:    IncomingMessage{MessageID: "msg", Chain: ChainReceiving, Counter: 5},
			wantReason: ReasonInOrderMessage,
			wantAccept: true,
		},
		{
			name:          "out of order within bounds",
			state:         RatchetState{ReceivingCounter: 2, TotalSkippedKeys: 1},
			message:       IncomingMessage{MessageID: "msg-b", Chain: ChainReceiving, Counter: 5},
			wantReason:    ReasonOutOfOrderWithinBounds,
			wantAccept:    true,
			wantGap:       3,
			wantSkipped:   4,
			wantSkipChain: ChainReceiving,
		},
		{
			name:       "skip window exceeded",
			state:      RatchetState{ReceivingCounter: 10},
			message:    IncomingMessage{MessageID: "msg", Chain: ChainReceiving, Counter: 15},
			config:     RatchetConfig{MaxSkippedKeysPerChain: 3},
			wantReason: ReasonSkipWindowExceeded,
			wantResync: true,
		},
		{
			name:       "total skip budget exceeded",
			state:      RatchetState{ReceivingCounter: 1, TotalSkippedKeys: 4},
			message:    IncomingMessage{MessageID: "msg", Chain: ChainReceiving, Counter: 8},
			config:     RatchetConfig{MaxTotalSkippedKeys: 5},
			wantReason: ReasonTotalSkipBudgetExceeded,
			wantResync: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := EvaluateIncomingMessage(tc.state, tc.message, tc.config)
			if decision.Reason != tc.wantReason {
				t.Fatalf("expected %s, got %+v", tc.wantReason, decision)
			}
			if tc.wantAccept && !decision.Accept {
				t.Fatalf("expected accept, got %+v", decision)
			}
			if !tc.wantAccept && decision.Accept {
				t.Fatalf("did not expect accept, got %+v", decision)
			}
			if tc.wantReplay != decision.Replay {
				t.Fatalf("expected replay=%t, got %+v", tc.wantReplay, decision)
			}
			if tc.wantResync != decision.Resync {
				t.Fatalf("expected resync=%t, got %+v", tc.wantResync, decision)
			}
			if tc.wantGap != 0 && decision.Gap != tc.wantGap {
				t.Fatalf("expected gap=%d, got %+v", tc.wantGap, decision)
			}
			if tc.wantSkipped != 0 && decision.NewSkippedTotal != tc.wantSkipped {
				t.Fatalf("expected skipped=%d, got %+v", tc.wantSkipped, decision)
			}
			if tc.wantSkipChain != "" && decision.SkipChain != tc.wantSkipChain {
				t.Fatalf("expected chain=%s, got %+v", tc.wantSkipChain, decision)
			}
		})
	}
}
