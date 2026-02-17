package securitymode

import "testing"

func TestDefaultModeFor(t *testing.T) {
	tests := []struct {
		channelType string
		want        ChannelMode
	}{
		{channelType: "tree", want: ModeTree},
		{channelType: "crowd", want: ModeCrowd},
		{channelType: "channel", want: ModeChannel},
		{channelType: "clear", want: ModeClear},
		{channelType: "e2ee", want: ModeE2EE},
		{channelType: "unmapped", want: ModeChannel},
	}

	for _, tt := range tests {
		t.Run(tt.channelType, func(t *testing.T) {
			if got := DefaultModeFor(tt.channelType); got != tt.want {
				t.Fatalf("DefaultModeFor(%s): got %s want %s", tt.channelType, got, tt.want)
			}
		})
	}
}

func TestAllowsTransition(t *testing.T) {
	tests := []struct {
		name string
		from ChannelMode
		to   ChannelMode
		want bool
	}{
		{name: "tree to crowd", from: ModeTree, to: ModeCrowd, want: true},
		{name: "crowd to clear", from: ModeCrowd, to: ModeClear, want: true},
		{name: "clear to crowd invalid", from: ModeClear, to: ModeCrowd, want: false},
		{name: "unknown from", from: ChannelMode("mystery"), to: ModeClear, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AllowsTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf("AllowsTransition(%s, %s): got %t want %t", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestDisclosureFor(t *testing.T) {
	cases := []struct {
		mode        ChannelMode
		description string
	}{
		{mode: ModeTree, description: "Topology-aware disclosure"},
		{mode: ModeCrowd, description: "Crowd policy mention"},
		{mode: ModeChannel, description: "Channel override disclosure"},
		{mode: ModeClear, description: "Clear channel audit-ready state"},
		{mode: ModeE2EE, description: "E2EE summary evidence"},
		{mode: ChannelMode("mystery"), description: "Unknown mode"},
	}

	for _, tt := range cases {
		t.Run(string(tt.mode), func(t *testing.T) {
			got := DisclosureFor(tt.mode)
			if got.Description != tt.description {
				t.Fatalf("DisclosureFor(%s): got %q want %q", tt.mode, got.Description, tt.description)
			}
			if got.Mode != tt.mode {
				t.Fatalf("mode mismatch: got %s want %s", got.Mode, tt.mode)
			}
		})
	}
}
