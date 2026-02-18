package v16

import (
	"testing"

	"github.com/aether/code_aether/pkg/v16/acl"
	"github.com/aether/code_aether/pkg/v16/enforcement"
	"github.com/aether/code_aether/pkg/v16/rbac"
)

func TestPermissionStepsLoop(t *testing.T) {
	mgr := rbac.NewRoleManager("founder")
	eng := enforcement.New(mgr)

	for i := 0; i < 100; i++ {
		user := "founder"
		decision := eng.Ensure(acl.ActionChatSend, user, "perf")
		if !decision.Allowed {
			t.Fatalf("founder should always send chat step %d", i)
		}
	}
}
