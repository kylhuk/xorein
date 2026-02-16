package phase9

import (
	"fmt"
	"sync"
	"time"
)

const (
	// RelayLogClassOperational marks counters and policy values that are safe for logs.
	RelayLogClassOperational = "operational"
	// RelayLogClassRestricted marks fields that must never contain payload/key material.
	RelayLogClassRestricted = "restricted"
)

// Config defines deterministic relay-service controls for Circuit Relay v2 fallback handling.
type Config struct {
	ReservationLimit int
	SessionTimeout   time.Duration
	MaxBytesPerSec   int64
}

// Normalize validates configuration and applies safe defaults for omitted fields.
func (c Config) Normalize() (Config, error) {
	normalized := c
	if normalized.ReservationLimit == 0 {
		normalized.ReservationLimit = 256
	}
	if normalized.SessionTimeout == 0 {
		normalized.SessionTimeout = 2 * time.Minute
	}
	if normalized.MaxBytesPerSec == 0 {
		normalized.MaxBytesPerSec = 1_000_000
	}
	if normalized.ReservationLimit < 1 {
		return Config{}, fmt.Errorf("reservation limit must be at least 1")
	}
	if normalized.SessionTimeout < 15*time.Second {
		return Config{}, fmt.Errorf("session timeout must be at least 15s")
	}
	if normalized.MaxBytesPerSec < 16_384 {
		return Config{}, fmt.Errorf("max bytes per second must be at least 16384")
	}
	return normalized, nil
}

// Service tracks bounded relay reservations and emits deterministic observability snapshots.
type Service struct {
	cfg         Config
	mu          sync.Mutex
	active      int
	rejected    uint64
	timedOut    uint64
	established uint64
}

// Snapshot exposes service limits and counters for logs/diagnostics.
type Snapshot struct {
	ReservationLimit int
	SessionTimeout   time.Duration
	MaxBytesPerSec   int64
	Active           int
	Rejected         uint64
	TimedOut         uint64
	Established      uint64
}

// PrivacyAuditRecord is a privacy-preserving relay audit row for diagnostics.
type PrivacyAuditRecord struct {
	Event           string
	Sensitivity     string
	ReservationCap  int
	SessionTimeout  string
	RateLimitBps    int64
	Active          int
	Rejected        uint64
	TimedOut        uint64
	Established     uint64
	ResidualRiskIDs []string
}

// NewService constructs a relay service policy handler.
func NewService(cfg Config) (*Service, error) {
	normalized, err := cfg.Normalize()
	if err != nil {
		return nil, err
	}
	return &Service{cfg: normalized}, nil
}

// Reserve attempts to admit a relay reservation and enforces configured limits.
func (s *Service) Reserve() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active >= s.cfg.ReservationLimit {
		s.rejected++
		return false
	}
	s.active++
	s.established++
	return true
}

// Release closes an active reservation if one exists.
func (s *Service) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active > 0 {
		s.active--
	}
}

// ExpireOne records an active reservation timeout event.
func (s *Service) ExpireOne() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active > 0 {
		s.active--
		s.timedOut++
	}
}

// Snapshot returns a stable view of relay limits and policy counters.
func (s *Service) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Snapshot{
		ReservationLimit: s.cfg.ReservationLimit,
		SessionTimeout:   s.cfg.SessionTimeout,
		MaxBytesPerSec:   s.cfg.MaxBytesPerSec,
		Active:           s.active,
		Rejected:         s.rejected,
		TimedOut:         s.timedOut,
		Established:      s.established,
	}
}

// PrivacyAuditSnapshot returns a log-safe record without identity or payload material.
func (s *Service) PrivacyAuditSnapshot() PrivacyAuditRecord {
	snap := s.Snapshot()
	return PrivacyAuditRecord{
		Event:           "relay_snapshot",
		Sensitivity:     RelayLogClassOperational,
		ReservationCap:  snap.ReservationLimit,
		SessionTimeout:  snap.SessionTimeout.String(),
		RateLimitBps:    snap.MaxBytesPerSec,
		Active:          snap.Active,
		Rejected:        snap.Rejected,
		TimedOut:        snap.TimedOut,
		Established:     snap.Established,
		ResidualRiskIDs: []string{"R6"},
	}
}
