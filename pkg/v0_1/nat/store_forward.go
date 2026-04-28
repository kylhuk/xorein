package nat

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// RelayQueueMaxPerRecipient is the maximum number of in-flight deliveries
	// stored per recipient (spec 32 §4, charter §3.4).
	RelayQueueMaxPerRecipient = 256

	// RelayQueueTTL is the time-to-live for a stored delivery.
	// After this period the entry is eligible for eviction (spec 32 §4).
	RelayQueueTTL = 24 * time.Hour
)

// ErrRelayOpacityViolation is returned by RelayQueue.Store when the delivery
// body appears to be plaintext rather than ciphertext.
var ErrRelayOpacityViolation = errors.New("relay queue: delivery body is not opaque ciphertext (opacity violation)")

// ErrRelayQueueFull is returned by RelayQueue.Store when the per-recipient
// limit would be exceeded.
var ErrRelayQueueFull = fmt.Errorf("relay queue: per-recipient limit of %d exceeded", RelayQueueMaxPerRecipient)

// RelayQueueEntry is a single store-and-forward delivery (spec 32 §4).
type RelayQueueEntry struct {
	ID          string
	RecipientID string
	Body        []byte    // must be base64url ciphertext (verified on Store)
	StoredAt    time.Time
	ExpiresAt   time.Time
}

// RelayQueue is the in-memory store-and-forward relay queue.
//
// Spec 32 §4 and charter §3.4: relay nodes MUST store encrypted deliveries
// for offline recipients and MUST NOT inspect or modify message content.
// The opacity check below enforces the invariant that only base64url-encoded
// ciphertext (AEAD tag minimum 16 bytes) is ever queued.
type RelayQueue struct {
	mu      sync.Mutex
	entries map[string][]*RelayQueueEntry // recipientID → entries
}

// NewRelayQueue creates an empty RelayQueue.
func NewRelayQueue() *RelayQueue {
	return &RelayQueue{
		entries: make(map[string][]*RelayQueueEntry),
	}
}

// Store adds a delivery to the queue for recipientID.
//
// Returns ErrRelayOpacityViolation if body is not opaque ciphertext.
// Returns ErrRelayQueueFull if the per-recipient limit (256) would be exceeded
// after pruning expired entries.
func (q *RelayQueue) Store(recipientID string, id string, body []byte) error {
	if err := checkRelayBodyOpacity(body); err != nil {
		return err
	}

	now := time.Now()
	entry := &RelayQueueEntry{
		ID:          id,
		RecipientID: recipientID,
		Body:        body,
		StoredAt:    now,
		ExpiresAt:   now.Add(RelayQueueTTL),
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Purge expired entries before checking the limit.
	q.evictLocked(recipientID, now)

	if len(q.entries[recipientID]) >= RelayQueueMaxPerRecipient {
		return ErrRelayQueueFull
	}

	q.entries[recipientID] = append(q.entries[recipientID], entry)
	return nil
}

// Drain removes and returns all pending entries for recipientID, purging
// expired entries first.
func (q *RelayQueue) Drain(recipientID string) []*RelayQueueEntry {
	now := time.Now()
	q.mu.Lock()
	defer q.mu.Unlock()

	q.evictLocked(recipientID, now)

	out := q.entries[recipientID]
	delete(q.entries, recipientID)
	return out
}

// Snapshot returns a summary of all queued (non-expired) entries keyed by
// recipientID. Useful for control API observability.
func (q *RelayQueue) Snapshot() map[string]int {
	now := time.Now()
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make(map[string]int, len(q.entries))
	for id, entries := range q.entries {
		count := 0
		for _, e := range entries {
			if now.Before(e.ExpiresAt) {
				count++
			}
		}
		if count > 0 {
			result[id] = count
		}
	}
	return result
}

// Evict removes all expired entries across all recipients.
// Intended to be called periodically by a maintenance goroutine.
func (q *RelayQueue) Evict() {
	now := time.Now()
	q.mu.Lock()
	defer q.mu.Unlock()

	for recipientID := range q.entries {
		q.evictLocked(recipientID, now)
		if len(q.entries[recipientID]) == 0 {
			delete(q.entries, recipientID)
		}
	}
}

// evictLocked removes expired entries for a single recipient.
// Must be called with q.mu held.
func (q *RelayQueue) evictLocked(recipientID string, now time.Time) {
	entries := q.entries[recipientID]
	if len(entries) == 0 {
		return
	}
	valid := entries[:0]
	for _, e := range entries {
		if now.Before(e.ExpiresAt) {
			valid = append(valid, e)
		}
	}
	if len(valid) == 0 {
		delete(q.entries, recipientID)
	} else {
		q.entries[recipientID] = valid
	}
}

// checkRelayBodyOpacity verifies that body is opaque ciphertext and not
// plaintext. This mirrors the logic in pkg/v0_1/family/peer.CheckRelayOpacity
// but operates on raw bytes (not a JSON-encoded RelayDelivery) to avoid an
// import cycle between the nat and peer packages.
//
// Rules (spec 32 §4 opacity requirement):
//  1. Body must be valid base64url-without-padding (raw URL encoding).
//  2. The decoded payload must be at least 16 bytes (AEAD tag minimum).
//  3. The decoded payload must not look like printable UTF-8 text (>70% printable
//     ASCII suggests plaintext rather than ciphertext).
func checkRelayBodyOpacity(body []byte) error {
	if len(body) == 0 {
		// Empty body is technically opaque; let the operation handler decide.
		return nil
	}

	trimmed := strings.TrimSpace(string(body))
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return ErrRelayOpacityViolation
	}
	if len(decoded) < 16 {
		return ErrRelayOpacityViolation
	}
	if bodyLooksLikePlaintext(decoded) {
		return ErrRelayOpacityViolation
	}
	return nil
}

// bodyLooksLikePlaintext returns true if >70% of bytes are printable ASCII,
// suggesting the decoded body is plaintext rather than ciphertext.
func bodyLooksLikePlaintext(b []byte) bool {
	if len(b) < 16 {
		return false
	}
	printable := 0
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			printable++
		}
	}
	return printable*10 > len(b)*7
}
