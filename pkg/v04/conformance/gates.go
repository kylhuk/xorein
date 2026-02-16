package conformance

import "fmt"

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

var GateScopeMapping = map[GateID][]string{
	GateV4G0: {"P0-T1", "P0-T2", "P0-T3"},
	GateV4G1: {"P1-T1", "P1-T2", "P1-T3", "P1-T4"},
	GateV4G2: {"P2-T1", "P2-T2", "P2-T3"},
	GateV4G3: {"P3-T1", "P3-T2", "P3-T3"},
	GateV4G4: {"P4-T1", "P4-T2", "P4-T3", "P4-T4"},
	GateV4G5: {"P5-T1", "P5-T2"},
	GateV4G6: {"P5-T3"},
}

type GateChecklist struct {
	Gate      GateID
	Completed bool
	Evidence  map[string]string
}

func NewGateChecklist(id GateID) GateChecklist {
	return GateChecklist{Gate: id, Evidence: make(map[string]string)}
}

func (g GateChecklist) RequiredTasks() []string {
	tasks, ok := GateScopeMapping[g.Gate]
	if !ok {
		return nil
	}
	return append([]string(nil), tasks...)
}

func (g GateChecklist) IsSatisfied() bool {
	required := g.RequiredTasks()
	if len(required) == 0 {
		return false
	}
	if !g.Completed || g.Evidence == nil {
		return false
	}
	for _, task := range required {
		if _, ok := g.Evidence[task]; !ok {
			return false
		}
	}
	return true
}

func (g *GateChecklist) RecordEvidence(task, evidence string) {
	if g.Evidence == nil {
		g.Evidence = make(map[string]string)
	}
	g.Evidence[task] = evidence
}

func GateSummary(id GateID) string {
	tasks := append([]string(nil), GateScopeMapping[id]...)
	return fmt.Sprintf("gate %s: %d tasks", id, len(tasks))
}

func EvidenceMatrix(positive, adverse, degraded, recovery []string) map[string][]string {
	return map[string][]string{
		"positive": append([]string(nil), positive...),
		"adverse":  append([]string(nil), adverse...),
		"degraded": append([]string(nil), degraded...),
		"recovery": append([]string(nil), recovery...),
	}
}

var TraceableScopes = []string{
	"S4-01", "S4-02", "S4-03", "S4-04", "S4-05", "S4-06", "S4-07",
}
