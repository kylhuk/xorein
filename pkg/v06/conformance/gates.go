package conformance

type GateID string

const (
	GateV6G0 GateID = "V6-G0"
	GateV6G1 GateID = "V6-G1"
	GateV6G2 GateID = "V6-G2"
	GateV6G3 GateID = "V6-G3"
	GateV6G4 GateID = "V6-G4"
	GateV6G5 GateID = "V6-G5"
	GateV6G6 GateID = "V6-G6"
)

var ScopeIDs = []string{
	"S6-01",
	"S6-02",
	"S6-03",
	"S6-04",
	"S6-05",
	"S6-06",
	"S6-07",
	"S6-08",
	"S6-09",
	"S6-10",
}

var GateScopeMapping = map[GateID][]string{
	GateV6G0: {"S6-01", "S6-02"},
	GateV6G1: {"S6-01", "S6-02", "S6-03"},
	GateV6G2: {"S6-02", "S6-03", "S6-04"},
	GateV6G3: {"S6-05", "S6-06", "S6-07"},
	GateV6G4: {"S6-08", "S6-09", "S6-10"},
	GateV6G5: {"S6-01", "S6-05", "S6-09"},
	GateV6G6: {"S6-02", "S6-04", "S6-07", "S6-10"},
}

type GateChecklist struct {
	Gate      GateID
	Completed bool
	Evidence  map[string]string
}

func NewGateChecklist(id GateID) GateChecklist {
	return GateChecklist{
		Gate:     id,
		Evidence: make(map[string]string),
	}
}

func (g GateChecklist) RequiredScopes() []string {
	scopes, ok := GateScopeMapping[g.Gate]
	if !ok {
		return nil
	}
	return append([]string(nil), scopes...)
}

func (g GateChecklist) IsSatisfied() bool {
	if !g.Completed {
		return false
	}
	required := g.RequiredScopes()
	if len(required) == 0 {
		return false
	}
	for _, scope := range required {
		if _, ok := g.Evidence[scope]; !ok {
			return false
		}
	}
	return true
}

func (g *GateChecklist) RecordEvidence(scope, detail string) {
	if g.Evidence == nil {
		g.Evidence = make(map[string]string)
	}
	g.Evidence[scope] = detail
}

func GateCoverageScore(g GateChecklist) float64 {
	required := g.RequiredScopes()
	total := len(required)
	if total == 0 {
		return 0
	}

	covered := 0
	for _, scope := range required {
		if _, ok := g.Evidence[scope]; ok {
			covered++
		}
	}

	return float64(covered) / float64(total)
}
