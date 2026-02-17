package conformance

import (
	"strings"
	"testing"
)

func TestGatesListAndSummary(t *testing.T) {
	gates := Gates()
	if len(gates) != 7 {
		t.Fatalf("expected 7 gates, got %d", len(gates))
	}
	if gates[0].ID != GateS801 {
		t.Fatalf("unexpected gate ordering: %s", gates[0].ID)
	}
	summary := GateSummary(gates)
	if summary != "v0.8 gates defined: 7" {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

func TestValidateChecklist(t *testing.T) {
	cases := []struct {
		name  string
		gates []Gate
		want  []string
	}{
		{"complete set", Gates(), nil},
		{"missing id", []Gate{{Description: "desc", Scope: "scope"}}, []string{"gate[0] missing ID"}},
		{"missing scope", []Gate{{ID: GateS801, Description: "desc"}}, []string{"S8-01 missing scope"}},
		{"missing description", []Gate{{ID: GateS801, Scope: "scope"}}, []string{"S8-01 missing description"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateChecklist(tc.gates)
			if len(errs) != len(tc.want) {
				t.Fatalf("expected %d errors, got %d: %v", len(tc.want), len(errs), errs)
			}
			for i, msg := range tc.want {
				if !strings.Contains(errs[i], msg) {
					t.Fatalf("expected error %q, got %q", msg, errs[i])
				}
			}
		})
	}
}
