package conformance

type GateID string

const (
	GateV7G0 GateID = "V7-G0"
	GateV7G1 GateID = "V7-G1"
	GateV7G2 GateID = "V7-G2"
	GateV7G3 GateID = "V7-G3"
	GateV7G4 GateID = "V7-G4"
	GateV7G5 GateID = "V7-G5"
	GateV7G6 GateID = "V7-G6"
)

var ScopeIDs = []string{
	"S7-01",
	"S7-02",
	"S7-03",
	"S7-04",
	"S7-05",
	"S7-06",
	"S7-07",
	"S7-08",
	"S7-09",
}

var GateScopeMapping = map[GateID][]string{
	GateV7G0: {"S7-01", "S7-02"},
	GateV7G1: {"S7-02", "S7-03", "S7-04"},
	GateV7G2: {"S7-03", "S7-05"},
	GateV7G3: {"S7-04", "S7-06"},
	GateV7G4: {"S7-05", "S7-07", "S7-08"},
	GateV7G5: {"S7-06", "S7-08", "S7-09"},
	GateV7G6: {"S7-01", "S7-05", "S7-09"},
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

func (g GateChecklist) MissingScopes() []string {
	required := g.RequiredScopes()
	if len(required) == 0 {
		return nil
	}
	var missing []string
	for _, scope := range required {
		if _, ok := g.Evidence[scope]; !ok {
			missing = append(missing, scope)
		}
	}
	return missing
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

func GateIsComplete(g GateChecklist) bool {
	return g.IsSatisfied() && len(g.MissingScopes()) == 0
}
