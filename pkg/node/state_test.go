package node

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aether/code_aether/pkg/storage"
)

func TestLoadStatePersistsAcrossRestart(t *testing.T) {
	dataDir := t.TempDir()
	state, err := loadState(dataDir)
	if err != nil {
		t.Fatalf("loadState() initial error = %v", err)
	}
	state.Settings["listen_address"] = "127.0.0.1:4242"
	state.Deliveries["msg-1"] = struct{}{}
	state.Telemetry = append(state.Telemetry, "persisted")
	if err := saveState(dataDir, state); err != nil {
		t.Fatalf("saveState() error = %v", err)
	}
	restarted, err := loadState(dataDir)
	if err != nil {
		t.Fatalf("loadState() restart error = %v", err)
	}
	if restarted.Identity.PeerID != state.Identity.PeerID {
		t.Fatalf("restarted peer id = %s want %s", restarted.Identity.PeerID, state.Identity.PeerID)
	}
	if _, ok := restarted.Deliveries["msg-1"]; !ok {
		t.Fatal("expected delivery dedupe state to persist")
	}
	if restarted.Settings["listen_address"] != "127.0.0.1:4242" {
		t.Fatalf("listen_address = %q want %q", restarted.Settings["listen_address"], "127.0.0.1:4242")
	}
	if _, err := os.Stat(filepath.Join(dataDir, storage.StoreFileName)); err != nil {
		t.Fatalf("Stat(state.db) error = %v", err)
	}
}

func TestLoadStateMigratesLegacyJSONAndRecoversFromCorruption(t *testing.T) {
	dataDir := t.TempDir()
	identity, err := GenerateIdentity("legacy")
	if err != nil {
		t.Fatalf("GenerateIdentity() error = %v", err)
	}
	legacy := legacyStateV1{Identity: identity, KnownPeers: map[string]PeerRecord{}, Servers: map[string]ServerRecord{}, DMs: map[string]DMRecord{}, Messages: map[string]MessageRecord{}, Settings: map[string]string{"k": "v"}, ControlToken: "legacy-token"}
	raw, _ := json.Marshal(legacy)
	legacyPath := filepath.Join(dataDir, legacyStateFile)
	if err := os.WriteFile(legacyPath, raw, 0o600); err != nil {
		t.Fatalf("WriteFile(state.json) error = %v", err)
	}
	migrated, err := loadState(dataDir)
	if err != nil {
		t.Fatalf("loadState() migration error = %v", err)
	}
	if migrated.Identity.PeerID != identity.PeerID {
		t.Fatalf("migrated peer id = %s want %s", migrated.Identity.PeerID, identity.PeerID)
	}
	backups, err := filepath.Glob(filepath.Join(dataDir, legacyBackupTag+"*.json"))
	if err != nil {
		t.Fatalf("Glob(migrated backup) error = %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("expected migrated legacy backup")
	}
	if err := os.WriteFile(filepath.Join(dataDir, storage.StoreFileName), []byte("broken"), 0o600); err != nil {
		t.Fatalf("WriteFile(corrupt state.db) error = %v", err)
	}
	recovered, err := loadState(dataDir)
	if err != nil {
		t.Fatalf("loadState() recovered error = %v", err)
	}
	if recovered.Identity.PeerID != identity.PeerID {
		t.Fatalf("recovered peer id = %s want %s", recovered.Identity.PeerID, identity.PeerID)
	}
	quarantine, err := filepath.Glob(filepath.Join(dataDir, storage.StoreFileName+".corrupt-*"))
	if err != nil {
		t.Fatalf("Glob(corrupt store) error = %v", err)
	}
	if len(quarantine) == 0 {
		t.Fatal("expected corrupt store quarantine")
	}
}

func TestLoadStateRejectsWrongKeyAndQuarantinesCorruptLegacyJSON(t *testing.T) {
	dataDir := t.TempDir()
	state, err := loadState(dataDir)
	if err != nil {
		t.Fatalf("loadState() error = %v", err)
	}
	if err := saveState(dataDir, state); err != nil {
		t.Fatalf("saveState() error = %v", err)
	}
	wrongSecret := make([]byte, 32)
	copy(wrongSecret, []byte("this-is-a-different-storage-secret!!"))
	if err := os.WriteFile(filepath.Join(dataDir, "state.key"), []byte(base64.RawURLEncoding.EncodeToString(wrongSecret)+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(state.key) error = %v", err)
	}
	if _, err := loadState(dataDir); err == nil || err.Error() == "" {
		t.Fatal("expected wrong-key failure")
	}

	legacyDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(legacyDir, legacyStateFile), []byte("{broken"), 0o600); err != nil {
		t.Fatalf("WriteFile(corrupt state.json) error = %v", err)
	}
	recovered, err := loadState(legacyDir)
	if err != nil {
		t.Fatalf("loadState() corrupt legacy error = %v", err)
	}
	if recovered.Identity.PeerID == "" {
		t.Fatal("expected fresh identity after corrupt legacy recovery")
	}
	quarantine, err := filepath.Glob(filepath.Join(legacyDir, "state.corrupt-*.json"))
	if err != nil {
		t.Fatalf("Glob(corrupt legacy) error = %v", err)
	}
	if len(quarantine) == 0 {
		t.Fatal("expected corrupt legacy quarantine")
	}
}

func TestLoadStateRejectsCorruptLegacyBackupInsteadOfFreshState(t *testing.T) {
	dataDir := t.TempDir()
	backupPath := filepath.Join(dataDir, legacyBackupTag+"20260421T000000000.json")
	if err := os.WriteFile(backupPath, []byte("{broken"), 0o600); err != nil {
		t.Fatalf("WriteFile(legacy backup) error = %v", err)
	}
	if _, err := loadState(dataDir); err == nil {
		t.Fatal("expected corrupt legacy backup to fail closed")
	}
}

func TestEnsureStateNormalizesKnownPeerAddresses(t *testing.T) {
	state := persistedState{KnownPeers: map[string]PeerRecord{
		"peer-1": {PeerID: "peer-1", Addresses: []string{"http://127.0.0.1:1234/path", "192.168.0.1:80"}},
	}}
	if err := ensureState(&state); err != nil {
		t.Fatalf("ensureState() error = %v", err)
	}
	peer := state.KnownPeers["peer-1"]
	if !sameStringSet(peer.Addresses, []string{"/ip4/127.0.0.1/tcp/1234"}) {
		t.Fatalf("normalized addresses = %#v want %#v", peer.Addresses, []string{"/ip4/127.0.0.1/tcp/1234"})
	}
}
