package conformance

import (
	"math"
	"reflect"
	"testing"
)

func TestGateChecklistRequiredScopes(t *testing.T) {
	tests := []struct {
		name string
		gate GateID
		want []string
	}{
		{name: "gate0", gate: GateV6G0, want: []string{"S6-01", "S6-02"}},
		{name: "gate3", gate: GateV6G3, want: []string{"S6-05", "S6-06", "S6-07"}},
		{name: "gate6", gate: GateV6G6, want: []string{"S6-02", "S6-04", "S6-07", "S6-10"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewGateChecklist(tt.gate).RequiredScopes()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}

			// mutate the returned slice and ensure the underlying map remains untouched
			if len(got) > 0 {
				got[0] = "mutated"
			}
			again := NewGateChecklist(tt.gate).RequiredScopes()
			if !reflect.DeepEqual(again, tt.want) {
				t.Fatalf("slice copy mutated original; want %v, got %v", tt.want, again)
			}
		})
	}
}

func TestGateChecklistSatisfactionAndCoverage(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() GateChecklist
		wantSatisfied bool
		wantCoverage  float64
	}{
		{
			name: "not completed",
			setup: func() GateChecklist {
				g := NewGateChecklist(GateV6G0)
				g.RecordEvidence("S6-01", "doc")
				return g
			},
			wantSatisfied: false,
			wantCoverage:  0.5,
		},
		{
			name: "missing scopes",
			setup: func() GateChecklist {
				g := NewGateChecklist(GateV6G0)
				g.Completed = true
				g.RecordEvidence("S6-01", "doc")
				return g
			},
			wantSatisfied: false,
			wantCoverage:  0.5,
		},
		{
			name: "all evidence",
			setup: func() GateChecklist {
				g := NewGateChecklist(GateV6G0)
				g.Completed = true
				g.RecordEvidence("S6-01", "doc1")
				g.RecordEvidence("S6-02", "doc2")
				return g
			},
			wantSatisfied: true,
			wantCoverage:  1.0,
		},
		{
			name: "idempotent record",
			setup: func() GateChecklist {
				g := NewGateChecklist(GateV6G0)
				g.Completed = true
				g.RecordEvidence("S6-01", "doc1")
				g.RecordEvidence("S6-01", "doc2")
				return g
			},
			wantSatisfied: false,
			wantCoverage:  0.5,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := tt.setup()
			if got := g.IsSatisfied(); got != tt.wantSatisfied {
				t.Fatalf("expected satisfied=%v, got %v", tt.wantSatisfied, got)
			}
			if got := GateCoverageScore(g); math.Abs(got-tt.wantCoverage) > 1e-9 {
				t.Fatalf("coverage mismatch: want %v got %v", tt.wantCoverage, got)
			}
		})
	}
}
