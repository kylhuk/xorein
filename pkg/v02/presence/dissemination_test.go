package presence

import (
	"testing"
	"time"
)

func TestEvaluatePublishCadence(t *testing.T) {
	now := time.Unix(1000, 0)
	policy := CadencePolicy{
		DebounceWindow:     2 * time.Second,
		MinPublishInterval: 5 * time.Second,
		HeartbeatInterval:  15 * time.Second,
	}

	initial := EvaluatePublishCadence(policy, PublishTracker{}, PublishInput{ObservedAt: now, State: StateOnline, Status: "hello"})
	if !initial.Publish || initial.Reason != PublishReasonInitial {
		t.Fatalf("initial decision mismatch: %+v", initial)
	}

	debounced := EvaluatePublishCadence(policy, initial.Next, PublishInput{ObservedAt: now.Add(time.Second), State: StateIdle, Status: "hello"})
	if debounced.Publish || debounced.Reason != PublishReasonDebounced {
		t.Fatalf("debounced decision mismatch: %+v", debounced)
	}

	throttled := EvaluatePublishCadence(policy, initial.Next, PublishInput{ObservedAt: now.Add(3 * time.Second), State: StateIdle, Status: "hello"})
	if throttled.Publish || throttled.Reason != PublishReasonThrottled {
		t.Fatalf("throttled decision mismatch: %+v", throttled)
	}

	stateChanged := EvaluatePublishCadence(policy, initial.Next, PublishInput{ObservedAt: now.Add(6 * time.Second), State: StateIdle, Status: "hello"})
	if !stateChanged.Publish || stateChanged.Reason != PublishReasonStateChanged {
		t.Fatalf("state change decision mismatch: %+v", stateChanged)
	}

	unchanged := EvaluatePublishCadence(policy, stateChanged.Next, PublishInput{ObservedAt: now.Add(10 * time.Second), State: StateIdle, Status: "hello"})
	if unchanged.Publish || unchanged.Reason != PublishReasonUnchanged {
		t.Fatalf("unchanged decision mismatch: %+v", unchanged)
	}

	heartbeat := EvaluatePublishCadence(policy, stateChanged.Next, PublishInput{ObservedAt: now.Add(22 * time.Second), State: StateIdle, Status: "hello"})
	if !heartbeat.Publish || heartbeat.Reason != PublishReasonHeartbeat {
		t.Fatalf("heartbeat decision mismatch: %+v", heartbeat)
	}
}

func TestApplyStatusUpdate(t *testing.T) {
	now := time.Unix(1000, 0)
	local := StatusRecord{Value: "hello", Version: 2, UpdatedAt: now}

	applied := ApplyStatusUpdate(local, StatusRecord{Value: "updated", Version: 3, UpdatedAt: now.Add(time.Second)})
	if !applied.Applied || !applied.Invalidated || applied.Reason != StatusReasonApplied || applied.Next.Value != "updated" {
		t.Fatalf("applied decision mismatch: %+v", applied)
	}

	stale := ApplyStatusUpdate(local, StatusRecord{Value: "stale", Version: 1, UpdatedAt: now.Add(time.Second)})
	if stale.Applied || stale.Reason != StatusReasonStaleVersion {
		t.Fatalf("stale decision mismatch: %+v", stale)
	}

	redacted := ApplyStatusUpdate(local, StatusRecord{Value: "secret", Version: 3, UpdatedAt: now.Add(2 * time.Second), Redacted: true})
	if !redacted.Applied || !redacted.Invalidated || redacted.Reason != StatusReasonRedacted || redacted.Next.Value != "" || !redacted.Next.Redacted {
		t.Fatalf("redacted decision mismatch: %+v", redacted)
	}

	noop := ApplyStatusUpdate(local, StatusRecord{Value: "hello", Version: 2, UpdatedAt: now})
	if noop.Applied || noop.Reason != StatusReasonNoop {
		t.Fatalf("noop decision mismatch: %+v", noop)
	}
}

func TestEvaluateStatusPropagation(t *testing.T) {
	now := time.Unix(1000, 0)
	policy := PropagationPolicy{MinInterval: 5 * time.Second, MaxLatency: 20 * time.Second}

	initial := EvaluateStatusPropagation(policy, time.Time{}, now, true)
	if !initial.Propagate || initial.Reason != PropagationReasonInitial {
		t.Fatalf("initial decision mismatch: %+v", initial)
	}

	throttled := EvaluateStatusPropagation(policy, now, now.Add(3*time.Second), true)
	if throttled.Propagate || throttled.Reason != PropagationReasonThrottled {
		t.Fatalf("throttled decision mismatch: %+v", throttled)
	}

	changed := EvaluateStatusPropagation(policy, now, now.Add(6*time.Second), true)
	if !changed.Propagate || changed.Reason != PropagationReasonChanged {
		t.Fatalf("changed decision mismatch: %+v", changed)
	}

	unchanged := EvaluateStatusPropagation(policy, now, now.Add(8*time.Second), false)
	if unchanged.Propagate || unchanged.Reason != PropagationReasonUnchanged {
		t.Fatalf("unchanged decision mismatch: %+v", unchanged)
	}

	heartbeat := EvaluateStatusPropagation(policy, now, now.Add(25*time.Second), false)
	if !heartbeat.Propagate || heartbeat.Reason != PropagationReasonHeartbeat {
		t.Fatalf("heartbeat decision mismatch: %+v", heartbeat)
	}
}
