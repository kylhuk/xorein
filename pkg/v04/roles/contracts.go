package roles

import "sort"

type Role string
type Permission string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
)

const (
	ResolutionPrimary   MergeResolution = "primary"
	ResolutionSecondary MergeResolution = "secondary"
)

type MergeResolution string

type RoleAssignment struct {
	Role        Role
	Permissions []Permission
	Priority    int
}

type MergeResult struct {
	Role        Role
	Permissions []Permission
	Conflict    bool
	Resolution  MergeResolution
}

func MergeAssignments(primary, secondary RoleAssignment) MergeResult {
	combined := append([]Permission(nil), primary.Permissions...)
	combined = append(combined, secondary.Permissions...)
	permissions := uniquePermissions(combined)
	conflict := permissionsConflict(primary.Permissions, secondary.Permissions)
	resolution := ResolutionPrimary
	role := primary.Role
	if secondary.Priority > primary.Priority {
		resolution = ResolutionSecondary
		role = secondary.Role
	}
	return MergeResult{Role: role, Permissions: permissions, Conflict: conflict, Resolution: resolution}
}

func (a RoleAssignment) HasPermission(permission Permission) bool {
	for _, p := range a.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func uniquePermissions(perms []Permission) []Permission {
	seen := make(map[Permission]struct{}, len(perms))
	result := make([]Permission, 0, len(perms))
	for _, p := range perms {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func permissionsConflict(primary, secondary []Permission) bool {
	primarySet := make(map[Permission]struct{}, len(primary))
	for _, p := range primary {
		primarySet[p] = struct{}{}
	}
	for _, p := range secondary {
		if _, ok := primarySet[p]; ok {
			return true
		}
	}
	return false
}
