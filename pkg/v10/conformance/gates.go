package conformance

import "fmt"

// Gate represents a release gate for v1.0.
type Gate struct {
	ID                string
	Name              string
	Description       string
	RequiredArtifacts []string
}

// Checklist maps artifact IDs to completion state.
type Checklist map[string]bool

// ValidationResult reports gate readiness and missing artifacts.
type ValidationResult struct {
	Gate    Gate
	Missing []string
	Ready   bool
}

var gateCatalog = []Gate{
	{ID: "V10-G0", Name: "Scope & Evidence", Description: "Scope, governance, and evidence schema", RequiredArtifacts: []string{"VA-G1", "VA-G2", "VA-G3", "VA-G4", "VA-G5", "VA-G6"}},
	{ID: "V10-G1", Name: "Protocol Freeze", Description: "Protocol surface, compatibility, and governance readiness", RequiredArtifacts: []string{"VA-P1", "VA-P2", "VA-P3", "VA-P4", "VA-P5", "VA-P6"}},
	{ID: "V10-G2", Name: "Security Audit", Description: "Audit scope, threat, and engagement governance", RequiredArtifacts: []string{"VA-S1", "VA-S2", "VA-S3", "VA-S4", "VA-S5", "VA-S6"}},
	{ID: "V10-G3", Name: "Spec Publication", Description: "Spec architecture, terminology, and publication workflow", RequiredArtifacts: []string{"VA-P7", "VA-P8", "VA-P9", "VA-P10", "VA-P11", "VA-P12"}},
	{ID: "V10-G4", Name: "Docs Publication", Description: "Documentation quality + publication control", RequiredArtifacts: []string{"VA-D1", "VA-D2", "VA-D3", "VA-D4", "VA-D5", "VA-D6"}},
	{ID: "V10-G5", Name: "Landing Surface", Description: "Landing and comparison surface governance", RequiredArtifacts: []string{"VA-W1", "VA-W2", "VA-W3"}},
	{ID: "V10-G6", Name: "Client Distribution", Description: "App store dossiers and compliance mapping", RequiredArtifacts: []string{"VA-A1", "VA-A2", "VA-A3"}},
	{ID: "V10-G7", Name: "Bootstrap Expansion", Description: "Bootstrap topology, monitoring, operator continuity", RequiredArtifacts: []string{"VA-N1", "VA-N2", "VA-N3"}},
	{ID: "V10-G8", Name: "Relay Program", Description: "Relay policy, onboarding, abuse response", RequiredArtifacts: []string{"VA-N4", "VA-N5", "VA-N6"}},
	{ID: "V10-G9", Name: "Repro Build", Description: "Deterministic builds, signing, container evidence", RequiredArtifacts: []string{"VA-R1", "VA-R2", "VA-B1", "VA-B7"}},
	{ID: "V10-G10", Name: "Integrated Go/No-Go", Description: "Integrated validation + go/no-go handoff", RequiredArtifacts: []string{"VA-X1", "VA-X8", "VA-H1", "VA-H2"}},
}

// Catalog returns the gates for v1.0.
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

// Summaries returns validation results for every gate.
func Summaries(checklist Checklist) []ValidationResult {
	var results []ValidationResult
	for _, gate := range gateCatalog {
		result, _ := ValidateChecklist(gate.ID, checklist)
		results = append(results, result)
	}
	return results
}

// ValidateChecklist inspects an individual gate using the checklist.
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

// SummaryText renders a human-friendly validation summary.
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
