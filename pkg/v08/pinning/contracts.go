package pinning

import (
	"sort"
)

// Scope defines the deterministic scope of a pin operation.
type Scope string

const (
	ScopePersonal Scope = "personal"
	ScopeTeam     Scope = "team"
	ScopeGlobal   Scope = "global"
)

// PinAuthority declares the authority that can approve pin scope changes.
type PinAuthority struct {
	ID       string
	Scope    Scope
	Priority int
}

// ValidatePinScope ensures the scope matches the additive policy.
func ValidatePinScope(scope Scope) bool {
	switch scope {
	case ScopePersonal, ScopeTeam, ScopeGlobal:
		return true
	default:
		return false
	}
}

// DeterministicOrder returns a stable ordering for authorities.
func DeterministicOrder(authorities []PinAuthority) []PinAuthority {
	copyAuth := make([]PinAuthority, len(authorities))
	copy(copyAuth, authorities)
	sort.SliceStable(copyAuth, func(i, j int) bool {
		if copyAuth[i].Scope != copyAuth[j].Scope {
			return copyAuth[i].Scope < copyAuth[j].Scope
		}
		if copyAuth[i].Priority != copyAuth[j].Priority {
			return copyAuth[i].Priority < copyAuth[j].Priority
		}
		return copyAuth[i].ID < copyAuth[j].ID
	})
	return copyAuth
}
