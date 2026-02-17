package relay

import "testing"

func TestReliabilityAndAbuseContracts(t *testing.T) {
	t.Parallel()

	scores := ReliabilityScores()
	if scores["bootstrap-node-a"] < 90 {
		t.Fatalf("bootstrap node score too low: %d", scores["bootstrap-node-a"])
	}

	abuse := AbuseResponseClass()
	if abuse["high"] != "immediate review" {
		t.Fatalf("unexpected high severity response %q", abuse["high"])
	}
}
