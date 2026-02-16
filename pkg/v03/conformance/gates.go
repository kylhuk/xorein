package conformance

import "fmt"

type GateChecklist struct {
	GateID      string
	ScopeItems  []string
	EvidenceMap map[string]string
	Completed   bool
}

func NewGateChecklist(id string, scopeItems []string) GateChecklist {
	return GateChecklist{GateID: id, ScopeItems: append([]string(nil), scopeItems...), EvidenceMap: make(map[string]string)}
}

func (g GateChecklist) IsSatisfied() bool {
	return g.Completed && len(g.EvidenceMap) == len(g.ScopeItems)
}

func (g *GateChecklist) RecordEvidence(scope string, evidence string) {
	if g.EvidenceMap == nil {
		g.EvidenceMap = make(map[string]string)
	}
	g.EvidenceMap[scope] = evidence
}

func (g GateChecklist) TraceSummary() string {
	return fmt.Sprintf("Gate %s covers %d items; completed=%t", g.GateID, len(g.ScopeItems), g.Completed)
}

var TraceableBullets = []string{
	"S3-01", "S3-02", "S3-03", "S3-04", "S3-05", "S3-06",
	"S3-07", "S3-08", "S3-09", "S3-10", "S3-11",
	"S3-12", "S3-13",
	"S3-14", "S3-15", "S3-16",
	"S3-17",
}

func EvidenceMatrix(positive, adverse, degraded, recovery []string) map[string][]string {
	return map[string][]string{
		"positive": append([]string(nil), positive...),
		"adverse":  append([]string(nil), adverse...),
		"degraded": append([]string(nil), degraded...),
		"recovery": append([]string(nil), recovery...),
	}
}
