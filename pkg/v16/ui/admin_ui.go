package ui

import (
	"fmt"
	"sort"

	"github.com/aether/code_aether/pkg/v16/rbac"
)

// RoleSummary describes UI-ready metadata for a role.
type RoleSummary struct {
	Name        rbac.Role
	Description string
	Permissions []rbac.Permission
}

// RoleSummaries reports summaries for all configured roles.
func RoleSummaries(mgr *rbac.RoleManager) []RoleSummary {
	roles := mgr.RoleDefinitions()
	summaries := make([]RoleSummary, 0, len(roles))
	for _, role := range roles {
		desc, _ := mgr.RoleDescription(role)
		perms, err := mgr.RolePermissions(role)
		if err != nil {
			perms = []rbac.Permission{}
		}
		summaries = append(summaries, RoleSummary{Name: role, Description: desc, Permissions: perms})
	}
	return summaries
}

// PermissionHint returns a deterministic hint string for UI workflows.
func PermissionHint(role rbac.Role, perm rbac.Permission, allowed bool) string {
	status := "cannot"
	if allowed {
		status = "can"
	}
	return fmt.Sprintf("Role %s %s perform %s", role, status, perm)
}

// RecommendUpgrade returns a sorted recommendation for a user to bridge to a new role.
func RecommendUpgrade(current, target rbac.Role, allowed bool) string {
	if allowed {
		return fmt.Sprintf("upgrade to %s is already allowed", target)
	}
	hints := []string{"contact owner", "escalate via admin", "review policy"}
	sort.Strings(hints)
	return fmt.Sprintf("consider %s to reach %s", hints[0], target)
}
