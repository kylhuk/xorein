package protocol

import (
	"sort"
	"strings"
)

const featureFlagPrefix = "cap."

// FeatureFlag is the canonical capability naming type exchanged during
// negotiation. v0.1 requires lowercase namespaced names with a "cap." prefix.
type FeatureFlag string

const (
	FeatureChat       FeatureFlag = "cap.chat"
	FeatureVoice      FeatureFlag = "cap.voice"
	FeatureManagement FeatureFlag = "cap.management"
	FeatureManifest   FeatureFlag = "cap.manifest"
	FeatureIdentity   FeatureFlag = "cap.identity"
	FeatureSync       FeatureFlag = "cap.sync"
)

var defaultFeatureFlags = []FeatureFlag{
	FeatureChat,
	FeatureVoice,
	FeatureManagement,
	FeatureManifest,
	FeatureIdentity,
	FeatureSync,
}

// DefaultFeatureFlags returns the canonical v0.1 capability names.
func DefaultFeatureFlags() []FeatureFlag {
	out := make([]FeatureFlag, len(defaultFeatureFlags))
	copy(out, defaultFeatureFlags)
	return out
}

// ValidFeatureFlagName validates v0.1 feature naming conventions.
func ValidFeatureFlagName(flag string) bool {
	if !strings.HasPrefix(flag, featureFlagPrefix) {
		return false
	}
	if flag != strings.ToLower(flag) {
		return false
	}
	tail := strings.TrimPrefix(flag, featureFlagPrefix)
	if tail == "" {
		return false
	}
	for _, r := range tail {
		switch {
		case r >= 'a' && r <= 'z':
			continue
		case r >= '0' && r <= '9':
			continue
		case r == '.' || r == '-':
			continue
		default:
			return false
		}
	}
	return true
}

// CapabilityFeedback is the user-facing deterministic outcome category for
// capability compatibility checks.
type CapabilityFeedback string

const (
	CapabilityFeedbackNone                  CapabilityFeedback = "none"
	CapabilityFeedbackRemoteFeaturesIgnored CapabilityFeedback = "remote-features-ignored"
	CapabilityFeedbackUpgradeRequired       CapabilityFeedback = "upgrade-required"
)

// CapabilityNegotiationResult captures accepted, ignored, and incompatible
// capabilities in deterministic ordering.
type CapabilityNegotiationResult struct {
	Accepted        []FeatureFlag
	IgnoredRemote   []string
	MissingRequired []string
	Feedback        CapabilityFeedback
}

// NegotiateCapabilities evaluates remote advertised/required capabilities
// against local support.
//
// Unknown advertised capabilities are ignored.
// Unknown or unsupported required capabilities are treated as incompatible and
// produce upgrade-required feedback.
func NegotiateCapabilities(localSupported []FeatureFlag, remoteAdvertised []string, remoteRequired []string) CapabilityNegotiationResult {
	localSet := make(map[FeatureFlag]struct{}, len(localSupported))
	for _, flag := range localSupported {
		localSet[flag] = struct{}{}
	}

	acceptedSet := make(map[FeatureFlag]struct{})
	ignoredSet := make(map[string]struct{})
	missingSet := make(map[string]struct{})

	for _, raw := range remoteAdvertised {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if !ValidFeatureFlagName(name) {
			ignoredSet[name] = struct{}{}
			continue
		}
		flag := FeatureFlag(name)
		if _, ok := localSet[flag]; ok {
			acceptedSet[flag] = struct{}{}
			continue
		}
		ignoredSet[name] = struct{}{}
	}

	for _, raw := range remoteRequired {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if !ValidFeatureFlagName(name) {
			missingSet[name] = struct{}{}
			continue
		}
		if _, ok := localSet[FeatureFlag(name)]; !ok {
			missingSet[name] = struct{}{}
		}
	}

	result := CapabilityNegotiationResult{
		Accepted:        sortFeatureFlags(acceptedSet),
		IgnoredRemote:   sortStrings(ignoredSet),
		MissingRequired: sortStrings(missingSet),
		Feedback:        CapabilityFeedbackNone,
	}

	if len(result.MissingRequired) > 0 {
		result.Feedback = CapabilityFeedbackUpgradeRequired
		return result
	}
	if len(result.IgnoredRemote) > 0 {
		result.Feedback = CapabilityFeedbackRemoteFeaturesIgnored
	}
	return result
}

func sortFeatureFlags(set map[FeatureFlag]struct{}) []FeatureFlag {
	out := make([]FeatureFlag, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func sortStrings(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
