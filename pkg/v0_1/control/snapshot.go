package control

import (
	"time"

	voicepkg "github.com/aether/code_aether/pkg/v0_1/family/voice"
)

// VoiceSessionSnapshot is a summary of an active voice session.
type VoiceSessionSnapshot struct {
	ID           string    `json:"id"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
}

// NodeSnapshot is the full local state as returned by GET /v1/state.
type NodeSnapshot struct {
	PeerID        string                  `json:"peer_id"`
	DisplayName   string                  `json:"display_name,omitempty"`
	Servers       []*ServerRecord         `json:"servers"`
	Channels      []*ChannelRecord        `json:"channels"`
	DMs           []*DMRecord             `json:"dms"`
	Messages      []*MessageRecord        `json:"messages"`
	Friends       []*FriendRequestRecord  `json:"friends"`
	FriendReqs    []*FriendRequestRecord  `json:"friend_requests"`
	Notifications []*NotificationRecord   `json:"notifications"`
	ManualPeers   []string                `json:"manual_peers"`
	RelayAddrs    []string                `json:"relay_addrs"`
	VoiceSessions []*VoiceSessionSnapshot `json:"voice_sessions"`
	RelayQueue    map[string]int          `json:"relay_queue"`
	SnapshotAt    time.Time               `json:"snapshot_at"`
}

// Snapshot aggregates state from all registered handlers and the control-layer state.
func (s *Server) Snapshot() NodeSnapshot {
	s.st.mu.RLock()
	snap := NodeSnapshot{
		PeerID:        s.hs.PeerID,
		DisplayName:   s.hs.DisplayName,
		Servers:       append([]*ServerRecord(nil), s.st.servers...),
		Channels:      append([]*ChannelRecord(nil), s.st.channels...),
		DMs:           append([]*DMRecord(nil), s.st.dms...),
		Messages:      append([]*MessageRecord(nil), s.st.messages...),
		Friends:       append([]*FriendRequestRecord(nil), s.st.friendReqs...),
		FriendReqs:    []*FriendRequestRecord{},
		Notifications: append([]*NotificationRecord(nil), s.st.notifications...),
		ManualPeers:   append([]string(nil), s.st.manualPeers...),
		RelayAddrs:    append([]string(nil), s.st.relayAddrs...),
		SnapshotAt:    timeNow(),
	}
	s.st.mu.RUnlock()

	// Filter: friends = accepted, friend_requests = pending
	var accepted, pending []*FriendRequestRecord
	for _, fr := range snap.Friends {
		if fr.Status == "accepted" {
			accepted = append(accepted, fr)
		} else if fr.Status == "pending" {
			pending = append(pending, fr)
		}
	}
	snap.Friends = accepted
	snap.FriendReqs = pending

	// Voice sessions from voice.Handler.
	if s.hs != nil && s.hs.Voice != nil {
		for _, vs := range s.hs.Voice.Sessions() {
			snap.VoiceSessions = append(snap.VoiceSessions, voiceSessionSnap(vs))
		}
	}

	// Relay queue counts from peer.Handler.RelayQueue.
	if s.hs != nil && s.hs.Peer != nil && s.hs.Peer.RelayQueue != nil {
		snap.RelayQueue = s.hs.Peer.RelayQueue.Snapshot()
	}
	if snap.VoiceSessions == nil {
		snap.VoiceSessions = []*VoiceSessionSnapshot{}
	}
	if snap.RelayQueue == nil {
		snap.RelayQueue = map[string]int{}
	}

	return snap
}

func voiceSessionSnap(s *voicepkg.Session) *VoiceSessionSnapshot {
	return &VoiceSessionSnapshot{
		ID:           s.ID,
		Participants: append([]string(nil), s.Participants...),
		CreatedAt:    s.CreatedAt,
	}
}

// timeNow is a variable for testing.
var timeNow = func() time.Time { return time.Now().UTC() }
