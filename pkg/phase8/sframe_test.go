package phase8

import (
	"errors"
	"testing"

	"github.com/aether/code_aether/pkg/phase7"
)

func TestSFrameKeyDistributorIssueAndRefresh(t *testing.T) {
	channel, err := phase7.NewChannelModel("srv-a")
	if err != nil {
		t.Fatalf("NewChannelModel() error = %v", err)
	}
	if _, err := channel.RegisterChannel(phase7.ChannelMetadata{ID: phase7.ChannelID("voice_main")}); err != nil {
		t.Fatalf("RegisterChannel() error = %v", err)
	}
	if _, err := channel.Join(phase7.ParticipantID("alice"), phase7.ChannelID("voice_main")); err != nil {
		t.Fatalf("Join(alice) error = %v", err)
	}

	distributor, err := NewSFrameKeyDistributor(channel, phase7.NewBootstrapper())
	if err != nil {
		t.Fatalf("NewSFrameKeyDistributor() error = %v", err)
	}

	record, err := distributor.Issue("voice-session-1", "srv-a", phase7.ChannelID("voice_main"), phase7.ParticipantID("alice"))
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if record.Epoch != 0 {
		t.Fatalf("Issue().Epoch = %d, want 0", record.Epoch)
	}
	if len(record.KeyMaterial) != 32 {
		t.Fatalf("Issue().KeyMaterial length = %d, want 32", len(record.KeyMaterial))
	}

	updatedEpoch, err := distributor.RefreshForParticipantChange("voice-session-1", "srv-a", phase7.ChannelID("voice_main"))
	if err != nil {
		t.Fatalf("RefreshForParticipantChange() error = %v", err)
	}
	if updatedEpoch != 1 {
		t.Fatalf("RefreshForParticipantChange() epoch = %d, want 1", updatedEpoch)
	}

	current, ok := distributor.Current("voice-session-1", phase7.ParticipantID("alice"))
	if !ok {
		t.Fatalf("Current() missing record after refresh")
	}
	if current.Epoch != 1 {
		t.Fatalf("Current().Epoch = %d, want 1", current.Epoch)
	}
	if current.KeyID == record.KeyID {
		t.Fatalf("Current().KeyID = %q, want different key id after refresh", current.KeyID)
	}
}

func TestSFrameKeyDistributorValidationAndMembershipGuards(t *testing.T) {
	channel, err := phase7.NewChannelModel("srv-a")
	if err != nil {
		t.Fatalf("NewChannelModel() error = %v", err)
	}
	if _, err := channel.RegisterChannel(phase7.ChannelMetadata{ID: phase7.ChannelID("voice_main")}); err != nil {
		t.Fatalf("RegisterChannel() error = %v", err)
	}

	distributor, err := NewSFrameKeyDistributor(channel, phase7.NewBootstrapper())
	if err != nil {
		t.Fatalf("NewSFrameKeyDistributor() error = %v", err)
	}

	tests := []struct {
		name          string
		voiceSession  string
		serverID      string
		channelID     phase7.ChannelID
		participantID phase7.ParticipantID
		wantErr       error
	}{
		{name: "missing voice session", voiceSession: "", serverID: "srv-a", channelID: phase7.ChannelID("voice_main"), participantID: phase7.ParticipantID("alice"), wantErr: ErrSFrameSessionRequired},
		{name: "missing server", voiceSession: "voice-session-1", serverID: "", channelID: phase7.ChannelID("voice_main"), participantID: phase7.ParticipantID("alice"), wantErr: ErrSFrameServerRequired},
		{name: "missing channel", voiceSession: "voice-session-1", serverID: "srv-a", channelID: "", participantID: phase7.ParticipantID("alice"), wantErr: ErrSFrameChannelRequired},
		{name: "missing participant", voiceSession: "voice-session-1", serverID: "srv-a", channelID: phase7.ChannelID("voice_main"), participantID: "", wantErr: ErrSFrameParticipantRequired},
		{name: "not a member", voiceSession: "voice-session-1", serverID: "srv-a", channelID: phase7.ChannelID("voice_main"), participantID: phase7.ParticipantID("alice"), wantErr: ErrSFrameUnauthorized},
	}

	for _, tc := range tests {
		_, err := distributor.Issue(tc.voiceSession, tc.serverID, tc.channelID, tc.participantID)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("Issue(%s) error = %v, want %v", tc.name, err, tc.wantErr)
		}
	}
}
