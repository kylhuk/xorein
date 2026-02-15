package phase4

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTraversalRunnerFallbackOrder(t *testing.T) {
	t.Parallel()
	var called []TraversalStage
	hooks := TraversalHooks{
		StageActions: map[TraversalStage]StageAction{
			TraversalStageDirect: func(_ context.Context) StageResult {
				called = append(called, TraversalStageDirect)
				return StageResult{Success: false, Reason: ReasonDirectUnavailable}
			},
			TraversalStageAutoNAT: func(_ context.Context) StageResult {
				called = append(called, TraversalStageAutoNAT)
				return StageResult{Success: false, Reason: ReasonAutoNATFailure}
			},
			TraversalStageHolePunch: func(_ context.Context) StageResult {
				called = append(called, TraversalStageHolePunch)
				return StageResult{Success: true, Reason: ReasonHolePunchSuccess}
			},
		},
	}
	runner := NewTraversalRunner(FallbackTimeoutPolicy{}, hooks)
	report := runner.Run(context.Background())
	wantStages := []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch}
	if !reflect.DeepEqual(called, wantStages) {
		t.Fatalf("called stages = %v, want %v", called, wantStages)
	}
	if report.Stage != TraversalStageHolePunch {
		t.Fatalf("final stage = %s, want %s", report.Stage, TraversalStageHolePunch)
	}
	if report.Reason != ReasonHolePunchSuccess {
		t.Fatalf("final reason = %s, want %s", report.Reason, ReasonHolePunchSuccess)
	}
	if len(report.Events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(report.Events))
	}
	if got, want := report.Events[0].Stage, TraversalStageDirect; got != want {
		t.Fatalf("event[0] stage = %s, want %s", got, want)
	}
	if got, want := report.Events[1].Stage, TraversalStageAutoNAT; got != want {
		t.Fatalf("event[1] stage = %s, want %s", got, want)
	}
	if got, want := report.Events[2].Stage, TraversalStageHolePunch; got != want {
		t.Fatalf("event[2] stage = %s, want %s", got, want)
	}
	lrec := report.Events[len(report.Events)-1]
	if lrec.Reason != ReasonRecoveryTriggered {
		t.Fatalf("last event reason = %s, want %s", lrec.Reason, ReasonRecoveryTriggered)
	}
	if !strings.Contains(lrec.Message, string(TraversalStageHolePunch)) {
		t.Fatalf("recovery message %q does not mention %s", lrec.Message, TraversalStageHolePunch)
	}
}

func TestFallbackTimeoutPolicyTimeoutFor(t *testing.T) {
	policy := FallbackTimeoutPolicy{
		StageTimeouts: map[TraversalStage]time.Duration{
			TraversalStageDirect:    15 * time.Second,
			TraversalStageHolePunch: 8 * time.Second,
		},
		RelayTimeout: 20 * time.Second,
	}
	tests := []struct {
		name  string
		stage TraversalStage
		want  time.Duration
	}{
		{name: "direct override", stage: TraversalStageDirect, want: 15 * time.Second},
		{name: "auto default", stage: TraversalStageAutoNAT, want: 5 * time.Second},
		{name: "hole override", stage: TraversalStageHolePunch, want: 8 * time.Second},
		{name: "relay override", stage: TraversalStageRelay, want: 20 * time.Second},
		{name: "unknown default", stage: TraversalStage("custom"), want: 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policy.TimeoutFor(tt.stage); got != tt.want {
				t.Fatalf("TimeoutFor(%s) = %s, want %s", tt.stage, got, tt.want)
			}
		})
	}
}

func TestTraversalRunnerStageTimeoutEvent(t *testing.T) {
	policy := FallbackTimeoutPolicy{
		StageTimeouts: map[TraversalStage]time.Duration{
			TraversalStageDirect: time.Millisecond,
		},
	}
	hooks := TraversalHooks{
		StageActions: map[TraversalStage]StageAction{
			TraversalStageDirect: func(ctx context.Context) StageResult {
				<-ctx.Done()
				return StageResult{Success: false, Reason: ReasonTraversalTimeout}
			},
		},
	}
	report := NewTraversalRunner(policy, hooks).Run(context.Background())
	if len(report.Events) == 0 {
		t.Fatal("expected at least one event")
	}
	evt := report.Events[0]
	if evt.Stage != TraversalStageDirect {
		t.Fatalf("stage = %s, want %s", evt.Stage, TraversalStageDirect)
	}
	if evt.Reason != ReasonTraversalTimeout {
		t.Fatalf("reason = %s, want %s", evt.Reason, ReasonTraversalTimeout)
	}
	if evt.Message != context.DeadlineExceeded.Error() {
		t.Fatalf("message = %q, want %q", evt.Message, context.DeadlineExceeded.Error())
	}
}

func TestTraversalRunnerRelayReservationLifecycle(t *testing.T) {
	cases := []struct {
		name       string
		outcome    RelayReservationOutcome
		wantReason ConnectivityReasonCode
		wantActive bool
	}{
		{
			name: "active reservation",
			outcome: RelayReservationOutcome{
				State: RelayReservationStateActive,
				Events: []RelayReservationEvent{
					{State: RelayReservationStatePending, Reason: ReasonRelayFallbackActive, Message: "pending"},
					{State: RelayReservationStateReserved, Reason: ReasonRelayFallbackActive, Message: "reserved"},
					{State: RelayReservationStateActive, Reason: ReasonRelayFallbackActive, Message: "active"},
				},
			},
			wantReason: ReasonRelayFallbackActive,
			wantActive: true,
		},
		{
			name: "failed reservation",
			outcome: RelayReservationOutcome{
				State: RelayReservationStateFailed,
				Events: []RelayReservationEvent{
					{State: RelayReservationStateFailed, Reason: ReasonRelayReservationFailed, Message: "failed"},
				},
			},
			wantReason: ReasonRelayReservationFailed,
			wantActive: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hooks := TraversalHooks{
				RelayAction: func(_ context.Context) RelayReservationOutcome {
					return tc.outcome
				},
			}
			report := NewTraversalRunner(FallbackTimeoutPolicy{}, hooks).Run(context.Background())
			if report.Stage != TraversalStageRelay {
				t.Fatalf("stage = %s, want %s", report.Stage, TraversalStageRelay)
			}
			if report.Reason != tc.wantReason {
				t.Fatalf("reason = %s, want %s", report.Reason, tc.wantReason)
			}
			if got, want := report.Events[len(report.Events)-1].Stage, TraversalStageRelay; got != want {
				t.Fatalf("last event stage = %s, want %s", got, want)
			}
			if got, want := report.Events[len(report.Events)-1].Reason, tc.wantReason; got != want {
				t.Fatalf("last event reason = %s, want %s", got, want)
			}
			if report.Relay.State != tc.outcome.State {
				t.Fatalf("relay state = %s, want %s", report.Relay.State, tc.outcome.State)
			}
			if report.Relay.Active != tc.wantActive {
				t.Fatalf("relay active = %t, want %t", report.Relay.Active, tc.wantActive)
			}
			if len(report.Relay.Events) != len(tc.outcome.Events) {
				t.Fatalf("relay events = %d, want %d", len(report.Relay.Events), len(tc.outcome.Events))
			}
			for i, evt := range report.Relay.Events {
				want := tc.outcome.Events[i]
				if evt.State != want.State {
					t.Fatalf("relay event[%d] state = %s, want %s", i, evt.State, want.State)
				}
				if evt.Reason != want.Reason {
					t.Fatalf("relay event[%d] reason = %s, want %s", i, evt.Reason, want.Reason)
				}
				if evt.Message != want.Message {
					t.Fatalf("relay event[%d] message = %q, want %q", i, evt.Message, want.Message)
				}
				if evt.Time.IsZero() {
					t.Fatalf("relay event[%d] time is zero", i)
				}
			}
		})
	}
}
