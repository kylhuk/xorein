package phase9

import (
	"fmt"
	"sort"
	"time"
)

const (
	// StoreLogClassOperational marks counters and bounds that are safe for logs.
	StoreLogClassOperational = "operational"
	// StoreLogClassRestricted marks relay metadata that should be tightly controlled.
	StoreLogClassRestricted = "restricted"
)

// StoreConfig defines deterministic store-and-forward retention controls.
type StoreConfig struct {
	RetentionTTL    time.Duration
	MaxMessages     int
	MaxBytes        int64
	MaxPayloadBytes int
}

// Normalize validates store configuration and applies safe defaults.
func (c StoreConfig) Normalize() (StoreConfig, error) {
	normalized := c
	if normalized.RetentionTTL == 0 {
		normalized.RetentionTTL = 24 * time.Hour
	}
	if normalized.MaxMessages == 0 {
		normalized.MaxMessages = 10_000
	}
	if normalized.MaxBytes == 0 {
		normalized.MaxBytes = 128 * 1024 * 1024
	}
	if normalized.MaxPayloadBytes == 0 {
		normalized.MaxPayloadBytes = 64 * 1024
	}

	if normalized.RetentionTTL < time.Minute {
		return StoreConfig{}, fmt.Errorf("retention ttl must be at least 1m")
	}
	if normalized.MaxMessages < 1 {
		return StoreConfig{}, fmt.Errorf("max messages must be at least 1")
	}
	if normalized.MaxBytes < 1 {
		return StoreConfig{}, fmt.Errorf("max bytes must be at least 1")
	}
	if normalized.MaxPayloadBytes < 1 {
		return StoreConfig{}, fmt.Errorf("max payload bytes must be at least 1")
	}

	return normalized, nil
}

// OpaqueEnvelope stores only relay-safe opaque payload metadata.
type OpaqueEnvelope struct {
	RecipientID string
	Ciphertext  []byte
	StoredAt    time.Time
	ExpiresAt   time.Time
}

// StoreResult reports deterministic retention and quota outcomes per append.
type StoreResult struct {
	Stored       bool
	Reason       string
	DroppedByTTL int
	DroppedByCap int
}

// StoreSnapshot exposes deterministic queue and quota counters.
type StoreSnapshot struct {
	RetentionTTL    time.Duration
	MaxMessages     int
	MaxBytes        int64
	MaxPayloadBytes int
	QueuedMessages  int
	QueuedBytes     int64
	DroppedByTTL    uint64
	DroppedByCap    uint64
	RejectedWrites  uint64
}

// StorePrivacyAuditRecord is a privacy-preserving retention/logging audit row.
type StorePrivacyAuditRecord struct {
	Event            string
	Sensitivity      string
	QueueSensitivity string
	RetentionTTL     string
	MaxMessages      int
	MaxBytes         int64
	MaxPayloadBytes  int
	QueuedMessages   int
	QueuedBytes      int64
	DroppedByTTL     uint64
	DroppedByCap     uint64
	RejectedWrites   uint64
	ResidualRiskIDs  []string
}

type storedEnvelope struct {
	recipientID string
	ciphertext  []byte
	storedAt    time.Time
	expiresAt   time.Time
	seq         uint64
}

// StoreService is an in-memory deterministic store-and-forward policy model.
type StoreService struct {
	cfg StoreConfig

	nextSeq uint64
	items   []storedEnvelope

	queuedBytes int64

	droppedByTTL   uint64
	droppedByCap   uint64
	rejectedWrites uint64
}

// NewStoreService constructs a bounded store-and-forward service.
func NewStoreService(cfg StoreConfig) (*StoreService, error) {
	normalized, err := cfg.Normalize()
	if err != nil {
		return nil, err
	}
	return &StoreService{cfg: normalized}, nil
}

// Store appends an opaque ciphertext envelope, applying ttl and quota policies.
func (s *StoreService) Store(now time.Time, recipientID string, ciphertext []byte) StoreResult {
	if recipientID == "" {
		s.rejectedWrites++
		return StoreResult{Stored: false, Reason: "recipient required"}
	}
	if len(ciphertext) == 0 {
		s.rejectedWrites++
		return StoreResult{Stored: false, Reason: "ciphertext required"}
	}
	if len(ciphertext) > s.cfg.MaxPayloadBytes {
		s.rejectedWrites++
		return StoreResult{Stored: false, Reason: "payload exceeds max"}
	}

	result := s.compact(now)

	if int64(len(ciphertext)) > s.cfg.MaxBytes {
		s.rejectedWrites++
		result.Stored = false
		result.Reason = "payload exceeds store quota"
		return result
	}

	env := storedEnvelope{
		recipientID: recipientID,
		ciphertext:  append([]byte(nil), ciphertext...),
		storedAt:    now.UTC(),
		expiresAt:   now.UTC().Add(s.cfg.RetentionTTL),
		seq:         s.nextSeq,
	}
	s.nextSeq++

	s.items = append(s.items, env)
	s.queuedBytes += int64(len(ciphertext))

	for len(s.items) > s.cfg.MaxMessages || s.queuedBytes > s.cfg.MaxBytes {
		dropped := s.items[0]
		s.items = s.items[1:]
		s.queuedBytes -= int64(len(dropped.ciphertext))
		s.droppedByCap++
		result.DroppedByCap++
	}

	result.Stored = true
	result.Reason = "stored"
	return result
}

// DrainRecipient returns and removes unexpired queued ciphertexts for recipient.
func (s *StoreService) DrainRecipient(now time.Time, recipientID string) [][]byte {
	s.compact(now)

	if recipientID == "" {
		return nil
	}

	out := make([][]byte, 0)
	kept := make([]storedEnvelope, 0, len(s.items))
	for _, item := range s.items {
		if item.recipientID == recipientID {
			out = append(out, append([]byte(nil), item.ciphertext...))
			s.queuedBytes -= int64(len(item.ciphertext))
			continue
		}
		kept = append(kept, item)
	}
	s.items = kept

	return out
}

// Snapshot returns deterministic store retention/quota counters.
func (s *StoreService) Snapshot(now time.Time) StoreSnapshot {
	s.compact(now)
	return StoreSnapshot{
		RetentionTTL:    s.cfg.RetentionTTL,
		MaxMessages:     s.cfg.MaxMessages,
		MaxBytes:        s.cfg.MaxBytes,
		MaxPayloadBytes: s.cfg.MaxPayloadBytes,
		QueuedMessages:  len(s.items),
		QueuedBytes:     s.queuedBytes,
		DroppedByTTL:    s.droppedByTTL,
		DroppedByCap:    s.droppedByCap,
		RejectedWrites:  s.rejectedWrites,
	}
}

// PrivacyAuditSnapshot returns a log-safe retention snapshot with no payload exposure.
func (s *StoreService) PrivacyAuditSnapshot(now time.Time) StorePrivacyAuditRecord {
	snap := s.Snapshot(now)
	return StorePrivacyAuditRecord{
		Event:            "storeforward_snapshot",
		Sensitivity:      StoreLogClassOperational,
		QueueSensitivity: StoreLogClassRestricted,
		RetentionTTL:     snap.RetentionTTL.String(),
		MaxMessages:      snap.MaxMessages,
		MaxBytes:         snap.MaxBytes,
		MaxPayloadBytes:  snap.MaxPayloadBytes,
		QueuedMessages:   snap.QueuedMessages,
		QueuedBytes:      snap.QueuedBytes,
		DroppedByTTL:     snap.DroppedByTTL,
		DroppedByCap:     snap.DroppedByCap,
		RejectedWrites:   snap.RejectedWrites,
		ResidualRiskIDs:  []string{"R6"},
	}
}

func (s *StoreService) compact(now time.Time) StoreResult {
	if len(s.items) == 0 {
		return StoreResult{}
	}

	sort.SliceStable(s.items, func(i, j int) bool {
		if s.items[i].expiresAt.Equal(s.items[j].expiresAt) {
			return s.items[i].seq < s.items[j].seq
		}
		return s.items[i].expiresAt.Before(s.items[j].expiresAt)
	})

	var dropped int
	var droppedCounter uint64
	cut := 0
	for cut < len(s.items) {
		if s.items[cut].expiresAt.After(now) {
			break
		}
		s.queuedBytes -= int64(len(s.items[cut].ciphertext))
		cut++
		dropped++
		droppedCounter++
	}
	if cut > 0 {
		s.items = append([]storedEnvelope(nil), s.items[cut:]...)
		s.droppedByTTL += droppedCounter
	}

	return StoreResult{DroppedByTTL: dropped}
}
