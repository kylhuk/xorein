package rbac

import (
	"testing"
)

func TestRoleManagerDefaultFounder(t *testing.T) {
	mgr := NewRoleManager("alice")
	role, ok := mgr.Assignment("alice")
	if !ok || role != RoleFounder {
		t.Fatalf("expected alice as founder, got %q ok=%v", role, ok)
	}
	if mgr.Founder() != "alice" {
		t.Fatalf("founder should remain alice")
	}
}

func TestCustomRoleLifecycle(t *testing.T) {
	mgr := NewRoleManager("root")
	role := Role("support")
	def := RoleDefinition{Description: "support", Permissions: map[Permission]struct{}{PermissionChatSend: {}}}
	if err := mgr.CreateCustomRole(role, def); err != nil {
		t.Fatalf("create custom role: %v", err)
	}
	if err := mgr.AssignRole("bob", role); err != nil {
		t.Fatalf("assign custom role: %v", err)
	}
	perms, err := mgr.CustomRolePermissions(role)
	if err != nil {
		t.Fatalf("custom permissions: %v", err)
	}
	if len(perms) != 1 || perms[0] != PermissionChatSend {
		t.Fatalf("unexpected perms: %v", perms)
	}
	updated := RoleDefinition{Description: "support", Permissions: map[Permission]struct{}{PermissionVoiceJoin: {}}}
	if err := mgr.UpdateCustomRole(role, updated); err != nil {
		t.Fatalf("update custom role: %v", err)
	}
	perms, err = mgr.CustomRolePermissions(role)
	if err != nil {
		t.Fatalf("custom permissions: %v", err)
	}
	if len(perms) != 1 || perms[0] != PermissionVoiceJoin {
		t.Fatalf("unexpected perms after update: %v", perms)
	}
	if err := mgr.DeleteCustomRole(role); err != nil {
		t.Fatalf("delete custom role: %v", err)
	}
}

func TestFounderCannotBeRevoked(t *testing.T) {
	mgr := NewRoleManager("founder")
	if err := mgr.RevokeRole("founder"); err != ErrFounderImmutable {
		t.Fatalf("expected founder immutable, got %v", err)
	}
}

func TestPermissionsSortOrder(t *testing.T) {
	mgr := NewRoleManager("alice")
	perms, err := mgr.PermissionsFor("alice")
	if err != nil {
		t.Fatalf("permissions: %v", err)
	}
	if len(perms) < 1 {
		t.Fatalf("expected founder permissions")
	}
	if perms[0] != PermissionAdminConfig {
		t.Fatalf("expected deterministic order, got %v", perms)
	}
}
