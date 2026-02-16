package rbac

import "time"

// Role is the baseline v0.2 RBAC hierarchy.
type Role string

const (
	RoleMember    Role = "member"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
	RoleOwner     Role = "owner"
)

func rank(role Role) int {
	switch role {
	case RoleOwner:
		return 4
	case RoleAdmin:
		return 3
	case RoleModerator:
		return 2
	default:
		return 1
	}
}

// HigherThan reports whether a role outranks another role.
func (r Role) HigherThan(other Role) bool {
	return rank(r) > rank(other)
}

// CanActOnTarget enforces deterministic actor-target authority checks.
func CanActOnTarget(actor Role, target Role) bool {
	return rank(actor) > rank(target)
}

// RoleState stores role convergence metadata for reconnect/stale handling.
type RoleState struct {
	Role      Role
	Version   uint64
	UpdatedAt time.Time
}

// ResolveRoleState deterministically resolves stale or concurrent role views.
func ResolveRoleState(a RoleState, b RoleState) RoleState {
	if a.Version > b.Version {
		return a
	}
	if b.Version > a.Version {
		return b
	}
	if a.UpdatedAt.After(b.UpdatedAt) {
		return a
	}
	if b.UpdatedAt.After(a.UpdatedAt) {
		return b
	}
	if rank(a.Role) < rank(b.Role) {
		return a
	}
	return b
}
