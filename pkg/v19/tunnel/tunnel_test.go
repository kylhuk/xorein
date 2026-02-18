package tunnel

import "testing"

func TestPolicyEvaluateTransitions(t *testing.T) {
	policy := DefaultPolicy()

	dec := policy.Evaluate(0, true, true)
	if dec.Action != ActionNoOp || dec.Reason != "healthy" {
		t.Fatalf("healthy path should noop, got %v", dec)
	}

	dec = policy.Evaluate(1, true, false)
	if dec.Action != ActionRetry || dec.Reason != "heartbeat-missed" {
		t.Fatalf("expect retry on heartbeat miss, got %v", dec)
	}

	dec = policy.Evaluate(policy.MaxRetries, false, false)
	if dec.Action != ActionClose || dec.Reason != "retry-limit" {
		t.Fatalf("limit reached should close, got %v", dec)
	}
}
