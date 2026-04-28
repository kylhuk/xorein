package discovery

import (
	"sync"
	"time"
)

const (
	addrTTL    = 24 * time.Hour  // spec 31 §3.2 provider record TTL
	peerTTL    = 15 * time.Minute // mDNS peer TTL (spec 31 §2.4)
	manualPeerTTL = 0             // manual peers never expire
)

// Cache is the Layer 1 live peer registry (spec 31 §1, layer 1).
// It caches peer records from all discovery sources with per-source TTLs.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry // peerID → entry
}

// CacheEntry holds a peer record with metadata.
type CacheEntry struct {
	Record    PeerRecord
	FirstSeen time.Time
	UpdatedAt time.Time
	// backoff state for reconnect attempts
	failures  int
	nextRetry time.Time
}

// NewCache returns an empty Cache.
func NewCache() *Cache {
	return &Cache{entries: make(map[string]*CacheEntry)}
}

// Put adds or refreshes a peer record. ttl=0 means never expires.
// If the record carries a non-empty Signature, it is verified before admission;
// records with an invalid signature are silently dropped.
func (c *Cache) Put(r PeerRecord, ttl time.Duration) {
	// Verify signatures when present; drop records whose signature is invalid.
	if r.Signature != "" {
		if err := VerifyPeerRecord(&r); err != nil {
			return
		}
	}

	now := time.Now()
	var exp time.Time
	if ttl > 0 {
		exp = now.Add(ttl)
	}
	r.ExpiresAt = exp
	r.LastSeen = now.UnixMilli()

	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[r.PeerID]; ok {
		e.Record = r
		e.UpdatedAt = now
	} else {
		c.entries[r.PeerID] = &CacheEntry{
			Record:    r,
			FirstSeen: now,
			UpdatedAt: now,
		}
	}
}

// Get returns the entry for peerID, or nil if not present or expired.
func (c *Cache) Get(peerID string) *CacheEntry {
	c.mu.RLock()
	e := c.entries[peerID]
	c.mu.RUnlock()
	if e == nil || e.Record.IsExpired() {
		return nil
	}
	return e
}

// All returns all non-expired entries.
func (c *Cache) All() []*PeerRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]*PeerRecord, 0, len(c.entries))
	now := time.Now()
	for id, e := range c.entries {
		if !e.Record.ExpiresAt.IsZero() && now.After(e.Record.ExpiresAt) {
			delete(c.entries, id)
			continue
		}
		cp := e.Record
		out = append(out, &cp)
	}
	return out
}

// Remove deletes a peer from the cache.
func (c *Cache) Remove(peerID string) {
	c.mu.Lock()
	delete(c.entries, peerID)
	c.mu.Unlock()
}

// Len returns the number of live entries.
func (c *Cache) Len() int {
	c.mu.RLock()
	n := len(c.entries)
	c.mu.RUnlock()
	return n
}

// RecordFailure increments the failure counter and schedules a retry with
// exponential backoff (spec 31 §2.4 and plan: immediate/5s/30s/2× to 10min).
func (c *Cache) RecordFailure(peerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[peerID]
	if e == nil {
		return
	}
	e.failures++
	backoffs := []time.Duration{0, 5 * time.Second, 30 * time.Second, 2 * time.Minute, 5 * time.Minute, 10 * time.Minute}
	idx := e.failures - 1
	if idx >= len(backoffs) {
		idx = len(backoffs) - 1
	}
	e.nextRetry = time.Now().Add(backoffs[idx])
}

// ReadyForRetry reports whether the peer can be retried now.
func (c *Cache) ReadyForRetry(peerID string) bool {
	c.mu.RLock()
	e := c.entries[peerID]
	c.mu.RUnlock()
	return e == nil || time.Now().After(e.nextRetry)
}
