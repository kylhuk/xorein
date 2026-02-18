package v16

import (
	"testing"

	"github.com/aether/code_aether/pkg/v16/acl"
	"github.com/aether/code_aether/pkg/v16/enforcement"
	"github.com/aether/code_aether/pkg/v16/rbac"
)

func TestPermissionMatrix(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	mgr.AssignRole("mod", rbac.RoleModerator)
	mgr.AssignRole("guest", rbac.RoleGuest)
	eng := enforcement.New(mgr)

	cases := []struct {
		user   string
		action acl.Action
		allow  bool
	}{
		{user: "founder", action: acl.ActionScreenshareView, allow: true},
		{user: "mod", action: acl.ActionVoiceJoin, allow: true},
		{user: "guest", action: acl.ActionVoiceJoin, allow: false},
	}

	for _, test := range cases {
		decision := eng.Ensure(test.action, test.user, "")
		if decision.Allowed != test.allow {
			t.Fatalf("expected %s allow=%v got=%v", test.user, test.allow, decision.Allowed)
		}
	}
}
