package phase4

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDHTServiceRecordsSuccessToCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")

	cfg := DHTConfig{
		Namespace:      "test-ns",
		BootstrapPeers: []string{"bootstrap-peer"},
		PeerCachePath:  cachePath,
	}

	var probed []string
	svc, err := NewDHTService(cfg, func(_ context.Context, peer string) bool {
		probed = append(probed, peer)
		return peer == "bootstrap-peer"
	})
	if err != nil {
		t.Fatalf("new DHT service: %v", err)
	}

	statuses := svc.Bootstrap(context.Background())
	if len(statuses) == 0 || !statuses[0].Success {
		t.Fatalf("expected bootstrap success, got %#v", statuses)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read peer cache: %v", err)
	}
	var record PeerCacheFile
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("unmarshal peer cache: %v", err)
	}

	if len(record.Entries) != 1 || record.Entries[0].Address != "bootstrap-peer" {
		t.Fatalf("unexpected cache entries: %#v", record.Entries)
	}
	if record.Entries[0].LastSuccessUnix <= 0 {
		t.Fatalf("expected positive timestamp, got %d", record.Entries[0].LastSuccessUnix)
	}
	if len(probed) == 0 || probed[0] != "bootstrap-peer" {
		t.Fatalf("expected probe order to include bootstrap peer, got %v", probed)
	}
}

func TestDHTServicePrefersCachedPeers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")

	cache := NewLocalPeerCache(cachePath, 0, 0)
	cache.now = func() time.Time { return time.Unix(1_000, 0) }
	if err := cache.writeEntries([]PeerCacheEntry{{Address: "cached-peer", LastSuccessUnix: cache.now().Unix()}}); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	cfg := DHTConfig{
		Namespace:      "test-ns",
		BootstrapPeers: []string{"bootstrap-peer"},
		PeerCachePath:  cachePath,
	}

	var probed []string
	svc, err := NewDHTService(cfg, func(_ context.Context, peer string) bool {
		probed = append(probed, peer)
		return false
	})
	if err != nil {
		t.Fatalf("new DHT service: %v", err)
	}

	svc.Bootstrap(context.Background())
	cachedIndex, bootstrapIndex := -1, -1
	for i, peer := range probed {
		if peer == "cached-peer" {
			cachedIndex = i
		}
		if peer == "bootstrap-peer" {
			bootstrapIndex = i
		}
	}
	if cachedIndex == -1 || bootstrapIndex == -1 {
		t.Fatalf("missing expected peers during bootstrap, probed=%v", probed)
	}
	if cachedIndex > bootstrapIndex {
		t.Fatalf("cached peer should be probed before bootstrap peers, order=%v", probed)
	}
}

func TestLocalPeerCacheRetentionFiltersOutdatedEntries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")

	base := time.Unix(10_000, 0)
	cache := NewLocalPeerCache(cachePath, 30*time.Second, 0)
	cache.now = func() time.Time { return base }
	entries := []PeerCacheEntry{
		{Address: "old-peer", LastSuccessUnix: base.Unix() - 31},
		{Address: "fresh-peer", LastSuccessUnix: base.Unix() - 10},
	}
	if err := cache.writeEntries(entries); err != nil {
		t.Fatalf("write cache file: %v", err)
	}

	peers, err := cache.Load()
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if len(peers) != 1 || peers[0] != "fresh-peer" {
		t.Fatalf("expected only fresh entry, got %v", peers)
	}
}

func TestLocalPeerCacheRejectsInvalidAddresses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")

	cache := NewLocalPeerCache(cachePath, 0, 0)
	inputs := []string{"valid-peer", " invalid ", "bad\npeer"}
	for _, addr := range inputs {
		if err := cache.RecordSuccess(addr); err != nil {
			t.Fatalf("record %q: %v", addr, err)
		}
	}

	peers, err := cache.Load()
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if len(peers) != 2 {
		t.Fatalf("expected canonicalized+valid entries only, got %v", peers)
	}
	if !containsPeer(peers, "valid-peer") || !containsPeer(peers, "invalid") {
		t.Fatalf("expected peers to include valid-peer and canonicalized invalid, got %v", peers)
	}
}

func TestLocalPeerCacheRecoverCorruptFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")

	cache := NewLocalPeerCache(cachePath, 0, 0)
	cache.now = func() time.Time { return time.Unix(1_000, 0).UTC() }
	if err := os.WriteFile(cachePath, []byte("not json"), 0o600); err != nil {
		t.Fatalf("write invalid cache: %v", err)
	}

	peers, err := cache.Load()
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if len(peers) != 0 {
		t.Fatalf("expected empty peer list after recovery, got %v", peers)
	}

	backups, err := filepath.Glob(cachePath + ".corrupt-*")
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected corrupt backup, got %v", backups)
	}
	data, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != "not json" {
		t.Fatalf("expected original payload in backup, got %q", data)
	}
}
