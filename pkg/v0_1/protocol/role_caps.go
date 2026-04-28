package protocol

import "fmt"

// roleCapsTable encodes the role→capability matrix per spec 03 §3.4.
// Keys are normalised role strings matching Config.Role in the v0.1 runtime.
var roleCapsTable = map[string][]FeatureFlag{
	"client": {
		"cap.chat",
		"cap.dm",
		"cap.voice",
		"cap.identity",
		"cap.manifest",
		"cap.friends",
		"cap.presence",
		"cap.notify",
		"cap.group-dm",
		"cap.sync",
		"cap.rbac",
		"cap.moderation",
		"cap.peer.transport",
		"mode.seal",
		"mode.tree",
		"mode.crowd",
		"mode.channel",
		"mode.mediashield",
	},
	"relay": {
		"cap.peer.transport",
		"cap.peer.relay",
		"cap.peer.delivery",
		"cap.peer.manifest",
		"cap.peer.bootstrap",
	},
	"bootstrap": {
		"cap.peer.transport",
		"cap.peer.bootstrap",
	},
	"archivist": {
		"cap.chat",
		"cap.sync",
		"cap.archivist",
		"cap.manifest",
		"cap.peer.transport",
		"cap.identity",
	},
}

// RoleCapabilities returns the canonical capability set for a node role per spec 03 §3.4.
// Returns nil for unknown roles.
func RoleCapabilities(role string) []FeatureFlag {
	caps, ok := roleCapsTable[role]
	if !ok {
		return nil
	}
	out := make([]FeatureFlag, len(caps))
	copy(out, caps)
	return out
}

// ValidateRoleCapabilities returns an error if any required cap for the role is absent
// from the provided list. Extra caps beyond the role's canonical set are silently ignored
// (forward-compatibility). An unknown role is itself an error.
func ValidateRoleCapabilities(role string, caps []FeatureFlag) error {
	required, ok := roleCapsTable[role]
	if !ok {
		return fmt.Errorf("protocol: unknown role %q", role)
	}

	have := make(map[FeatureFlag]struct{}, len(caps))
	for _, c := range caps {
		have[c] = struct{}{}
	}

	for _, req := range required {
		if _, found := have[req]; !found {
			return fmt.Errorf("protocol: role %q requires capability %q which is missing", role, req)
		}
	}
	return nil
}
