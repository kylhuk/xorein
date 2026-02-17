package conformance

import "fmt"

// Gate represents a review gate for v0.9. RequiredArtifacts lists the VA IDs tracked per gate.
type Gate struct {
	ID                string
	Name              string
	Description       string
	RequiredArtifacts []string
}

// Checklist maps VA IDs to completion state.
type Checklist map[string]bool

// ValidationResult reports if a gate is ready and which artifacts are missing.
type ValidationResult struct {
	Gate    Gate
	Missing []string
	Ready   bool
}

var gateCatalog = []Gate{
	{ID: "V9-G0", Name: "Scope & Guardrails", Description: "Scope lock and evidence taxonomy", RequiredArtifacts: []string{"VA-G1", "VA-G2"}},
	{ID: "V9-G1", Name: "IPFS Contract Freeze", Description: "Persistent-hosting contracts", RequiredArtifacts: []string{"VA-I1", "VA-I4"}},
	{ID: "V9-G2", Name: "Large Server Contract Freeze", Description: "Scale fanout and lazy loading", RequiredArtifacts: []string{"VA-L1", "VA-L4", "VA-L7"}},
	{ID: "V9-G3", Name: "SFU Mesh Contract", Description: "Cascading SFU topology", RequiredArtifacts: []string{"VA-S1", "VA-S4"}},
	{ID: "V9-G4", Name: "Profiling Contract", Description: "Profiling/optimization taxonomy", RequiredArtifacts: []string{"VA-P1", "VA-P4"}},
	{ID: "V9-G5", Name: "Stress Test Contract", Description: "Stress campaign definitions", RequiredArtifacts: []string{"VA-T1", "VA-T7"}},
	{ID: "V9-G6", Name: "Relay Contract", Description: "Relay load & recovery", RequiredArtifacts: []string{"VA-R1", "VA-R3"}},
	{ID: "V9-G7", Name: "Mobile Contract", Description: "Battery optimization", RequiredArtifacts: []string{"VA-B1", "VA-B7"}},
	{ID: "V9-G8", Name: "Integrated Validation", Description: "Governance readiness and validation matrix", RequiredArtifacts: []string{"VA-X1", "VA-X8"}},
	{ID: "V9-G9", Name: "Release Handoff", Description: "Release conformance and handoff", RequiredArtifacts: []string{"VA-H1", "VA-H2"}},
}

// Catalog returns the list of gates.
func Catalog() []Gate {
	return gateCatalog
}

// GateByID looks up a gate by ID.
func GateByID(id string) (Gate, bool) {
	for _, gate := range gateCatalog {
		if gate.ID == id {
			return gate, true
		}
	}
	return Gate{}, false
}

// Summaries returns validation status for every gate given a checklist.
func Summaries(checklist Checklist) []ValidationResult {
	var results []ValidationResult
	for _, gate := range gateCatalog {
		result, _ := ValidateChecklist(gate.ID, checklist)
		results = append(results, result)
	}
	return results
}

// ValidateChecklist inspects an individual gate.
func ValidateChecklist(gateID string, checklist Checklist) (ValidationResult, error) {
	gate, ok := GateByID(gateID)
	if !ok {
		return ValidationResult{}, fmt.Errorf("unknown gate %q", gateID)
	}
	var missing []string
	for _, artifact := range gate.RequiredArtifacts {
		if !checklist[artifact] {
			missing = append(missing, artifact)
		}
	}
	return ValidationResult{Gate: gate, Missing: missing, Ready: len(missing) == 0}, nil
}

// SummaryText renders a human-friendly summary for gate reviews.
func SummaryText(results []ValidationResult) string {
	text := ""
	for _, result := range results {
		status := "pending"
		if result.Ready {
			status = "ready"
		}
		text += fmt.Sprintf("%s (%s): %s\n", result.Gate.ID, status, result.Gate.Description)
		if len(result.Missing) > 0 {
			text += fmt.Sprintf("  missing artifacts: %v\n", result.Missing)
		}
	}
	return text
}
