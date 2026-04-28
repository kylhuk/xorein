package protocol

import (
	"fmt"
	"strings"
)

type NegotiationErrorCode string

const (
	NegotiationErrorUnsupportedProtocol   NegotiationErrorCode = "unsupported-protocol"
	NegotiationErrorUnsupportedCapability NegotiationErrorCode = "unsupported-capability"
)

type NegotiationError struct {
	Code             NegotiationErrorCode
	Message          string
	OfferedProtocols []string
	MissingRequired  []string
}

func (e *NegotiationError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return string(e.Code)
}

type PeerTransportNegotiation struct {
	Protocol         ProtocolID
	CapabilityResult CapabilityNegotiationResult
}

func CanonicalProtocolStrings(family ProtocolFamily) []string {
	ids := CanonicalByFamily(family)
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func FeatureFlagStrings(flags []FeatureFlag) []string {
	set := make(map[string]struct{}, len(flags))
	for _, flag := range flags {
		name := strings.TrimSpace(string(flag))
		if name == "" {
			continue
		}
		set[name] = struct{}{}
	}
	return sortStrings(set)
}

func NegotiatePeerTransport(offeredProtocols, remoteAdvertised, remoteRequired []string) (PeerTransportNegotiation, error) {
	offers := make([]ProtocolID, 0, len(offeredProtocols))
	trimmedOffers := make([]string, 0, len(offeredProtocols))
	for _, raw := range offeredProtocols {
		candidate := strings.TrimSpace(raw)
		if candidate == "" {
			continue
		}
		trimmedOffers = append(trimmedOffers, candidate)
		parsed, err := ParseProtocolID(candidate)
		if err != nil {
			continue
		}
		offers = append(offers, parsed)
	}

	negotiated, ok := NegotiateProtocol(FamilyPeer, offers, nil)
	if !ok {
		return PeerTransportNegotiation{}, &NegotiationError{
			Code:             NegotiationErrorUnsupportedProtocol,
			Message:          fmt.Sprintf("peer transport negotiation failed for offers %v", trimmedOffers),
			OfferedProtocols: sortStrings(toStringSet(trimmedOffers)),
		}
	}

	capabilities := NegotiateCapabilities(DefaultPeerTransportFeatureFlags(), remoteAdvertised, remoteRequired)
	if len(capabilities.MissingRequired) > 0 {
		return PeerTransportNegotiation{}, &NegotiationError{
			Code:             NegotiationErrorUnsupportedCapability,
			Message:          fmt.Sprintf("peer transport requires unsupported capabilities %v", capabilities.MissingRequired),
			OfferedProtocols: sortStrings(toStringSet(trimmedOffers)),
			MissingRequired:  append([]string(nil), capabilities.MissingRequired...),
		}
	}

	return PeerTransportNegotiation{Protocol: negotiated, CapabilityResult: capabilities}, nil
}

func toStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	return set
}
