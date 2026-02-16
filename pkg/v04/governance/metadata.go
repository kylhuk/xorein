package governance

type GateID string

const (
	GateV4G0 GateID = "V4-G0"
	GateV4G1 GateID = "V4-G1"
	GateV4G2 GateID = "V4-G2"
	GateV4G3 GateID = "V4-G3"
	GateV4G4 GateID = "V4-G4"
	GateV4G5 GateID = "V4-G5"
	GateV4G6 GateID = "V4-G6"
)

var RequiredGates = []GateID{GateV4G0, GateV4G1, GateV4G2, GateV4G3, GateV4G4, GateV4G5, GateV4G6}

var ReleaseChecklistItems = []string{
	"All seven v0.4 scope bullets are mapped to tasks and artifacts.",
	"No earlier-version baseline capability is re-introduced as first-introduction scope in v0.4.",
	"No v0.6 hardening/scaling scope is imported.",
	"Role/override, policy versioning, and auto-mod contracts are deterministic and test-mapped.",
	"Audit traceability includes signed policy linkage and authorized visibility rules.",
	"Compatibility/governance/open-decision checks are complete.",
	"Planned-vs-implemented distinction remains explicit.",
}

type ReleaseChecklist struct {
	Items map[string]bool
}

func NewReleaseChecklist() *ReleaseChecklist {
	items := make(map[string]bool, len(ReleaseChecklistItems))
	for _, item := range ReleaseChecklistItems {
		items[item] = false
	}
	return &ReleaseChecklist{Items: items}
}

func (r *ReleaseChecklist) Mark(item string) {
	if _, ok := r.Items[item]; ok {
		r.Items[item] = true
	}
}

func (r ReleaseChecklist) IsComplete() bool {
	for _, done := range r.Items {
		if !done {
			return false
		}
	}
	return true
}

type OpenDecisionReminder struct {
	DecisionID string
	Status     string
	Notes      string
	Gate       GateID
}

func (d OpenDecisionReminder) IsOpen() bool {
	return d.Status == "Open"
}
