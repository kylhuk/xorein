package security

// AssistedSearchGate represents an explicit opt-in gate for assisted search.
type AssistedSearchGate struct {
	Enabled      bool
	ConsentToken string
	Requirement  string
}

// Enable returns a gate configured for an explicit consent token.
func (g AssistedSearchGate) Enable(token string) AssistedSearchGate {
	g.Enabled = true
	g.ConsentToken = token
	return g
}

// Allows reports whether assisted search is permitted with the provided token.
func (g AssistedSearchGate) Allows(token string) bool {
	if !g.Enabled {
		return false
	}
	return token != "" && g.ConsentToken == token
}

// Info returns human-readable hints about the gate state.
func (g AssistedSearchGate) Info() string {
	if !g.Enabled {
		return "assisted search is disabled"
	}
	if g.ConsentToken == "" {
		return "assisted search requires a consent token"
	}
	return "assisted search gated"
}
