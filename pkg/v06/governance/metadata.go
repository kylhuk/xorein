package governance

import "github.com/aether/code_aether/pkg/v06/conformance"

type OpenDecisionID string

const (
	OpenDecisionOD601 OpenDecisionID = "OD6-01"
	OpenDecisionOD602 OpenDecisionID = "OD6-02"
	OpenDecisionOD603 OpenDecisionID = "OD6-03"
)

var OpenDecisions = []OpenDecisionID{
	OpenDecisionOD601,
	OpenDecisionOD602,
	OpenDecisionOD603,
}

type ReleaseChecklist struct {
	Items map[string]bool
}

var ReleaseChecklistItems = []string{
	"Map all ten v0.6 bullets to tasks, artifacts, and VA traces",
	"Frame v0.6 as hardening/scaling/reliability in every artifact",
	"Describe discovery/search/explore/preview hardening language",
	"Keep optional filters non-authoritative",
	"Capture deterministic reliability for anti-abuse/reporting",
	"Prevent v0.7+/archive scope creep",
	"Finish compatibility/governance/open-decision checks",
	"Keep planned-vs-implemented cues explicit",
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
