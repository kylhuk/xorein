// Package governance implements the /aether/governance/0.2.0 protocol family.
// Source: docs/spec/v0.1/51-family-governance.md
package governance

import "time"

// Role represents the RBAC rank per spec 51 §3.1 and proto V02Role.
type Role uint8

const (
	RoleMember    Role = 1
	RoleModerator Role = 2
	RoleAdmin     Role = 3
	RoleOwner     Role = 4
)

// Outranks returns true when r strictly outranks other.
func (r Role) Outranks(other Role) bool { return r > other }

// CanActOnTarget returns true when actor strictly outranks target (spec 51 §3.1 rule 2).
func CanActOnTarget(actor, target Role) bool { return actor.Outranks(target) }

// Valid returns true for a known, non-zero role value.
func (r Role) Valid() bool { return r >= RoleMember && r <= RoleOwner }

// String returns the wire name for logging.
func (r Role) String() string {
	switch r {
	case RoleMember:
		return "member"
	case RoleModerator:
		return "moderator"
	case RoleAdmin:
		return "admin"
	case RoleOwner:
		return "owner"
	default:
		return "unknown"
	}
}

// baseRoles is the set of protected role names that MUST NOT be overridden.
var baseRoles = map[string]struct{}{
	"member": {}, "moderator": {}, "admin": {}, "owner": {},
}

// IsBaseRoleName returns true when name is one of the four protected base roles.
func IsBaseRoleName(name string) bool {
	_, ok := baseRoles[name]
	return ok
}

// PermissionsReservedMask covers bits 10–63 which MUST be zero in v0.1 (spec 51 §4.5).
const PermissionsReservedMask uint64 = 0xFFFFFFFFFFFFFC00

// RoleState is the per-member RBAC record with convergence metadata (spec 51 §4.10, §6).
type RoleState struct {
	PeerID    string // the peer this role belongs to; set by StateStore
	Role      Role
	Version   uint64
	UpdatedAt time.Time
}

// ResolveRoleState deterministically selects the canonical state when two concurrent
// records exist for the same (server_id, peer_id) key (spec 51 §6):
//  1. Higher Version wins.
//  2. Tie: newer UpdatedAt wins.
//  3. Tie: lower-ranked Role wins (conservative resolution).
func ResolveRoleState(a, b RoleState) RoleState {
	if a.Version != b.Version {
		if a.Version > b.Version {
			return a
		}
		return b
	}
	if !a.UpdatedAt.Equal(b.UpdatedAt) {
		if a.UpdatedAt.After(b.UpdatedAt) {
			return a
		}
		return b
	}
	if a.Role <= b.Role {
		return a
	}
	return b
}
