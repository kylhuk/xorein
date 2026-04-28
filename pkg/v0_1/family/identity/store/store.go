// Package store defines the interface for prekey bundle persistence.
package store

import (
	"encoding/json"
	"sync"

	seal "github.com/aether/code_aether/pkg/v0_1/mode/seal"
)

// BundleStore is the persistence interface for prekey bundles.
type BundleStore interface {
	Get(peerID string) (*seal.PrekeyBundle, error)
	Put(peerID string, b *seal.PrekeyBundle) error
	Delete(peerID string) error
}

// MemStore is an in-memory BundleStore for testing.
type MemStore struct {
	mu      sync.RWMutex
	bundles map[string][]byte
}

var _ BundleStore = (*MemStore)(nil)

func NewMemStore() *MemStore { return &MemStore{bundles: make(map[string][]byte)} }

func (m *MemStore) Get(peerID string) (*seal.PrekeyBundle, error) {
	m.mu.RLock()
	data := m.bundles[peerID]
	m.mu.RUnlock()
	if data == nil {
		return nil, nil
	}
	var b seal.PrekeyBundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (m *MemStore) Put(peerID string, b *seal.PrekeyBundle) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.bundles[peerID] = data
	m.mu.Unlock()
	return nil
}

func (m *MemStore) Delete(peerID string) error {
	m.mu.Lock()
	delete(m.bundles, peerID)
	m.mu.Unlock()
	return nil
}
