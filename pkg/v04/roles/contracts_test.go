package roles

import (
	"reflect"
	"testing"
)

func TestMergeAssignments(t *testing.T) {
	tests := []struct {
		name           string
		primary        RoleAssignment
		secondary      RoleAssignment
		wantRole       Role
		wantResolution MergeResolution
		wantConflict   bool
		wantPerms      []Permission
	}{
		{
			name: "primary priority wins with unique permissions",
			primary: RoleAssignment{
				Role:        RoleOwner,
				Permissions: []Permission{"read", "write"},
				Priority:    5,
			},
			secondary: RoleAssignment{
				Role:        RoleAdmin,
				Permissions: []Permission{"delete"},
				Priority:    3,
			},
			wantRole:       RoleOwner,
			wantResolution: ResolutionPrimary,
			wantConflict:   false,
			wantPerms:      []Permission{"delete", "read", "write"},
		},
		{
			name: "secondary priority wins with conflict and dedup",
			primary: RoleAssignment{
				Role:        RoleMember,
				Permissions: []Permission{"audit", "audit", "common"},
				Priority:    1,
			},
			secondary: RoleAssignment{
				Role:        RoleModerator,
				Permissions: []Permission{"common", "block"},
				Priority:    10,
			},
			wantRole:       RoleModerator,
			wantResolution: ResolutionSecondary,
			wantConflict:   true,
			wantPerms:      []Permission{"audit", "block", "common"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeAssignments(tt.primary, tt.secondary)
			if got.Role != tt.wantRole {
				t.Fatalf("role: got %s want %s", got.Role, tt.wantRole)
			}
			if got.Resolution != tt.wantResolution {
				t.Fatalf("resolution: got %s want %s", got.Resolution, tt.wantResolution)
			}
			if got.Conflict != tt.wantConflict {
				t.Fatalf("conflict: got %t want %t", got.Conflict, tt.wantConflict)
			}
			if !reflect.DeepEqual(got.Permissions, tt.wantPerms) {
				t.Fatalf("permissions: got %v want %v", got.Permissions, tt.wantPerms)
			}
		})
	}
}

func TestRoleAssignmentHasPermission(t *testing.T) {
	assignment := RoleAssignment{Permissions: []Permission{"alpha", "beta"}}
	tests := []struct {
		permission Permission
		want       bool
	}{
		{permission: "alpha", want: true},
		{permission: "beta", want: true},
		{permission: "gamma", want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.permission), func(t *testing.T) {
			if got := assignment.HasPermission(tt.permission); got != tt.want {
				t.Fatalf("HasPermission(%s): got %t want %t", tt.permission, got, tt.want)
			}
		})
	}
}
