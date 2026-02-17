package relay

import "testing"

func TestCapacityEnvelopeUtilization(t *testing.T) {
	t.Parallel()

	envelope := CapacityEnvelope{Limit: 0}
	if got := envelope.UtilizationPercent(); got != 0 {
		t.Fatalf("utilization = %d, want 0 when limit zero and no usage", got)
	}
	envelope = CapacityEnvelope{Limit: 500, Control: 50, ActiveMedia: 150, StoreForward: 50, BulkSync: 50}
	if got := envelope.UtilizationPercent(); got != 60 {
		t.Fatalf("utilization = %d, want 60", got)
	}
}

func TestOverloadPolicyOrdering(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		util       int
		wantLength int
	}{
		{name: "moderate load", util: 80, wantLength: 3},
		{name: "critical load", util: 92, wantLength: 4},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			policy := OverloadPolicy(c.util)
			if len(policy) != c.wantLength {
				t.Fatalf("policy len = %d, want %d", len(policy), c.wantLength)
			}
		})
	}
}

func TestRecoveryPlanSteps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		util int
		want string
	}{
		{name: "very high", util: 97, want: "throttle bulk sync and escalate"},
		{name: "high", util: 80, want: "throttle store-forward"},
		{name: "steady", util: 50, want: "steady"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := RecoveryPlan(c.util); got.NextStep != c.want {
				t.Fatalf("next step = %s, want %s", got.NextStep, c.want)
			}
		})
	}
}
