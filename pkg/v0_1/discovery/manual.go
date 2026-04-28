package discovery

import (
	"strings"
	"sync"
)

// ManualPeers tracks a set of manually configured peer addresses.
// Manual peers have no TTL and are given Source="manual" priority.
type ManualPeers struct {
	mu    sync.RWMutex
	addrs []string // raw multiaddrs or "host:port" addresses
}

// NewManualPeers creates a ManualPeers from a CSV string or slice of addresses.
func NewManualPeers(addrs []string) *ManualPeers {
	var cleaned []string
	for _, a := range addrs {
		a = strings.TrimSpace(a)
		if a != "" {
			cleaned = append(cleaned, a)
		}
	}
	return &ManualPeers{addrs: cleaned}
}

// Addrs returns all configured manual peer addresses.
func (m *ManualPeers) Addrs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, len(m.addrs))
	copy(out, m.addrs)
	return out
}

// Add appends an address if not already present.
func (m *ManualPeers) Add(addr string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.addrs {
		if a == addr {
			return
		}
	}
	m.addrs = append(m.addrs, addr)
}

// Remove deletes an address from the list.
func (m *ManualPeers) Remove(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.addrs[:0]
	for _, a := range m.addrs {
		if a != addr {
			out = append(out, a)
		}
	}
	m.addrs = out
}

// SeedCache loads all manual peers into the cache with manualPeerTTL (never expires).
func (m *ManualPeers) SeedCache(c *Cache) {
	for _, addr := range m.Addrs() {
		r := PeerRecord{
			Addresses: []string{addr},
			Source:    "manual",
		}
		c.Put(r, manualPeerTTL)
	}
}
