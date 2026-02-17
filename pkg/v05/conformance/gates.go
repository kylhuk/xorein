package conformance

type GateID string

const (
	GateV5G0 GateID = "V5-G0"
	GateV5G1 GateID = "V5-G1"
	GateV5G2 GateID = "V5-G2"
	GateV5G3 GateID = "V5-G3"
	GateV5G4 GateID = "V5-G4"
	GateV5G5 GateID = "V5-G5"
	GateV5G6 GateID = "V5-G6"
	GateV5G7 GateID = "V5-G7"
)

var ScopeIDs = []string{
	"S5-01",
	"S5-02",
	"S5-03",
	"S5-04",
	"S5-05",
	"S5-06",
	"S5-07",
	"S5-08",
}

var GateScopeMapping = map[GateID][]string{
	GateV5G0: {"S5-01", "S5-02"},
	GateV5G1: {"S5-03"},
	GateV5G2: {"S5-04", "S5-05"},
	GateV5G3: {"S5-06"},
	GateV5G4: {"S5-07"},
	GateV5G5: {"S5-08"},
	GateV5G6: {"S5-01", "S5-04"},
	GateV5G7: {"S5-02", "S5-03", "S5-05"},
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
