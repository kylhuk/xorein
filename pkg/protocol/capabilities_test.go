package protocol

import (
	"reflect"
	"testing"
)

func TestValidFeatureFlagName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "valid chat", input: "cap.chat", valid: true},
		{name: "valid dotted", input: "cap.voice.low-latency", valid: true},
		{name: "valid numeric segment", input: "cap.voice.v2", valid: true},
		{name: "valid mode flag", input: "mode.tree", valid: true},
		{name: "missing prefix", input: "chat", valid: false},
		{name: "upper-case", input: "cap.Chat", valid: false},
		{name: "empty tail", input: "cap.", valid: false},
		{name: "mode trailing separator", input: "mode.tree.", valid: false},
		{name: "invalid rune", input: "cap.voice/opus", valid: false},
		{name: "double dot", input: "cap.voice..opus", valid: false},
		{name: "double hyphen separator", input: "cap.voice--opus", valid: false},
		{name: "leading separator", input: "cap.-voice", valid: false},
		{name: "trailing separator", input: "cap.voice-", valid: false},
		{name: "unicode letter rejected", input: "cap.voicé", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidFeatureFlagName(tt.input); got != tt.valid {
				t.Fatalf("ValidFeatureFlagName(%q)=%v want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestNegotiateCapabilities(t *testing.T) {
	local := []FeatureFlag{FeatureChat, FeatureVoice, FeatureIdentity}

	t.Run("accepts-known-ignores-unknown-advertised", func(t *testing.T) {
		result := NegotiateCapabilities(
			local,
			[]string{"cap.chat", "cap.voice", "cap.future", "bad"},
			nil,
		)

		wantAccepted := []FeatureFlag{FeatureChat, FeatureVoice}
		if !reflect.DeepEqual(result.Accepted, wantAccepted) {
			t.Fatalf("accepted mismatch: got %#v want %#v", result.Accepted, wantAccepted)
		}

		wantIgnored := []string{"bad", "cap.future"}
		if !reflect.DeepEqual(result.IgnoredRemote, wantIgnored) {
			t.Fatalf("ignored mismatch: got %#v want %#v", result.IgnoredRemote, wantIgnored)
		}

		if len(result.MissingRequired) != 0 {
			t.Fatalf("unexpected missing required: %#v", result.MissingRequired)
		}

		if result.Feedback != CapabilityFeedbackRemoteFeaturesIgnored {
			t.Fatalf("feedback mismatch: got %q want %q", result.Feedback, CapabilityFeedbackRemoteFeaturesIgnored)
		}
	})

	t.Run("fails-on-unsupported-required", func(t *testing.T) {
		result := NegotiateCapabilities(
			local,
			[]string{"cap.chat", "cap.voice"},
			[]string{"cap.identity", "cap.sync", "bad", "cap.voice..opus"},
		)

		wantAccepted := []FeatureFlag{FeatureChat, FeatureVoice}
		if !reflect.DeepEqual(result.Accepted, wantAccepted) {
			t.Fatalf("accepted mismatch: got %#v want %#v", result.Accepted, wantAccepted)
		}

		wantMissing := []string{"bad", "cap.sync", "cap.voice..opus"}
		if !reflect.DeepEqual(result.MissingRequired, wantMissing) {
			t.Fatalf("missing mismatch: got %#v want %#v", result.MissingRequired, wantMissing)
		}

		if result.Feedback != CapabilityFeedbackUpgradeRequired {
			t.Fatalf("feedback mismatch: got %q want %q", result.Feedback, CapabilityFeedbackUpgradeRequired)
		}
	})

	t.Run("deduplicates-and-trims-capability-inputs", func(t *testing.T) {
		result := NegotiateCapabilities(
			local,
			[]string{" cap.chat ", "cap.chat", "", "cap.unknown", "cap.unknown"},
			[]string{" cap.sync ", "cap.sync", " ", "bad"},
		)

		if !reflect.DeepEqual(result.Accepted, []FeatureFlag{FeatureChat}) {
			t.Fatalf("accepted mismatch: got %#v", result.Accepted)
		}
		if !reflect.DeepEqual(result.IgnoredRemote, []string{"cap.unknown"}) {
			t.Fatalf("ignored mismatch: got %#v", result.IgnoredRemote)
		}
		if !reflect.DeepEqual(result.MissingRequired, []string{"bad", "cap.sync"}) {
			t.Fatalf("missing mismatch: got %#v", result.MissingRequired)
		}
		if result.Feedback != CapabilityFeedbackUpgradeRequired {
			t.Fatalf("feedback mismatch: got %q want %q", result.Feedback, CapabilityFeedbackUpgradeRequired)
		}
	})
}

func TestNegotiateConversationSecurityModeV02(t *testing.T) {
	result := NegotiateConversationSecurityMode(
		[]SecurityMode{SecurityModeSeal, SecurityModeTree},
		[]SecurityMode{SecurityModeTree, SecurityModeClear},
	)
	if result.Mode != SecurityModeTree {
		t.Fatalf("unexpected mode: %q", result.Mode)
	}
	if result.Reason != ModeNegotiationReasonMatched {
		t.Fatalf("unexpected reason: %q", result.Reason)
	}

	rejected := NegotiateConversationSecurityMode(
		[]SecurityMode{SecurityModeSeal},
		[]SecurityMode{SecurityModeClear},
	)
	if rejected.Mode != SecurityModeUnspecified {
		t.Fatalf("expected unspecified mode, got %q", rejected.Mode)
	}
	if rejected.Reason != ModeNegotiationReasonUnsupported {
		t.Fatalf("unexpected reason: %q", rejected.Reason)
	}
}

func TestValidFeatureFlagNameAcceptsArchivistCapability(t *testing.T) {
	if !ValidFeatureFlagName(string(FeatureArchivist)) {
		t.Fatalf("expected %q to be a valid feature flag", FeatureArchivist)
	}
}

func TestDefaultPeerTransportFeatureFlagsAreValid(t *testing.T) {
	flags := DefaultPeerTransportFeatureFlags()
	if len(flags) == 0 {
		t.Fatal("expected peer transport feature flags")
	}
	for _, flag := range flags {
		if !ValidFeatureFlagName(string(flag)) {
			t.Fatalf("expected %q to be a valid feature flag", flag)
		}
	}
}
