package v20

import "testing"

type sloMetric struct {
	name     string
	target   float64
	observed float64
}

func (s sloMetric) meets() bool {
	return s.observed >= s.target
}

func TestSLOScorecard(t *testing.T) {
	metrics := []sloMetric{
		{name: "login_success_rate", target: 99.5, observed: 99.72},
		{name: "call_connectivity_score", target: 92.0, observed: 94.5},
		{name: "relay_stability", target: 99.9, observed: 99.95},
	}

	for _, m := range metrics {
		if !m.meets() {
			t.Fatalf("metric %s failed to meet target", m.name)
		}
	}
}
