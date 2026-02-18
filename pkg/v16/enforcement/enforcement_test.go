package enforcement

import (
	"testing"

	"github.com/aether/code_aether/pkg/v16/acl"
	"github.com/aether/code_aether/pkg/v16/rbac"
)

func TestEnforcementBlocksWithoutPermission(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	mgr.AssignRole("guest", rbac.RoleGuest)
	eng := New(mgr)
	decision := eng.Ensure(acl.ActionVoiceJoin, "guest", "")
	if decision.Allowed {
		t.Fatalf("guest should not join voice")
	}
}

func TestChannelPolicyOverrides(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	mgr.AssignRole("member", rbac.RoleMember)
	eng := New(mgr)
	policy := acl.New()
	policy.Deny(acl.ActionChatSend, "channel policy")
	eng.SetChannelPolicy("ch1", policy)
	decision := eng.Ensure(acl.ActionChatSend, "member", "ch1")
	if decision.Allowed {
		t.Fatalf("channel deny should block chat send")
	}
	trace, err := eng.ExplainChannelPolicy("ch1")
	if err != nil {
		t.Fatalf("explain policy: %v", err)
	}
	if len(trace) != 1 {
		t.Fatalf("expected policy trace entry, got %v", trace)
	}
}

func TestAdminActionPermission(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	eng := New(mgr)
	decision := eng.EnsureAdminAction("founder")
	if !decision.Allowed {
		t.Fatalf("founder should run admin action")
	}
	if decision.Reason == "" {
		t.Fatalf("reason expected")
	}
}
