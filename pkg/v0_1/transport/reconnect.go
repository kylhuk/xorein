package transport

import (
	"sync"
	"time"
)

// backoffSchedule maps consecutive failure counts to the delay before the next
// retry (spec 30 §4.2): immediate / 5s / 30s / 2min / 5min / capped at 10min.
var backoffSchedule = []time.Duration{
	0,               // 0 failures → retry immediately
	5 * time.Second, // 1 failure
	30 * time.Second, // 2 failures
	2 * time.Minute,  // 3 failures
	5 * time.Minute,  // 4 failures
	10 * time.Minute, // 5+ failures (cap)
}

// degradedThreshold is the number of consecutive failures after which a peer
// is considered "degraded" (spec 30 §4.2).
const degradedThreshold = 3

// ReconnectSchedule tracks per-peer reconnect state and implements the spec 30
// §4.2 backoff schedule: immediate / 5s / 30s / 2min / 5min / 10min (cap).
type ReconnectSchedule struct {
	mu      sync.Mutex
	entries map[string]*reconnectEntry // keyed by peer ID string
}

type reconnectEntry struct {
	failures  int
	nextRetry time.Time
	degraded  bool // true after 3+ consecutive failures
}

// NewReconnectSchedule creates a new, empty schedule.
func NewReconnectSchedule() *ReconnectSchedule {
	return &ReconnectSchedule{
		entries: make(map[string]*reconnectEntry),
	}
}

// ShouldRetry returns true if the peer is eligible for reconnection right now.
// A peer with no recorded failures is always eligible.
func (s *ReconnectSchedule) ShouldRetry(peerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[peerID]
	if !ok {
		return true
	}
	return !time.Now().Before(e.nextRetry)
}

// RecordFailure records a failed connection attempt for peerID and advances the
// backoff timer.  Once failures reaches degradedThreshold the peer is marked
// degraded.
func (s *ReconnectSchedule) RecordFailure(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[peerID]
	if !ok {
		e = &reconnectEntry{}
		s.entries[peerID] = e
	}
	e.failures++

	// Pick delay from schedule; cap at the last entry.
	idx := e.failures
	if idx >= len(backoffSchedule) {
		idx = len(backoffSchedule) - 1
	}
	e.nextRetry = time.Now().Add(backoffSchedule[idx])

	if e.failures >= degradedThreshold {
		e.degraded = true
	}
}

// RecordSuccess clears all failure state for peerID.
func (s *ReconnectSchedule) RecordSuccess(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, peerID)
}

// IsDegraded returns true if peerID has had 3 or more consecutive failures.
func (s *ReconnectSchedule) IsDegraded(peerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[peerID]
	return ok && e.degraded
}
