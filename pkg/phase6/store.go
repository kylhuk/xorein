package phase6

import (
	"errors"
	"sync"
	"time"
)

// DefaultManifestTTL describes the default cache lifetime for published manifests.
const DefaultManifestTTL = 5 * time.Minute

var (
	ErrManifestNotFound = errors.New("manifest not found")
	ErrManifestExpired  = errors.New("manifest entry expired")
)

type manifestEntry struct {
	manifest  *Manifest
	expiresAt time.Time
}

// ManifestStore publishes manifests and resolves them via deterministic server IDs.
type ManifestStore struct {
	mu      sync.RWMutex
	entries map[string]*manifestEntry
	ttl     time.Duration
}

// NewManifestStore constructs a store with a configurable TTL. If ttl <= 0 it falls back to DefaultManifestTTL.
func NewManifestStore(ttl time.Duration) *ManifestStore {
	if ttl <= 0 {
		ttl = DefaultManifestTTL
	}
	return &ManifestStore{entries: map[string]*manifestEntry{}, ttl: ttl}
}

// Publish adds or updates the manifest for a server ID. Entries are invalidated
// when the manifest version increases, or when the version is unchanged and the
// timestamp is newer.
func (s *ManifestStore) Publish(m *Manifest) error {
	if err := m.ValidateFields(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	existing, ok := s.entries[m.ServerID]

	if ok {
		prev := existing.manifest
		if m.Version < prev.Version {
			return nil
		}
		if m.Version == prev.Version && !m.UpdatedAt.After(prev.UpdatedAt) {
			return nil
		}
	}

	s.entries[m.ServerID] = &manifestEntry{
		manifest:  m.Clone(),
		expiresAt: now.Add(s.ttl),
	}
	return nil
}

// Resolve returns the cached manifest for the provided server ID if it has not expired.
func (s *ManifestStore) Resolve(serverID string) (*Manifest, error) {
	s.mu.RLock()
	entry, ok := s.entries[serverID]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrManifestNotFound
	}

	now := time.Now().UTC()
	if now.After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.entries, serverID)
		s.mu.Unlock()
		return nil, ErrManifestExpired
	}

	return entry.manifest.Clone(), nil
}

// Invalidate removes the manifest entry for the provided server ID.
func (s *ManifestStore) Invalidate(serverID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, serverID)
}
