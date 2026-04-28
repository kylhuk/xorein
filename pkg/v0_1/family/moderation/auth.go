package moderation

import (
	"fmt"

	gov "github.com/aether/code_aether/pkg/v0_1/family/governance"
)

// RoleSource is the interface moderation uses to look up RBAC roles.
type RoleSource interface {
	RoleOf(serverID, peerID string) gov.Role
}

// ErrUnauthorized signals the actor's role is insufficient.
var ErrUnauthorized = fmt.Errorf("moderation: unauthorized")

// ErrForbiddenTarget signals the actor cannot act on the target.
var ErrForbiddenTarget = fmt.Errorf("moderation: forbidden target")

// authorize checks that actorID holds at least minRole for serverID, and when
// targetID is non-empty that actor strictly outranks target (spec 50 §3.1).
func authorize(roles RoleSource, serverID, actorID, targetID string, minRole gov.Role) error {
	actorRole := roles.RoleOf(serverID, actorID)
	if actorRole < minRole {
		return ErrUnauthorized
	}
	if targetID == "" {
		return nil
	}
	targetRole := roles.RoleOf(serverID, targetID)
	if !gov.CanActOnTarget(actorRole, targetRole) {
		return ErrForbiddenTarget
	}
	return nil
}
