package hardening

import "testing"

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy(ContinuitySteady)
	if policy.MaxRestarts != 1 || policy.MaxDowntimeSec != 30 {
		t.Fatalf("expected steady limits, got %+v", policy)
	}

	resilient := DefaultPolicy(ContinuityResilient)
	if resilient.MaxRestarts != 3 || resilient.MaxDowntimeSec != 120 {
		t.Fatalf("expected resilient limits, got %+v", resilient)
	}
}

func TestContinuityReportIssues(t *testing.T) {
	report := ContinuityReport{
		Policy: DefaultPolicy(ContinuitySteady),
		Observations: []ContinuityObservation{
			{Description: "cold-start", Restarts: 1, DowntimeSec: 10},
			{Description: "panic-restart", Restarts: 2, DowntimeSec: 40},
		},
	}

	if report.IsCompliant() {
		t.Fatalf("expected non-compliant report")
	}
	issues := report.Issues()
	if len(issues) != 1 {
		t.Fatalf("expected single issue, got %d", len(issues))
	}
}

func TestContinuityReportPasses(t *testing.T) {
	report := ContinuityReport{
		Policy: DefaultPolicy(ContinuityResilient),
		Observations: []ContinuityObservation{
			{Description: "blue-green", Restarts: 2, DowntimeSec: 100},
		},
	}

	if !report.IsCompliant() {
		t.Fatalf("expected compliant report")
	}
}
