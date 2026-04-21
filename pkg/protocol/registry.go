package protocol

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const multistreamNamespace = "/aether"

type ProtocolFamily string

const (
	FamilyChat       ProtocolFamily = "chat"
	FamilyPeer       ProtocolFamily = "peer"
	FamilyVoice      ProtocolFamily = "voice"
	FamilyManifest   ProtocolFamily = "manifest"
	FamilyIdentity   ProtocolFamily = "identity"
	FamilySync       ProtocolFamily = "sync"
	FamilyDM         ProtocolFamily = "dm"
	FamilyGroupDM    ProtocolFamily = "groupdm"
	FamilyFriends    ProtocolFamily = "friends"
	FamilyPresence   ProtocolFamily = "presence"
	FamilyNotify     ProtocolFamily = "notify"
	FamilyModeration ProtocolFamily = "moderation"
	FamilyGovernance ProtocolFamily = "governance"
)

type ProtocolVersion struct {
	Major int
	Minor int
	Patch int
}

type ProtocolID struct {
	Family  ProtocolFamily
	Version ProtocolVersion
	Name    string
}

func (p ProtocolID) String() string {
	return fmt.Sprintf("%s/%s/%d.%d.%d", multistreamNamespace, strings.ToLower(string(p.Family)), p.Version.Major, p.Version.Minor, p.Version.Patch)
}

var canonicalRegistry = map[ProtocolFamily][]ProtocolID{}

func registerCanonical(id ProtocolID) {
	if id.Name == "" {
		panic("protocol id requires name")
	}
	canonicalRegistry[id.Family] = append(canonicalRegistry[id.Family], id)
	sort.SliceStable(canonicalRegistry[id.Family], func(i, j int) bool {
		left, right := canonicalRegistry[id.Family][i].Version, canonicalRegistry[id.Family][j].Version
		if left.Major != right.Major {
			return left.Major > right.Major
		}
		if left.Minor != right.Minor {
			return left.Minor > right.Minor
		}
		return left.Patch > right.Patch
	})
}

var (
	chatV010     = ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "chat/v0.1.0"}
	peerV010     = ProtocolID{Family: FamilyPeer, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "peer/v0.1.0"}
	voiceV010    = ProtocolID{Family: FamilyVoice, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "voice/v0.1.0"}
	manifestV010 = ProtocolID{Family: FamilyManifest, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "manifest/v0.1.0"}
	identityV010 = ProtocolID{Family: FamilyIdentity, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "identity/v0.1.0"}
	syncV010     = ProtocolID{Family: FamilySync, Version: ProtocolVersion{Major: 0, Minor: 1, Patch: 0}, Name: "sync/v0.1.0"}
	dmV020       = ProtocolID{Family: FamilyDM, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "dm/v0.2.0"}
	groupDMV020  = ProtocolID{Family: FamilyGroupDM, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "groupdm/v0.2.0"}
	friendsV020  = ProtocolID{Family: FamilyFriends, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "friends/v0.2.0"}
	presenceV020 = ProtocolID{Family: FamilyPresence, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "presence/v0.2.0"}
	notifyV020   = ProtocolID{Family: FamilyNotify, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "notify/v0.2.0"}
	modV020      = ProtocolID{Family: FamilyModeration, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "moderation/v0.2.0"}
	govV020      = ProtocolID{Family: FamilyGovernance, Version: ProtocolVersion{Major: 0, Minor: 2, Patch: 0}, Name: "governance/v0.2.0"}
)

func init() {
	registerCanonical(chatV010)
	registerCanonical(peerV010)
	registerCanonical(voiceV010)
	registerCanonical(manifestV010)
	registerCanonical(identityV010)
	registerCanonical(syncV010)
	registerCanonical(dmV020)
	registerCanonical(groupDMV020)
	registerCanonical(friendsV020)
	registerCanonical(presenceV020)
	registerCanonical(notifyV020)
	registerCanonical(modV020)
	registerCanonical(govV020)
}

func CanonicalByFamily(family ProtocolFamily) []ProtocolID {
	slice := make([]ProtocolID, len(canonicalRegistry[family]))
	copy(slice, canonicalRegistry[family])
	return slice
}

func CanonicalProtocolByVersion(family ProtocolFamily, version ProtocolVersion) (ProtocolID, bool) {
	for _, id := range canonicalRegistry[family] {
		if id.Version == version {
			return id, true
		}
	}
	return ProtocolID{}, false
}

func ParseProtocolID(input string) (ProtocolID, error) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, multistreamNamespace+"/") {
		return ProtocolID{}, fmt.Errorf("invalid protocol namespace: %s", trimmed)
	}
	parts := strings.Split(strings.TrimPrefix(trimmed, multistreamNamespace+"/"), "/")
	if len(parts) != 2 {
		return ProtocolID{}, fmt.Errorf("malformed protocol identifier: %s", trimmed)
	}
	family := ProtocolFamily(strings.ToLower(strings.TrimSpace(parts[0])))
	if family == "" {
		return ProtocolID{}, fmt.Errorf("protocol family required")
	}
	versionParts := strings.Split(parts[1], ".")
	if len(versionParts) != 3 {
		return ProtocolID{}, fmt.Errorf("unexpected version syntax: %s", parts[1])
	}
	major, err := parseProtocolVersionPart(versionParts[0], "major")
	if err != nil {
		return ProtocolID{}, err
	}
	minor, err := parseProtocolVersionPart(versionParts[1], "minor")
	if err != nil {
		return ProtocolID{}, err
	}
	patch, err := parseProtocolVersionPart(versionParts[2], "patch")
	if err != nil {
		return ProtocolID{}, err
	}
	return ProtocolID{Family: family, Version: ProtocolVersion{Major: major, Minor: minor, Patch: patch}, Name: fmt.Sprintf("%s/%d.%d.%d", family, major, minor, patch)}, nil
}

func parseProtocolVersionPart(part string, label string) (int, error) {
	if part == "" {
		return 0, fmt.Errorf("invalid %s version: empty", label)
	}
	for _, r := range part {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid %s version: %s", label, part)
		}
	}
	value, err := strconv.Atoi(part)
	if err != nil {
		return 0, fmt.Errorf("invalid %s version: %w", label, err)
	}
	return value, nil
}

func IntersectByFamily(local, offered ProtocolID) bool {
	if local.Family != offered.Family {
		return false
	}
	if local.Version.Major != offered.Version.Major {
		return false
	}
	if offered.Version.Minor > local.Version.Minor {
		return false
	}
	if offered.Version.Minor == local.Version.Minor && offered.Version.Patch > local.Version.Patch {
		return false
	}
	return true
}

type CompatibilityPolicy interface {
	Name() string
	Allows(candidate, offer ProtocolID) bool
}

type VersionCompatibilityPolicy struct {
	allowMinorDowngrade bool
	allowMajorDowngrade bool
	minimumMinor        int
}

func DefaultCompatibilityPolicy() CompatibilityPolicy {
	return VersionCompatibilityPolicy{allowMinorDowngrade: true, minimumMinor: 1}
}

func (p VersionCompatibilityPolicy) Name() string {
	return "phase3:compatibility-default"
}

func (p VersionCompatibilityPolicy) Allows(candidate, offer ProtocolID) bool {
	if candidate.Family != offer.Family {
		return false
	}
	if offer.Version.Major > candidate.Version.Major {
		return false
	}
	if offer.Version.Major < candidate.Version.Major {
		return p.allowMajorDowngrade
	}
	if candidate.Version.Minor < p.minimumMinor {
		return false
	}
	if offer.Version.Minor > candidate.Version.Minor {
		return p.allowMinorDowngrade
	}
	if offer.Version.Minor == candidate.Version.Minor && offer.Version.Patch > candidate.Version.Patch {
		return p.allowMinorDowngrade
	}
	return true
}

type DeprecationGuard struct {
	anchors map[ProtocolFamily]ProtocolVersion
}

func NewDeprecationGuard(anchors map[ProtocolFamily]ProtocolVersion) DeprecationGuard {
	copyAnchors := make(map[ProtocolFamily]ProtocolVersion, len(anchors))
	for family, version := range anchors {
		copyAnchors[family] = version
	}
	return DeprecationGuard{anchors: copyAnchors}
}

func (g DeprecationGuard) IsDeprecated(id ProtocolID) bool {
	if len(g.anchors) == 0 {
		return false
	}
	anchor, ok := g.anchors[id.Family]
	if !ok {
		return false
	}
	if id.Version.Major < anchor.Major {
		return true
	}
	if id.Version.Major > anchor.Major {
		return false
	}
	if id.Version.Minor < anchor.Minor {
		return true
	}
	if id.Version.Minor > anchor.Minor {
		return false
	}
	return id.Version.Patch <= anchor.Patch
}

var defaultDeprecationGuard = NewDeprecationGuard(nil)

func NegotiateProtocol(family ProtocolFamily, offers []ProtocolID, policy CompatibilityPolicy) (ProtocolID, bool) {
	if len(offers) == 0 {
		return ProtocolID{}, false
	}
	if policy == nil {
		policy = DefaultCompatibilityPolicy()
	}
	for _, candidate := range canonicalRegistry[family] {
		if defaultDeprecationGuard.IsDeprecated(candidate) {
			continue
		}
		for _, offer := range offers {
			if candidate.Family != offer.Family {
				continue
			}
			if !policy.Allows(candidate, offer) {
				continue
			}
			return candidate, true
		}
	}
	return ProtocolID{}, false
}
