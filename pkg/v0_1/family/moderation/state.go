package moderation

import (
	"fmt"
	"sync"
	"time"
)

// MaxBanListSize is the implementation limit per spec 50 §7.
const MaxBanListSize = 4096

// MaxMuteDurationMS is the max mute duration (30 days) per spec 50 §3.5.
const MaxMuteDurationMS = 2_592_000_000

// MaxSlowModeMS is the max slow-mode delay (6 hours) per spec 50 §3.6.
const MaxSlowModeMS = 21_600_000

// ServerRecord holds per-server moderation state.
type ServerRecord struct {
	Members       []string
	BannedPeerIDs []string
	SlowModeMS    map[string]uint64 // channelID → minimum delay ms
}

// MessageRecord is a minimal view of a chat message for moderation purposes.
type MessageRecord struct {
	ID      string
	Deleted bool
}

// State is the thread-safe moderation state for all servers.
type State struct {
	mu       sync.Mutex
	servers  map[string]*ServerRecord        // serverID → ServerRecord
	mutes    map[string]map[string]time.Time // serverID → peerID → expiry (ephemeral, not persisted)
	messages map[string]*MessageRecord       // messageID → MessageRecord
}

func newState() *State {
	return &State{
		servers:  make(map[string]*ServerRecord),
		mutes:    make(map[string]map[string]time.Time),
		messages: make(map[string]*MessageRecord),
	}
}

func (s *State) server(serverID string) *ServerRecord {
	srv := s.servers[serverID]
	if srv == nil {
		srv = &ServerRecord{SlowModeMS: make(map[string]uint64)}
		s.servers[serverID] = srv
	}
	return srv
}

// AddMember adds a peer to the server member list.
func (s *State) AddMember(serverID, peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	for _, m := range srv.Members {
		if m == peerID {
			return
		}
	}
	srv.Members = append(srv.Members, peerID)
}

// Kick removes a peer from the member list. Returns error if not found.
func (s *State) Kick(serverID, peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	for i, m := range srv.Members {
		if m == peerID {
			srv.Members = append(srv.Members[:i], srv.Members[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

// Ban adds a peer to the ban list.
func (s *State) Ban(serverID, peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	if len(srv.BannedPeerIDs) >= MaxBanListSize {
		return fmt.Errorf("ban list full")
	}
	for _, b := range srv.BannedPeerIDs {
		if b == peerID {
			return nil // already banned
		}
	}
	srv.BannedPeerIDs = append(srv.BannedPeerIDs, peerID)
	return nil
}

// Unban removes a peer from the ban list.
func (s *State) Unban(serverID, peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	for i, b := range srv.BannedPeerIDs {
		if b == peerID {
			srv.BannedPeerIDs = append(srv.BannedPeerIDs[:i], srv.BannedPeerIDs[i+1:]...)
			return
		}
	}
}

// Mute sets a timed mute for peerID in serverID.
func (s *State) Mute(serverID, peerID string, durationMS uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mutes[serverID] == nil {
		s.mutes[serverID] = make(map[string]time.Time)
	}
	s.mutes[serverID][peerID] = time.Now().Add(time.Duration(durationMS) * time.Millisecond)
}

// IsMuted returns true when peerID is currently muted in serverID.
func (s *State) IsMuted(serverID, peerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.mutes[serverID]
	if m == nil {
		return false
	}
	expiry, ok := m[peerID]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(m, peerID)
		return false
	}
	return true
}

// SetSlowMode updates the per-channel slow-mode delay.
func (s *State) SetSlowMode(serverID, channelID string, minDelayMS uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.server(serverID)
	srv.SlowModeMS[channelID] = minDelayMS
}

// SlowModeMS returns the current slow-mode delay for a channel (0 = disabled).
func (s *State) GetSlowModeMS(serverID, channelID string) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.servers[serverID]
	if srv == nil {
		return 0
	}
	return srv.SlowModeMS[channelID]
}

// StoreMessage adds a message for later deletion.
func (s *State) StoreMessage(messageID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[messageID] = &MessageRecord{ID: messageID}
}

// DeleteMessage marks a message as deleted.
func (s *State) DeleteMessage(messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.messages[messageID]
	if m == nil {
		return fmt.Errorf("not found")
	}
	m.Deleted = true
	return nil
}

// IsDeleted reports whether a message has been remotely deleted.
func (s *State) IsDeleted(messageID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.messages[messageID]
	return m != nil && m.Deleted
}

// IsBanned checks if peerID is in the ban list.
func (s *State) IsBanned(serverID, peerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	srv := s.servers[serverID]
	if srv == nil {
		return false
	}
	for _, b := range srv.BannedPeerIDs {
		if b == peerID {
			return true
		}
	}
	return false
}
