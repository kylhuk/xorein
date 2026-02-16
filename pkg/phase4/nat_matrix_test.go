package phase4

import (
	"reflect"
	"testing"
)

func TestDefaultNATScenarioMatrix_ValidationAndHighRiskMitigations(t *testing.T) {
	t.Parallel()

	matrix := DefaultNATScenarioMatrix()
	if err := ValidateNATScenarioMatrix(matrix); err != nil {
		t.Fatalf("ValidateNATScenarioMatrix(default) returned error: %v", err)
	}

	if len(matrix) == 0 {
		t.Fatal("DefaultNATScenarioMatrix returned empty matrix")
	}

	highRiskWithMitigations := 0
	for _, scenario := range matrix {
		if !scenario.HighRisk {
			continue
		}
		highRiskWithMitigations++
		if len(scenario.Mitigations) == 0 {
			t.Fatalf("high-risk scenario %q has no mitigations", scenario.ID)
		}
		for _, mitigation := range scenario.Mitigations {
			if mitigation.Diagnostic == "" {
				t.Fatalf("scenario %q mitigation %q has empty diagnostic reason code", scenario.ID, mitigation.ID)
			}
			if mitigation.ValidationRef == "" {
				t.Fatalf("scenario %q mitigation %q has empty validation reference", scenario.ID, mitigation.ID)
			}
		}
	}

	if highRiskWithMitigations == 0 {
		t.Fatal("expected at least one high-risk scenario with mitigations")
	}
}

func TestNATScenarioByID_ReturnsExpectedScenario(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantOK  bool
		wantSeq []TraversalStage
	}{
		{
			name:    "known scenario",
			id:      "nat-symmetric-hard-fallback",
			wantOK:  true,
			wantSeq: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch, TraversalStageRelay},
		},
		{
			name:   "unknown scenario",
			id:     "not-present",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, ok := NATScenarioByID(tt.id)
			if ok != tt.wantOK {
				t.Fatalf("NATScenarioByID(%q) ok = %t, want %t", tt.id, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if !reflect.DeepEqual(scenario.TraversalSequence, tt.wantSeq) {
				t.Fatalf("NATScenarioByID(%q) traversal sequence = %v, want %v", tt.id, scenario.TraversalSequence, tt.wantSeq)
			}
		})
	}
}

func TestValidateNATScenarioMatrix_FailureCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		matrix []NATScenarioEntry
	}{
		{
			name:   "empty matrix",
			matrix: nil,
		},
		{
			name: "missing traversal sequence",
			matrix: []NATScenarioEntry{{
				ID:         "missing-sequence",
				Topology:   NATTopologySymmetric,
				Conditions: "strict mapping",
			}},
		},
		{
			name: "sequence must start direct",
			matrix: []NATScenarioEntry{{
				ID:                "bad-order",
				Topology:          NATTopologySymmetric,
				Conditions:        "strict mapping",
				TraversalSequence: []TraversalStage{TraversalStageAutoNAT, TraversalStageRelay},
			}},
		},
		{
			name: "high risk missing mitigation metadata",
			matrix: []NATScenarioEntry{{
				ID:                "missing-metadata",
				Topology:          NATTopologyCarrierGradeNAT,
				Conditions:        "carrier nat",
				TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageRelay},
				HighRisk:          true,
				RiskReason:        "relay dependency",
				Mitigations: []NATMitigation{{
					ID:         "m1",
					Priority:   MitigationPriorityP0,
					Action:     "do thing",
					Diagnostic: ReasonRelayFallbackActive,
				}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateNATScenarioMatrix(tt.matrix); err == nil {
				t.Fatalf("ValidateNATScenarioMatrix(%s) expected error, got nil", tt.name)
			}
		})
	}
}
