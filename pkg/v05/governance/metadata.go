package governance

import "github.com/aether/code_aether/pkg/v05/conformance"

type OpenDecisionID string

const (
	OpenDecisionOD501 OpenDecisionID = "OD5-01"
	OpenDecisionOD502 OpenDecisionID = "OD5-02"
	OpenDecisionOD503 OpenDecisionID = "OD5-03"
	OpenDecisionOD504 OpenDecisionID = "OD5-04"
	OpenDecisionOD505 OpenDecisionID = "OD5-05"
)

var OpenDecisions = []OpenDecisionID{
	OpenDecisionOD501,
	OpenDecisionOD502,
	OpenDecisionOD503,
	OpenDecisionOD504,
	OpenDecisionOD505,
}

type ReleaseChecklist struct {
	Items map[string]bool
}

var ReleaseChecklistItems = []string{
	"Track v0.5 scope per TODO references",
	"Document gating evidence for all conformance scopes",
	"Preserve planned-vs-implemented cues for v0.5 artifacts",
	"Open decisions remain tracked and unresolved",
	"Governance metrics include SDK/community conformance helpers",
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
