package protocol

import (
	"strings"

	legacy "github.com/aether/code_aether/pkg/protocol"
)

// Re-export legacy capability flags and negotiation for use in v0.1 code.
type (
	FeatureFlag         = legacy.FeatureFlag
	CapabilityNegResult = legacy.CapabilityNegotiationResult
)

// SecurityMode is the explicit conversation mode, extended to all 6 spec values.
// Source: docs/spec/v0.1/04-security-modes.md §1.
type SecurityMode string

const (
	ModeSeal        SecurityMode = "seal"
	ModeTree        SecurityMode = "tree"
	ModeCrowd       SecurityMode = "crowd"
	ModeChannel     SecurityMode = "channel"
	ModeMediaShield SecurityMode = "mediashield"
	ModeClear       SecurityMode = "clear"
	ModeUnspecified SecurityMode = ""
)

// Mode capability flags per spec 03 §3.3.
const (
	ModeFlagSeal        FeatureFlag = "mode.seal"
	ModeFlagTree        FeatureFlag = "mode.tree"
	ModeFlagCrowd       FeatureFlag = "mode.crowd"
	ModeFlagChannel     FeatureFlag = "mode.channel"
	ModeFlagMediaShield FeatureFlag = "mode.mediashield"
	ModeFlagClear       FeatureFlag = "mode.clear"
)

var modeFromFlag = map[FeatureFlag]SecurityMode{
	ModeFlagSeal:        ModeSeal,
	ModeFlagTree:        ModeTree,
	ModeFlagCrowd:       ModeCrowd,
	ModeFlagChannel:     ModeChannel,
	ModeFlagMediaShield: ModeMediaShield,
	ModeFlagClear:       ModeClear,
}

// ModeFromFlag maps a mode.* capability flag to the corresponding SecurityMode.
func ModeFromFlag(flag FeatureFlag) (SecurityMode, bool) {
	m, ok := modeFromFlag[flag]
	return m, ok
}

// ValidFlagName returns true for valid cap.* or mode.* flags (spec 03 §3.1 + §3.3).
// Lowercase, no whitespace, must start with "cap." or "mode.".
func ValidFlagName(flag string) bool {
	if strings.HasPrefix(flag, "cap.") || strings.HasPrefix(flag, "mode.") {
		return flag == strings.ToLower(flag) && !strings.ContainsAny(flag, " \t\n")
	}
	return false
}

// NegotiateCapabilities delegates to the legacy negotiation which computes
// Accepted, IgnoredRemote, and MissingRequired capability sets.
func NegotiateCapabilities(local, remote, required []FeatureFlag) CapabilityNegResult {
	// Legacy signature: (localSupported []FeatureFlag, remoteAdvertised []string, remoteRequired []string)
	remoteStrs := make([]string, len(remote))
	for i, f := range remote {
		remoteStrs[i] = string(f)
	}
	requiredStrs := make([]string, len(required))
	for i, f := range required {
		requiredStrs[i] = string(f)
	}
	return legacy.NegotiateCapabilities(local, remoteStrs, requiredStrs)
}
