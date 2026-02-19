package v26

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v12/backup"
	"github.com/aether/code_aether/pkg/v12/identity"
	"github.com/aether/code_aether/pkg/v13/chat"
	"github.com/aether/code_aether/pkg/v13/joinpolicy"
	"github.com/aether/code_aether/pkg/v13/spaces"
	"github.com/aether/code_aether/pkg/v14/voice"
	"github.com/aether/code_aether/pkg/v15/screenshare"
	"github.com/aether/code_aether/pkg/v25/assets"
	"github.com/aether/code_aether/pkg/v25/bridge"
)

func TestRegressionMatrixIdentityJourney(t *testing.T) {
	seed := bytes.Repeat([]byte{0x01}, identity.SeedSize)
	now := time.Unix(1_800_000_000, 0).UTC()
	record, _, err := identity.CreateFromSeed(seed, now, "")
	if err != nil {
		t.Fatalf("create identity: %v", err)
	}
	if record.MetadataVersion != identity.CurrentMetadataVersion {
		t.Fatalf("metadata version mismatch: %d", record.MetadataVersion)
	}
	envelope, err := backup.Export(backup.Payload{Identity: record}, "correct horse battery staple", now)
	if err != nil {
		t.Fatalf("export backup: %v", err)
	}
	data, err := backup.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	restored, err := backup.Restore(data, "correct horse battery staple", nil)
	if err != nil {
		t.Fatalf("restore backup: %v", err)
	}
	if restored.Identity.IdentityID != record.IdentityID {
		t.Fatalf("restored identity id mismatch: %s vs %s", restored.Identity.IdentityID, record.IdentityID)
	}
}

func TestRegressionMatrixSpaceJourney(t *testing.T) {
	policy, err := joinpolicy.ParseMode(string(joinpolicy.ModeInviteOnly))
	if err != nil {
		t.Fatalf("parse policy: %v", err)
	}
	space, err := spaces.NewSpace("space-001", "Matrix Labs", "alice", policy)
	if err != nil {
		t.Fatalf("create space: %v", err)
	}
	if !space.IsMember("alice") {
		t.Fatalf("founder should be a member")
	}
	if err := space.AddMember("bob"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if err := space.PromoteToAdmin("bob"); err != nil {
		t.Fatalf("promote admin: %v", err)
	}
	if err := space.TransferFounder("bob"); err != nil {
		t.Fatalf("transfer founder: %v", err)
	}
	if err := space.Validate(); err != nil {
		t.Fatalf("validate space: %v", err)
	}
	if err := joinpolicy.ValidateRequest(policy, "carol", ""); !errors.Is(err, joinpolicy.ErrInviteTokenMissing) {
		t.Fatalf("expected invite token missing, got %v", err)
	}
	if err := joinpolicy.ValidateRequest(policy, "carol", "invite-123"); err != nil {
		t.Fatalf("validate join request: %v", err)
	}
}

func TestRegressionMatrixChatJourney(t *testing.T) {
	now := time.Unix(1_800_000_100, 0).UTC()
	msg, err := chat.NewMessage("msg-123", "channel-1", "alice", "hi matrix", now)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	if msg.State != chat.DeliveryStatePending {
		t.Fatalf("unexpected state: %s", msg.State)
	}
	if err := msg.Acknowledge(); err != nil {
		t.Fatalf("acknowledge message: %v", err)
	}
	if err := msg.Fail("network"); !errors.Is(err, chat.ErrInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
	var marker chat.ReadMarker
	if err := marker.Touch("msg-123", "bob", now); err != nil {
		t.Fatalf("touch read marker: %v", err)
	}
	if marker.LatestMessage != "msg-123" {
		t.Fatalf("marker latest message mismatch: %s", marker.LatestMessage)
	}
}

func TestRegressionMatrixMediaJourney(t *testing.T) {
	session := voice.NewSession()
	codec, transport, err := session.Negotiate([]string{"opus/48000/2"}, []string{"udp", "tcp"})
	if err != nil {
		t.Fatalf("negotiate voice: %v", err)
	}
	if codec != "opus/48000/2" || transport != "udp" {
		t.Fatalf("unexpected negotiation result: %s %s", codec, transport)
	}
	if next := session.AdvanceFallback(); next != voice.FallbackMesh {
		t.Fatalf("unexpected fallback step: %s", next)
	}
	if backoff := session.Reconnect(); backoff != 250*time.Millisecond {
		t.Fatalf("expect first backoff 250ms, got %s", backoff)
	}
	if session.CurrentFallback() != voice.FallbackMesh {
		t.Fatalf("fallback should remain mesh, got %s", session.CurrentFallback())
	}
	if screenshare.Transition(screenshare.StateIdle, screenshare.EventStart) != screenshare.StateConnecting {
		t.Fatalf("start transition broken")
	}
	desc := screenshare.Summarize(screenshare.StateActive, 2600)
	if desc.Decision.Layer != 2 {
		t.Fatalf("unexpected screen share layer: %d", desc.Decision.Layer)
	}
	if label := screenshare.Label(desc.State, desc.Decision); label != "state=active.layer=2.bitrate=2500kbps" {
		t.Fatalf("unexpected label: %s", label)
	}
}

func TestRegressionMatrixAssetJourney(t *testing.T) {
	preview := assets.PlanPreview(assets.KindAttachment, true)
	if preview.State != assets.StatePreview {
		t.Fatalf("unexpected preview state: %s", preview.State)
	}
	down := assets.PlanDownload(assets.KindAttachment, true)
	if down.State != assets.StateDownload {
		t.Fatalf("unexpected download state: %s", down.State)
	}
	if down.Action != "download" {
		t.Fatalf("unexpected download action: %s", down.Action)
	}
	degraded := assets.PlanPreview(assets.KindEmoji, false)
	if degraded.State != assets.StateDegraded {
		t.Fatalf("expected degraded state, got %s", degraded.State)
	}
	if degraded.Reason != assets.ReasonMissingBlob {
		t.Fatalf("unexpected degrade reason: %s", degraded.Reason)
	}
}

func TestRegressionMatrixBridgeJourney(t *testing.T) {
	allowed := bridge.EvaluateAssetBridgeRequest(
		bridge.ProviderCapabilities{SupportsAssetBridges: true},
		bridge.BotCapabilities{AllowsAssetBridge: true},
		bridge.TransferModeMetadata,
	)
	if !allowed.Allowed || allowed.Reason != bridge.ReasonBridgeAllowed {
		t.Fatalf("expected metadata bridge allowed, got %v", allowed)
	}
	raw := bridge.EvaluateAssetBridgeRequest(
		bridge.ProviderCapabilities{SupportsAssetBridges: true},
		bridge.BotCapabilities{AllowsAssetBridge: true},
		bridge.TransferModeRaw,
	)
	if raw.Allowed || raw.Reason != bridge.ReasonRawBlobNotAllowed {
		t.Fatalf("expected raw bridge refusal, got %v", raw)
	}
	providerDenied := bridge.EvaluateAssetBridgeRequest(
		bridge.ProviderCapabilities{SupportsAssetBridges: false},
		bridge.BotCapabilities{AllowsAssetBridge: true},
		bridge.TransferModeMetadata,
	)
	if providerDenied.Allowed || providerDenied.Reason != bridge.ReasonProviderBridgeUnsupported {
		t.Fatalf("expected provider refusal, got %v", providerDenied)
	}
	agentDenied := bridge.EvaluateAssetBridgeRequest(
		bridge.ProviderCapabilities{SupportsAssetBridges: true},
		bridge.BotCapabilities{AllowsAssetBridge: false},
		bridge.TransferModeMetadata,
	)
	if agentDenied.Allowed || agentDenied.Reason != bridge.ReasonBotBridgeDenied {
		t.Fatalf("expected bot refusal, got %v", agentDenied)
	}
}
