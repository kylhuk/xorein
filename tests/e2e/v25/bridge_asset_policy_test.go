package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/bridge"
)

func TestBridgeAssetMetadataOnlySuccess(t *testing.T) {
	provider := bridge.ProviderCapabilities{SupportsAssetBridges: true}
	bot := bridge.BotCapabilities{AllowsAssetBridge: true}
	decision := bridge.EvaluateAssetBridgeRequest(provider, bot, bridge.TransferModeMetadata)
	if !decision.Allowed {
		t.Fatalf("expected metadata-only bridge request to be allowed")
	}
	if decision.Reason != bridge.ReasonBridgeAllowed {
		t.Fatalf("unexpected reason %q", decision.Reason)
	}
	if decision.Message != bridge.MessageBridgeAllowed {
		t.Fatalf("unexpected message %q", decision.Message)
	}
	if !decision.ForwardMetadataToken {
		t.Fatalf("expected metadata token to forward")
	}
	if decision.ForwardRawBytes {
		t.Fatalf("bridge policy must never forward raw bytes")
	}
}

func TestBridgeAssetRawBlobRefused(t *testing.T) {
	provider := bridge.ProviderCapabilities{SupportsAssetBridges: true}
	bot := bridge.BotCapabilities{AllowsAssetBridge: true}
	decision := bridge.EvaluateAssetBridgeRequest(provider, bot, bridge.TransferModeRaw)
	if decision.Allowed {
		t.Fatalf("raw blob requests must be denied")
	}
	if decision.Reason != bridge.ReasonRawBlobNotAllowed {
		t.Fatalf("unexpected reason %q", decision.Reason)
	}
	if decision.Message != bridge.MessageRawBlobNotAllowed {
		t.Fatalf("unexpected message %q", decision.Message)
	}
	if decision.ForwardMetadataToken {
		t.Fatalf("denied requests must not forward tokens")
	}
	if decision.ForwardRawBytes {
		t.Fatalf("bridge policy must never forward raw bytes")
	}
}

func TestBridgeAssetProviderUnsupportedRefused(t *testing.T) {
	provider := bridge.ProviderCapabilities{SupportsAssetBridges: false}
	bot := bridge.BotCapabilities{AllowsAssetBridge: true}
	decision := bridge.EvaluateAssetBridgeRequest(provider, bot, bridge.TransferModeMetadata)
	if decision.Allowed {
		t.Fatalf("unsupported providers must deny bridge requests")
	}
	if decision.Reason != bridge.ReasonProviderBridgeUnsupported {
		t.Fatalf("unexpected reason %q", decision.Reason)
	}
	if decision.Message != bridge.MessageProviderBridgeUnsupported {
		t.Fatalf("unexpected message %q", decision.Message)
	}
	if decision.ForwardMetadataToken {
		t.Fatalf("denied requests must not forward tokens")
	}
	if decision.ForwardRawBytes {
		t.Fatalf("denied requests must never forward raw bytes")
	}
}

func TestBridgeAssetBotCapabilityDeniedRefused(t *testing.T) {
	provider := bridge.ProviderCapabilities{SupportsAssetBridges: true}
	bot := bridge.BotCapabilities{AllowsAssetBridge: false}
	decision := bridge.EvaluateAssetBridgeRequest(provider, bot, bridge.TransferModeMetadata)
	if decision.Allowed {
		t.Fatalf("bot capability denial must refuse the request")
	}
	if decision.Reason != bridge.ReasonBotBridgeDenied {
		t.Fatalf("unexpected reason %q", decision.Reason)
	}
	if decision.Message != bridge.MessageBotBridgeDenied {
		t.Fatalf("unexpected message %q", decision.Message)
	}
	if decision.ForwardMetadataToken {
		t.Fatalf("denied requests must not forward tokens")
	}
	if decision.ForwardRawBytes {
		t.Fatalf("denied requests must never forward raw bytes")
	}
}
