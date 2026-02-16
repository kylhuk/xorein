package governance

import "sort"

type GateID string

const (
	GateV3G0 GateID = "V3-G0"
	GateV3G1 GateID = "V3-G1"
	GateV3G2 GateID = "V3-G2"
	GateV3G3 GateID = "V3-G3"
	GateV3G4 GateID = "V3-G4"
	GateV3G5 GateID = "V3-G5"
	GateV3G6 GateID = "V3-G6"
)

type GateEvidence struct {
	Gate      GateID
	Artifacts []string
	Completed bool
	Notes     string
}

func (e GateEvidence) Normalized() GateEvidence {
	result := GateEvidence{
		Gate:      e.Gate,
		Artifacts: append([]string(nil), e.Artifacts...),
		Completed: e.Completed,
		Notes:     e.Notes,
	}
	sort.Strings(result.Artifacts)
	return result
}

func GateEvidenceFor(id GateID, artifacts []string, completed bool, notes string) GateEvidence {
	return GateEvidence{Gate: id, Artifacts: artifacts, Completed: completed, Notes: notes}
}

var RequiredGates = []GateID{GateV3G0, GateV3G1, GateV3G2, GateV3G3, GateV3G4, GateV3G5, GateV3G6}

type OpenDecisionReminder struct {
	DecisionID string
	Status     string
	Notes      string
}

func (d OpenDecisionReminder) IsOpen() bool {
	return d.Status == "Open"
}
