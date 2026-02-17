package governance

import (
	"testing"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestReleaseChecklistCompletion(t *testing.T) {
	tests := []struct {
		name         string
		marks        []string
		extraMarks   []string
		wantComplete bool
	}{
		{name: "none", wantComplete: false},
		{name: "partial", marks: []string{ReleaseChecklistItems[0]}, wantComplete: false},
		{name: "all", marks: ReleaseChecklistItems, wantComplete: true},
		{name: "all plus extra", marks: ReleaseChecklistItems, extraMarks: []string{"unused"}, wantComplete: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewReleaseChecklist()
			for _, item := range tt.marks {
				r.Mark(item)
			}
			for _, item := range tt.extraMarks {
				r.Mark(item)
			}
			if got := r.IsComplete(); got != tt.wantComplete {
				t.Fatalf("expected complete=%v, got %v", tt.wantComplete, got)
			}
		})
	}
}

func TestOpenDecisionRecordIsOpen(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "open", status: "Open", want: true},
		{name: "closed", status: "Closed", want: false},
		{name: "empty", status: "", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			record := OpenDecisionRecord{Status: tt.status}
			if got := record.IsOpen(); got != tt.want {
				t.Fatalf("status=%q, want %v got %v", tt.status, tt.want, got)
			}
		})
	}
}

func TestReleaseMetadataRecordsGateEvidenceAndDecisions(t *testing.T) {
	meta := NewReleaseMetadata()
	detail := "evidence-bridge"
	meta.RecordGateEvidence(conformance.GateV6G2, detail)
	if got := meta.GateEvidence[conformance.GateV6G2]; got != detail {
		t.Fatalf("expected gate detail stored")
	}

	decision := OpenDecisionRecord{ID: OpenDecisionOD601, Gate: conformance.GateV6G0, Status: "Open", Notes: "note"}
	meta.AddDecision(decision)
	if len(meta.Decisions) != 1 {
		t.Fatalf("expected decision appended")
	}
	if meta.Decisions[0] != decision {
		t.Fatalf("decision identity preserved")
	}
}
