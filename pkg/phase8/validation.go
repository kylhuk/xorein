package phase8

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ValidationScenario captures one deterministic voice validation execution sample.
type ValidationScenario struct {
	Name                   string
	Participants           int
	JoinEvents             int
	LeaveEvents            int
	ReconnectAttempts      int
	ReconnectSuccesses     int
	ControlOps             int
	ControlStateMismatches int
	SessionCollapses       int
	Duration               time.Duration
}

// ValidationReport is a compact aggregate of voice quality/stability checks.
type ValidationReport struct {
	Scenarios               []ValidationScenario
	TotalScenarios          int
	TotalReconnectAttempts  int
	TotalReconnectSuccesses int
	TotalSessionCollapses   int
	TotalControlMismatches  int
	LongestScenario         time.Duration
	AverageScenario         time.Duration
}

// BuildValidationReport aggregates deterministic scenario metrics for P8-T6 checks.
func BuildValidationReport(scenarios []ValidationScenario) (ValidationReport, error) {
	if len(scenarios) == 0 {
		return ValidationReport{}, fmt.Errorf("phase8: validation scenarios required")
	}

	report := ValidationReport{Scenarios: make([]ValidationScenario, 0, len(scenarios))}
	var totalDuration time.Duration

	for _, scenario := range scenarios {
		if err := validateScenario(scenario); err != nil {
			return ValidationReport{}, err
		}
		report.Scenarios = append(report.Scenarios, scenario)
		report.TotalReconnectAttempts += scenario.ReconnectAttempts
		report.TotalReconnectSuccesses += scenario.ReconnectSuccesses
		report.TotalSessionCollapses += scenario.SessionCollapses
		report.TotalControlMismatches += scenario.ControlStateMismatches
		if scenario.Duration > report.LongestScenario {
			report.LongestScenario = scenario.Duration
		}
		totalDuration += scenario.Duration
	}

	sort.Slice(report.Scenarios, func(i, j int) bool {
		return strings.Compare(report.Scenarios[i].Name, report.Scenarios[j].Name) < 0
	})

	report.TotalScenarios = len(report.Scenarios)
	report.AverageScenario = time.Duration(int64(totalDuration) / int64(report.TotalScenarios))
	return report, nil
}

func validateScenario(s ValidationScenario) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("phase8: validation scenario name required")
	}
	if s.Participants < 2 || s.Participants > 8 {
		return fmt.Errorf("phase8: participants out of range for scenario %q: %d", s.Name, s.Participants)
	}
	if s.JoinEvents < 0 || s.LeaveEvents < 0 || s.ReconnectAttempts < 0 || s.ReconnectSuccesses < 0 || s.ControlOps < 0 || s.ControlStateMismatches < 0 || s.SessionCollapses < 0 {
		return fmt.Errorf("phase8: negative counters in scenario %q", s.Name)
	}
	if s.ReconnectSuccesses > s.ReconnectAttempts {
		return fmt.Errorf("phase8: reconnect successes exceed attempts in scenario %q", s.Name)
	}
	if s.Duration <= 0 {
		return fmt.Errorf("phase8: duration must be > 0 in scenario %q", s.Name)
	}
	return nil
}
