package ui

import "testing"

func TestEvaluateStateTransitions(t *testing.T) {
	state, action := EvaluateState(StateInCall, false, false)
	if state != StateRecovering || action != ActionShowRecover {
		t.Fatalf("expected recover guidance, got %v %v", state, action)
	}

	state, action = EvaluateState(StateIdle, true, true)
	if state != StateInCall || action != ActionShowInCall {
		t.Fatalf("wake should go to in-call, got %v %v", state, action)
	}

	state, action = EvaluateState(StateConnecting, true, false)
	if state != StateConnecting || action != ActionNoVisualWork {
		t.Fatalf("healthy silent state should no-op, got %v %v", state, action)
	}
}
