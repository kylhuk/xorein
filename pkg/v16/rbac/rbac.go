package rbac

import (
	"errors"
	"fmt"
	"sort"
)

// Role identifies a bounded role in the system.
type Role string

const (
	RoleFounder   Role = "founder"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleMember    Role = "member"
	RoleGuest     Role = "guest"
)

// Permission enumerates actionable permissions attached to roles.
type Permission string

const (
	PermissionChatSend        Permission = "chat:send"
	PermissionVoiceJoin       Permission = "voice:join"
	PermissionScreenshareView Permission = "screenshare:view"
	PermissionAdminConfig     Permission = "admin:config"
)

// RoleDefinition describes the permissions attached to a role.
type RoleDefinition struct {
	Description string
	Permissions map[Permission]struct{}
}

// DefaultRoles exposes the shipments for the default RBAC fleet.
var DefaultRoles = map[Role]RoleDefinition{
	RoleFounder: {
		Description: "ultimate control",
		Permissions: map[Permission]struct{}{
			PermissionChatSend:        {},
			PermissionVoiceJoin:       {},
			PermissionScreenshareView: {},
			PermissionAdminConfig:     {},
		},
	},
	RoleAdmin: {
		Description: "admin and policy",
		Permissions: map[Permission]struct{}{
			PermissionChatSend:        {},
			PermissionVoiceJoin:       {},
			PermissionScreenshareView: {},
			PermissionAdminConfig:     {},
		},
	},
	RoleModerator: {
		Description: "moderation and oversight",
		Permissions: map[Permission]struct{}{
			PermissionChatSend:        {},
			PermissionVoiceJoin:       {},
			PermissionScreenshareView: {},
		},
	},
	RoleMember: {
		Description: "standard member",
		Permissions: map[Permission]struct{}{
			PermissionChatSend:  {},
			PermissionVoiceJoin: {},
		},
	},
	RoleGuest: {
		Description: "light spectator",
		Permissions: map[Permission]struct{}{
			PermissionChatSend: {},
		},
	},
}

var (
	ErrFounderImmutable = errors.New("founder role is immutable")
	ErrRoleNotFound     = errors.New("role not found")
	ErrRoleReserved     = errors.New("role name is reserved")
)

// RoleManager handles assignments and custom roles while enforcing safety.
type RoleManager struct {
	assignments  map[string]Role
	customRoles  map[Role]RoleDefinition
	defaultRoles map[Role]RoleDefinition
	founder      string
}

// NewRoleManager seeds the manager with the initial founder assignment.
func NewRoleManager(founder string) *RoleManager {
	manager := &RoleManager{
		assignments:  make(map[string]Role),
		customRoles:  make(map[Role]RoleDefinition),
		defaultRoles: make(map[Role]RoleDefinition),
		founder:      founder,
	}
	for role, def := range DefaultRoles {
		manager.defaultRoles[role] = def
	}
	manager.assignments[founder] = RoleFounder
	return manager
}

// AssignRole assigns a known role to a user.
func (m *RoleManager) AssignRole(user string, role Role) error {
	if role == "" {
		return ErrRoleNotFound
	}
	if !m.isKnownRole(role) {
		return ErrRoleNotFound
	}
	if role == RoleFounder {
		m.founder = user
	}
	m.assignments[user] = role
	return nil
}

// RevokeRole removes a role assignment while preserving founder safety.
func (m *RoleManager) RevokeRole(user string) error {
	current, ok := m.assignments[user]
	if !ok {
		return nil
	}
	if current == RoleFounder {
		return ErrFounderImmutable
	}
	delete(m.assignments, user)
	return nil
}

// CreateCustomRole adds a new custom role definition.
func (m *RoleManager) CreateCustomRole(role Role, def RoleDefinition) error {
	if m.isDefaultRole(role) {
		return ErrRoleReserved
	}
	m.customRoles[role] = def
	return nil
}

// UpdateCustomRole overwrites an existing custom role definition.
func (m *RoleManager) UpdateCustomRole(role Role, def RoleDefinition) error {
	if !m.isCustomRole(role) {
		return ErrRoleNotFound
	}
	m.customRoles[role] = def
	return nil
}

// DeleteCustomRole removes a custom role if unused.
func (m *RoleManager) DeleteCustomRole(role Role) error {
	if !m.isCustomRole(role) {
		return ErrRoleNotFound
	}
	delete(m.customRoles, role)
	return nil
}

// PermissionsFor returns granted permissions for a given user.
func (m *RoleManager) PermissionsFor(user string) ([]Permission, error) {
	role, ok := m.assignments[user]
	if !ok {
		return nil, ErrRoleNotFound
	}
	def, err := m.definitionFor(role)
	if err != nil {
		return nil, err
	}
	perms := make([]Permission, 0, len(def.Permissions))
	for perm := range def.Permissions {
		perms = append(perms, perm)
	}
	sort.Slice(perms, func(i, j int) bool { return perms[i] < perms[j] })
	return perms, nil
}

// RoleDefinitions lists all configured role definitions in deterministic order.
func (m *RoleManager) RoleDefinitions() []Role {
	roles := make([]Role, 0, len(m.defaultRoles)+len(m.customRoles))
	for role := range m.defaultRoles {
		roles = append(roles, role)
	}
	for role := range m.customRoles {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return roles
}

func (m *RoleManager) definitionFor(role Role) (RoleDefinition, error) {
	if def, ok := m.customRoles[role]; ok {
		return def, nil
	}
	if def, ok := m.defaultRoles[role]; ok {
		return def, nil
	}
	return RoleDefinition{}, ErrRoleNotFound
}

func (m *RoleManager) isDefaultRole(role Role) bool {
	_, ok := m.defaultRoles[role]
	return ok
}

func (m *RoleManager) isCustomRole(role Role) bool {
	_, ok := m.customRoles[role]
	return ok
}

func (m *RoleManager) isKnownRole(role Role) bool {
	return m.isDefaultRole(role) || m.isCustomRole(role)
}

// Assignment returns the role currently assigned to a user.
func (m *RoleManager) Assignment(user string) (Role, bool) {
	role, ok := m.assignments[user]
	return role, ok
}

// Founder returns the current founder principal, always assigned.
func (m *RoleManager) Founder() string {
	return m.founder
}

// WithPermission checks whether a user has the requested permission.
func (m *RoleManager) WithPermission(user string, perm Permission) bool {
	role, ok := m.assignments[user]
	if !ok {
		return false
	}
	def, err := m.definitionFor(role)
	if err != nil {
		return false
	}
	_, has := def.Permissions[perm]
	return has
}

// RoleDescription returns the textual description for a role if configured.
func (m *RoleManager) RoleDescription(role Role) (string, error) {
	def, err := m.definitionFor(role)
	if err != nil {
		return "", err
	}
	return def.Description, nil
}

// CustomRolePermissions returns the custom permissions for a role.
func (m *RoleManager) CustomRolePermissions(role Role) ([]Permission, error) {
	if !m.isCustomRole(role) {
		return nil, ErrRoleNotFound
	}
	def := m.customRoles[role]
	perms := make([]Permission, 0, len(def.Permissions))
	for perm := range def.Permissions {
		perms = append(perms, perm)
	}
	sort.Slice(perms, func(i, j int) bool { return perms[i] < perms[j] })
	return perms, nil
}

// RolePermissions returns permissions for a named role.
func (m *RoleManager) RolePermissions(role Role) ([]Permission, error) {
	def, err := m.definitionFor(role)
	if err != nil {
		return nil, err
	}
	perms := make([]Permission, 0, len(def.Permissions))
	for perm := range def.Permissions {
		perms = append(perms, perm)
	}
	sort.Slice(perms, func(i, j int) bool { return perms[i] < perms[j] })
	return perms, nil
}

// Helper to format permission list for debugging.
func FormatPermissions(perms []Permission) string {
	if len(perms) == 0 {
		return "none"
	}
	sorted := append([]Permission{}, perms...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return fmt.Sprintf("[%s]", joinPerms(sorted))
}

func joinPerms(perms []Permission) string {
	labels := make([]string, len(perms))
	for i, perm := range perms {
		labels[i] = string(perm)
	}
	return fmt.Sprintf("%s", labels)
}
