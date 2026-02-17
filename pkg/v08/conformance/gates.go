package conformance

import "fmt"

// GateID identifies a v0.8 scope gate.
type GateID string

const (
	GateS801 GateID = "S8-01"
	GateS802 GateID = "S8-02"
	GateS803 GateID = "S8-03"
	GateS804 GateID = "S8-04"
	GateS805 GateID = "S8-05"
	GateS806 GateID = "S8-06"
	GateS807 GateID = "S8-07"
)

// Gate records the deterministic scope intent for a v0.8 gate.
type Gate struct {
	ID          GateID
	Description string
	Scope       string
	Owner       string
}

// Gates returns the scoped gate definitions for v0.8 execution.
func Gates() []Gate {
	return []Gate{
		{ID: GateS801, Description: "Deterministic contract helpers", Scope: "pkg/v08/*", Owner: "V8-G1"},
		{ID: GateS802, Description: "Minimal scenario hook for contract witness", Scope: "cmd/aether main.cases/v08-echo", Owner: "V8-G1"},
		{ID: GateS803, Description: "Bookmark privacy lifecycle semantics", Scope: "pkg/v08/bookmarks", Owner: "V8-G2"},
		{ID: GateS804, Description: "Link preview metadata precedence and normalization", Scope: "pkg/v08/linkpreview", Owner: "V8-G2"},
		{ID: GateS805, Description: "Theme/token validation with deterministic fallback", Scope: "pkg/v08/themes", Owner: "V8-G3"},
		{ID: GateS806, Description: "Accessibility role/announcement contracts", Scope: "pkg/v08/accessibility", Owner: "V8-G3"},
		{ID: GateS807, Description: "Voice noise suppression policy and DTLN fallback", Scope: "pkg/v08/voice", Owner: "V8-G4"},
	}
}

// ValidateChecklist returns every gate that lacks essential metadata.
func ValidateChecklist(gates []Gate) []string {
	var errs []string
	for i, gate := range gates {
		if gate.ID == "" {
			errs = append(errs, fmt.Sprintf("gate[%d] missing ID", i))
			continue
		}
		if gate.Description == "" {
			errs = append(errs, fmt.Sprintf("%s missing description", gate.ID))
		}
		if gate.Scope == "" {
			errs = append(errs, fmt.Sprintf("%s missing scope", gate.ID))
		}
	}
	return errs
}

// GateSummary returns a human-readable summary of the gate set.
func GateSummary(gates []Gate) string {
	return fmt.Sprintf("v0.8 gates defined: %d", len(gates))
}
