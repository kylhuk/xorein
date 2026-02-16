package phase9

import (
	"context"
	"testing"
	"time"

	phase4 "github.com/aether/code_aether/pkg/phase4"
)

func TestRelayFallbackValidationBlockedPathTriggersRelay(t *testing.T) {
	t.Helper()
	hooks := phase4.TraversalHooks{
		StageActions: map[phase4.TraversalStage]phase4.StageAction{
			phase4.TraversalStageDirect: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonDirectUnavailable, Message: "direct path blocked by symmetric NAT"}
			},
			phase4.TraversalStageAutoNAT: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonAutoNATFailure, Message: "AutoNAT reports private reachability"}
			},
			phase4.TraversalStageHolePunch: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonHolePunchFailure, Message: "hole punch timed out"}
			},
		},
		RelayAction: func(context.Context) phase4.RelayReservationOutcome {
			return phase4.RelayReservationOutcome{
				State: phase4.RelayReservationStateActive,
				Events: []phase4.RelayReservationEvent{
					{State: phase4.RelayReservationStatePending, Reason: phase4.ReasonRelayFallbackActive, Message: "request submitted"},
					{State: phase4.RelayReservationStateReserved, Reason: phase4.ReasonRelayFallbackActive, Message: "reservation accepted"},
					{State: phase4.RelayReservationStateActive, Reason: phase4.ReasonRelayFallbackActive, Message: "circuit carrying session"},
				},
			}
		},
	}
	report := phase4.NewTraversalRunner(phase4.FallbackTimeoutPolicy{}, hooks).Run(context.Background())
	if report.Stage != phase4.TraversalStageRelay {
		t.Fatalf("final stage = %s, want relay", report.Stage)
	}
	if report.Reason != phase4.ReasonRelayFallbackActive {
		t.Fatalf("final reason = %s, want %s", report.Reason, phase4.ReasonRelayFallbackActive)
	}
	wantReasons := []phase4.ConnectivityReasonCode{
		phase4.ReasonDirectUnavailable,
		phase4.ReasonAutoNATFailure,
		phase4.ReasonHolePunchFailure,
	}
	for i, want := range wantReasons {
		if got := report.Events[i].Reason; got != want {
			t.Fatalf("event[%d] reason = %s, want %s", i, got, want)
		}
	}
	if !report.Relay.Active {
		t.Fatalf("relay Active = false, want true")
	}
	if len(report.Relay.Events) != 3 {
		t.Fatalf("relay events = %d, want 3", len(report.Relay.Events))
	}
}

func TestRelayFallbackValidationSessionEstablishmentLifecycle(t *testing.T) {
	hooks := phase4.TraversalHooks{
		StageActions: map[phase4.TraversalStage]phase4.StageAction{
			phase4.TraversalStageDirect: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonDirectUnavailable}
			},
			phase4.TraversalStageAutoNAT: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonAutoNATFailure}
			},
			phase4.TraversalStageHolePunch: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonHolePunchFailure}
			},
		},
		RelayAction: func(context.Context) phase4.RelayReservationOutcome {
			return phase4.RelayReservationOutcome{
				State: phase4.RelayReservationStateActive,
				Events: []phase4.RelayReservationEvent{
					{State: phase4.RelayReservationStatePending, Reason: phase4.ReasonRelayFallbackActive, Message: "pending"},
					{State: phase4.RelayReservationStateReserved, Reason: phase4.ReasonRelayFallbackActive, Message: "reserved"},
					{State: phase4.RelayReservationStateActive, Reason: phase4.ReasonRelayFallbackActive, Message: "active"},
				},
			}
		},
	}
	report := phase4.NewTraversalRunner(phase4.FallbackTimeoutPolicy{}, hooks).Run(context.Background())
	if report.Relay.State != phase4.RelayReservationStateActive {
		t.Fatalf("relay state = %s, want active", report.Relay.State)
	}
	for i, evt := range report.Relay.Events {
		if evt.State != hooks.RelayAction(context.Background()).Events[i].State {
			t.Fatalf("event[%d] state = %s, want %s", i, evt.State, hooks.RelayAction(context.Background()).Events[i].State)
		}
		if evt.Time.IsZero() {
			t.Fatalf("event[%d] has zero timestamp", i)
		}
	}
}

func TestStoreForwardRetrievalAfterRelayFallback(t *testing.T) {
	now := time.Date(2026, time.February, 1, 12, 0, 0, 0, time.UTC)
	store, err := NewStoreService(StoreConfig{
		RetentionTTL:    time.Minute,
		MaxMessages:     4,
		MaxBytes:        16 * 1024,
		MaxPayloadBytes: 1024,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}
	msg := []byte("ciphertext-payload")
	res := store.Store(now, "recipient", msg)
	if !res.Stored {
		t.Fatalf("store result = %+v, want stored", res)
	}
	directBlockedHooks := phase4.TraversalHooks{
		StageActions: map[phase4.TraversalStage]phase4.StageAction{
			phase4.TraversalStageDirect: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonDirectUnavailable}
			},
			phase4.TraversalStageAutoNAT: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonAutoNATFailure}
			},
			phase4.TraversalStageHolePunch: func(context.Context) phase4.StageResult {
				return phase4.StageResult{Success: false, Reason: phase4.ReasonHolePunchFailure}
			},
		},
		RelayAction: func(context.Context) phase4.RelayReservationOutcome {
			return phase4.RelayReservationOutcome{State: phase4.RelayReservationStateActive}
		},
	}
	report := phase4.NewTraversalRunner(phase4.FallbackTimeoutPolicy{}, directBlockedHooks).Run(context.Background())
	if report.Stage != phase4.TraversalStageRelay {
		t.Fatalf("report.Stage = %s, want relay", report.Stage)
	}
	drained := store.DrainRecipient(now.Add(10*time.Second), "recipient")
	if len(drained) != 1 {
		t.Fatalf("drained len = %d, want 1", len(drained))
	}
	if string(drained[0]) != string(msg) {
		t.Fatalf("drained payload = %q, want %q", drained[0], msg)
	}
	snap := store.Snapshot(now.Add(20 * time.Second))
	if snap.QueuedMessages != 0 {
		t.Fatalf("queued messages = %d, want 0", snap.QueuedMessages)
	}
}
