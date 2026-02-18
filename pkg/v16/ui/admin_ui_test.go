package ui

import (
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v16/rbac"
)

func TestRoleSummaries(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	summaries := RoleSummaries(mgr)
	if len(summaries) == 0 {
		t.Fatalf("expected role summaries")
	}
	for _, summary := range summaries {
		if strings.TrimSpace(summary.Description) == "" {
			t.Fatalf("summary should describe %s", summary.Name)
		}
	}
}

func TestPermissionHint(t *testing.T) {
	hint := PermissionHint(rbac.RoleMember, rbac.PermissionChatSend, true)
	if !strings.Contains(hint, "can") {
		t.Fatalf("hint should mention can: %s", hint)
	}
}

func TestRecommendUpgrade(t *testing.T) {
	msg := RecommendUpgrade(rbac.RoleMember, rbac.RoleAdmin, false)
	if !strings.Contains(msg, "consider") {
		t.Fatalf("expected recommendation, got %s", msg)
	}
}
