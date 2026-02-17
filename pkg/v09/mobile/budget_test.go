package mobile

import "testing"

func TestBackgroundBudgetDeterminism(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cpu     int
		battery int
		want    BudgetClass
	}{
		{name: "critical battery", cpu: 20, battery: 10, want: BudgetCritical},
		{name: "high cpu remaining", cpu: 60, battery: 50, want: BudgetInteractive},
		{name: "background stay", cpu: 10, battery: 90, want: BudgetBackground},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := BackgroundBudget(c.cpu, c.battery); got != c.want {
				t.Fatalf("budget = %s, want %s", got, c.want)
			}
		})
	}
}

func TestEvaluateWakePolicy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		battery             int
		pendingHighPriority bool
		wantAllowAndReason  WakePolicy
	}{
		{name: "high priority", battery: 50, pendingHighPriority: true, wantAllowAndReason: WakePolicy{AllowWake: true, SuppressionReason: "high-priority"}},
		{name: "low battery", battery: 10, pendingHighPriority: false, wantAllowAndReason: WakePolicy{AllowWake: false, SuppressionReason: "low battery"}},
		{name: "budget ok", battery: 60, pendingHighPriority: false, wantAllowAndReason: WakePolicy{AllowWake: true, SuppressionReason: "budget available"}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := EvaluateWakePolicy(c.battery, c.pendingHighPriority); got != c.wantAllowAndReason {
				t.Fatalf("wake policy = %+v, want %+v", got, c.wantAllowAndReason)
			}
		})
	}
}

func TestBatteryOptimizationDecision(t *testing.T) {
	t.Parallel()

	if got := BatteryOptimizationDecision(60, false); got != "suppress syncs" {
		t.Fatalf("decision = %q, want suppress syncs", got)
	}
	if got := BatteryOptimizationDecision(80, true); got != "reduce render quality" {
		t.Fatalf("decision = %q, want reduce render quality", got)
	}
	if got := BatteryOptimizationDecision(30, true); got != "stay active" {
		t.Fatalf("decision = %q, want stay active", got)
	}
}
