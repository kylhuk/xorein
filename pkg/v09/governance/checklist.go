package governance

// ChecklistItem describes a governance requirement.
type ChecklistItem struct {
	ID          string
	Description string
	Completed   bool
}

// AdditiveChecklist lists in-scope additive requirements.
func AdditiveChecklist() []ChecklistItem {
	return []ChecklistItem{
		{ID: "GOV-ADD-01", Description: "All schema changes must add new fields only"},
		{ID: "GOV-ADD-02", Description: "Reserved fields are recorded for future reuse"},
		{ID: "GOV-ADD-03", Description: "Compatibility defaults documented"},
	}
}

// MajorPathTriggerClassifier explains when a breaking change escalates.
func MajorPathTriggerClassifier(breaking bool, context string) string {
	if !breaking {
		return "additive path"
	}
	switch context {
	case "multistream":
		return "requires new multistream ID + AEP"
	case "downgrade":
		return "requires downgrade negotiation + multi-impl proof"
	default:
		return "major-path treatment"
	}
}

// LicensingStatus returns the license narrative for traceability.
func LicensingStatus(codeLicense, specLicense string) string {
	return codeLicense + " / " + specLicense
}
