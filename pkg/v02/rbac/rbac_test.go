package rbac

import (
	"testing"
	"time"
)

func TestRoleHierarchy(t *testing.T) {
	if !RoleOwner.HigherThan(RoleAdmin) {
		t.Fatal("owner should outrank admin")
	}
	if RoleModerator.HigherThan(RoleOwner) {
		t.Fatal("moderator should not outrank owner")
	}
}

func TestCanActOnTarget(t *testing.T) {
	if !CanActOnTarget(RoleAdmin, RoleMember) {
		t.Fatal("admin should act on member")
	}
	if CanActOnTarget(RoleModerator, RoleAdmin) {
		t.Fatal("moderator should not act on admin")
	}
}

func TestResolveRoleStateDeterministic(t *testing.T) {
	a := RoleState{Role: RoleModerator, Version: 2, UpdatedAt: time.Unix(5, 0)}
	b := RoleState{Role: RoleAdmin, Version: 2, UpdatedAt: time.Unix(5, 0)}
	resolved := ResolveRoleState(a, b)
	if resolved.Role != RoleModerator {
		t.Fatalf("role=%q want %q", resolved.Role, RoleModerator)
	}
}
