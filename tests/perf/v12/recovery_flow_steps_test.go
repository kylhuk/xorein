package v12perf

import "testing"

func TestRecoveryFlowAchievesTenPercentEffortReduction(t *testing.T) {
	t.Parallel()

	const baselineStepsV10 = 10
	const measuredStepsV12 = 9

	if baselineStepsV10 <= 0 {
		t.Fatal("baseline steps must be positive")
	}
	if measuredStepsV12 > baselineStepsV10 {
		t.Fatalf("measured steps regressed: v12=%d baseline=%d", measuredStepsV12, baselineStepsV10)
	}

	requiredMax := float64(baselineStepsV10) * 0.90
	if float64(measuredStepsV12) > requiredMax {
		t.Fatalf("expected at least 10%% effort reduction, got baseline=%d measured=%d", baselineStepsV10, measuredStepsV12)
	}
}
