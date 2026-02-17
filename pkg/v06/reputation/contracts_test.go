package reputation

import (
	"math"
	"reflect"
	"testing"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestReputationContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   ReputationContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "weight",
			contract:   NewReputationContract("S6-08", "T6-30", "rep-doc", conformance.GateV6G4, ReputationReasonWeight),
			wantAnchor: "rep-doc#T6-30",
			wantReason: string(ReputationReasonWeight),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.contract.EvidenceAnchor(); got != tt.wantAnchor {
				t.Fatalf("anchor mismatch: want %q got %q", tt.wantAnchor, got)
			}
			if got := tt.contract.ReasonLabel(); got != tt.wantReason {
				t.Fatalf("reason label mismatch: want %q got %q", tt.wantReason, got)
			}
		})
	}
}

func TestReputationReasonClassesDeterministic(t *testing.T) {
	want := []ReputationReasonClass{ReputationReasonWeight}
	got := ReputationReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestComputeReputationWeightBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		raw         float64
		trustPaths  int
		uncertainty float64
		wantWeight  float64
		wantConf    float64
		wantReason  string
	}{
		{name: "strong", raw: 0.8, trustPaths: 5, uncertainty: 0.2, wantWeight: 0.8, wantConf: 1, wantReason: "reputation.weight.success"},
		{name: "low-trust", raw: 1, trustPaths: 1, uncertainty: 0.1, wantWeight: 0.8, wantConf: 0.2, wantReason: "reputation.weight.failure"},
		{name: "uncertain", raw: 0.9, trustPaths: 3, uncertainty: 0.6, wantWeight: 0.72, wantConf: 0.6, wantReason: "reputation.weight.failure"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeReputationWeight(tt.raw, tt.trustPaths, tt.uncertainty)
			if got.Reason != tt.wantReason || math.Abs(got.Weight-tt.wantWeight) > 1e-9 || math.Abs(got.Confidence-tt.wantConf) > 1e-9 {
				t.Fatalf("reputation mismatch: want %s/%f/%f got %s/%f/%f", tt.wantReason, tt.wantWeight, tt.wantConf, got.Reason, got.Weight, got.Confidence)
			}
		})
	}
}
