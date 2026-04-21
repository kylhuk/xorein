package protocol

import (
	"errors"
	"reflect"
	"testing"
)

func TestNegotiatePeerTransport(t *testing.T) {
	t.Run("supported version and capabilities", func(t *testing.T) {
		result, err := NegotiatePeerTransport(
			[]string{"/aether/peer/0.1.0"},
			[]string{string(FeaturePeerMetadata), string(FeaturePeerManifest), "cap.peer.future"},
			[]string{string(FeaturePeerMetadata), string(FeaturePeerManifest)},
		)
		if err != nil {
			t.Fatalf("NegotiatePeerTransport() error = %v", err)
		}
		if result.Protocol != peerV010 {
			t.Fatalf("protocol = %+v want %+v", result.Protocol, peerV010)
		}
		wantAccepted := []FeatureFlag{FeaturePeerManifest, FeaturePeerMetadata}
		if !reflect.DeepEqual(result.CapabilityResult.Accepted, wantAccepted) {
			t.Fatalf("accepted = %#v want %#v", result.CapabilityResult.Accepted, wantAccepted)
		}
		if result.CapabilityResult.Feedback != CapabilityFeedbackRemoteFeaturesIgnored {
			t.Fatalf("feedback = %q want %q", result.CapabilityResult.Feedback, CapabilityFeedbackRemoteFeaturesIgnored)
		}
	})

	t.Run("unsupported protocol fails closed", func(t *testing.T) {
		_, err := NegotiatePeerTransport([]string{"/aether/peer/1.0.0"}, nil, nil)
		if err == nil {
			t.Fatal("expected unsupported protocol error")
		}
		var negotiationErr *NegotiationError
		if !errors.As(err, &negotiationErr) {
			t.Fatalf("error type = %T want *NegotiationError", err)
		}
		if negotiationErr.Code != NegotiationErrorUnsupportedProtocol {
			t.Fatalf("code = %q want %q", negotiationErr.Code, NegotiationErrorUnsupportedProtocol)
		}
	})

	t.Run("unsupported required capability fails closed", func(t *testing.T) {
		_, err := NegotiatePeerTransport(
			[]string{"/aether/peer/0.1.0"},
			[]string{string(FeaturePeerMetadata)},
			[]string{string(FeaturePeerMetadata), "cap.peer.experimental"},
		)
		if err == nil {
			t.Fatal("expected unsupported capability error")
		}
		var negotiationErr *NegotiationError
		if !errors.As(err, &negotiationErr) {
			t.Fatalf("error type = %T want *NegotiationError", err)
		}
		if negotiationErr.Code != NegotiationErrorUnsupportedCapability {
			t.Fatalf("code = %q want %q", negotiationErr.Code, NegotiationErrorUnsupportedCapability)
		}
		if !reflect.DeepEqual(negotiationErr.MissingRequired, []string{"cap.peer.experimental"}) {
			t.Fatalf("missing required = %#v", negotiationErr.MissingRequired)
		}
	})
}

func TestPeerTransportCanonicalSurface(t *testing.T) {
	if got, want := CanonicalProtocolStrings(FamilyPeer), []string{"/aether/peer/0.1.0"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical peer protocols = %#v want %#v", got, want)
	}
	if got, want := FeatureFlagStrings(DefaultPeerTransportFeatureFlags()), []string{
		string(FeaturePeerBootstrap),
		string(FeaturePeerDelivery),
		string(FeaturePeerJoin),
		string(FeaturePeerManifest),
		string(FeaturePeerMetadata),
		string(FeaturePeerRelay),
		string(FeaturePeerTransport),
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical peer capabilities = %#v want %#v", got, want)
	}
}
