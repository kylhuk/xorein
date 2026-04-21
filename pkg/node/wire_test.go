package node

import (
	"bytes"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/protocol"
)

func TestManifestCanonicalBytesOmitUnsetHistorySemantics(t *testing.T) {
	identity, err := GenerateIdentity("legacy")
	if err != nil {
		t.Fatalf("GenerateIdentity() error = %v", err)
	}

	manifest := Manifest{
		ServerID:       "server-legacy",
		Name:           "legacy",
		OwnerPeerID:    identity.PeerID,
		OwnerPublicKey: identity.PublicKey,
		OwnerAddresses: []string{"127.0.0.1:9000"},
		Capabilities:   []string{"cap.chat"},
		IssuedAt:       time.Now().UTC().Add(-time.Minute),
		UpdatedAt:      time.Now().UTC(),
	}
	payload, err := manifest.canonicalBytes()
	if err != nil {
		t.Fatalf("canonicalBytes() error = %v", err)
	}
	if bytes.Contains(payload, []byte("history_retention_messages")) || bytes.Contains(payload, []byte("history_coverage")) || bytes.Contains(payload, []byte("history_durability")) {
		t.Fatalf("canonical payload unexpectedly included unset history semantics: %s", payload)
	}
	if err := manifest.Sign(identity); err != nil {
		t.Fatalf("manifest.Sign() error = %v", err)
	}
	if err := manifest.Verify(); err != nil {
		t.Fatalf("manifest.Verify() error = %v", err)
	}
}

func TestCreateServerManifestAdvertisesHistorySemantics(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 7})
	defer stop()

	server, err := service.CreateServer("coverage", "manifest semantics")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if server.Manifest.HistoryRetentionMessages != 7 {
		t.Fatalf("HistoryRetentionMessages = %d want %d", server.Manifest.HistoryRetentionMessages, 7)
	}
	if server.Manifest.HistoryCoverage != HistoryCoverageLocalWindow {
		t.Fatalf("HistoryCoverage = %q want %q", server.Manifest.HistoryCoverage, HistoryCoverageLocalWindow)
	}
	if server.Manifest.HistoryDurability != HistoryDurabilitySingleNode {
		t.Fatalf("HistoryDurability = %q want %q", server.Manifest.HistoryDurability, HistoryDurabilitySingleNode)
	}
	if err := server.Manifest.Verify(); err != nil {
		t.Fatalf("manifest.Verify() error = %v", err)
	}
}

func TestCreateServerManifestAdvertisesArchivistCapability(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleArchivist, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 7})
	defer stop()

	server, err := service.CreateServer("archive", "archivist manifest semantics")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if !contains(server.Manifest.Capabilities, string(protocol.FeatureArchivist)) {
		t.Fatalf("manifest capabilities = %v, want archivist capability", server.Manifest.Capabilities)
	}
	if err := server.Manifest.Verify(); err != nil {
		t.Fatalf("manifest.Verify() error = %v", err)
	}
}
