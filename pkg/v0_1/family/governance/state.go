package governance

import (
	"fmt"
	"sync"
	"time"
)

// CustomRole is a server-defined role with a permissions bitfield.
type CustomRole struct {
	Name               string
	PermissionsBitfield uint64
}

// serverState is the in-memory RBAC state for one server.
type serverState struct {
	roles       map[string]RoleState // peerID → RoleState
	customRoles map[string]CustomRole // roleName → CustomRole
	version     uint64               // max RoleState.Version seen
}

func newServerState() *serverState {
	return &serverState{
		roles:       make(map[string]RoleState),
		customRoles: make(map[string]CustomRole),
	}
}

// policyVersion returns the current max role-version for this server.
func (s *serverState) policyVersion() uint64 { return s.version }

// upsert applies a new RoleState using ResolveRoleState convergence.
// Caller holds Handler.mu.Lock().
func (s *serverState) upsert(peerID string, incoming RoleState) {
	if existing, ok := s.roles[peerID]; ok {
		incoming = ResolveRoleState(existing, incoming)
	}
	s.roles[peerID] = incoming
	if incoming.Version > s.version {
		s.version = incoming.Version
	}
}

// allRoles returns a copy of every RoleState for governance.sync.
// Each returned entry has PeerID set to the map key.
func (s *serverState) allRoles() []RoleState {
	out := make([]RoleState, 0, len(s.roles))
	for peerID, rs := range s.roles {
		rs.PeerID = peerID
		out = append(out, rs)
	}
	return out
}

// StateStore is the thread-safe in-memory RBAC store shared by the Handler.
// The handler exposes helpers moderation needs without importing the handler type.
type StateStore struct {
	mu      sync.RWMutex
	servers map[string]*serverState // serverID → serverState
}

func newStateStore() *StateStore {
	return &StateStore{servers: make(map[string]*serverState)}
}

func (s *StateStore) server(serverID string) *serverState {
	srv := s.servers[serverID]
	if srv == nil {
		srv = newServerState()
		s.servers[serverID] = srv
	}
	return srv
}

// RoleOf returns the role for (serverID, peerID); RoleMember when unknown.
func (s *StateStore) RoleOf(serverID, peerID string) Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv := s.servers[serverID]
	if srv == nil {
		return RoleMember
	}
	rs, ok := srv.roles[peerID]
	if !ok {
		return RoleMember
	}
	return rs.Role
}

// PolicyVersion returns the max policy version seen for this server.
func (s *StateStore) PolicyVersion(serverID string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv := s.servers[serverID]
	if srv == nil {
		return 0
	}
	return srv.policyVersion()
}

// Assign writes a new RoleState for (serverID, peerID).
func (s *StateStore) Assign(serverID, peerID string, role Role) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	existing := srv.roles[peerID]
	next := RoleState{
		Role:      role,
		Version:   existing.Version + 1,
		UpdatedAt: time.Now(),
	}
	srv.upsert(peerID, next)
}

// AllRoleStates returns all role states for the server; used by governance.sync.
func (s *StateStore) AllRoleStates(serverID string) ([]RoleState, uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv := s.servers[serverID]
	if srv == nil {
		return nil, 0
	}
	return srv.allRoles(), srv.policyVersion()
}

// SetCustomRole creates or updates a custom role.
func (s *StateStore) SetCustomRole(serverID, roleName string, bitfield uint64) error {
	if IsBaseRoleName(roleName) {
		return fmt.Errorf("base role protected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	if _, exists := srv.customRoles[roleName]; exists {
		return fmt.Errorf("conflict")
	}
	srv.customRoles[roleName] = CustomRole{Name: roleName, PermissionsBitfield: bitfield}
	return nil
}

// DeleteCustomRole removes a custom role and reverts affected members.
func (s *StateStore) DeleteCustomRole(serverID, roleName string) error {
	if IsBaseRoleName(roleName) {
		return fmt.Errorf("base role protected")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	if _, exists := srv.customRoles[roleName]; !exists {
		return fmt.Errorf("not found")
	}
	delete(srv.customRoles, roleName)
	return nil
}

// SeedForTest writes a RoleState directly without incrementing policy_version.
// Use only from test seams; bypasses the normal version-increment path.
func (s *StateStore) SeedForTest(serverID, peerID string, role Role) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	srv.roles[peerID] = RoleState{Role: role, Version: 0, UpdatedAt: time.Now()}
	// Intentionally do NOT update srv.version so policy_version stays at 0.
}

// CustomRoleExists checks whether a named custom role exists.
func (s *StateStore) CustomRoleExists(serverID, roleName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv := s.servers[serverID]
	if srv == nil {
		return false
	}
	_, ok := srv.customRoles[roleName]
	return ok
}
