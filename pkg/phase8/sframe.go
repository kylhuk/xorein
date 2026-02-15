package phase8

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"github.com/aether/code_aether/pkg/phase7"
)

var (
	ErrSFrameSessionRequired     = errors.New("phase8: sframe voice session id required")
	ErrSFrameServerRequired      = errors.New("phase8: sframe server id required")
	ErrSFrameChannelRequired     = errors.New("phase8: sframe channel id required")
	ErrSFrameParticipantRequired = errors.New("phase8: sframe participant id required")
	ErrSFrameUnauthorized        = errors.New("phase8: participant is not a channel member")
)

type SFrameKeyRecord struct {
	VoiceSessionID string
	ServerID       string
	ChannelID      phase7.ChannelID
	ParticipantID  phase7.ParticipantID
	Epoch          uint64
	KeyID          string
	KeyMaterial    []byte
}

type SFrameKeyDistributor struct {
	mu           sync.Mutex
	channel      *phase7.ChannelModel
	bootstrap    *phase7.Bootstrapper
	records      map[string]*SFrameKeyRecord
	channelEpoch map[string]uint64
}

func NewSFrameKeyDistributor(channel *phase7.ChannelModel, bootstrap *phase7.Bootstrapper) (*SFrameKeyDistributor, error) {
	if channel == nil {
		return nil, errors.New("phase8: channel model required")
	}
	if bootstrap == nil {
		return nil, errors.New("phase8: bootstrapper required")
	}
	return &SFrameKeyDistributor{
		channel:      channel,
		bootstrap:    bootstrap,
		records:      make(map[string]*SFrameKeyRecord),
		channelEpoch: make(map[string]uint64),
	}, nil
}

func (d *SFrameKeyDistributor) Issue(voiceSessionID, serverID string, channelID phase7.ChannelID, participantID phase7.ParticipantID) (*SFrameKeyRecord, error) {
	if voiceSessionID == "" {
		return nil, ErrSFrameSessionRequired
	}
	if serverID == "" {
		return nil, ErrSFrameServerRequired
	}
	if channelID == "" {
		return nil, ErrSFrameChannelRequired
	}
	if participantID == "" {
		return nil, ErrSFrameParticipantRequired
	}

	member, err := d.channel.HasMember(participantID, channelID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrSFrameUnauthorized
	}

	state, err := d.bootstrap.Bootstrap(participantID)
	if err != nil {
		return nil, fmt.Errorf("phase8: bootstrap participant: %w", err)
	}

	channelKey := channelKey(serverID, channelID)

	d.mu.Lock()
	defer d.mu.Unlock()

	epoch := d.channelEpoch[channelKey]
	record := newSFrameKeyRecord(voiceSessionID, serverID, channelID, participantID, epoch, state.SenderKey)
	d.records[recordKey(voiceSessionID, participantID)] = record
	return cloneSFrameKeyRecord(record), nil
}

func (d *SFrameKeyDistributor) RefreshForParticipantChange(voiceSessionID, serverID string, channelID phase7.ChannelID) (uint64, error) {
	if voiceSessionID == "" {
		return 0, ErrSFrameSessionRequired
	}
	if serverID == "" {
		return 0, ErrSFrameServerRequired
	}
	if channelID == "" {
		return 0, ErrSFrameChannelRequired
	}

	channelKey := channelKey(serverID, channelID)

	d.mu.Lock()
	defer d.mu.Unlock()

	d.channelEpoch[channelKey] = d.channelEpoch[channelKey] + 1
	updatedEpoch := d.channelEpoch[channelKey]

	for key, rec := range d.records {
		if rec.VoiceSessionID != voiceSessionID || rec.ServerID != serverID || rec.ChannelID != channelID {
			continue
		}

		state, err := d.bootstrap.Rotate(rec.ParticipantID)
		if err != nil {
			return 0, fmt.Errorf("phase8: rotate participant key %s: %w", rec.ParticipantID, err)
		}
		d.records[key] = newSFrameKeyRecord(voiceSessionID, serverID, channelID, rec.ParticipantID, updatedEpoch, state.SenderKey)
	}

	return updatedEpoch, nil
}

func (d *SFrameKeyDistributor) Current(voiceSessionID string, participantID phase7.ParticipantID) (*SFrameKeyRecord, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	record, ok := d.records[recordKey(voiceSessionID, participantID)]
	if !ok {
		return nil, false
	}
	return cloneSFrameKeyRecord(record), true
}

func newSFrameKeyRecord(voiceSessionID, serverID string, channelID phase7.ChannelID, participantID phase7.ParticipantID, epoch uint64, senderKey []byte) *SFrameKeyRecord {
	material := deriveSFrameKeyMaterial(voiceSessionID, serverID, channelID, participantID, epoch, senderKey)
	hash := sha256.Sum256(material)
	return &SFrameKeyRecord{
		VoiceSessionID: voiceSessionID,
		ServerID:       serverID,
		ChannelID:      channelID,
		ParticipantID:  participantID,
		Epoch:          epoch,
		KeyID:          fmt.Sprintf("%x", hash[:8]),
		KeyMaterial:    material,
	}
}

func deriveSFrameKeyMaterial(voiceSessionID, serverID string, channelID phase7.ChannelID, participantID phase7.ParticipantID, epoch uint64, senderKey []byte) []byte {
	input := fmt.Sprintf("%s|%s|%s|%s|%d", voiceSessionID, serverID, channelID, participantID, epoch)
	sum := sha256.Sum256(append([]byte(input), senderKey...))
	out := make([]byte, 32)
	copy(out, sum[:])
	return out
}

func channelKey(serverID string, channelID phase7.ChannelID) string {
	return fmt.Sprintf("%s|%s", serverID, channelID)
}

func recordKey(voiceSessionID string, participantID phase7.ParticipantID) string {
	return fmt.Sprintf("%s|%s", voiceSessionID, participantID)
}

func cloneSFrameKeyRecord(in *SFrameKeyRecord) *SFrameKeyRecord {
	if in == nil {
		return nil
	}
	copy := *in
	copy.KeyMaterial = append([]byte(nil), in.KeyMaterial...)
	return &copy
}
