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
	FeatureChat          FeatureFlag = "cap.chat"
	FeatureVoice         FeatureFlag = "cap.voice"
	FeatureManagement    FeatureFlag = "cap.management"
	FeatureManifest      FeatureFlag = "cap.manifest"
	FeatureIdentity      FeatureFlag = "cap.identity"
	FeatureSync          FeatureFlag = "cap.sync"
	FeaturePeerTransport FeatureFlag = "cap.peer.transport"
	FeaturePeerMetadata  FeatureFlag = "cap.peer.metadata"
	FeaturePeerBootstrap FeatureFlag = "cap.peer.bootstrap"
	FeaturePeerManifest  FeatureFlag = "cap.peer.manifest"
	FeaturePeerJoin      FeatureFlag = "cap.peer.join"
	FeaturePeerDelivery  FeatureFlag = "cap.peer.delivery"
	FeaturePeerRelay     FeatureFlag = "cap.peer.relay"
	FeatureArchivist     FeatureFlag = "cap.archivist"
	FeatureDM            FeatureFlag = "cap.dm"
	FeatureGroupDM       FeatureFlag = "cap.group-dm"
	FeatureFriends       FeatureFlag = "cap.friends"
	FeaturePresence      FeatureFlag = "cap.presence"
	FeatureNotify        FeatureFlag = "cap.notify"
	FeatureMentions      FeatureFlag = "cap.mentions"
	FeatureModeration    FeatureFlag = "cap.moderation"
	FeatureRBAC          FeatureFlag = "cap.rbac"
	FeatureSlowMode      FeatureFlag = "cap.slow-mode"
)

var defaultFeatureFlags = []FeatureFlag{
	FeatureChat,
	FeatureVoice,
	FeatureManagement,
	FeatureManifest,
	FeatureIdentity,
	FeatureSync,
	FeatureDM,
	FeatureGroupDM,
	FeatureFriends,
	FeaturePresence,
	FeatureNotify,
	FeatureMentions,
	FeatureModeration,
	FeatureRBAC,
	FeatureSlowMode,
}

var defaultPeerTransportFeatureFlags = []FeatureFlag{
	FeaturePeerTransport,
	FeaturePeerMetadata,
	FeaturePeerBootstrap,
	FeaturePeerManifest,
	FeaturePeerJoin,
	FeaturePeerDelivery,
	FeaturePeerRelay,
}

// SecurityMode is the negotiated conversation posture label.
type SecurityMode string

const (
	SecurityModeUnspecified SecurityMode = "unspecified"
	SecurityModeSeal        SecurityMode = "seal"
	SecurityModeTree        SecurityMode = "tree"
	SecurityModeClear       SecurityMode = "clear"
)

// ModeNegotiationReason provides deterministic mode negotiation outcomes.
type ModeNegotiationReason string

const (
	ModeNegotiationReasonMatched     ModeNegotiationReason = "matched"
	ModeNegotiationReasonUnsupported ModeNegotiationReason = "unsupported-mode"
	ModeNegotiationReasonNoOffer     ModeNegotiationReason = "no-offer"
)

// ModeNegotiationResult captures negotiated mode and deterministic reason.
type ModeNegotiationResult struct {
	Mode   SecurityMode
	Reason ModeNegotiationReason
}

// DefaultFeatureFlags returns the canonical v0.1 capability names.
func DefaultFeatureFlags() []FeatureFlag {
	out := make([]FeatureFlag, len(defaultFeatureFlags))
	copy(out, defaultFeatureFlags)
	return out
}

func DefaultPeerTransportFeatureFlags() []FeatureFlag {
	out := make([]FeatureFlag, len(defaultPeerTransportFeatureFlags))
	copy(out, defaultPeerTransportFeatureFlags)
	return out
}

// ValidFeatureFlagName validates v0.1 feature naming conventions.
// Accepts "cap.*" and "mode.*" prefixes (spec 03 §3.1 + §3.3).
func ValidFeatureFlagName(flag string) bool {
	var prefix string
	switch {
	case strings.HasPrefix(flag, "cap."):
		prefix = "cap."
	case strings.HasPrefix(flag, "mode."):
		prefix = "mode."
	default:
		return false
	}
	if flag != strings.ToLower(flag) {
		return false
	}
	tail := strings.TrimPrefix(flag, prefix)
	if tail == "" {
		return false
	}

	lastWasSeparator := true
	for _, r := range tail {
		switch {
		case r >= 'a' && r <= 'z':
			lastWasSeparator = false
		case r >= '0' && r <= '9':
			lastWasSeparator = false
		case r == '.' || r == '-':
			if lastWasSeparator {
				return false
			}
			lastWasSeparator = true
		default:
			return false
		}
	}

	return !lastWasSeparator
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

// NegotiateConversationSecurityMode returns the first shared mode based on
// deterministic local preference order.
func NegotiateConversationSecurityMode(localPreferred []SecurityMode, remoteOffered []SecurityMode) ModeNegotiationResult {
	if len(remoteOffered) == 0 {
		return ModeNegotiationResult{Mode: SecurityModeUnspecified, Reason: ModeNegotiationReasonNoOffer}
	}
	offered := make(map[SecurityMode]struct{}, len(remoteOffered))
	for _, mode := range remoteOffered {
		offered[mode] = struct{}{}
	}
	for _, mode := range localPreferred {
		if mode == SecurityModeUnspecified {
			continue
		}
		if _, ok := offered[mode]; ok {
			return ModeNegotiationResult{Mode: mode, Reason: ModeNegotiationReasonMatched}
		}
	}
	return ModeNegotiationResult{Mode: SecurityModeUnspecified, Reason: ModeNegotiationReasonUnsupported}
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
