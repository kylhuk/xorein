package phase8

import (
	"strings"
	"testing"
	"time"
)

func TestBuildValidationReport(t *testing.T) {
	t.Run("aggregates and sorts scenarios", func(t *testing.T) {
		report, err := BuildValidationReport([]ValidationScenario{
			{
				Name:                   "multi-peer-stability",
				Participants:           6,
				JoinEvents:             6,
				LeaveEvents:            2,
				ReconnectAttempts:      4,
				ReconnectSuccesses:     3,
				ControlOps:             10,
				ControlStateMismatches: 1,
				SessionCollapses:       0,
				Duration:               5 * time.Second,
			},
			{
				Name:                   "two-peer-baseline",
				Participants:           2,
				JoinEvents:             2,
				LeaveEvents:            1,
				ReconnectAttempts:      1,
				ReconnectSuccesses:     1,
				ControlOps:             4,
				ControlStateMismatches: 0,
				SessionCollapses:       0,
				Duration:               3 * time.Second,
			},
		})
		if err != nil {
			t.Fatalf("BuildValidationReport() error = %v", err)
		}

		if got, want := report.TotalScenarios, 2; got != want {
			t.Fatalf("TotalScenarios = %d, want %d", got, want)
		}
		if got, want := report.TotalReconnectAttempts, 5; got != want {
			t.Fatalf("TotalReconnectAttempts = %d, want %d", got, want)
		}
		if got, want := report.TotalReconnectSuccesses, 4; got != want {
			t.Fatalf("TotalReconnectSuccesses = %d, want %d", got, want)
		}
		if got, want := report.TotalControlMismatches, 1; got != want {
			t.Fatalf("TotalControlMismatches = %d, want %d", got, want)
		}
		if got, want := report.TotalSessionCollapses, 0; got != want {
			t.Fatalf("TotalSessionCollapses = %d, want %d", got, want)
		}
		if got, want := report.LongestScenario, 5*time.Second; got != want {
			t.Fatalf("LongestScenario = %v, want %v", got, want)
		}
		if got, want := report.AverageScenario, 4*time.Second; got != want {
			t.Fatalf("AverageScenario = %v, want %v", got, want)
		}
		if got, want := report.Scenarios[0].Name, "multi-peer-stability"; got != want {
			t.Fatalf("Scenarios[0].Name = %q, want %q", got, want)
		}
		if got, want := report.Scenarios[1].Name, "two-peer-baseline"; got != want {
			t.Fatalf("Scenarios[1].Name = %q, want %q", got, want)
		}
	})

	t.Run("validation failures", func(t *testing.T) {
		tests := []struct {
			name      string
			scenarios []ValidationScenario
			wantErr   string
		}{
			{
				name:      "empty scenario list",
				scenarios: nil,
				wantErr:   "validation scenarios required",
			},
			{
				name: "missing name",
				scenarios: []ValidationScenario{{
					Participants: 2,
					Duration:     time.Second,
				}},
				wantErr: "name required",
			},
			{
				name: "participants out of range",
				scenarios: []ValidationScenario{{
					Name:         "bad-peers",
					Participants: 9,
					Duration:     time.Second,
				}},
				wantErr: "participants out of range",
			},
			{
				name: "reconnect successes exceed attempts",
				scenarios: []ValidationScenario{{
					Name:               "bad-reconnect",
					Participants:       2,
					ReconnectAttempts:  1,
					ReconnectSuccesses: 2,
					Duration:           time.Second,
				}},
				wantErr: "successes exceed attempts",
			},
			{
				name: "non-positive duration",
				scenarios: []ValidationScenario{{
					Name:         "bad-duration",
					Participants: 2,
					Duration:     0,
				}},
				wantErr: "duration must be > 0",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := BuildValidationReport(tc.scenarios)
				if err == nil {
					t.Fatalf("BuildValidationReport() error = nil, want contains %q", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("BuildValidationReport() error = %q, want contains %q", err.Error(), tc.wantErr)
				}
			})
		}
	})
}
