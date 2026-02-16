package conformance

import "testing"

func TestValidateTraceMatrix(t *testing.T) {
	trace := map[ScopeBullet]string{}
	for _, bullet := range allBullets {
		trace[bullet] = "P0-T1"
	}
	if err := ValidateTraceMatrix(trace); err != nil {
		t.Fatalf("validate trace: %v", err)
	}
	delete(trace, BulletDM)
	if err := ValidateTraceMatrix(trace); err == nil {
		t.Fatal("expected missing mapping error")
	}
}

func TestValidateRequirementMatrix(t *testing.T) {
	items := map[string]RequirementEvidence{
		"dm": {Positive: true, Negative: true},
	}
	if err := ValidateRequirementMatrix(items); err != nil {
		t.Fatalf("validate matrix: %v", err)
	}
	items["dm"] = RequirementEvidence{Positive: true, Negative: false}
	if err := ValidateRequirementMatrix(items); err == nil {
		t.Fatal("expected coverage error")
	}
}

func TestGateStatusComplete(t *testing.T) {
	status := GateStatus{P0: true, P1: true, P2: true, P3: true, P4: true, P5: true, P6: true, P7: true}
	if !status.Complete() {
		t.Fatal("expected complete")
	}
}
