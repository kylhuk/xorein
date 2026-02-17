package conformance

import (
	"strings"
	"testing"
)

func TestValidateChecklist(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		gateID      string
		checklist   Checklist
		wantReady   bool
		wantMissing []string
		wantErr     bool
	}{
		{
			name:      "ready gate",
			gateID:    "V9-G0",
			checklist: Checklist{"VA-G1": true, "VA-G2": true},
			wantReady: true,
		},
		{
			name:        "partial artifacts",
			gateID:      "V9-G2",
			checklist:   Checklist{"VA-L1": true},
			wantReady:   false,
			wantMissing: []string{"VA-L4", "VA-L7"},
		},
		{
			name:    "unknown gate",
			gateID:  "V9-UNKNOWN",
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := ValidateChecklist(c.gateID, c.checklist)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error for gate %s", c.gateID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Ready != c.wantReady {
				t.Fatalf("ready = %t, want %t", got.Ready, c.wantReady)
			}
			if len(got.Missing) != len(c.wantMissing) {
				t.Fatalf("missing artifacts = %v, want %v", got.Missing, c.wantMissing)
			}
			for i, artifact := range got.Missing {
				if artifact != c.wantMissing[i] {
					t.Fatalf("missing artifact[%d] = %s, want %s", i, artifact, c.wantMissing[i])
				}
			}
		})
	}
}

func TestSummariesAndSummaryText(t *testing.T) {
	t.Parallel()

	results := Summaries(Checklist{})
	if len(results) != len(gateCatalog) {
		t.Fatalf("Summaries len = %d, want %d", len(results), len(gateCatalog))
	}

	text := SummaryText(results)
	if !strings.Contains(text, "pending") {
		t.Fatalf("summary text lacks pending marker: %q", text)
	}
	if !strings.Contains(text, "VA-G") {
		t.Fatalf("summary text lacks artifact references: %q", text)
	}
}
