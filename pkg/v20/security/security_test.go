package security

import "testing"

func TestComplianceScore(t *testing.T) {
	profile := HardeningProfile{
		Name: "release",
		Controls: []HardeningCheck{
			{Name: "crypto", Passed: true, Severity: SeverityHigh},
			{Name: "fips", Passed: false, Severity: SeverityCritical},
		},
	}

	if score := profile.ComplianceScore(); score != 50 {
		t.Fatalf("expected 50 got %d", score)
	}
}

func TestCriticalFailures(t *testing.T) {
	profile := HardeningProfile{
		Name: "relay",
		Controls: []HardeningCheck{
			{Name: "policy", Passed: false, Severity: SeverityHigh},
			{Name: "monitoring", Passed: false, Severity: SeverityMedium},
		},
	}

	failures := profile.CriticalFailures()
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
	if failures[0].Name != "policy" {
		t.Fatalf("expected policy failure")
	}
}

func TestEvaluatePassState(t *testing.T) {
	profile := HardeningProfile{Name: "empty"}
	result := Evaluate(profile)
	if !result.Pass {
		t.Fatalf("expected pass")
	}
	if result.Score != 100 {
		t.Fatalf("expected score 100 got %d", result.Score)
	}
	if result.Message != "no-controls" {
		t.Fatalf("unexpected message %q", result.Message)
	}
}
