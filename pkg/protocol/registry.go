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
}

type ProtocolID struct {
	Family  ProtocolFamily
	Version ProtocolVersion
	Name    string
}

func (p ProtocolID) String() string {
	return fmt.Sprintf("%s/%s/%d.%d", multistreamNamespace, strings.ToLower(string(p.Family)), p.Version.Major, p.Version.Minor)
}

var canonicalRegistry = map[ProtocolFamily][]ProtocolID{}

func registerCanonical(id ProtocolID) {
	if id.Name == "" {
		panic("protocol id requires name")
	}
	canonicalRegistry[id.Family] = append(canonicalRegistry[id.Family], id)
	sort.SliceStable(canonicalRegistry[id.Family], func(i, j int) bool {
		if canonicalRegistry[id.Family][i].Version.Major != canonicalRegistry[id.Family][j].Version.Major {
			return canonicalRegistry[id.Family][i].Version.Major > canonicalRegistry[id.Family][j].Version.Major
		}
		return canonicalRegistry[id.Family][i].Version.Minor > canonicalRegistry[id.Family][j].Version.Minor
	})
}

var (
	chatV01     = ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1}, Name: "chat/v0.1"}
	voiceV01    = ProtocolID{Family: FamilyVoice, Version: ProtocolVersion{Major: 0, Minor: 1}, Name: "voice/v0.1"}
	manifestV01 = ProtocolID{Family: FamilyManifest, Version: ProtocolVersion{Major: 0, Minor: 1}, Name: "manifest/v0.1"}
	identityV01 = ProtocolID{Family: FamilyIdentity, Version: ProtocolVersion{Major: 0, Minor: 1}, Name: "identity/v0.1"}
	syncV01     = ProtocolID{Family: FamilySync, Version: ProtocolVersion{Major: 0, Minor: 1}, Name: "sync/v0.1"}
	dmV02       = ProtocolID{Family: FamilyDM, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "dm/v0.2"}
	groupDMV02  = ProtocolID{Family: FamilyGroupDM, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "groupdm/v0.2"}
	friendsV02  = ProtocolID{Family: FamilyFriends, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "friends/v0.2"}
	presenceV02 = ProtocolID{Family: FamilyPresence, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "presence/v0.2"}
	notifyV02   = ProtocolID{Family: FamilyNotify, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "notify/v0.2"}
	modV02      = ProtocolID{Family: FamilyModeration, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "moderation/v0.2"}
	govV02      = ProtocolID{Family: FamilyGovernance, Version: ProtocolVersion{Major: 0, Minor: 2}, Name: "governance/v0.2"}
)

func init() {
	registerCanonical(chatV01)
	registerCanonical(voiceV01)
	registerCanonical(manifestV01)
	registerCanonical(identityV01)
	registerCanonical(syncV01)
	registerCanonical(dmV02)
	registerCanonical(groupDMV02)
	registerCanonical(friendsV02)
	registerCanonical(presenceV02)
	registerCanonical(notifyV02)
	registerCanonical(modV02)
	registerCanonical(govV02)
}

func CanonicalByFamily(family ProtocolFamily) []ProtocolID {
	slice := make([]ProtocolID, len(canonicalRegistry[family]))
	copy(slice, canonicalRegistry[family])
	return slice
}

// CanonicalProtocolByVersion returns the canonical entry for exactly matching family and version.
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
	family := ProtocolFamily(strings.ToLower(parts[0]))
	versionParts := strings.Split(parts[1], ".")
	if len(versionParts) != 2 {
		return ProtocolID{}, fmt.Errorf("unexpected version syntax: %s", parts[1])
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return ProtocolID{}, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return ProtocolID{}, fmt.Errorf("invalid minor version: %w", err)
	}
	return ProtocolID{Family: family, Version: ProtocolVersion{Major: major, Minor: minor}, Name: fmt.Sprintf("%s/%d.%d", family, major, minor)}, nil
}

func IntersectByFamily(local, offered ProtocolID) bool {
	if local.Family != offered.Family {
		return false
	}
	return local.Version.Major == offered.Version.Major && offered.Version.Minor <= local.Version.Minor
}

// CompatibilityPolicy decides whether a candidate protocol can satisfy a remote offer.
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
	return id.Version.Major == anchor.Major && id.Version.Minor <= anchor.Minor
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
