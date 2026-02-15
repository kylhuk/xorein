package phase4

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const peerCacheSchemaVersion = 1

// PeerCacheEntry stores one known-good peer address and its freshness metadata.
type PeerCacheEntry struct {
	Address         string `json:"address"`
	LastSuccessUnix int64  `json:"last_success_unix"`
}

// PeerCacheFile is the on-disk schema for local peer cache persistence.
type PeerCacheFile struct {
	Version     int              `json:"version"`
	UpdatedUnix int64            `json:"updated_unix"`
	Entries     []PeerCacheEntry `json:"entries"`
}

// LocalPeerCache provides minimal persistence for peer freshness tracking.
type LocalPeerCache struct {
	path       string
	retention  time.Duration
	maxEntries int
	now        func() time.Time
}

// NewLocalPeerCache creates a cache with retention and entry-limit controls.
func NewLocalPeerCache(path string, retention time.Duration, maxEntries int) *LocalPeerCache {
	if maxEntries < 0 {
		maxEntries = 0
	}
	return &LocalPeerCache{
		path:       strings.TrimSpace(path),
		retention:  retention,
		maxEntries: maxEntries,
		now:        time.Now,
	}
}

// Load returns cached peers ordered by freshness (newest first).
// Corrupt cache files are recovered by being renamed out of the active path.
func (c *LocalPeerCache) Load() ([]string, error) {
	if c.path == "" {
		return nil, nil
	}
	record, err := c.readRecord()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		if recErr := c.recoverCorruptFile(); recErr != nil {
			return nil, fmt.Errorf("recover corrupt peer cache: %w", recErr)
		}
		return nil, nil
	}

	entries := c.normalizeEntries(record.Entries)
	if len(entries) != len(record.Entries) {
		if err := c.writeEntries(entries); err != nil {
			return nil, fmt.Errorf("rewrite normalized peer cache: %w", err)
		}
	}

	peers := make([]string, 0, len(entries))
	for _, entry := range entries {
		peers = append(peers, entry.Address)
	}
	return peers, nil
}

// RecordSuccess updates freshness metadata for a successful connection.
func (c *LocalPeerCache) RecordSuccess(address string) error {
	if c.path == "" {
		return nil
	}
	address = strings.TrimSpace(address)
	if !isValidPeerAddress(address) {
		return nil
	}

	record, err := c.readRecord()
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			if recErr := c.recoverCorruptFile(); recErr != nil {
				return fmt.Errorf("recover corrupt peer cache: %w", recErr)
			}
		}
		record = PeerCacheFile{Version: peerCacheSchemaVersion}
	}

	nowUnix := c.now().UTC().Unix()
	entries := append([]PeerCacheEntry(nil), record.Entries...)
	updated := false
	for i := range entries {
		if entries[i].Address == address {
			entries[i].LastSuccessUnix = nowUnix
			updated = true
			break
		}
	}
	if !updated {
		entries = append(entries, PeerCacheEntry{Address: address, LastSuccessUnix: nowUnix})
	}

	entries = c.normalizeEntries(entries)
	if err := c.writeEntries(entries); err != nil {
		return fmt.Errorf("write peer cache: %w", err)
	}
	return nil
}

func (c *LocalPeerCache) readRecord() (PeerCacheFile, error) {
	raw, err := os.ReadFile(c.path)
	if err != nil {
		return PeerCacheFile{}, err
	}
	var record PeerCacheFile
	if err := json.Unmarshal(raw, &record); err != nil {
		return PeerCacheFile{}, err
	}
	if record.Version != peerCacheSchemaVersion {
		return PeerCacheFile{}, fmt.Errorf("unsupported peer cache schema version: %d", record.Version)
	}
	return record, nil
}

func (c *LocalPeerCache) writeEntries(entries []PeerCacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	record := PeerCacheFile{
		Version:     peerCacheSchemaVersion,
		UpdatedUnix: c.now().UTC().Unix(),
		Entries:     entries,
	}
	encoded, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(c.path, encoded, 0o600)
}

func (c *LocalPeerCache) recoverCorruptFile() error {
	stamp := c.now().UTC().Format("20060102T150405")
	backup := fmt.Sprintf("%s.corrupt-%s", c.path, stamp)
	if err := os.Rename(c.path, backup); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

func (c *LocalPeerCache) normalizeEntries(entries []PeerCacheEntry) []PeerCacheEntry {
	nowUnix := c.now().UTC().Unix()
	cutoff := int64(0)
	if c.retention > 0 {
		cutoff = nowUnix - int64(c.retention/time.Second)
	}

	merged := make(map[string]PeerCacheEntry, len(entries))
	for _, entry := range entries {
		address := strings.TrimSpace(entry.Address)
		if !isValidPeerAddress(address) {
			continue
		}
		if cutoff > 0 && entry.LastSuccessUnix < cutoff {
			continue
		}
		if entry.LastSuccessUnix <= 0 {
			continue
		}
		existing, ok := merged[address]
		if !ok || entry.LastSuccessUnix > existing.LastSuccessUnix {
			merged[address] = PeerCacheEntry{Address: address, LastSuccessUnix: entry.LastSuccessUnix}
		}
	}

	normalized := make([]PeerCacheEntry, 0, len(merged))
	for _, entry := range merged {
		normalized = append(normalized, entry)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].LastSuccessUnix == normalized[j].LastSuccessUnix {
			return normalized[i].Address < normalized[j].Address
		}
		return normalized[i].LastSuccessUnix > normalized[j].LastSuccessUnix
	})
	if c.maxEntries > 0 && len(normalized) > c.maxEntries {
		normalized = normalized[:c.maxEntries]
	}
	return normalized
}

func preferCachedPeers(cached []string, bootstrap []string) []string {
	ordered := make([]string, 0, len(cached)+len(bootstrap))
	seen := make(map[string]struct{}, len(cached)+len(bootstrap))
	for _, peer := range cached {
		peer = strings.TrimSpace(peer)
		if !isValidPeerAddress(peer) {
			continue
		}
		if _, ok := seen[peer]; ok {
			continue
		}
		seen[peer] = struct{}{}
		ordered = append(ordered, peer)
	}
	for _, peer := range bootstrap {
		peer = strings.TrimSpace(peer)
		if !isValidPeerAddress(peer) {
			continue
		}
		if _, ok := seen[peer]; ok {
			continue
		}
		seen[peer] = struct{}{}
		ordered = append(ordered, peer)
	}
	return ordered
}

func isValidPeerAddress(address string) bool {
	if strings.TrimSpace(address) == "" {
		return false
	}
	if len(address) > 512 {
		return false
	}
	if !utf8.ValidString(address) {
		return false
	}
	for _, r := range address {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
