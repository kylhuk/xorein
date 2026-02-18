package v20

import (
	"testing"

	"github.com/aether/code_aether/pkg/v20/hardening"
)

func TestRecoveryPathContinuity(t *testing.T) {
	report := hardening.ContinuityReport{
		Policy: hardening.DefaultPolicy(hardening.ContinuityResilient),
		Observations: []hardening.ContinuityObservation{
			{Description: "retry-loop", Restarts: 2, DowntimeSec: 60},
			{Description: "warm-standby", Restarts: 1, DowntimeSec: 20},
		},
	}

	if !report.IsCompliant() {
		t.Fatalf("expected resilient policy to accept measured recovery paths, got %v", report.Issues())
	}
}

func TestRecoveryPathFailsWhenDowntimeExceeds(t *testing.T) {
	report := hardening.ContinuityReport{
		Policy: hardening.DefaultPolicy(hardening.ContinuitySteady),
		Observations: []hardening.ContinuityObservation{
			{Description: "network-blip", Restarts: 1, DowntimeSec: 90},
		},
	}

	if report.IsCompliant() {
		t.Fatalf("expected steady policy to flag downtime breaches")
	}
}
