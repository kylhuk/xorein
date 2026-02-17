package governance

import "github.com/aether/code_aether/pkg/v07/conformance"

type OpenDecisionID string

const (
	OpenDecisionOD701 OpenDecisionID = "OD7-01"
	OpenDecisionOD702 OpenDecisionID = "OD7-02"
	OpenDecisionOD703 OpenDecisionID = "OD7-03"
	OpenDecisionOD704 OpenDecisionID = "OD7-04"
)

var OpenDecisions = []OpenDecisionID{
	OpenDecisionOD701,
	OpenDecisionOD702,
	OpenDecisionOD703,
	OpenDecisionOD704,
}

type ReleaseChecklist struct {
	Items map[string]bool
}

var ReleaseChecklistItems = []string{
	"Map v0.7 compliance scope to deterministic contracts",
	"Surface protocol compatibility checks for all gates",
	"Capture retention/storeforward boundary semantics",
	"Document archivist enrollment/withdrawal posture",
	"Describe history sync/resume expectations",
	"Keep authorization scope filters documented",
	"Require notification dedupe/action fallback rationale",
	"Validate push-relay packaging and CLI dispatch",
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

type OpenDecisionRecord struct {
	ID     OpenDecisionID
	Gate   conformance.GateID
	Status string
	Notes  string
}

func (d OpenDecisionRecord) IsOpen() bool {
	return d.Status == "Open"
}

type ReleaseMetadata struct {
	Checklist    *ReleaseChecklist
	GateEvidence map[conformance.GateID]string
	Decisions    []OpenDecisionRecord
}

func NewReleaseMetadata() *ReleaseMetadata {
	return &ReleaseMetadata{
		Checklist:    NewReleaseChecklist(),
		GateEvidence: make(map[conformance.GateID]string),
		Decisions:    []OpenDecisionRecord{},
	}
}

func (m *ReleaseMetadata) RecordGateEvidence(gate conformance.GateID, detail string) {
	m.GateEvidence[gate] = detail
}

func (m *ReleaseMetadata) AddDecision(decision OpenDecisionRecord) {
	m.Decisions = append(m.Decisions, decision)
}

func (m ReleaseMetadata) GateCoverage() int {
	return len(m.GateEvidence)
}
