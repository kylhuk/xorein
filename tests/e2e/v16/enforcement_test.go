package v16

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v16/acl"
	"github.com/aether/code_aether/pkg/v16/enforcement"
	"github.com/aether/code_aether/pkg/v16/rbac"
)

func TestEnforcementScenario(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	mgr.AssignRole("alice", rbac.RoleMember)
	eng := enforcement.New(mgr)
	eng.SetChannelPolicy("public", func() *acl.ACL {
		policy := acl.New()
		policy.Deny(acl.ActionVoiceJoin, "pod override")
		return policy
	}())
	if decision := eng.Ensure(acl.ActionVoiceJoin, "alice", "public"); decision.Allowed {
		t.Fatalf("voice channel enforce should deny")
	}
}

func TestRelayNoDataRegression(t *testing.T) {
	forbidden := relaypolicy.ForbiddenClasses()
	if len(forbidden) == 0 {
		t.Fatalf("relay must still forbid durable hosting")
	}
}
