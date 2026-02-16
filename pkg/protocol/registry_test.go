package protocol

import (
	"strings"
	"testing"
)

func TestCanonicalRegistryEntries(t *testing.T) {
	seen := make(map[string]struct{})
	for family, ids := range canonicalRegistry {
		if len(ids) == 0 {
			t.Fatalf("family %s has no canonical IDs", family)
		}
		for _, id := range ids {
			if id.Family == "" {
				t.Fatalf("canonical entry has empty family")
			}
			toks := strings.Split(id.String(), "/")
			if len(toks) != 4 {
				t.Fatalf("unexpected formatted protocol string %q", id.String())
			}
			if toks[0] != "" || toks[1] != strings.TrimPrefix(multistreamNamespace, "/") {
				t.Fatalf("unexpected namespace for %s", id.String())
			}
			if _, ok := seen[id.String()]; ok {
				t.Fatalf("duplicate canonical ID %s", id.String())
			}
			seen[id.String()] = struct{}{}
		}
	}
}

func TestParseProtocolID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		id, err := ParseProtocolID("/aether/chat/0.1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id.Family != FamilyChat || id.Version.Major != 0 || id.Version.Minor != 1 {
			t.Fatalf("wrong parse result: %+v", id)
		}
	})
	t.Run("invalid namespace", func(t *testing.T) {
		if _, err := ParseProtocolID("/bad/chat/0.1"); err == nil {
			t.Fatal("expected namespace error")
		}
	})
	t.Run("invalid version", func(t *testing.T) {
		if _, err := ParseProtocolID("/aether/chat/zero"); err == nil {
			t.Fatal("expected version error")
		}
	})
}

func TestVersionCompatibilityPolicy(t *testing.T) {
	cases := []struct {
		name           string
		policy         VersionCompatibilityPolicy
		candidate      ProtocolID
		offer          ProtocolID
		expectedAllows bool
	}{
		{
			name:           "identical",
			policy:         VersionCompatibilityPolicy{allowMinorDowngrade: true, minimumMinor: 1},
			candidate:      ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1}},
			offer:          ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1}},
			expectedAllows: true,
		},
		{
			name:           "major upgrade rejected",
			policy:         VersionCompatibilityPolicy{},
			candidate:      ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1}},
			offer:          ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 1, Minor: 0}},
			expectedAllows: false,
		},
		{
			name:           "major downgrade allowed",
			policy:         VersionCompatibilityPolicy{allowMajorDowngrade: true},
			candidate:      ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 1, Minor: 0}},
			offer:          ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 9}},
			expectedAllows: true,
		},
		{
			name:           "candidate below minimumMinor",
			policy:         VersionCompatibilityPolicy{minimumMinor: 1},
			candidate:      ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 0}},
			offer:          ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 0}},
			expectedAllows: false,
		},
		{
			name:           "minor downgrade allowed",
			policy:         VersionCompatibilityPolicy{allowMinorDowngrade: true, minimumMinor: 0},
			candidate:      ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 2}},
			offer:          ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 3}},
			expectedAllows: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			requires := tt.policy.Allows(tt.candidate, tt.offer)
			if requires != tt.expectedAllows {
				t.Fatalf("policy allows mismatch: got %v want %v", requires, tt.expectedAllows)
			}
		})
	}
}

func TestDeprecationGuard(t *testing.T) {
	guard := NewDeprecationGuard(map[ProtocolFamily]ProtocolVersion{FamilyChat: {Major: 0, Minor: 1}})
	cases := []struct {
		name       string
		id         ProtocolID
		deprecated bool
	}{
		{"before", ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 0}}, true},
		{"equal", ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 1}}, true},
		{"after", ProtocolID{Family: FamilyChat, Version: ProtocolVersion{Major: 0, Minor: 2}}, false},
		{"other family", ProtocolID{Family: FamilyVoice, Version: ProtocolVersion{Major: 0, Minor: 1}}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if guard.IsDeprecated(tt.id) != tt.deprecated {
				t.Fatalf("unexpected deprecation for %v", tt.id)
			}
		})
	}
}

func TestNegotiateProtocolDeprecationGuard(t *testing.T) {
	original := defaultDeprecationGuard
	defer func() { defaultDeprecationGuard = original }()
	defaultDeprecationGuard = NewDeprecationGuard(map[ProtocolFamily]ProtocolVersion{FamilyChat: {Major: 0, Minor: 1}})
	if _, ok := NegotiateProtocol(FamilyChat, []ProtocolID{chatV01}, nil); ok {
		t.Fatal("expected deprecated candidate to be skipped")
	}
	defaultDeprecationGuard = NewDeprecationGuard(nil)
	if _, ok := NegotiateProtocol(FamilyChat, []ProtocolID{chatV01}, nil); !ok {
		t.Fatal("expected canonical candidate to match")
	}
}

func TestNegotiateProtocolUnknownFamily(t *testing.T) {
	if _, ok := NegotiateProtocol(FamilyChat, []ProtocolID{{Family: FamilyVoice, Version: ProtocolVersion{Major: 0, Minor: 1}}}, nil); ok {
		t.Fatal("expected no negotiation on mismatched families")
	}
}
