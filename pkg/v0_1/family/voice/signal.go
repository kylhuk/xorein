package voice

import (
	"sync"
	"time"
)

// SignalState tracks the last seen sequence number for replay prevention.
type SignalState struct {
	mu      sync.Mutex
	entries map[string]map[string]uint64 // sessionID → peerID → lastSeq
}

func newSignalState() *SignalState {
	return &SignalState{entries: make(map[string]map[string]uint64)}
}

// Check validates and records a new signal sequence number.
// Returns true if the sequence is valid (strictly greater than previous).
func (s *SignalState) Check(sessionID, peerID string, seq uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries[sessionID] == nil {
		s.entries[sessionID] = make(map[string]uint64)
	}
	last, seen := s.entries[sessionID][peerID]
	if seen && seq <= last {
		return false
	}
	s.entries[sessionID][peerID] = seq
	return true
}

// Reset clears signal state for a session (used on restart).
func (s *SignalState) Reset(sessionID, peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries[sessionID] != nil {
		delete(s.entries[sessionID], peerID)
	}
}

// FrameCounterState tracks per-session/per-sender frame counters for replay rejection.
type FrameCounterState struct {
	mu      sync.Mutex
	counters map[string]map[string]uint64 // sessionID → senderID → lastCounter
}

func newFrameCounterState() *FrameCounterState {
	return &FrameCounterState{counters: make(map[string]map[string]uint64)}
}

// Check validates and records a new frame counter. Returns true if valid.
func (f *FrameCounterState) Check(sessionID, senderID string, counter uint64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counters[sessionID] == nil {
		f.counters[sessionID] = make(map[string]uint64)
	}
	last, seen := f.counters[sessionID][senderID]
	if seen && counter <= last {
		return false
	}
	f.counters[sessionID][senderID] = counter
	return true
}

// Remove deletes frame counter state for a session.
func (f *FrameCounterState) Remove(sessionID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.counters, sessionID)
}

// MuteState tracks muted peers per session.
type MuteState struct {
	mu   sync.Mutex
	muted map[string]map[string]bool // sessionID → peerID → muted
}

func newMuteState() *MuteState {
	return &MuteState{muted: make(map[string]map[string]bool)}
}

func (m *MuteState) SetMuted(sessionID, peerID string, muted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.muted[sessionID] == nil {
		m.muted[sessionID] = make(map[string]bool)
	}
	m.muted[sessionID][peerID] = muted
}

func (m *MuteState) IsMuted(sessionID, peerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.muted[sessionID][peerID]
}

// ICEState tracks ICE candidate gathering per session.
type ICEState struct {
	mu         sync.Mutex
	candidates map[string][]string // sessionID → candidates
	complete   map[string]bool     // sessionID → gathering complete
}

func newICEState() *ICEState {
	return &ICEState{
		candidates: make(map[string][]string),
		complete:   make(map[string]bool),
	}
}

func (s *ICEState) Add(sessionID, candidate string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.candidates[sessionID] = append(s.candidates[sessionID], candidate)
}

func (s *ICEState) Complete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.complete[sessionID] = true
}

func (s *ICEState) IsComplete(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.complete[sessionID]
}

func (s *ICEState) Remove(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.candidates, sessionID)
	delete(s.complete, sessionID)
}

// Ensure time is used (for import linting purposes).
var _ = time.Now
