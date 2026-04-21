package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/network"
	"github.com/aether/code_aether/pkg/storage"
)

func TestNetworkFormationAndPeerCache(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	clientA, stopA := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopA()

	clientBDir := t.TempDir()
	clientB, stopB := startService(t, Config{Role: RoleClient, DataDir: clientBDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})

	waitFor(t, 3*time.Second, func() bool {
		return hasPeer(clientA.Snapshot(), bootstrap.PeerID()) && hasPeer(clientA.Snapshot(), clientB.PeerID()) && hasPeer(clientB.Snapshot(), clientA.PeerID())
	})

	stopB()

	restarted, stopRestarted := startService(t, Config{Role: RoleClient, DataDir: clientBDir, ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopRestarted()
	waitFor(t, 3*time.Second, func() bool {
		snap := restarted.Snapshot()
		return hasPeer(snap, bootstrap.PeerID()) && hasPeer(snap, clientA.PeerID())
	})
}

func TestDiscoverySurfacesSaveFailuresAndRollsBackPeerCache(t *testing.T) {
	service, err := NewService(Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	originalSaveStateFn := saveStateFn
	saveStateFn = func(string, persistedState) error { return errors.New("forced discovery save failure") }
	defer func() { saveStateFn = originalSaveStateFn }()

	if err := service.mergeDiscoveredPeers("bootstrap", []PeerInfo{{PeerID: "peer-1", Role: RoleRelay, Addresses: []string{"127.0.0.1:1234"}, PublicKey: "peer-key"}}); err == nil || !strings.Contains(err.Error(), "forced discovery save failure") {
		t.Fatalf("mergeDiscoveredPeers() error = %v, want forced discovery save failure", err)
	}
	if hasPeer(service.Snapshot(), "peer-1") {
		t.Fatal("expected discovery rollback to remove unsaved peer cache entry")
	}
}

func TestCreateServerRollsBackOnSaveFailure(t *testing.T) {
	service, err := NewService(Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	originalSaveStateFn := saveStateFn
	saveStateFn = func(string, persistedState) error { return errors.New("forced server save failure") }
	defer func() { saveStateFn = originalSaveStateFn }()

	if _, err := service.CreateServer("rollback", "save failure"); err == nil || !strings.Contains(err.Error(), "forced server save failure") {
		t.Fatalf("CreateServer() error = %v, want forced server save failure", err)
	}
	if len(service.Snapshot().Servers) != 0 {
		t.Fatal("expected server creation rollback to leave no persisted server")
	}
}

func TestRecordTelemetryRollsBackOnSaveFailure(t *testing.T) {
	service, err := NewService(Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	originalSaveStateFn := saveStateFn
	saveStateFn = func(string, persistedState) error { return errors.New("forced telemetry save failure") }
	defer func() { saveStateFn = originalSaveStateFn }()

	service.recordTelemetry("test.telemetry entry")
	if containsTelemetry(service.Snapshot().Telemetry, "test.telemetry entry") {
		t.Fatal("expected telemetry rollback to drop unsaved entry")
	}
}

func TestPeerAddressFilteringNormalizesPersistenceAndTargets(t *testing.T) {
	service, err := NewService(Config{
		Role:              RoleClient,
		DataDir:           t.TempDir(),
		ListenAddr:        "127.0.0.1:0",
		BootstrapAddrs:    []string{"http://127.0.0.1:4242/path", "192.168.0.1:80"},
		RelayAddrs:        []string{"https://127.0.0.1:4343/relay", "10.0.0.1:90"},
		ManualPeers:       []string{"http://127.0.0.1:4444", "203.0.113.1:1"},
		DiscoveryInterval: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	service.mu.Lock()
	service.upsertPeerLocked(PeerRecord{PeerID: "peer-1", Role: RoleRelay, Addresses: []string{"https://127.0.0.1:5555/path", "127.0.0.1:5555", "10.0.0.2:55"}, PublicKey: "peer-key"})
	service.mu.Unlock()

	peer, ok := peerByID(service.Snapshot(), "peer-1")
	if !ok {
		t.Fatal("expected peer-1 to be stored")
	}
	if !sameStringSet(peer.Addresses, []string{"/ip4/127.0.0.1/tcp/5555"}) {
		t.Fatalf("peer addresses = %#v want normalized loopback target", peer.Addresses)
	}
	if !sameStringSet(service.bootstrapTargets(), []string{"/ip4/127.0.0.1/tcp/4242"}) {
		t.Fatalf("bootstrap targets = %#v want filtered loopback target", service.bootstrapTargets())
	}
	if !sameStringSet(service.relayTargets(), []string{"/ip4/127.0.0.1/tcp/4343", "/ip4/127.0.0.1/tcp/5555"}) {
		t.Fatalf("relay targets = %#v want filtered loopback targets", service.relayTargets())
	}
}

func TestConfiguredManualPeersAreDiscovered(t *testing.T) {
	manualTarget, stopManualTarget := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopManualTarget()

	manualJoiner, stopManualJoiner := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", ManualPeers: []string{manualTarget.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopManualJoiner()

	waitFor(t, 3*time.Second, func() bool {
		peer, ok := peerByID(manualJoiner.Snapshot(), manualTarget.PeerID())
		return ok && peer.Source == "manual" && len(peer.Addresses) > 0
	})
}

func TestPeerExchangePrefersSharedServerPeersAndFiltersKnownPeers(t *testing.T) {
	service, err := NewService(Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	service.mu.Lock()
	service.state.Identity.PeerID = "self"
	service.state.KnownPeers["owner"] = PeerRecord{PeerID: "owner", Role: RoleArchivist, Addresses: []string{"owner.example:1"}, PublicKey: "owner-key"}
	service.state.KnownPeers["member"] = PeerRecord{PeerID: "member", Role: RoleClient, Addresses: []string{"member.example:2"}, PublicKey: "member-key"}
	service.state.KnownPeers["other"] = PeerRecord{PeerID: "other", Role: RoleRelay, Addresses: []string{"other.example:3"}, PublicKey: "other-key"}
	service.state.Servers["srv"] = ServerRecord{ID: "srv", OwnerPeerID: "owner", Members: []string{"member", "self", "owner"}}
	service.mu.Unlock()

	peers := service.peerExchange(PeerExchangeRequest{KnownPeerIDs: []string{"other"}, ServerIDs: []string{"srv"}, Limit: 4})
	if len(peers) != 2 {
		t.Fatalf("peerExchange() returned %d peers want 2 (%+v)", len(peers), peers)
	}
	if peers[0].PeerID != "owner" || peers[1].PeerID != "member" {
		t.Fatalf("peerExchange() order = %#v want owner/member first", peers)
	}
}

func TestOfflineBootstrapBackoffPreservesKnownPeersAcrossRestart(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	clientDir := t.TempDir()
	bootstrapAddr := bootstrap.ListenAddress()
	client, stopClient := startService(t, Config{Role: RoleClient, DataDir: clientDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrapAddr}, DiscoveryInterval: 40 * time.Millisecond})

	waitFor(t, 3*time.Second, func() bool {
		return hasPeer(client.Snapshot(), bootstrap.PeerID())
	})

	stopBootstrap()

	waitFor(t, 3*time.Second, func() bool {
		snapshot := client.Snapshot()
		return hasPeer(snapshot, bootstrap.PeerID()) && containsTelemetry(snapshot.Telemetry, "discovery.backoff layer=bootstrap-register")
	})

	stopClient()

	restarted, stopRestarted := startService(t, Config{Role: RoleClient, DataDir: clientDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrapAddr}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopRestarted()

	waitFor(t, 3*time.Second, func() bool {
		snapshot := restarted.Snapshot()
		return hasPeer(snapshot, bootstrap.PeerID()) && containsTelemetry(snapshot.Telemetry, "discovery.backoff")
	})
}

func TestRestartRefreshesOwnedServerManifestAndInviteAddresses(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	hostDir := t.TempDir()
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: hostDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	server, err := host.CreateServer("restartable", "refresh addresses")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	oldInvite := server.Invite
	oldAddr := host.ListenAddress()
	stopHost()

	restartAddr := reserveListenAddr(t)
	for restartAddr == oldAddr {
		restartAddr = reserveListenAddr(t)
	}
	restartedHost, stopRestartedHost := startService(t, Config{Role: RoleClient, DataDir: hostDir, ListenAddr: restartAddr, BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopRestartedHost()

	restartedSnapshot := restartedHost.Snapshot()
	if len(restartedSnapshot.Servers) != 1 {
		t.Fatalf("expected one owned server after restart, got %d", len(restartedSnapshot.Servers))
	}
	restartedServer := restartedSnapshot.Servers[0]
	if restartedServer.Invite == oldInvite {
		t.Fatal("expected restart to refresh stored invite")
	}
	if !sameStringSet(restartedServer.Manifest.OwnerAddresses, []string{restartedHost.ListenAddress()}) {
		t.Fatalf("manifest owner addresses = %#v want %#v", restartedServer.Manifest.OwnerAddresses, []string{restartedHost.ListenAddress()})
	}
	invite, err := ParseDeeplink(restartedServer.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	if !sameStringSet(invite.ServerAddrs, []string{restartedHost.ListenAddress()}) {
		t.Fatalf("invite server addrs = %#v want %#v", invite.ServerAddrs, []string{restartedHost.ListenAddress()})
	}
	if invite.ManifestHash != restartedServer.Manifest.Hash() {
		t.Fatalf("invite manifest hash = %q want %q", invite.ManifestHash, restartedServer.Manifest.Hash())
	}

	waitFor(t, 3*time.Second, func() bool {
		for _, discovered := range bootstrap.Snapshot().Servers {
			if discovered.ID == restartedServer.ID {
				return sameStringSet(discovered.Manifest.OwnerAddresses, []string{restartedHost.ListenAddress()})
			}
		}
		return false
	})

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	if _, err := guest.JoinByDeeplink(restartedServer.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() with refreshed invite error = %v", err)
	}
}

func TestInviteJoinMessageRelayFallbackAndHistory(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	guestDir := t.TempDir()
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: guestDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})

	server, err := host.CreateServer("alpha", "test server")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}

	stopGuest()
	msg, err := host.SendChannelMessage(general, "hello while offline")
	if err != nil {
		t.Fatalf("SendChannelMessage() error = %v", err)
	}

	restartedGuest, stopRestartedGuest := startService(t, Config{Role: RoleClient, DataDir: guestDir, ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopRestartedGuest()
	waitFor(t, 3*time.Second, func() bool {
		return hasMessage(restartedGuest.Snapshot(), msg.ID)
	})

	lateGuest, stopLateGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopLateGuest()
	joined, err := lateGuest.JoinByDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("late JoinByDeeplink() error = %v", err)
	}
	if joined.ID != server.ID {
		t.Fatalf("joined server id mismatch: got %s want %s", joined.ID, server.ID)
	}
	waitFor(t, 2*time.Second, func() bool {
		return hasMessage(lateGuest.Snapshot(), msg.ID)
	})

	badInvite, err := ParseDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	badInvite.ManifestHash = "tampered"
	badRaw, _ := json.Marshal(badInvite)
	badLink := "aether://join/" + badInvite.ServerID + "?invite=" + base64.RawURLEncoding.EncodeToString(badRaw)
	if _, err := lateGuest.JoinByDeeplink(badLink); err == nil {
		t.Fatal("expected invalid signature path to fail")
	}

	expiredInvite, err := ParseDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	expiredInvite.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if err := expiredInvite.Sign(host.Snapshot().Identity); err != nil {
		t.Fatalf("expiredInvite.Sign() error = %v", err)
	}
	expiredLink, _ := expiredInvite.Deeplink()
	if _, err := lateGuest.JoinByDeeplink(expiredLink); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired invite error, got %v", err)
	}
}

func TestPreviewAndJoinRejectSignedInviteOwnerMismatch(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("owner-check", "signed invite owner mismatch")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	invite, err := ParseDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	invite.OwnerPeerID = "attacker"
	if err := invite.Sign(host.Snapshot().Identity); err != nil {
		t.Fatalf("invite.Sign() error = %v", err)
	}
	badLink, err := invite.Deeplink()
	if err != nil {
		t.Fatalf("invite.Deeplink() error = %v", err)
	}

	if _, err := guest.PreviewDeeplink(badLink); err == nil || !strings.Contains(err.Error(), "owner peer mismatch") {
		t.Fatalf("PreviewDeeplink() error = %v, want owner peer mismatch", err)
	}
	if _, err := guest.JoinByDeeplink(badLink); err == nil || !strings.Contains(err.Error(), "owner peer mismatch") {
		t.Fatalf("JoinByDeeplink() error = %v, want owner peer mismatch", err)
	}
}

func TestCreateChannelPropagatesToJoinedPeers(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("propagation", "channel propagation")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}

	channel, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		for _, srv := range guest.Snapshot().Servers {
			if srv.ID == server.ID {
				_, ok := srv.Channels[channel.ID]
				return ok
			}
		}
		return false
	})

	msg, err := guest.SendChannelMessage(channel.ID, "hello alerts")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		for _, m := range host.Snapshot().Messages {
			if m.ID == msg.ID && m.ScopeID == channel.ID {
				return true
			}
		}
		return false
	})
}

func TestPreviewDeeplinkFetchesManifestWithoutJoining(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 3})
	defer stopHost()
	server, err := host.CreateServer("previewable", "preview me")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := host.CreateChannel(server.ID, "voice", true); err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	preview, err := guest.PreviewDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("PreviewDeeplink() error = %v", err)
	}
	if preview.Invite.ServerID != server.ID {
		t.Fatalf("preview invite server id = %s want %s", preview.Invite.ServerID, server.ID)
	}
	if preview.Manifest.Hash() != server.Manifest.Hash() {
		t.Fatalf("preview manifest hash = %s want %s", preview.Manifest.Hash(), server.Manifest.Hash())
	}
	if !contains(preview.SafetyLabels, "history-local-window") || !contains(preview.SafetyLabels, "history-single-node") {
		t.Fatalf("preview safety labels = %v", preview.SafetyLabels)
	}
	if preview.MemberCount != 1 {
		t.Fatalf("preview member count = %d want 1", preview.MemberCount)
	}
	if len(preview.Channels) != 2 {
		t.Fatalf("preview channels = %d want 2", len(preview.Channels))
	}
	if !hasChannelNamed(preview.Channels, "general") || !hasChannelNamed(preview.Channels, "voice") {
		t.Fatalf("preview channels = %#v", preview.Channels)
	}
	if len(guest.Snapshot().Servers) != 0 {
		t.Fatalf("preview should not join server, guest servers = %d", len(guest.Snapshot().Servers))
	}
	if got := len(host.Snapshot().Servers[0].Members); got != 1 {
		t.Fatalf("preview should not mutate host membership, got %d members", got)
	}
}

func TestPreviewDeeplinkAddsArchivistSafetyLabel(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleArchivist, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	server, err := host.CreateServer("archivist-preview", "preview archivist owner")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	preview, err := guest.PreviewDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("PreviewDeeplink() error = %v", err)
	}
	if !contains(preview.SafetyLabels, "owner-archivist") {
		t.Fatalf("preview safety labels = %v", preview.SafetyLabels)
	}
	if preview.OwnerRole != RoleArchivist {
		t.Fatalf("preview owner role = %q want %q", preview.OwnerRole, RoleArchivist)
	}
}

func TestResolveServerPreviewIncludesArchivistMetadata(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()
	host, stopHost := startService(t, Config{Role: RoleArchivist, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	server, err := host.CreateServer("preview-info-archivist", "preview info archivist")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	invite, err := ParseDeeplink(server.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	previewInfo, err := guest.resolveServerPreview(server.ID, invite.ServerAddrs)
	if err != nil {
		t.Fatalf("resolveServerPreview() error = %v", err)
	}
	if previewInfo.OwnerRole != RoleArchivist {
		t.Fatalf("preview info owner role = %q want %q", previewInfo.OwnerRole, RoleArchivist)
	}
	if !contains(previewInfo.SafetyLabels, "owner-archivist") {
		t.Fatalf("preview info safety labels = %v", previewInfo.SafetyLabels)
	}
}

func TestControlAPIPreviewReturnsArchivistOwnerRole(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()
	host, stopHost := startService(t, Config{Role: RoleArchivist, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	server, err := host.CreateServer("control-archivist-preview", "control archivist preview")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	token, err := ControlTokenFromDataDir(guest.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var preview ServerPreview
	if err := CallControlJSON(guest.ControlEndpoint(), token, http.MethodPost, "/v1/servers/preview", PreviewServerRequest{Deeplink: server.Invite}, &preview); err != nil {
		t.Fatalf("POST /v1/servers/preview error = %v", err)
	}
	if preview.OwnerRole != RoleArchivist {
		t.Fatalf("preview owner role = %q want %q", preview.OwnerRole, RoleArchivist)
	}
}

func TestControlAPIPreviewsSignedInvite(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	server, err := host.CreateServer("control-preview", "control preview flow")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	token, err := ControlTokenFromDataDir(guest.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var preview ServerPreview
	if err := CallControlJSON(guest.ControlEndpoint(), token, http.MethodPost, "/v1/servers/preview", PreviewServerRequest{Deeplink: server.Invite}, &preview); err != nil {
		t.Fatalf("POST /v1/servers/preview error = %v", err)
	}
	if preview.Manifest.ServerID != server.ID {
		t.Fatalf("preview manifest server id = %s want %s", preview.Manifest.ServerID, server.ID)
	}
	if preview.MemberCount != 1 {
		t.Fatalf("preview member count = %d want 1", preview.MemberCount)
	}
	if len(preview.Channels) != 1 || !hasChannelNamed(preview.Channels, "general") {
		t.Fatalf("preview channels = %#v", preview.Channels)
	}
	if len(guest.Snapshot().Servers) != 0 {
		t.Fatalf("control preview should not join server, guest servers = %d", len(guest.Snapshot().Servers))
	}
}

func TestPresenceReflectsVoiceParticipationAndStalePeers(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	waitFor(t, 3*time.Second, func() bool {
		return hasPeer(host.Snapshot(), guest.PeerID()) && hasPeer(host.Snapshot(), bootstrap.PeerID())
	})
	server, err := host.CreateServer("presence", "presence surface")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	voiceChannel, err := host.CreateChannel(server.ID, "voice", true)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	if err := guest.JoinVoice(voiceChannel.ID, true); err != nil {
		t.Fatalf("JoinVoice() error = %v", err)
	}
	waitFor(t, 3*time.Second, func() bool {
		p, ok := presenceByPeerID(host.Presence(), guest.PeerID())
		return ok && contains(p.ActiveVoiceChannels, voiceChannel.ID)
	})
	host.mu.Lock()
	stale := host.state.KnownPeers[bootstrap.PeerID()]
	stale.LastSeenAt = time.Now().UTC().Add(-20 * time.Minute)
	host.state.KnownPeers[bootstrap.PeerID()] = stale
	host.mu.Unlock()
	guestPresence, ok := presenceByPeerID(host.Presence(), guest.PeerID())
	if !ok {
		t.Fatal("expected guest presence")
	}
	if guestPresence.Status != "online" {
		t.Fatalf("guest presence status = %q want online", guestPresence.Status)
	}
	if !contains(guestPresence.ActiveVoiceChannels, voiceChannel.ID) {
		t.Fatalf("guest active voice channels = %#v want %q", guestPresence.ActiveVoiceChannels, voiceChannel.ID)
	}
	bootstrapPresence, ok := presenceByPeerID(host.Presence(), bootstrap.PeerID())
	if !ok {
		t.Fatal("expected bootstrap presence")
	}
	if bootstrapPresence.Status != "offline" {
		t.Fatalf("bootstrap presence status = %q want offline", bootstrapPresence.Status)
	}
}

func TestControlAPIPresenceReturnsDerivedStatuses(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	server, err := service.CreateServer("control-presence", "presence api")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	voiceChannel, err := service.CreateChannel(server.ID, "voice", true)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if err := service.JoinVoice(voiceChannel.ID, false); err != nil {
		t.Fatalf("JoinVoice() error = %v", err)
	}
	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var resp PresenceResponse
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodGet, "/v1/presence", nil, &resp); err != nil {
		t.Fatalf("GET /v1/presence error = %v", err)
	}
	selfPresence, ok := presenceByPeerID(resp.Peers, service.PeerID())
	if !ok {
		t.Fatal("expected self presence")
	}
	if selfPresence.Status != "online" {
		t.Fatalf("self presence status = %q want online", selfPresence.Status)
	}
	if !contains(selfPresence.ActiveVoiceChannels, voiceChannel.ID) {
		t.Fatalf("self active voice channels = %#v want %q", selfPresence.ActiveVoiceChannels, voiceChannel.ID)
	}
}

func TestSearchMessagesScopesAndOrdersLocalHistory(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	server, err := service.CreateServer("searchable", "search local history")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}
	ops, err := service.CreateChannel(server.ID, "ops", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	first, err := service.SendChannelMessage(general, "deploy alpha")
	if err != nil {
		t.Fatalf("SendChannelMessage(first) error = %v", err)
	}
	second, err := service.SendChannelMessage(general, "deploy beta")
	if err != nil {
		t.Fatalf("SendChannelMessage(second) error = %v", err)
	}
	if _, err := service.SendChannelMessage(ops.ID, "deploy gamma"); err != nil {
		t.Fatalf("SendChannelMessage(ops) error = %v", err)
	}
	results, err := service.SearchMessages(SearchMessagesRequest{Query: "deploy", ScopeType: "channel", ScopeID: general, Limit: 10})
	if err != nil {
		t.Fatalf("SearchMessages() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("result count = %d want 2", len(results))
	}
	if results[0].ID != second.ID || results[1].ID != first.ID {
		t.Fatalf("result ids = %v want [%s %s]", []string{results[0].ID, results[1].ID}, second.ID, first.ID)
	}
	for _, msg := range results {
		if msg.ScopeID != general {
			t.Fatalf("unexpected scope id %s in results", msg.ScopeID)
		}
	}
}

func TestNormalizeNotificationScopePrefersDMServerErrorBeforeMissingScopeID(t *testing.T) {
	_, _, _, err := normalizeNotificationScope("server-1", "dm", "")
	if err == nil || !strings.Contains(err.Error(), "server id is not valid for dm scope") {
		t.Fatalf("normalizeNotificationScope() error = %v", err)
	}
}

func TestSearchMessagesRejectsInvalidScopeCombinations(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()

	if _, err := service.SearchMessages(SearchMessagesRequest{ScopeType: "dm", ServerID: "server-1"}); err == nil || !strings.Contains(err.Error(), "server id is not valid for dm scope") {
		t.Fatalf("SearchMessages(dm with server) error = %v", err)
	}
	if _, err := service.SearchMessages(SearchMessagesRequest{ScopeType: "channel"}); err == nil || !strings.Contains(err.Error(), "scope id is required when scope type is set") {
		t.Fatalf("SearchMessages(scope without id) error = %v", err)
	}
}

func TestControlAPISearchMessagesIncludesDeletedWhenRequested(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	server, err := service.CreateServer("search-api", "search through control API")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}
	live, err := service.SendChannelMessage(general, "alpha ready")
	if err != nil {
		t.Fatalf("SendChannelMessage(live) error = %v", err)
	}
	deleted, err := service.SendChannelMessage(general, "beta tombstone")
	if err != nil {
		t.Fatalf("SendChannelMessage(deleted) error = %v", err)
	}
	if err := service.DeleteMessage(deleted.ID); err != nil {
		t.Fatalf("DeleteMessage() error = %v", err)
	}
	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var hidden SearchMessagesResponse
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/messages/search", SearchMessagesRequest{ServerID: server.ID, Limit: 10}, &hidden); err != nil {
		t.Fatalf("POST /v1/messages/search hidden error = %v", err)
	}
	if len(hidden.Messages) != 1 || hidden.Messages[0].ID != live.ID {
		t.Fatalf("hidden search results = %#v", hidden.Messages)
	}
	if len(hidden.Results) != 1 || hidden.Results[0].Message.ID != live.ID {
		t.Fatalf("hidden search result views = %#v", hidden.Results)
	}
	if hidden.Results[0].ServerName != "search-api" || hidden.Results[0].ScopeName != "general" {
		t.Fatalf("hidden message search labels = server=%q scope=%q", hidden.Results[0].ServerName, hidden.Results[0].ScopeName)
	}
	var visible SearchMessagesResponse
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/messages/search", SearchMessagesRequest{ServerID: server.ID, IncludeDeleted: true, Limit: 10}, &visible); err != nil {
		t.Fatalf("POST /v1/messages/search visible error = %v", err)
	}
	if len(visible.Messages) != 2 {
		t.Fatalf("visible search count = %d want 2", len(visible.Messages))
	}
	if visible.Messages[0].ID != deleted.ID || !visible.Messages[0].Deleted {
		t.Fatalf("expected deleted message first, got %#v", visible.Messages[0])
	}
	if visible.Messages[1].ID != live.ID || visible.Messages[1].Deleted {
		t.Fatalf("expected live message second, got %#v", visible.Messages[1])
	}
	if len(visible.Results) != 2 {
		t.Fatalf("visible result views count = %d want 2", len(visible.Results))
	}
	if visible.Results[0].Message.ID != deleted.ID || visible.Results[1].Message.ID != live.ID {
		t.Fatalf("visible result views ordering = %#v", visible.Results)
	}
}

func TestControlAPISearchMessagesIncludesDMParticipantIDs(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	selfIdentity, err := service.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("service.CreateIdentity() error = %v", err)
	}
	peer, stopPeer := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopPeer()
	peerIdentity, err := peer.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("peer.CreateIdentity() error = %v", err)
	}

	dm, err := service.CreateDM(peerIdentity.PeerID)
	if err != nil {
		t.Fatalf("service.CreateDM() error = %v", err)
	}
	if _, err := peer.CreateDM(selfIdentity.PeerID); err != nil {
		t.Fatalf("peer.CreateDM() error = %v", err)
	}
	msg, err := peer.SendDMMessage(dm.ID, "hello alice")
	if err != nil {
		t.Fatalf("peer.SendDMMessage() error = %v", err)
	}

	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var resp SearchMessagesResponse
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/messages/search", SearchMessagesRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10}, &resp); err != nil {
		t.Fatalf("POST /v1/messages/search dm error = %v", err)
	}
	if len(resp.Messages) != 1 || len(resp.Results) != 1 {
		t.Fatalf("dm search response = %#v", resp)
	}
	if resp.Results[0].Message.ID != msg.ID {
		t.Fatalf("dm result message id = %q want %q", resp.Results[0].Message.ID, msg.ID)
	}
	if resp.Results[0].ScopeName != peerIdentity.PeerID {
		t.Fatalf("dm result scope name = %q want %q", resp.Results[0].ScopeName, peerIdentity.PeerID)
	}
	if len(resp.Results[0].ParticipantIDs) != 1 || resp.Results[0].ParticipantIDs[0] != peerIdentity.PeerID {
		t.Fatalf("dm result participant ids = %#v want [%q]", resp.Results[0].ParticipantIDs, peerIdentity.PeerID)
	}
	if resp.Results[0].ServerName != "" {
		t.Fatalf("unexpected server name on dm result = %q", resp.Results[0].ServerName)
	}
}

func TestCannotEditOrDeleteAnotherPeersMessage(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("ownership", "message ownership")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	remote, err := guest.SendChannelMessage(channelID, "guest owned")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		for _, msg := range host.Snapshot().Messages {
			if msg.ID == remote.ID && msg.Body == "guest owned" && !msg.Deleted {
				return true
			}
		}
		return false
	})

	if _, err := host.EditMessage(remote.ID, "tampered"); err == nil || !strings.Contains(err.Error(), "owned by another peer") {
		t.Fatalf("host.EditMessage() error = %v, want ownership failure", err)
	}
	if err := host.DeleteMessage(remote.ID); err == nil || !strings.Contains(err.Error(), "owned by another peer") {
		t.Fatalf("host.DeleteMessage() error = %v, want ownership failure", err)
	}

	for _, msg := range host.Snapshot().Messages {
		if msg.ID == remote.ID {
			if msg.Body != "guest owned" || msg.Deleted {
				t.Fatalf("message mutated after rejected ownership checks: %#v", msg)
			}
			return
		}
	}
	t.Fatalf("message %s not found in host snapshot", remote.ID)
}

func TestApplyDeliveryRejectsForgedMessageMutations(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("forgery", "forged message mutation")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	original, err := host.SendChannelMessage(channelID, "host owned")
	if err != nil {
		t.Fatalf("host.SendChannelMessage() error = %v", err)
	}

	guestIdentity := guest.Snapshot().Identity
	forgedEdit := Delivery{
		ID:               original.ID,
		Kind:             "message_edit",
		ScopeID:          original.ScopeID,
		ScopeType:        original.ScopeType,
		ServerID:         original.ServerID,
		SenderPeerID:     guest.PeerID(),
		SenderPublicKey:  guestIdentity.PublicKey,
		RecipientPeerIDs: []string{host.PeerID()},
		Body:             "forged edit",
		CreatedAt:        time.Now().UTC(),
	}
	if err := forgedEdit.Sign(guestIdentity); err != nil {
		t.Fatalf("forgedEdit.Sign() error = %v", err)
	}
	if err := host.applyDelivery(forgedEdit); err == nil || !strings.Contains(err.Error(), "sender mismatch") {
		t.Fatalf("host.applyDelivery(forgedEdit) error = %v, want sender mismatch", err)
	}

	forgedDelete := Delivery{
		ID:               original.ID,
		Kind:             "message_delete",
		ScopeID:          original.ScopeID,
		ScopeType:        original.ScopeType,
		ServerID:         original.ServerID,
		SenderPeerID:     guest.PeerID(),
		SenderPublicKey:  guestIdentity.PublicKey,
		RecipientPeerIDs: []string{host.PeerID()},
		CreatedAt:        time.Now().UTC().Add(time.Millisecond),
	}
	if err := forgedDelete.Sign(guestIdentity); err != nil {
		t.Fatalf("forgedDelete.Sign() error = %v", err)
	}
	if err := host.applyDelivery(forgedDelete); err == nil || !strings.Contains(err.Error(), "sender mismatch") {
		t.Fatalf("host.applyDelivery(forgedDelete) error = %v, want sender mismatch", err)
	}

	for _, msg := range host.Snapshot().Messages {
		if msg.ID == original.ID {
			if msg.Body != "host owned" || msg.Deleted {
				t.Fatalf("message mutated after forged deliveries: %#v", msg)
			}
			return
		}
	}
	t.Fatalf("message %s not found in host snapshot", original.ID)
}

func TestPeerTransportRejectsBadDeliverySignature(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	participants := dedupeSorted([]string{host.PeerID(), guest.PeerID()})
	delivery := Delivery{
		ID:               randomID("dm"),
		Kind:             "dm_create",
		ScopeID:          strings.Join(participants, ":"),
		ScopeType:        "dm",
		SenderPeerID:     guest.PeerID(),
		SenderPublicKey:  guest.Snapshot().Identity.PublicKey,
		RecipientPeerIDs: []string{host.PeerID()},
		CreatedAt:        time.Now().UTC(),
	}
	if err := delivery.Sign(guest.Snapshot().Identity); err != nil {
		t.Fatalf("delivery.Sign() error = %v", err)
	}
	delivery.CreatedAt = delivery.CreatedAt.Add(time.Second)

	_, err := network.NewClient(time.Second).Call(context.Background(), host.ListenAddress(), network.OperationDeliver, delivery, nil)
	if err == nil {
		t.Fatal("expected peer transport to reject tampered delivery signature")
	}
	transportErr, ok := err.(*network.Error)
	if !ok {
		t.Fatalf("transport error type = %T want *network.Error", err)
	}
	if transportErr.Code != "invalid_signature" {
		t.Fatalf("transport error code = %q want invalid_signature", transportErr.Code)
	}
}

func TestIncomingMentionEmitsNotificationCreatedEvent(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("CreateIdentity(host) error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	if _, err := guest.CreateIdentity("bob", ""); err != nil {
		t.Fatalf("CreateIdentity(guest) error = %v", err)
	}

	server, err := host.CreateServer("notif-events", "incoming mention events")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	events, cancel := host.Subscribe()
	defer cancel()
	msg, err := guest.SendChannelMessage(channelID, "hello @alice")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type != "notification.created" {
				continue
			}
			if got, _ := event.Payload["message_id"].(string); got != msg.ID {
				continue
			}
			tokens, _ := event.Payload["tokens"].([]string)
			if !contains(tokens, "@alice") {
				t.Fatalf("notification tokens = %#v", tokens)
			}
			if unread, ok := event.Payload["unread"].(bool); !ok || !unread {
				t.Fatalf("notification unread payload = %#v", event.Payload["unread"])
			}
			if got, _ := event.Payload["sender_peer_id"].(string); got != guest.PeerID() {
				t.Fatalf("notification sender_peer_id = %q want %q", got, guest.PeerID())
			}
			if got, _ := event.Payload["scope_type"].(string); got != "channel" {
				t.Fatalf("notification scope_type = %q want channel", got)
			}
			if got, _ := event.Payload["server_id"].(string); got != server.ID {
				t.Fatalf("notification server_id = %q want %q", got, server.ID)
			}
			if got, _ := event.Payload["server_name"].(string); got != "notif-events" {
				t.Fatalf("notification server_name = %q want notif-events", got)
			}
			if got, _ := event.Payload["scope_name"].(string); got != "general" {
				t.Fatalf("notification scope_name = %q want general", got)
			}
			return
		case <-deadline:
			t.Fatal("expected notification.created event for incoming mention")
		}
	}
}

func TestMentionEditEmitsNotificationCreatedEvent(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("CreateIdentity(host) error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	if _, err := guest.CreateIdentity("bob", ""); err != nil {
		t.Fatalf("CreateIdentity(guest) error = %v", err)
	}

	server, err := host.CreateServer("notif-edit-events", "edit mention events")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	msg, err := guest.SendChannelMessage(channelID, "hello everyone")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10})
		return err == nil && len(resp.Notifications) == 0
	})

	events, cancel := host.Subscribe()
	defer cancel()
	if _, err := guest.EditMessage(msg.ID, "hello @alice"); err != nil {
		t.Fatalf("guest.EditMessage() error = %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type != "notification.created" {
				continue
			}
			if got, _ := event.Payload["message_id"].(string); got == msg.ID {
				return
			}
		case <-deadline:
			t.Fatal("expected notification.created event for mention-introducing edit")
		}
	}
}

func TestDMMentionNotificationCreatedEventIncludesParticipantIDs(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}
	if _, err := guest.CreateDM(hostIdentity.PeerID); err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}

	events, cancel := host.Subscribe()
	defer cancel()
	msg, err := guest.SendDMMessage(dm.ID, "hello @alice")
	if err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type != "notification.created" {
				continue
			}
			if got, _ := event.Payload["message_id"].(string); got != msg.ID {
				continue
			}
			if got, _ := event.Payload["scope_type"].(string); got != "dm" {
				t.Fatalf("notification scope_type = %q want dm", got)
			}
			if got, _ := event.Payload["scope_id"].(string); got != dm.ID {
				t.Fatalf("notification scope_id = %q want %q", got, dm.ID)
			}
			if got, _ := event.Payload["scope_name"].(string); got != guestIdentity.PeerID {
				t.Fatalf("notification scope_name = %q want %q", got, guestIdentity.PeerID)
			}
			if got, _ := event.Payload["sender_peer_id"].(string); got != guestIdentity.PeerID {
				t.Fatalf("notification sender_peer_id = %q want %q", got, guestIdentity.PeerID)
			}
			participantIDs, _ := event.Payload["participant_ids"].([]string)
			if len(participantIDs) != 1 || participantIDs[0] != guestIdentity.PeerID {
				t.Fatalf("notification participant_ids = %#v want [%q]", participantIDs, guestIdentity.PeerID)
			}
			return
		case <-deadline:
			t.Fatal("expected notification.created event for dm mention")
		}
	}
}

func TestSearchNotificationsTracksUnreadMentions(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	hostIdentity, err := host.CreateIdentity("alice", "host")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	if hostIdentity.Profile.DisplayName != "alice" {
		t.Fatalf("display name = %q want alice", hostIdentity.Profile.DisplayName)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("notifications", "notification surface")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	if _, err := host.SendChannelMessage(channelID, "self ping @alice"); err != nil {
		t.Fatalf("host.SendChannelMessage() error = %v", err)
	}
	msg, err := guest.SendChannelMessage(channelID, "hello @alice")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10})
		return err == nil && len(resp.Notifications) == 1
	})

	resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications() error = %v", err)
	}
	if resp.UnreadCount != 1 {
		t.Fatalf("unread count = %d want 1", resp.UnreadCount)
	}
	if len(resp.Notifications) != 1 {
		t.Fatalf("notification count = %d want 1", len(resp.Notifications))
	}
	note := resp.Notifications[0]
	if note.Kind != "mention" {
		t.Fatalf("notification kind = %q want mention", note.Kind)
	}
	if note.Message.ID != msg.ID {
		t.Fatalf("notification message id = %s want %s", note.Message.ID, msg.ID)
	}
	if !note.Unread {
		t.Fatalf("expected notification to be unread")
	}
	if !contains(note.Tokens, "@alice") {
		t.Fatalf("notification tokens = %#v", note.Tokens)
	}
	if note.ServerName != "notifications" || note.ScopeName != "general" {
		t.Fatalf("notification labels = server=%q scope=%q", note.ServerName, note.ScopeName)
	}
	if len(note.ParticipantIDs) != 0 {
		t.Fatalf("unexpected participant ids for channel notification = %#v", note.ParticipantIDs)
	}

	readThrough, err := host.MarkNotificationsRead(note.CreatedAt)
	if err != nil {
		t.Fatalf("MarkNotificationsRead() error = %v", err)
	}
	if readThrough.IsZero() {
		t.Fatal("expected non-zero read through timestamp")
	}

	resp, err = host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications() after read error = %v", err)
	}
	if resp.UnreadCount != 0 {
		t.Fatalf("unread count after read = %d want 0", resp.UnreadCount)
	}
	if len(resp.Notifications) != 1 || resp.Notifications[0].Unread {
		t.Fatalf("expected one read notification after mark read, got %#v", resp.Notifications)
	}

	resp, err = host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10, UnreadOnly: true})
	if err != nil {
		t.Fatalf("SearchNotifications(unread only) error = %v", err)
	}
	if len(resp.Notifications) != 0 {
		t.Fatalf("unread-only notifications = %d want 0", len(resp.Notifications))
	}
}

func TestControlAPINotificationsSearchAndRead(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("notifications-api", "notifications through control api")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	msg, err := guest.SendChannelMessage(channelID, "hey @alice")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp SearchNotificationsResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: server.ID, Limit: 10}, &resp) == nil && len(resp.Notifications) == 1
	})

	var resp SearchNotificationsResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: server.ID, Limit: 10}, &resp); err != nil {
		t.Fatalf("POST /v1/notifications/search error = %v", err)
	}
	if resp.UnreadCount != 1 || len(resp.Notifications) != 1 {
		t.Fatalf("search response = %#v", resp)
	}
	if resp.Notifications[0].Message.ID != msg.ID {
		t.Fatalf("notification message id = %s want %s", resp.Notifications[0].Message.ID, msg.ID)
	}

	var markResp MarkNotificationsReadResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/read", MarkNotificationsReadRequest{Through: resp.Notifications[0].CreatedAt}, &markResp); err != nil {
		t.Fatalf("POST /v1/notifications/read error = %v", err)
	}
	if markResp.ReadThrough.IsZero() {
		t.Fatal("expected non-zero read through timestamp")
	}

	var unreadOnly SearchNotificationsResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: server.ID, Limit: 10, UnreadOnly: true}, &unreadOnly); err != nil {
		t.Fatalf("POST /v1/notifications/search unread only error = %v", err)
	}
	if unreadOnly.UnreadCount != 0 || len(unreadOnly.Notifications) != 0 {
		t.Fatalf("unread-only notifications after mark read = %#v", unreadOnly)
	}
}

func TestServerScopedMarkNotificationsReadLeavesOtherServersUnread(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	serverA, err := host.CreateServer("server-a", "first server")
	if err != nil {
		t.Fatalf("CreateServer(serverA) error = %v", err)
	}
	serverB, err := host.CreateServer("server-b", "second server")
	if err != nil {
		t.Fatalf("CreateServer(serverB) error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(serverA.Invite); err != nil {
		t.Fatalf("JoinByDeeplink(serverA) error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(serverB.Invite); err != nil {
		t.Fatalf("JoinByDeeplink(serverB) error = %v", err)
	}
	channelA := firstChannelID(serverA)
	channelB := firstChannelID(serverB)
	if channelA == "" || channelB == "" {
		t.Fatal("expected default channels")
	}

	msgA, err := guest.SendChannelMessage(channelA, "hello server-a @alice")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage(serverA) error = %v", err)
	}
	if _, err := guest.SendChannelMessage(channelB, "hello server-b @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(serverB) error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{Limit: 10, UnreadOnly: true})
		return err == nil && resp.UnreadCount == 2 && len(resp.Notifications) == 2
	})

	through, err := host.MarkNotificationsReadScoped(MarkNotificationsReadRequest{Through: msgA.CreatedAt, ServerID: serverA.ID})
	if err != nil {
		t.Fatalf("MarkNotificationsReadScoped(server) error = %v", err)
	}
	if through.IsZero() {
		t.Fatal("expected non-zero server-scoped read through")
	}

	serverAUnread, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: serverA.ID, Limit: 10, UnreadOnly: true})
	if err != nil {
		t.Fatalf("SearchNotifications(serverA unread) error = %v", err)
	}
	if serverAUnread.UnreadCount != 0 || len(serverAUnread.Notifications) != 0 {
		t.Fatalf("serverA unread notifications = %#v", serverAUnread)
	}

	serverBUnread, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: serverB.ID, Limit: 10, UnreadOnly: true})
	if err != nil {
		t.Fatalf("SearchNotifications(serverB unread) error = %v", err)
	}
	if serverBUnread.UnreadCount != 1 || len(serverBUnread.Notifications) != 1 {
		t.Fatalf("serverB unread notifications = %#v", serverBUnread)
	}

	summary := host.NotificationSummary()
	if summary.TotalUnread != 1 || len(summary.Buckets) != 1 {
		t.Fatalf("summary after server-scoped read = %#v", summary)
	}
	if summary.Buckets[0].ServerID != serverB.ID {
		t.Fatalf("remaining unread bucket = %#v want server %q", summary.Buckets[0], serverB.ID)
	}
}

func TestControlAPIServerScopedNotificationsReadOnlyClearsSelectedServer(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	serverA, err := host.CreateServer("server-a-api", "first server api")
	if err != nil {
		t.Fatalf("CreateServer(serverA) error = %v", err)
	}
	serverB, err := host.CreateServer("server-b-api", "second server api")
	if err != nil {
		t.Fatalf("CreateServer(serverB) error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(serverA.Invite); err != nil {
		t.Fatalf("JoinByDeeplink(serverA) error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(serverB.Invite); err != nil {
		t.Fatalf("JoinByDeeplink(serverB) error = %v", err)
	}
	channelA := firstChannelID(serverA)
	channelB := firstChannelID(serverB)
	if channelA == "" || channelB == "" {
		t.Fatal("expected default channels")
	}

	msgA, err := guest.SendChannelMessage(channelA, "hello server-a @alice")
	if err != nil {
		t.Fatalf("guest.SendChannelMessage(serverA) error = %v", err)
	}
	if _, err := guest.SendChannelMessage(channelB, "hello server-b @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(serverB) error = %v", err)
	}

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp SearchNotificationsResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: serverA.ID, Limit: 10}, &resp) == nil && len(resp.Notifications) == 1
	})

	var markResp MarkNotificationsReadResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/read", MarkNotificationsReadRequest{Through: msgA.CreatedAt, ServerID: serverA.ID}, &markResp); err != nil {
		t.Fatalf("POST /v1/notifications/read server-scoped error = %v", err)
	}
	if markResp.ReadThrough.IsZero() {
		t.Fatal("expected non-zero server-scoped read through")
	}
	if markResp.ServerID != serverA.ID || markResp.ServerName != "server-a-api" {
		t.Fatalf("server-scoped read response = %#v", markResp)
	}

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if summary.TotalUnread != 1 || len(summary.Buckets) != 1 {
		t.Fatalf("summary after server-scoped read = %#v", summary)
	}
	if summary.Buckets[0].ServerID != serverB.ID {
		t.Fatalf("remaining summary bucket = %#v want server %q", summary.Buckets[0], serverB.ID)
	}
}

func TestScopedMarkNotificationsReadLeavesOtherScopesUnread(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("notifications-scoped-read", "scoped read state")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	generalID := firstChannelID(server)
	if generalID == "" {
		t.Fatal("expected default channel")
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}

	if _, err := guest.SendChannelMessage(generalID, "general @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(general) error = %v", err)
	}
	if _, err := guest.SendChannelMessage(alerts.ID, "alerts @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(alerts) error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, Limit: 10, UnreadOnly: true})
		return err == nil && resp.UnreadCount == 2 && len(resp.Notifications) == 2
	})

	generalResp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: generalID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications(general) error = %v", err)
	}
	if len(generalResp.Notifications) != 1 {
		t.Fatalf("general notification count = %d want 1", len(generalResp.Notifications))
	}
	through, err := host.MarkNotificationsReadScoped(MarkNotificationsReadRequest{Through: generalResp.Notifications[0].CreatedAt, ServerID: server.ID, ScopeType: "channel", ScopeID: generalID})
	if err != nil {
		t.Fatalf("MarkNotificationsReadScoped() error = %v", err)
	}
	if through.IsZero() {
		t.Fatal("expected non-zero scoped read through")
	}

	generalUnread, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: generalID, Limit: 10, UnreadOnly: true})
	if err != nil {
		t.Fatalf("SearchNotifications(general unread) error = %v", err)
	}
	if generalUnread.UnreadCount != 0 || len(generalUnread.Notifications) != 0 {
		t.Fatalf("general unread notifications = %#v", generalUnread)
	}

	alertsUnread, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: alerts.ID, Limit: 10, UnreadOnly: true})
	if err != nil {
		t.Fatalf("SearchNotifications(alerts unread) error = %v", err)
	}
	if alertsUnread.UnreadCount != 1 || len(alertsUnread.Notifications) != 1 {
		t.Fatalf("alerts unread notifications = %#v", alertsUnread)
	}

	summary := host.NotificationSummary()
	if summary.TotalUnread != 1 || len(summary.Buckets) != 1 {
		t.Fatalf("summary after scoped read = %#v", summary)
	}
	if summary.Buckets[0].ScopeID != alerts.ID {
		t.Fatalf("remaining unread bucket = %#v want scope %q", summary.Buckets[0], alerts.ID)
	}
}

func TestControlAPIScopedNotificationsReadOnlyClearsSelectedBucket(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("notifications-scoped-api", "scoped api read state")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	generalID := firstChannelID(server)
	if generalID == "" {
		t.Fatal("expected default channel")
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}

	if _, err := guest.SendChannelMessage(generalID, "general @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(general) error = %v", err)
	}
	if _, err := guest.SendChannelMessage(alerts.ID, "alerts @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage(alerts) error = %v", err)
	}

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp SearchNotificationsResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: generalID, Limit: 10}, &resp) == nil && len(resp.Notifications) == 1
	})

	var generalResp SearchNotificationsResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: generalID, Limit: 10}, &generalResp); err != nil {
		t.Fatalf("POST /v1/notifications/search general error = %v", err)
	}
	if len(generalResp.Notifications) != 1 {
		t.Fatalf("general response = %#v", generalResp)
	}

	var markResp MarkNotificationsReadResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/read", MarkNotificationsReadRequest{Through: generalResp.Notifications[0].CreatedAt, ServerID: server.ID, ScopeType: "channel", ScopeID: generalID}, &markResp); err != nil {
		t.Fatalf("POST /v1/notifications/read scoped error = %v", err)
	}
	if markResp.ReadThrough.IsZero() {
		t.Fatal("expected non-zero scoped read through")
	}
	if markResp.ServerID != server.ID || markResp.ServerName != "notifications-scoped-api" || markResp.ScopeType != "channel" || markResp.ScopeID != generalID || markResp.ScopeName != "general" {
		t.Fatalf("scoped read response = %#v", markResp)
	}

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if summary.TotalUnread != 1 || len(summary.Buckets) != 1 {
		t.Fatalf("summary after scoped read = %#v", summary)
	}
	if summary.Buckets[0].ScopeID != alerts.ID {
		t.Fatalf("remaining summary bucket = %#v want scope %q", summary.Buckets[0], alerts.ID)
	}
}

func TestScopedNotificationsReadEventIncludesServerAndScopeLabels(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("read-event-server", "notifications read event")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if _, err := guest.SendChannelMessage(alerts.ID, "hello @alice"); err != nil {
		t.Fatalf("guest.SendChannelMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: alerts.ID, Limit: 10})
		return err == nil && len(resp.Notifications) == 1
	})
	resp, err := host.SearchNotifications(SearchNotificationsRequest{ServerID: server.ID, ScopeType: "channel", ScopeID: alerts.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications() error = %v", err)
	}
	if len(resp.Notifications) != 1 {
		t.Fatalf("notifications = %#v", resp)
	}

	events, cancel := host.Subscribe()
	defer cancel()
	if _, err := host.MarkNotificationsReadScoped(MarkNotificationsReadRequest{Through: resp.Notifications[0].CreatedAt, ServerID: server.ID, ScopeType: "channel", ScopeID: alerts.ID}); err != nil {
		t.Fatalf("MarkNotificationsReadScoped() error = %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type != "notifications.read" {
				continue
			}
			if got, _ := event.Payload["server_id"].(string); got != server.ID {
				t.Fatalf("server_id = %q want %q", got, server.ID)
			}
			if got, _ := event.Payload["server_name"].(string); got != "read-event-server" {
				t.Fatalf("server_name = %q want read-event-server", got)
			}
			if got, _ := event.Payload["scope_type"].(string); got != "channel" {
				t.Fatalf("scope_type = %q want channel", got)
			}
			if got, _ := event.Payload["scope_id"].(string); got != alerts.ID {
				t.Fatalf("scope_id = %q want %q", got, alerts.ID)
			}
			if got, _ := event.Payload["scope_name"].(string); got != "alerts" {
				t.Fatalf("scope_name = %q want alerts", got)
			}
			return
		case <-deadline:
			t.Fatal("expected notifications.read event for scoped channel read")
		}
	}
}

func TestDMNotificationsReadEventIncludesParticipantIDs(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}
	if _, err := guest.CreateDM(hostIdentity.PeerID); err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}
	if _, err := guest.SendDMMessage(dm.ID, "hello @alice"); err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10})
		return err == nil && len(resp.Notifications) == 1
	})
	resp, err := host.SearchNotifications(SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications() error = %v", err)
	}
	if len(resp.Notifications) != 1 {
		t.Fatalf("notifications = %#v", resp)
	}

	events, cancel := host.Subscribe()
	defer cancel()
	if _, err := host.MarkNotificationsReadScoped(MarkNotificationsReadRequest{Through: resp.Notifications[0].CreatedAt, ScopeType: "dm", ScopeID: dm.ID}); err != nil {
		t.Fatalf("MarkNotificationsReadScoped() error = %v", err)
	}

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-events:
			if event.Type != "notifications.read" {
				continue
			}
			if got, _ := event.Payload["scope_type"].(string); got != "dm" {
				t.Fatalf("scope_type = %q want dm", got)
			}
			if got, _ := event.Payload["scope_id"].(string); got != dm.ID {
				t.Fatalf("scope_id = %q want %q", got, dm.ID)
			}
			if got, _ := event.Payload["scope_name"].(string); got != guestIdentity.PeerID {
				t.Fatalf("scope_name = %q want %q", got, guestIdentity.PeerID)
			}
			participantIDs, _ := event.Payload["participant_ids"].([]string)
			if len(participantIDs) != 1 || participantIDs[0] != guestIdentity.PeerID {
				t.Fatalf("participant_ids = %#v want [%q]", participantIDs, guestIdentity.PeerID)
			}
			if _, ok := event.Payload["server_id"]; ok {
				t.Fatalf("unexpected server_id in dm notifications.read payload: %#v", event.Payload["server_id"])
			}
			return
		case <-deadline:
			t.Fatal("expected notifications.read event for dm read")
		}
	}
}

func TestSearchNotificationsDMResultsIncludeParticipantIDs(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}
	if _, err := guest.CreateDM(hostIdentity.PeerID); err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}
	if _, err := guest.SendDMMessage(dm.ID, "hello @alice"); err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		resp, err := host.SearchNotifications(SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10})
		return err == nil && len(resp.Notifications) == 1
	})

	resp, err := host.SearchNotifications(SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchNotifications(dm) error = %v", err)
	}
	if len(resp.Notifications) != 1 {
		t.Fatalf("dm notifications = %#v", resp)
	}
	note := resp.Notifications[0]
	if note.ScopeName != guestIdentity.PeerID {
		t.Fatalf("dm notification scope name = %q want %q", note.ScopeName, guestIdentity.PeerID)
	}
	if len(note.ParticipantIDs) != 1 || note.ParticipantIDs[0] != guestIdentity.PeerID {
		t.Fatalf("dm notification participant ids = %#v want [%q]", note.ParticipantIDs, guestIdentity.PeerID)
	}
	if note.ServerName != "" {
		t.Fatalf("unexpected server name on dm notification = %q", note.ServerName)
	}
}

func TestControlAPINotificationsReadReturnsDMParticipantIDs(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}
	if _, err := guest.CreateDM(hostIdentity.PeerID); err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}
	if _, err := guest.SendDMMessage(dm.ID, "hello @alice"); err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp SearchNotificationsResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10}, &resp) == nil && len(resp.Notifications) == 1
	})

	var resp SearchNotificationsResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/search", SearchNotificationsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10}, &resp); err != nil {
		t.Fatalf("POST /v1/notifications/search dm error = %v", err)
	}
	if len(resp.Notifications) != 1 {
		t.Fatalf("dm notifications = %#v", resp)
	}

	var markResp MarkNotificationsReadResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/read", MarkNotificationsReadRequest{Through: resp.Notifications[0].CreatedAt, ScopeType: "dm", ScopeID: dm.ID}, &markResp); err != nil {
		t.Fatalf("POST /v1/notifications/read dm error = %v", err)
	}
	if markResp.ReadThrough.IsZero() {
		t.Fatal("expected non-zero dm read through")
	}
	if markResp.ScopeType != "dm" || markResp.ScopeID != dm.ID || markResp.ScopeName != guestIdentity.PeerID {
		t.Fatalf("dm read response = %#v", markResp)
	}
	if len(markResp.ParticipantIDs) != 1 || markResp.ParticipantIDs[0] != guestIdentity.PeerID {
		t.Fatalf("dm read participants = %#v", markResp.ParticipantIDs)
	}
	if markResp.ServerID != "" || markResp.ServerName != "" {
		t.Fatalf("unexpected server metadata in dm read response = %#v", markResp)
	}
}

func TestNotificationSummaryIncludesDirectAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	bob := newTestIdentity(t, "bob")
	carol := newTestIdentity(t, "carol")
	bobDM := dmScopeID(host.PeerID(), bob.PeerID)
	carolDM := dmScopeID(host.PeerID(), carol.PeerID)
	applyRemoteMessage(t, host, bob, "dm", bobDM, "", "hello @alice one")
	applyRemoteMessage(t, host, bob, "dm", bobDM, "", "hello @alice two")
	applyRemoteMessage(t, host, carol, "dm", carolDM, "", "hello @alice three")

	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 3 })
	summary := host.NotificationSummary()
	if len(summary.Directs) != 2 {
		t.Fatalf("direct aggregate count = %d want 2; summary=%#v", len(summary.Directs), summary)
	}
	bobBucket, ok := notificationDirectBucketByID(summary.Directs, bobDM)
	if !ok {
		t.Fatalf("bob direct aggregate missing in %#v", summary.Directs)
	}
	if bobBucket.ScopeName != bob.PeerID || bobBucket.UnreadCount != 2 {
		t.Fatalf("bob direct aggregate = %#v", bobBucket)
	}
	carolBucket, ok := notificationDirectBucketByID(summary.Directs, carolDM)
	if !ok {
		t.Fatalf("carol direct aggregate missing in %#v", summary.Directs)
	}
	if carolBucket.ScopeName != carol.PeerID || carolBucket.UnreadCount != 1 {
		t.Fatalf("carol direct aggregate = %#v", carolBucket)
	}
}

func TestControlAPINotificationSummaryIncludesDirectAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	bob := newTestIdentity(t, "bob")
	carol := newTestIdentity(t, "carol")
	bobDM := dmScopeID(host.PeerID(), bob.PeerID)
	carolDM := dmScopeID(host.PeerID(), carol.PeerID)
	applyRemoteMessage(t, host, bob, "dm", bobDM, "", "hello @alice one")
	applyRemoteMessage(t, host, carol, "dm", carolDM, "", "hello @alice two")

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		var resp NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &resp) == nil && resp.TotalUnread == 2 && len(resp.Directs) == 2
	})
	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	bobBucket, ok := notificationDirectBucketByID(summary.Directs, bobDM)
	if !ok {
		t.Fatalf("bob direct aggregate missing in %#v", summary.Directs)
	}
	if bobBucket.ScopeName != bob.PeerID || bobBucket.UnreadCount != 1 {
		t.Fatalf("bob direct aggregate = %#v", bobBucket)
	}
	carolBucket, ok := notificationDirectBucketByID(summary.Directs, carolDM)
	if !ok {
		t.Fatalf("carol direct aggregate missing in %#v", summary.Directs)
	}
	if carolBucket.ScopeName != carol.PeerID || carolBucket.UnreadCount != 1 {
		t.Fatalf("carol direct aggregate = %#v", carolBucket)
	}
}

func TestNotificationSummaryIncludesPerServerAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	serverA, err := host.CreateServer("alpha", "first aggregate")
	if err != nil {
		t.Fatalf("CreateServer(serverA) error = %v", err)
	}
	serverB, err := host.CreateServer("beta", "second aggregate")
	if err != nil {
		t.Fatalf("CreateServer(serverB) error = %v", err)
	}
	channelA := firstChannelID(serverA)
	channelB := firstChannelID(serverB)
	if channelA == "" || channelB == "" {
		t.Fatal("expected default channels")
	}

	applyRemoteMessage(t, host, guest, "channel", channelA, serverA.ID, "one @alice")
	applyRemoteMessage(t, host, guest, "channel", channelA, serverA.ID, "two @alice")
	applyRemoteMessage(t, host, guest, "channel", channelB, serverB.ID, "three @alice")

	waitFor(t, 2*time.Second, func() bool {
		return host.NotificationSummary().TotalUnread == 3
	})

	summary := host.NotificationSummary()
	if len(summary.Servers) != 2 {
		t.Fatalf("server aggregate count = %d want 2; summary=%#v", len(summary.Servers), summary)
	}
	alpha, ok := notificationServerBucketByID(summary.Servers, serverA.ID)
	if !ok {
		t.Fatalf("alpha server aggregate missing in %#v", summary.Servers)
	}
	if alpha.ServerName != "alpha" || alpha.UnreadCount != 2 {
		t.Fatalf("alpha aggregate = %#v", alpha)
	}
	if alpha.LatestScopeType != "channel" || alpha.LatestScopeID != channelA || alpha.LatestScopeName != "general" {
		t.Fatalf("alpha latest scope context = %#v", alpha)
	}
	beta, ok := notificationServerBucketByID(summary.Servers, serverB.ID)
	if !ok {
		t.Fatalf("beta server aggregate missing in %#v", summary.Servers)
	}
	if beta.ServerName != "beta" || beta.UnreadCount != 1 {
		t.Fatalf("beta aggregate = %#v", beta)
	}
	if beta.LatestScopeType != "channel" || beta.LatestScopeID != channelB || beta.LatestScopeName != "general" {
		t.Fatalf("beta latest scope context = %#v", beta)
	}
}

func TestControlAPINotificationSummaryIncludesServerAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	serverA, err := host.CreateServer("alpha-api", "first api aggregate")
	if err != nil {
		t.Fatalf("CreateServer(serverA) error = %v", err)
	}
	serverB, err := host.CreateServer("beta-api", "second api aggregate")
	if err != nil {
		t.Fatalf("CreateServer(serverB) error = %v", err)
	}
	channelA := firstChannelID(serverA)
	channelB := firstChannelID(serverB)
	if channelA == "" || channelB == "" {
		t.Fatal("expected default channels")
	}

	applyRemoteMessage(t, host, guest, "channel", channelA, serverA.ID, "one @alice")
	applyRemoteMessage(t, host, guest, "channel", channelB, serverB.ID, "two @alice")

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &resp) == nil && resp.TotalUnread == 2 && len(resp.Servers) == 2
	})

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	alpha, ok := notificationServerBucketByID(summary.Servers, serverA.ID)
	if !ok {
		t.Fatalf("alpha server aggregate missing in %#v", summary.Servers)
	}
	if alpha.ServerName != "alpha-api" || alpha.UnreadCount != 1 {
		t.Fatalf("alpha aggregate = %#v", alpha)
	}
	beta, ok := notificationServerBucketByID(summary.Servers, serverB.ID)
	if !ok {
		t.Fatalf("beta server aggregate missing in %#v", summary.Servers)
	}
	if beta.ServerName != "beta-api" || beta.UnreadCount != 1 {
		t.Fatalf("beta aggregate = %#v", beta)
	}
	if beta.LatestScopeType != "channel" || beta.LatestScopeID != channelB || beta.LatestScopeName != "general" {
		t.Fatalf("beta latest scope context = %#v", beta)
	}
}

func TestCreateDMEmitsEventsForCreatorAndRecipient(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	hostEvents, cancelHost := host.Subscribe()
	defer cancelHost()
	guestEvents, cancelGuest := guest.Subscribe()
	defer cancelGuest()

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}

	waitDMCreated := func(ch <-chan Event, wantID string) Event {
		deadline := time.After(2 * time.Second)
		for {
			select {
			case ev := <-ch:
				if ev.Type != "dm.created" {
					continue
				}
				if got, _ := ev.Payload["dm_id"].(string); got == wantID {
					return ev
				}
			case <-deadline:
				t.Fatalf("expected dm.created event for %q", wantID)
			}
		}
	}

	hostEvent := waitDMCreated(hostEvents, dm.ID)
	guestEvent := waitDMCreated(guestEvents, dm.ID)

	checkParticipants := func(ev Event) {
		parts, _ := ev.Payload["participant_ids"].([]string)
		if len(parts) != 2 {
			t.Fatalf("dm.created participants = %#v", parts)
		}
		if !contains(parts, host.PeerID()) || !contains(parts, guest.PeerID()) {
			t.Fatalf("dm.created participants = %#v", parts)
		}
	}
	checkParticipants(hostEvent)
	checkParticipants(guestEvent)
}

func TestCreateDMPropagatesToPeer(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		for _, existing := range guest.Snapshot().DMs {
			if existing.ID == dm.ID {
				return true
			}
		}
		return false
	})

	msg, err := guest.SendDMMessage(dm.ID, "hello alice")
	if err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		for _, existing := range host.Snapshot().Messages {
			if existing.ID == msg.ID && existing.ScopeType == "dm" && existing.ScopeID == dm.ID {
				return true
			}
		}
		return false
	})

	if len(guest.Snapshot().DMs) != 1 {
		t.Fatalf("guest DM snapshot = %#v", guest.Snapshot().DMs)
	}
	participants := guest.Snapshot().DMs[0].Participants
	if !contains(participants, hostIdentity.PeerID) || !contains(participants, guestIdentity.PeerID) {
		t.Fatalf("guest propagated dm participants = %#v", participants)
	}
}

func TestIncomingDMMessageMaterializesDMForRecipient(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}

	dm, err := guest.CreateDM(hostIdentity.PeerID)
	if err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		for _, candidate := range host.Snapshot().DMs {
			if candidate.ID == dm.ID {
				return true
			}
		}
		return false
	})
	host.mu.Lock()
	delete(host.state.DMs, dm.ID)
	host.mu.Unlock()
	if len(host.Snapshot().DMs) != 0 {
		t.Fatalf("host should not have dm before first incoming message: %#v", host.Snapshot().DMs)
	}

	msg, err := guest.SendDMMessage(dm.ID, "hello alice")
	if err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		snap := host.Snapshot()
		hasDM := false
		for _, candidate := range snap.DMs {
			if candidate.ID == dm.ID {
				hasDM = true
				break
			}
		}
		if !hasDM {
			return false
		}
		for _, stored := range snap.Messages {
			if stored.ID == msg.ID && stored.ScopeID == dm.ID && stored.ScopeType == "dm" {
				return true
			}
		}
		return false
	})

	var hostDM DMRecord
	foundDM := false
	for _, candidate := range host.Snapshot().DMs {
		if candidate.ID == dm.ID {
			hostDM = candidate
			foundDM = true
			break
		}
	}
	if !foundDM {
		t.Fatalf("host missing materialized dm %q", dm.ID)
	}
	if !contains(hostDM.Participants, hostIdentity.PeerID) || !contains(hostDM.Participants, guestIdentity.PeerID) {
		t.Fatalf("host materialized dm participants = %#v", hostDM.Participants)
	}

	reply, err := host.SendDMMessage(dm.ID, "hello bob")
	if err != nil {
		t.Fatalf("host.SendDMMessage() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		for _, stored := range guest.Snapshot().Messages {
			if stored.ID == reply.ID && stored.ScopeID == dm.ID && stored.SenderPeerID == host.PeerID() {
				return true
			}
		}
		return false
	})
}

func TestSendDMMessageUsesLivePeerRegistry(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	hostIdentity, err := host.CreateIdentity("alice", "host")
	if err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	guestIdentity, err := guest.CreateIdentity("bob", "guest")
	if err != nil {
		t.Fatalf("guest.CreateIdentity() error = %v", err)
	}
	dm, err := host.CreateDM(guestIdentity.PeerID)
	if err != nil {
		t.Fatalf("host.CreateDM() error = %v", err)
	}
	if _, err := guest.CreateDM(hostIdentity.PeerID); err != nil {
		t.Fatalf("guest.CreateDM() error = %v", err)
	}
	msg, err := guest.SendDMMessage(dm.ID, "hello alice")
	if err != nil {
		t.Fatalf("guest.SendDMMessage() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		for _, existing := range host.Snapshot().Messages {
			if existing.ID == msg.ID && existing.ScopeType == "dm" && existing.ScopeID == dm.ID {
				return true
			}
		}
		return false
	})
}

func TestNotificationSummaryDirectAggregatesIncludeParticipantIDs(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guestIdentity := newTestIdentity(t, "bob")
	dmID := dmScopeID(host.PeerID(), guestIdentity.PeerID)
	applyRemoteMessage(t, host, guestIdentity, "dm", dmID, "", "hello @alice")
	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 1 })
	summary := host.NotificationSummary()
	if len(summary.Directs) != 1 {
		t.Fatalf("direct aggregate count = %d want 1; summary=%#v", len(summary.Directs), summary)
	}
	if !contains(summary.Directs[0].ParticipantIDs, guestIdentity.PeerID) {
		t.Fatalf("direct aggregate participant ids = %#v want %q", summary.Directs[0].ParticipantIDs, guestIdentity.PeerID)
	}
	if summary.Directs[0].ScopeName == "" {
		t.Fatal("expected direct aggregate scope label")
	}
}

func TestControlAPIDirectNotificationAggregatesIncludeParticipantIDs(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guestIdentity := newTestIdentity(t, "bob")
	dmID := dmScopeID(host.PeerID(), guestIdentity.PeerID)
	applyRemoteMessage(t, host, guestIdentity, "dm", dmID, "", "hello @alice")
	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		var summary NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary) == nil && len(summary.Directs) == 1
	})
	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if !contains(summary.Directs[0].ParticipantIDs, guestIdentity.PeerID) {
		t.Fatalf("direct aggregate participant ids = %#v want %q", summary.Directs[0].ParticipantIDs, guestIdentity.PeerID)
	}
}

func TestShouldAdvanceNotificationSummaryMessagePrefersLargerIDOnTie(t *testing.T) {
	now := time.Now().UTC()
	if !shouldAdvanceNotificationSummaryMessage(now, "msg-1", now, "msg-2") {
		t.Fatal("expected larger candidate id to win when timestamps tie")
	}
	if shouldAdvanceNotificationSummaryMessage(now, "msg-2", now, "msg-1") {
		t.Fatal("did not expect smaller candidate id to win when timestamps tie")
	}
}

func TestNotificationSummaryLatestMessageIDUsesLargerIDOnTiedTimestamp(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("tiebreak", "latest id tiebreak")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	base := time.Now().UTC()
	msgA := applyRemoteMessageAt(t, host, guest, "channel", channelID, server.ID, "hello @alice one", base)
	msgB := applyRemoteMessageAt(t, host, guest, "channel", channelID, server.ID, "hello @alice two", base.Add(time.Millisecond))

	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 2 })

	tiedAt := time.Now().UTC()
	host.mu.Lock()
	first := host.state.Messages[msgA.ID]
	first.CreatedAt = tiedAt
	first.UpdatedAt = time.Time{}
	host.state.Messages[msgA.ID] = first
	second := host.state.Messages[msgB.ID]
	second.CreatedAt = tiedAt
	second.UpdatedAt = time.Time{}
	host.state.Messages[msgB.ID] = second
	host.mu.Unlock()

	summary := host.NotificationSummary()
	bucket, ok := notificationBucketByScope(summary.Buckets, channelID)
	if !ok {
		t.Fatalf("expected bucket for scope %q in %#v", channelID, summary.Buckets)
	}
	expected := msgA.ID
	if msgB.ID > expected {
		expected = msgB.ID
	}
	if bucket.LatestMessageID != expected {
		t.Fatalf("latest message id on tie = %q want %q (msgA=%q msgB=%q)", bucket.LatestMessageID, expected, msgA.ID, msgB.ID)
	}
}

func TestNotificationSummaryBucketsExposeLatestMessageIDs(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("summary-latest", "latest message ids")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	generalID := firstChannelID(server)
	if generalID == "" {
		t.Fatal("expected default channel")
	}

	base := time.Now().UTC()
	msg1 := applyRemoteMessageAt(t, host, guest, "channel", generalID, server.ID, "one @alice", base)
	msg2 := applyRemoteMessageAt(t, host, guest, "channel", generalID, server.ID, "two @alice", base.Add(time.Millisecond))
	msg3 := applyRemoteMessageAt(t, host, guest, "channel", alerts.ID, server.ID, "three @alice", base.Add(2*time.Millisecond))

	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 3 })
	summary := host.NotificationSummary()

	generalBucket, ok := notificationBucketByScope(summary.Buckets, generalID)
	if !ok {
		t.Fatalf("expected general bucket in %#v", summary.Buckets)
	}
	generalLatestID, ok := latestMessageIDForIDs(host, msg1.ID, msg2.ID)
	if !ok {
		t.Fatal("expected general messages in host state")
	}
	if generalBucket.LatestMessageID != generalLatestID {
		t.Fatalf("general latest message id = %q want %q", generalBucket.LatestMessageID, generalLatestID)
	}
	if generalBucket.LatestSenderPeerID != guest.PeerID {
		t.Fatalf("general latest sender peer id = %q want %q", generalBucket.LatestSenderPeerID, guest.PeerID)
	}

	alertsBucket, ok := notificationBucketByScope(summary.Buckets, alerts.ID)
	if !ok {
		t.Fatalf("expected alerts bucket in %#v", summary.Buckets)
	}
	alertsLatestID, ok := latestMessageIDForIDs(host, msg3.ID)
	if !ok {
		t.Fatal("expected alerts message in host state")
	}
	if alertsBucket.LatestMessageID != alertsLatestID {
		t.Fatalf("alerts latest message id = %q want %q", alertsBucket.LatestMessageID, alertsLatestID)
	}
	if alertsBucket.LatestSenderPeerID != guest.PeerID {
		t.Fatalf("alerts latest sender peer id = %q want %q", alertsBucket.LatestSenderPeerID, guest.PeerID)
	}

	serverBucket, ok := notificationServerBucketByID(summary.Servers, server.ID)
	if !ok {
		t.Fatalf("expected server aggregate in %#v", summary.Servers)
	}
	serverLatestID, ok := latestMessageIDForIDs(host, msg1.ID, msg2.ID, msg3.ID)
	if !ok {
		t.Fatal("expected server messages in host state")
	}
	if serverBucket.LatestMessageID != serverLatestID {
		t.Fatalf("server latest message id = %q want %q", serverBucket.LatestMessageID, serverLatestID)
	}
	if serverBucket.LatestSenderPeerID != guest.PeerID {
		t.Fatalf("server latest sender peer id = %q want %q", serverBucket.LatestSenderPeerID, guest.PeerID)
	}

	_ = msg1
}

func TestControlAPINotificationSummaryDirectAggregatesExposeLatestMessageID(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guestIdentity := newTestIdentity(t, "bob")
	dmID := dmScopeID(host.PeerID(), guestIdentity.PeerID)
	base := time.Now().UTC()
	msg1 := applyRemoteMessageAt(t, host, guestIdentity, "dm", dmID, "", "hello @alice one", base)
	msg2 := applyRemoteMessageAt(t, host, guestIdentity, "dm", dmID, "", "hello @alice two", base.Add(time.Millisecond))

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var summary NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary) == nil && len(summary.Directs) == 1
	})
	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	directBucket, ok := notificationDirectBucketByID(summary.Directs, dmID)
	if !ok {
		t.Fatalf("expected direct aggregate in %#v", summary.Directs)
	}
	directLatestID, ok := latestMessageIDForIDs(host, msg1.ID, msg2.ID)
	if !ok {
		t.Fatal("expected direct messages in host state")
	}
	if directBucket.LatestMessageID != directLatestID {
		t.Fatalf("direct latest message id = %q want %q", directBucket.LatestMessageID, directLatestID)
	}
}

func TestNotificationSummaryIncludesKindAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("kind-summary", "kind aggregates")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	base := time.Now().UTC()
	msg1 := applyRemoteMessageAt(t, host, guest, "channel", channelID, server.ID, "one @alice", base)
	msg2 := applyRemoteMessageAt(t, host, guest, "channel", channelID, server.ID, "two @alice", base.Add(time.Millisecond))

	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 2 })
	summary := host.NotificationSummary()
	if len(summary.Kinds) != 1 {
		t.Fatalf("kind aggregate count = %d want 1; summary=%#v", len(summary.Kinds), summary)
	}
	if summary.Kinds[0].Kind != "mention" || summary.Kinds[0].UnreadCount != 2 {
		t.Fatalf("kind aggregate = %#v", summary.Kinds[0])
	}
	kindLatestID, ok := latestMessageIDForIDs(host, msg1.ID, msg2.ID)
	if !ok {
		t.Fatal("expected kind messages in host state")
	}
	if summary.Kinds[0].LatestMessageID != kindLatestID {
		t.Fatalf("kind latest message id = %q want %q", summary.Kinds[0].LatestMessageID, kindLatestID)
	}
	if summary.Kinds[0].LatestServerID != server.ID || summary.Kinds[0].LatestServerName != "kind-summary" {
		t.Fatalf("kind latest server context = %#v", summary.Kinds[0])
	}
	if summary.Kinds[0].LatestScopeType != "channel" || summary.Kinds[0].LatestScopeID != channelID || summary.Kinds[0].LatestScopeName != "general" {
		t.Fatalf("kind latest scope context = %#v", summary.Kinds[0])
	}
	if len(summary.Kinds[0].LatestParticipantIDs) != 0 {
		t.Fatalf("unexpected latest participant ids on channel kind bucket = %#v", summary.Kinds[0].LatestParticipantIDs)
	}
	_ = msg1
}

func TestControlAPINotificationSummaryIncludesKindAggregates(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("kind-summary-api", "kind aggregates api")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	msg := applyRemoteMessage(t, host, guest, "channel", channelID, server.ID, "api @alice")

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var summary NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary) == nil && len(summary.Kinds) == 1
	})
	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if len(summary.Kinds) != 1 {
		t.Fatalf("kind aggregate count = %d want 1; summary=%#v", len(summary.Kinds), summary)
	}
	if summary.Kinds[0].Kind != "mention" || summary.Kinds[0].UnreadCount != 1 || summary.Kinds[0].LatestMessageID != msg.ID {
		t.Fatalf("kind aggregate = %#v", summary.Kinds[0])
	}
}

func TestControlAPINotificationSummaryKindAggregateIncludesDMContext(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guestIdentity := newTestIdentity(t, "bob")
	dmID := dmScopeID(host.PeerID(), guestIdentity.PeerID)
	msg := applyRemoteMessage(t, host, guestIdentity, "dm", dmID, "", "hello @alice")

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var summary NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary) == nil && len(summary.Kinds) == 1
	})

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if len(summary.Kinds) != 1 {
		t.Fatalf("kind aggregate count = %d want 1; summary=%#v", len(summary.Kinds), summary)
	}
	kind := summary.Kinds[0]
	if kind.Kind != "mention" || kind.UnreadCount != 1 || kind.LatestMessageID != msg.ID {
		t.Fatalf("kind aggregate = %#v", kind)
	}
	if kind.LatestServerID != "" || kind.LatestServerName != "" {
		t.Fatalf("unexpected server context on dm kind aggregate = %#v", kind)
	}
	if kind.LatestScopeType != "dm" || kind.LatestScopeID != dmID || kind.LatestScopeName != guestIdentity.PeerID {
		t.Fatalf("dm kind latest scope context = %#v", kind)
	}
	if len(kind.LatestParticipantIDs) != 1 || kind.LatestParticipantIDs[0] != guestIdentity.PeerID {
		t.Fatalf("dm kind latest participant ids = %#v want [%q]", kind.LatestParticipantIDs, guestIdentity.PeerID)
	}
}

func TestNotificationSummaryServerAggregatesExposeLatestScopeContext(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("server-scope", "server aggregate scope context")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	generalID := firstChannelID(server)
	if generalID == "" {
		t.Fatal("expected default channel")
	}

	base := time.Now().UTC()
	msg1 := applyRemoteMessageAt(t, host, guest, "channel", generalID, server.ID, "general @alice", base)
	msg := applyRemoteMessageAt(t, host, guest, "channel", alerts.ID, server.ID, "alerts @alice", base.Add(time.Millisecond))

	waitFor(t, 2*time.Second, func() bool { return host.NotificationSummary().TotalUnread == 2 })
	summary := host.NotificationSummary()
	bucket, ok := notificationServerBucketByID(summary.Servers, server.ID)
	if !ok {
		t.Fatalf("expected server aggregate in %#v", summary.Servers)
	}
	serverLatest, ok := latestMessageRecordForIDs(t, host, msg1.ID, msg.ID)
	if !ok {
		t.Fatal("expected server messages in host state")
	}
	if bucket.LatestMessageID != serverLatest.ID {
		t.Fatalf("latest message id = %q want %q", bucket.LatestMessageID, serverLatest.ID)
	}
	expectedScopeName := "general"
	if serverLatest.ID == msg.ID {
		expectedScopeName = "alerts"
	}
	if bucket.LatestScopeType != "channel" || bucket.LatestScopeID != serverLatest.ScopeID || bucket.LatestScopeName != expectedScopeName {
		t.Fatalf("server latest scope context = %#v", bucket)
	}
}

func TestControlAPINotificationSummaryDirectAggregatesExposeLatestSenderPeerID(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guestIdentity := newTestIdentity(t, "bob")
	dmID := dmScopeID(host.PeerID(), guestIdentity.PeerID)
	base := time.Now().UTC()
	msg1 := applyRemoteMessageAt(t, host, guestIdentity, "dm", dmID, "", "hello @alice one", base)
	msg2 := applyRemoteMessageAt(t, host, guestIdentity, "dm", dmID, "", "hello @alice two", base.Add(time.Millisecond))

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var summary NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary) == nil && len(summary.Directs) == 1
	})

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	direct, ok := notificationDirectBucketByID(summary.Directs, dmID)
	if !ok {
		t.Fatalf("expected direct aggregate in %#v", summary.Directs)
	}
	directLatestID, ok := latestMessageIDForIDs(host, msg1.ID, msg2.ID)
	if !ok {
		t.Fatal("expected direct messages in host state")
	}
	if direct.LatestMessageID != directLatestID {
		t.Fatalf("direct latest message id = %q want %q", direct.LatestMessageID, directLatestID)
	}
	if direct.LatestSenderPeerID != guestIdentity.PeerID {
		t.Fatalf("direct latest sender peer id = %q want %q", direct.LatestSenderPeerID, guestIdentity.PeerID)
	}
}

func TestNotificationSummaryIncludesServerAndScopeLabels(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("ops", "labels")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	channel, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	applyRemoteMessage(t, host, guest, "channel", channel.ID, server.ID, "ping @alice")

	waitFor(t, 2*time.Second, func() bool {
		summary := host.NotificationSummary()
		bucket, ok := notificationBucketByScope(summary.Buckets, channel.ID)
		return ok && bucket.UnreadCount == 1
	})

	summary := host.NotificationSummary()
	bucket, ok := notificationBucketByScope(summary.Buckets, channel.ID)
	if !ok {
		t.Fatalf("expected summary bucket for channel %s, buckets=%#v", channel.ID, summary.Buckets)
	}
	if bucket.ServerName != "ops" {
		t.Fatalf("bucket.ServerName = %q want %q", bucket.ServerName, "ops")
	}
	if bucket.ScopeName != "alerts" {
		t.Fatalf("bucket.ScopeName = %q want %q", bucket.ScopeName, "alerts")
	}
}

func TestNotificationSummaryGroupsUnreadByScope(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("notifications-summary", "summary surface")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	generalID := firstChannelID(server)
	if generalID == "" {
		t.Fatal("expected default channel")
	}
	alerts, err := host.CreateChannel(server.ID, "alerts", false)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}

	applyRemoteMessage(t, host, guest, "channel", generalID, server.ID, "one @alice")
	applyRemoteMessage(t, host, guest, "channel", alerts.ID, server.ID, "two @alice")
	applyRemoteMessage(t, host, guest, "channel", generalID, server.ID, "three @alice")

	waitFor(t, 2*time.Second, func() bool {
		return host.NotificationSummary().TotalUnread == 3
	})

	summary := host.NotificationSummary()
	if summary.TotalUnread != 3 {
		t.Fatalf("total unread = %d want 3", summary.TotalUnread)
	}
	if len(summary.Buckets) != 2 {
		t.Fatalf("bucket count = %d want 2", len(summary.Buckets))
	}
	generalBucket, ok := notificationBucketByScope(summary.Buckets, generalID)
	if !ok {
		t.Fatalf("general bucket not found in %#v", summary.Buckets)
	}
	if generalBucket.UnreadCount != 2 {
		t.Fatalf("general unread count = %d want 2", generalBucket.UnreadCount)
	}
	alertsBucket, ok := notificationBucketByScope(summary.Buckets, alerts.ID)
	if !ok {
		t.Fatalf("alerts bucket not found in %#v", summary.Buckets)
	}
	if alertsBucket.UnreadCount != 1 {
		t.Fatalf("alerts unread count = %d want 1", alertsBucket.UnreadCount)
	}
}

func TestControlAPINotificationSummaryReflectsMarkRead(t *testing.T) {
	host, stopHost := startControlOnlyService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", "host"); err != nil {
		t.Fatalf("host.CreateIdentity() error = %v", err)
	}
	guest := newTestIdentity(t, "guest")

	server, err := host.CreateServer("notifications-summary-api", "summary api surface")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}
	applyRemoteMessage(t, host, guest, "channel", channelID, server.ID, "ping @alice")

	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		var resp NotificationSummaryResponse
		return CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &resp) == nil && resp.TotalUnread == 1
	})

	var summary NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &summary); err != nil {
		t.Fatalf("GET /v1/notifications/summary error = %v", err)
	}
	if summary.TotalUnread != 1 || len(summary.Buckets) != 1 {
		t.Fatalf("summary response = %#v", summary)
	}
	if summary.Buckets[0].ServerName != "notifications-summary-api" {
		t.Fatalf("bucket.ServerName = %q", summary.Buckets[0].ServerName)
	}
	if summary.Buckets[0].ScopeName == "" {
		t.Fatalf("bucket.ScopeName should not be empty: %#v", summary.Buckets[0])
	}

	var markResp MarkNotificationsReadResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/notifications/read", MarkNotificationsReadRequest{Through: summary.Buckets[0].LatestAt}, &markResp); err != nil {
		t.Fatalf("POST /v1/notifications/read error = %v", err)
	}

	var after NotificationSummaryResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodGet, "/v1/notifications/summary", nil, &after); err != nil {
		t.Fatalf("GET /v1/notifications/summary after read error = %v", err)
	}
	if after.TotalUnread != 0 || len(after.Buckets) != 0 {
		t.Fatalf("summary after read = %#v", after)
	}
}

func notificationBucketByScope(buckets []NotificationSummaryBucket, scopeID string) (NotificationSummaryBucket, bool) {
	for _, bucket := range buckets {
		if bucket.ScopeID == scopeID {
			return bucket, true
		}
	}
	return NotificationSummaryBucket{}, false
}

func notificationDirectBucketByID(buckets []NotificationSummaryDirectBucket, scopeID string) (NotificationSummaryDirectBucket, bool) {
	for _, bucket := range buckets {
		if bucket.ScopeID == scopeID {
			return bucket, true
		}
	}
	return NotificationSummaryDirectBucket{}, false
}

func notificationServerBucketByID(buckets []NotificationSummaryServerBucket, serverID string) (NotificationSummaryServerBucket, bool) {
	for _, bucket := range buckets {
		if bucket.ServerID == serverID {
			return bucket, true
		}
	}
	return NotificationSummaryServerBucket{}, false
}

func TestSearchMentionsFindsPeerIDAndDisplayNameMentions(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("CreateIdentity() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	server, err := host.CreateServer("mentions", "mention surface")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}
	byName, err := guest.SendChannelMessage(general, "hello @alice")
	if err != nil {
		t.Fatalf("SendChannelMessage(byName) error = %v", err)
	}
	byPeer, err := guest.SendChannelMessage(general, "ping @"+host.PeerID())
	if err != nil {
		t.Fatalf("SendChannelMessage(byPeer) error = %v", err)
	}
	if _, err := guest.SendChannelMessage(general, "ignore @nobody"); err != nil {
		t.Fatalf("SendChannelMessage(ignore) error = %v", err)
	}
	mentions, err := host.SearchMentions(SearchMentionsRequest{ServerID: server.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchMentions() error = %v", err)
	}
	if len(mentions) != 2 {
		t.Fatalf("mention count = %d want 2", len(mentions))
	}
	if mentions[0].Message.ID != byPeer.ID || mentions[1].Message.ID != byName.ID {
		t.Fatalf("mention ids = %v want [%s %s]", []string{mentions[0].Message.ID, mentions[1].Message.ID}, byPeer.ID, byName.ID)
	}
	if !contains(mentions[0].Tokens, "@"+host.PeerID()) {
		t.Fatalf("peer-id mention tokens = %#v", mentions[0].Tokens)
	}
	if !contains(mentions[1].Tokens, "@alice") {
		t.Fatalf("display-name mention tokens = %#v", mentions[1].Tokens)
	}
	if mentions[0].ServerName != "mentions" || mentions[0].ScopeName != "general" {
		t.Fatalf("mention labels = server=%q scope=%q", mentions[0].ServerName, mentions[0].ScopeName)
	}
	if len(mentions[0].ParticipantIDs) != 0 {
		t.Fatalf("unexpected participant ids on channel mention = %#v", mentions[0].ParticipantIDs)
	}
	for _, mention := range mentions {
		if mention.Message.SenderPeerID != guest.PeerID() {
			t.Fatalf("unexpected sender %q", mention.Message.SenderPeerID)
		}
	}
}

func TestSearchMentionsRejectsInvalidScopeCombinations(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()

	if _, err := service.SearchMentions(SearchMentionsRequest{ScopeType: "dm", ServerID: "server-1"}); err == nil || !strings.Contains(err.Error(), "server id is not valid for dm scope") {
		t.Fatalf("SearchMentions(dm with server) error = %v", err)
	}
	if _, err := service.SearchMentions(SearchMentionsRequest{ScopeType: "channel"}); err == nil || !strings.Contains(err.Error(), "scope id is required when scope type is set") {
		t.Fatalf("SearchMentions(scope without id) error = %v", err)
	}
}

func TestSearchMentionsDMResultsIncludeParticipantIDs(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	selfIdentity, err := service.CreateIdentity("alice", "")
	if err != nil {
		t.Fatalf("service.CreateIdentity() error = %v", err)
	}
	peer, stopPeer := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopPeer()
	peerIdentity, err := peer.CreateIdentity("bob", "")
	if err != nil {
		t.Fatalf("peer.CreateIdentity() error = %v", err)
	}

	dm, err := service.CreateDM(peerIdentity.PeerID)
	if err != nil {
		t.Fatalf("service.CreateDM() error = %v", err)
	}
	if _, err := peer.CreateDM(selfIdentity.PeerID); err != nil {
		t.Fatalf("peer.CreateDM() error = %v", err)
	}
	if _, err := peer.SendDMMessage(dm.ID, "hello @alice"); err != nil {
		t.Fatalf("peer.SendDMMessage() error = %v", err)
	}

	mentions, err := service.SearchMentions(SearchMentionsRequest{ScopeType: "dm", ScopeID: dm.ID, Limit: 10})
	if err != nil {
		t.Fatalf("SearchMentions(dm) error = %v", err)
	}
	if len(mentions) != 1 {
		t.Fatalf("dm mention count = %d want 1", len(mentions))
	}
	mention := mentions[0]
	if mention.ScopeName != peerIdentity.PeerID {
		t.Fatalf("dm mention scope name = %q want %q", mention.ScopeName, peerIdentity.PeerID)
	}
	if len(mention.ParticipantIDs) != 1 || mention.ParticipantIDs[0] != peerIdentity.PeerID {
		t.Fatalf("dm mention participant ids = %#v want [%q]", mention.ParticipantIDs, peerIdentity.PeerID)
	}
	if mention.ServerName != "" {
		t.Fatalf("unexpected server name on dm mention = %q", mention.ServerName)
	}
}

func TestControlAPISearchMentionsExcludesSelfMessages(t *testing.T) {
	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	if _, err := host.CreateIdentity("alice", ""); err != nil {
		t.Fatalf("CreateIdentity() error = %v", err)
	}
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()
	server, err := host.CreateServer("mentions-api", "mentions through control api")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}
	if _, err := host.SendChannelMessage(general, "self @alice should stay hidden"); err != nil {
		t.Fatalf("SendChannelMessage(self) error = %v", err)
	}
	guestMention, err := guest.SendChannelMessage(general, "guest sees @alice")
	if err != nil {
		t.Fatalf("SendChannelMessage(guest mention) error = %v", err)
	}
	token, err := ControlTokenFromDataDir(host.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	var resp SearchMentionsResponse
	if err := CallControlJSON(host.ControlEndpoint(), token, http.MethodPost, "/v1/mentions/search", SearchMentionsRequest{ServerID: server.ID, Limit: 10}, &resp); err != nil {
		t.Fatalf("POST /v1/mentions/search error = %v", err)
	}
	if len(resp.Mentions) != 1 {
		t.Fatalf("mention count = %d want 1", len(resp.Mentions))
	}
	if resp.Mentions[0].Message.ID != guestMention.ID {
		t.Fatalf("mention id = %s want %s", resp.Mentions[0].Message.ID, guestMention.ID)
	}
	if !contains(resp.Mentions[0].Tokens, "@alice") {
		t.Fatalf("mention tokens = %#v", resp.Mentions[0].Tokens)
	}
}

func TestLocalControlAPIBackupRestoreAndEvents(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()

	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	if token == "" {
		t.Fatal("expected control token")
	}

	client, base := NewControlClient(service.ControlEndpoint(), token)
	req, _ := http.NewRequest(http.MethodGet, base+"/v1/state", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unauthorized GET /v1/state error = %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	var identity Identity
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/identities", CreateIdentityRequest{DisplayName: "alice", Bio: "hello"}, &identity); err != nil {
		t.Fatalf("POST /v1/identities error = %v", err)
	}
	if identity.Profile.DisplayName != "alice" {
		t.Fatalf("identity profile = %+v", identity.Profile)
	}

	clientSSE, baseSSE := NewControlClient(service.ControlEndpoint(), token)
	reqSSE, _ := http.NewRequest(http.MethodGet, baseSSE+"/v1/events", nil)
	reqSSE.Header.Set("Authorization", "Bearer "+token)
	respSSE, err := clientSSE.Do(reqSSE)
	if err != nil {
		t.Fatalf("GET /v1/events error = %v", err)
	}
	defer respSSE.Body.Close()

	var server ServerRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/servers", CreateServerRequest{Name: "control-server", Description: "from control api"}, &server); err != nil {
		t.Fatalf("POST /v1/servers error = %v", err)
	}
	voiceReqPath := "/v1/servers/" + server.ID + "/channels"
	var voiceChannel ChannelRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, voiceReqPath, CreateChannelRequest{Name: "voice", Voice: true}, &voiceChannel); err != nil {
		t.Fatalf("POST %s error = %v", voiceReqPath, err)
	}
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/voice/"+voiceChannel.ID+"/join", VoiceJoinRequest{Muted: false}, nil); err != nil {
		t.Fatalf("POST voice join error = %v", err)
	}
	var dm DMRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/dms", CreateDMRequest{PeerID: service.PeerID()}, &dm); err != nil {
		t.Fatalf("POST /v1/dms error = %v", err)
	}
	var dmMessage MessageRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/dms/"+dm.ID+"/messages", SendMessageRequest{Body: "self dm"}, &dmMessage); err != nil {
		t.Fatalf("POST /v1/dms/{id}/messages error = %v", err)
	}
	if dmMessage.ScopeID != dm.ID {
		t.Fatalf("dm message scope mismatch: %+v", dmMessage)
	}

	reader := bufio.NewReader(respSSE.Body)
	line1, _ := reader.ReadString('\n')
	line2, _ := reader.ReadString('\n')
	if !strings.Contains(line1, "event: ready") || !strings.Contains(line2, "version") {
		t.Fatalf("unexpected sse prelude: %q %q", line1, line2)
	}
	_, _ = reader.ReadString('\n')
	gotEvent := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.HasPrefix(line, "event: server.created") || strings.HasPrefix(line, "event: voice.join") {
			gotEvent = true
			break
		}
	}
	if !gotEvent {
		t.Fatal("expected control event stream to emit follow-up event")
	}

	backupRaw, err := service.BackupIdentity()
	if err != nil {
		t.Fatalf("BackupIdentity() error = %v", err)
	}
	restoreService, stopRestore := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopRestore()
	if _, err := restoreService.RestoreIdentity(backupRaw); err != nil {
		t.Fatalf("RestoreIdentity() error = %v", err)
	}
	if restoreService.PeerID() != identity.PeerID {
		t.Fatalf("restored peer id = %s want %s", restoreService.PeerID(), identity.PeerID)
	}
}

func TestLocalControlAPIIdentityRestoreKeepsStateAndControlsAligned(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()

	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}

	var original Identity
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/identities", CreateIdentityRequest{DisplayName: "alice", Bio: "first"}, &original); err != nil {
		t.Fatalf("POST /v1/identities original error = %v", err)
	}
	var server ServerRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/servers", CreateServerRequest{Name: "restorable", Description: "owned locally"}, &server); err != nil {
		t.Fatalf("POST /v1/servers error = %v", err)
	}
	channelID := firstChannelID(server)
	if channelID == "" {
		t.Fatal("expected default channel")
	}

	backupRaw := fetchControlRaw(t, service.ControlEndpoint(), token, http.MethodGet, "/v1/identities/backup", nil)

	var changed Identity
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/identities", CreateIdentityRequest{DisplayName: "bob", Bio: "second"}, &changed); err != nil {
		t.Fatalf("POST /v1/identities changed error = %v", err)
	}
	if changed.PeerID == original.PeerID {
		t.Fatalf("expected changed peer id, got %s", changed.PeerID)
	}

	var restored Identity
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/identities/restore", json.RawMessage(backupRaw), &restored); err != nil {
		t.Fatalf("POST /v1/identities/restore error = %v", err)
	}
	if restored.PeerID != original.PeerID {
		t.Fatalf("restored peer id = %s want %s", restored.PeerID, original.PeerID)
	}

	var snapshot Snapshot
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodGet, "/v1/state", nil, &snapshot); err != nil {
		t.Fatalf("GET /v1/state error = %v", err)
	}
	if snapshot.Identity.PeerID != original.PeerID {
		t.Fatalf("snapshot identity peer id = %s want %s", snapshot.Identity.PeerID, original.PeerID)
	}
	owned, ok := serverByID(snapshot.Servers, server.ID)
	if !ok {
		t.Fatalf("snapshot missing server %s", server.ID)
	}
	if owned.OwnerPeerID != original.PeerID {
		t.Fatalf("server owner = %s want %s", owned.OwnerPeerID, original.PeerID)
	}
	if !contains(owned.Members, original.PeerID) || contains(owned.Members, changed.PeerID) {
		t.Fatalf("server members = %#v want restored self only", owned.Members)
	}
	if owned.Manifest.OwnerPeerID != original.PeerID {
		t.Fatalf("manifest owner = %s want %s", owned.Manifest.OwnerPeerID, original.PeerID)
	}
	invite, err := ParseDeeplink(owned.Invite)
	if err != nil {
		t.Fatalf("ParseDeeplink() error = %v", err)
	}
	if invite.OwnerPeerID != original.PeerID {
		t.Fatalf("invite owner = %s want %s", invite.OwnerPeerID, original.PeerID)
	}
	if hasPeer(snapshot, changed.PeerID) {
		t.Fatalf("snapshot should not retain stale self peer %s", changed.PeerID)
	}

	var created MessageRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPost, "/v1/channels/"+channelID+"/messages", SendMessageRequest{Body: "after restore"}, &created); err != nil {
		t.Fatalf("POST /v1/channels/{id}/messages error = %v", err)
	}
	var edited MessageRecord
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodPatch, "/v1/messages/"+created.ID, EditMessageRequest{Body: "edited after restore"}, &edited); err != nil {
		t.Fatalf("PATCH /v1/messages/{id} error = %v", err)
	}
	if edited.Body != "edited after restore" {
		t.Fatalf("edited body = %q want edited after restore", edited.Body)
	}
	if err := CallControlJSON(service.ControlEndpoint(), token, http.MethodDelete, "/v1/messages/"+created.ID, nil, nil); err != nil {
		t.Fatalf("DELETE /v1/messages/{id} error = %v", err)
	}
	deleted, ok := messageByID(service.Snapshot(), created.ID)
	if !ok || !deleted.Deleted || deleted.Body != "" {
		t.Fatalf("expected deleted message state, got %#v ok=%v", deleted, ok)
	}
}

func TestLocalControlAPIRejectsRemoteAndBadBearer(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()

	token, err := ControlTokenFromDataDir(service.cfg.DataDir)
	if err != nil {
		t.Fatalf("ControlTokenFromDataDir() error = %v", err)
	}
	handler := service.controlMux()

	remoteReq := httptest.NewRequest(http.MethodGet, "/v1/state", nil)
	remoteReq.RemoteAddr = "203.0.113.10:4242"
	remoteReq.Header.Set("Authorization", "Bearer "+token)
	remoteResp := httptest.NewRecorder()
	handler.ServeHTTP(remoteResp, remoteReq)
	if remoteResp.Code != http.StatusForbidden {
		t.Fatalf("remote status = %d want %d", remoteResp.Code, http.StatusForbidden)
	}
	var remoteErr APIError
	if err := json.NewDecoder(remoteResp.Body).Decode(&remoteErr); err != nil {
		t.Fatalf("decode remote error = %v", err)
	}
	if remoteErr.Code != "forbidden" {
		t.Fatalf("remote error code = %q want forbidden", remoteErr.Code)
	}

	badTokenReq := httptest.NewRequest(http.MethodGet, "/v1/state", nil)
	badTokenReq.RemoteAddr = "127.0.0.1:4242"
	badTokenReq.Header.Set("Authorization", "Bearer wrong-token")
	badTokenResp := httptest.NewRecorder()
	handler.ServeHTTP(badTokenResp, badTokenReq)
	if badTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("bad token status = %d want %d", badTokenResp.Code, http.StatusUnauthorized)
	}
	var authErr APIError
	if err := json.NewDecoder(badTokenResp.Body).Decode(&authErr); err != nil {
		t.Fatalf("decode auth error = %v", err)
	}
	if authErr.Code != "unauthorized" {
		t.Fatalf("auth error code = %q want unauthorized", authErr.Code)
	}
}

func TestStateMigrationAndCorruptRecovery(t *testing.T) {
	legacyDir := t.TempDir()
	identity, err := GenerateIdentity("legacy")
	if err != nil {
		t.Fatalf("GenerateIdentity() error = %v", err)
	}
	legacy := legacyStateV1{Identity: identity, KnownPeers: map[string]PeerRecord{}, Servers: map[string]ServerRecord{}, DMs: map[string]DMRecord{}, Messages: map[string]MessageRecord{}, Settings: map[string]string{}, ControlToken: "legacy-token"}
	rawLegacy, _ := json.Marshal(legacy)
	if err := os.WriteFile(filepath.Join(legacyDir, "state.json"), rawLegacy, 0o600); err != nil {
		t.Fatalf("Write legacy state error = %v", err)
	}
	service, stop := startService(t, Config{Role: RoleClient, DataDir: legacyDir, ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stop()
	if service.PeerID() != identity.PeerID {
		t.Fatalf("migrated peer id = %s want %s", service.PeerID(), identity.PeerID)
	}
	backups, err := filepath.Glob(filepath.Join(legacyDir, "state.migrated-*.json"))
	if err != nil {
		t.Fatalf("Glob migrated state backups error = %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("expected migrated legacy state backup")
	}
	if _, err := os.Stat(filepath.Join(legacyDir, storage.StoreFileName)); err != nil {
		t.Fatalf("Stat(state.db) error = %v", err)
	}

	corruptDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(corruptDir, "state.json"), []byte("{broken"), 0o600); err != nil {
		t.Fatalf("Write corrupt state error = %v", err)
	}
	corruptService, stopCorrupt := startService(t, Config{Role: RoleClient, DataDir: corruptDir, ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopCorrupt()
	if corruptService.PeerID() == "" {
		t.Fatal("expected fresh identity after corrupt recovery")
	}
	matches, err := filepath.Glob(filepath.Join(corruptDir, "state.corrupt-*.json"))
	if err != nil {
		t.Fatalf("Glob corrupt state files error = %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected quarantined corrupt state copy")
	}
}

func TestDuplicateSuppressionAndVoiceTransport(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()
	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	server, err := host.CreateServer("voice", "voice test")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	voiceChannel, err := host.CreateChannel(server.ID, "voice", true)
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}
	if err := host.JoinVoice(voiceChannel.ID, false); err != nil {
		t.Fatalf("host JoinVoice() error = %v", err)
	}
	if err := guest.JoinVoice(voiceChannel.ID, true); err != nil {
		t.Fatalf("guest JoinVoice() error = %v", err)
	}
	if err := host.SendVoiceFrame(voiceChannel.ID, []byte{1, 2, 3, 4}); err != nil {
		t.Fatalf("SendVoiceFrame() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool {
		snap := guest.Snapshot()
		for _, session := range snap.VoiceSessions {
			if session.ChannelID == voiceChannel.ID {
				participant, ok := session.Participants[host.PeerID()]
				return ok && !participant.LastFrameAt.IsZero()
			}
		}
		return false
	})

	identity := host.Snapshot().Identity
	delivery := Delivery{ID: randomID("msg"), Kind: "channel_message", ScopeID: voiceChannel.ID, ScopeType: "channel", ServerID: server.ID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: []string{guest.PeerID()}, Body: "dedupe", CreatedAt: time.Now().UTC()}
	if err := delivery.Sign(identity); err != nil {
		t.Fatalf("delivery.Sign() error = %v", err)
	}
	if err := guest.applyDelivery(delivery); err != nil {
		t.Fatalf("applyDelivery() first error = %v", err)
	}
	if err := guest.applyDelivery(delivery); err != nil {
		t.Fatalf("applyDelivery() second error = %v", err)
	}
	snap := guest.Snapshot()
	count := 0
	for _, msg := range snap.Messages {
		if msg.ID == delivery.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("duplicate suppression count = %d, want 1", count)
	}
}

type stubRuntime struct {
	listenAddr string
}

func (r *stubRuntime) Start(context.Context) error { return nil }
func (r *stubRuntime) Close() error               { return nil }
func (r *stubRuntime) ListenAddress() string {
	if strings.TrimSpace(r.listenAddr) == "" {
		return "127.0.0.1:0"
	}
	return r.listenAddr
}

func startControlOnlyService(t *testing.T, cfg Config) (*Service, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	service, err := NewService(cfg, WithPeerRuntime(&stubRuntime{listenAddr: cfg.ListenAddr}))
	if err != nil {
		cancel()
		t.Fatalf("NewService() error = %v", err)
	}
	if err := service.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Start() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool { return service.ControlEndpoint() != "" })
	return service, func() {
		cancel()
		_ = service.Close()
	}
}

func newTestIdentity(t *testing.T, displayName string) Identity {
	t.Helper()
	identity, err := GenerateIdentity(displayName)
	if err != nil {
		t.Fatalf("GenerateIdentity(%q) error = %v", displayName, err)
	}
	return identity
}

func dmScopeID(hostPeerID, remotePeerID string) string {
	return strings.Join(dedupeSorted([]string{hostPeerID, remotePeerID}), ":")
}

func applyRemoteMessage(t *testing.T, host *Service, sender Identity, scopeType, scopeID, serverID, body string) MessageRecord {
	return applyRemoteMessageAt(t, host, sender, scopeType, scopeID, serverID, body, time.Now().UTC())
}

func applyRemoteMessageAt(t *testing.T, host *Service, sender Identity, scopeType, scopeID, serverID, body string, createdAt time.Time) MessageRecord {
	t.Helper()
	delivery := Delivery{
		ID:               randomID("msg"),
		Kind:             scopeType + "_message",
		ScopeID:          scopeID,
		ScopeType:        scopeType,
		ServerID:         serverID,
		SenderPeerID:     sender.PeerID,
		SenderPublicKey:  sender.PublicKey,
		RecipientPeerIDs: []string{host.PeerID()},
		Body:             body,
		CreatedAt:        createdAt,
	}
	if err := delivery.Sign(sender); err != nil {
		t.Fatalf("delivery.Sign() error = %v", err)
	}
	if err := host.applyDelivery(delivery); err != nil {
		t.Fatalf("host.applyDelivery() error = %v", err)
	}
	return MessageRecord{ID: delivery.ID, ScopeType: scopeType, ScopeID: scopeID, ServerID: serverID, SenderPeerID: sender.PeerID, Body: body, CreatedAt: createdAt}
}

func startService(t *testing.T, cfg Config) (*Service, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	runtime, err := network.NewP2PRuntime(network.Config{Mode: network.Mode(cfg.Role), ListenAddr: cfg.ListenAddr})
	if err != nil {
		t.Fatalf("NewP2PRuntime() error = %v", err)
	}
	runtime.SetHandler(service)
	service.peerRuntime = runtime
	if err := service.Start(ctx); err != nil {
		cancel()
		t.Fatalf("Start() error = %v", err)
	}
	waitFor(t, 2*time.Second, func() bool { return service.ListenAddress() != "" && service.ControlEndpoint() != "" })
	return service, func() {
		cancel()
		_ = service.Close()
	}
}

func TestStartRequiresInjectedPeerRuntime(t *testing.T) {
	service, err := NewService(Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if err := service.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "peer runtime is required") {
		t.Fatalf("Start() error = %v, want peer runtime is required", err)
	}
}

func waitFor(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func latestMessageRecordForIDs(t *testing.T, svc *Service, ids ...string) (MessageRecord, bool) {
	t.Helper()
	latestID, ok := latestMessageIDForIDs(svc, ids...)
	if !ok {
		return MessageRecord{}, false
	}
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	msg, ok := svc.state.Messages[latestID]
	if !ok {
		return MessageRecord{}, false
	}
	return msg, true
}

func latestMessageIDForIDs(svc *Service, ids ...string) (string, bool) {
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	var latest MessageRecord
	for _, id := range ids {
		msg, ok := svc.state.Messages[id]
		if !ok {
			return "", false
		}
		if shouldAdvanceNotificationSummaryMessage(messageSortTime(latest), latest.ID, messageSortTime(msg), msg.ID) {
			latest = msg
		}
	}
	if latest.ID == "" {
		return "", false
	}
	return latest.ID, true
}

func hasPeer(snapshot Snapshot, peerID string) bool {
	_, ok := peerByID(snapshot, peerID)
	return ok
}

func peerByID(snapshot Snapshot, peerID string) (PeerRecord, bool) {
	for _, peer := range snapshot.KnownPeers {
		if peer.PeerID == peerID {
			return peer, true
		}
	}
	return PeerRecord{}, false
}

func serverByID(servers []ServerRecord, serverID string) (ServerRecord, bool) {
	for _, server := range servers {
		if server.ID == serverID {
			return server, true
		}
	}
	return ServerRecord{}, false
}

func messageByID(snapshot Snapshot, messageID string) (MessageRecord, bool) {
	for _, msg := range snapshot.Messages {
		if msg.ID == messageID {
			return msg, true
		}
	}
	return MessageRecord{}, false
}

func fetchControlRaw(t *testing.T, endpoint, token, method, path string, body []byte) []byte {
	t.Helper()
	client, base := NewControlClient(endpoint, token)
	var reader io.Reader
	if body == nil {
		reader = nil
	} else {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, base+path, reader)
	if err != nil {
		t.Fatalf("http.NewRequest(%s %s) error = %v", method, path, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do(%s %s) error = %v", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		t.Fatalf("%s %s status = %d", method, path, resp.StatusCode)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll(%s %s) error = %v", method, path, err)
	}
	return payload
}

func hasMessage(snapshot Snapshot, messageID string) bool {
	for _, msg := range snapshot.Messages {
		if msg.ID == messageID {
			return true
		}
	}
	return false
}

func firstChannelID(server ServerRecord) string {
	for _, channel := range server.Channels {
		return channel.ID
	}
	return ""
}

func reserveListenAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return addr
}

func TestJoinByDeeplinkInfersArchivistOwnerRoleFromManifest(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleArchivist, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopHost()

	server, err := host.CreateServer("archive-host", "owned by archivist")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: 40 * time.Millisecond})
	defer stopGuest()

	if _, err := guest.JoinByDeeplink(server.Invite); err != nil {
		t.Fatalf("JoinByDeeplink() error = %v", err)
	}

	waitFor(t, 2*time.Second, func() bool {
		peer, ok := peerByID(guest.Snapshot(), host.PeerID())
		return ok && peer.Role == RoleArchivist
	})
}

func TestRelayStoreRequiresRelayRoleAndBoundsQueues(t *testing.T) {
	client, stopClient := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopClient()

	recipientIdentity, err := GenerateIdentity("recipient")
	if err != nil {
		t.Fatalf("GenerateIdentity() error = %v", err)
	}

	relay, stopRelay := startService(t, Config{Role: RoleRelay, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond})
	defer stopRelay()

	identity := client.Snapshot().Identity
	for i := 0; i < relayQueueLimit+5; i++ {
		delivery := Delivery{
			ID:               randomID("msg"),
			Kind:             "channel_message",
			ScopeID:          "channel-1",
			ScopeType:        "channel",
			ServerID:         "server-1",
			SenderPeerID:     identity.PeerID,
			SenderPublicKey:  identity.PublicKey,
			RecipientPeerIDs: []string{recipientIdentity.PeerID},
			Body:             "queued",
			CreatedAt:        time.Now().UTC().Add(time.Duration(i) * time.Millisecond),
		}
		if err := delivery.Sign(identity); err != nil {
			t.Fatalf("delivery.Sign() error = %v", err)
		}
		if transportErr := relay.peerRelayStore(delivery); transportErr != nil {
			t.Fatalf("relay.peerRelayStore() error = %+v", transportErr)
		}
	}

	relay.mu.RLock()
	queued := append([]RelayQueueEntry(nil), relay.state.RelayQueues[recipientIdentity.PeerID]...)
	relay.mu.RUnlock()
	if len(queued) != relayQueueLimit {
		t.Fatalf("relay queue length = %d want %d", len(queued), relayQueueLimit)
	}
	if len(queued[0].Payload) == 0 || bytes.Contains(queued[0].Payload, []byte("queued")) {
		t.Fatalf("expected encrypted relay payload, got %q", string(queued[0].Payload))
	}
	decoded, err := decodeRelayDelivery(relay.cfg.DataDir, queued[0])
	if err != nil {
		t.Fatalf("decodeRelayDelivery() error = %v", err)
	}
	if decoded.Body != "queued" {
		t.Fatalf("decoded relay body = %q want %q", decoded.Body, "queued")
	}

	invalid := Delivery{ID: randomID("msg"), Kind: "channel_message", SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, CreatedAt: time.Now().UTC()}
	if err := invalid.Sign(identity); err != nil {
		t.Fatalf("invalid.Sign() error = %v", err)
	}
	if transportErr := relay.peerRelayStore(invalid); transportErr == nil || transportErr.Code != "invalid_request" {
		t.Fatalf("relay.peerRelayStore() invalid request = %+v, want invalid_request", transportErr)
	}

	valid := Delivery{
		ID:               randomID("msg"),
		Kind:             "channel_message",
		ScopeID:          "channel-1",
		ScopeType:        "channel",
		ServerID:         "server-1",
		SenderPeerID:     identity.PeerID,
		SenderPublicKey:  identity.PublicKey,
		RecipientPeerIDs: []string{recipientIdentity.PeerID},
		Body:             "plaintext stays opaque to relay runtime",
		CreatedAt:        time.Now().UTC(),
	}
	if err := valid.Sign(identity); err != nil {
		t.Fatalf("valid.Sign() error = %v", err)
	}
	if transportErr := client.peerRelayStore(valid); transportErr == nil || transportErr.Code != "unsupported_operation" {
		t.Fatalf("client.peerRelayStore() error = %+v, want unsupported_operation", transportErr)
	}

	drainReq, err := signedDrainRequest(recipientIdentity, RoleClient, "")
	if err != nil {
		t.Fatalf("signedDrainRequest() error = %v", err)
	}
	drained, transportErr := relay.peerRelayDrain(drainReq)
	if transportErr != nil {
		t.Fatalf("relay.peerRelayDrain() error = %+v", transportErr)
	}
	if len(drained) != relayQueueLimit {
		t.Fatalf("drained queue length = %d want %d", len(drained), relayQueueLimit)
	}
	if drained[0].Body != "queued" {
		t.Fatalf("drained relay body = %q want %q", drained[0].Body, "queued")
	}
	if drained[0].Kind != "channel_message" {
		t.Fatalf("drained kind = %q want channel_message", drained[0].Kind)
	}
	relay.mu.RLock()
	remaining := len(relay.state.RelayQueues[recipientIdentity.PeerID])
	relay.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("remaining relay queue length = %d want 0", remaining)
	}
}

func TestRelayDrainPrunesExpiredEntries(t *testing.T) {
	relay, stopRelay := startService(t, Config{Role: RoleRelay, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	defer stopRelay()

	recipientIdentity, err := GenerateIdentity("recipient-expired")
	if err != nil {
		t.Fatalf("GenerateIdentity() error = %v", err)
	}
	delivery := Delivery{ID: randomID("msg"), Kind: "channel_message", ScopeID: "channel-1", ScopeType: "channel", ServerID: "server-1", SenderPeerID: recipientIdentity.PeerID, SenderPublicKey: recipientIdentity.PublicKey, RecipientPeerIDs: []string{recipientIdentity.PeerID}, Body: "expired", CreatedAt: time.Now().UTC().Add(-relayQueueTTL - time.Minute)}
	if err := delivery.Sign(recipientIdentity); err != nil {
		t.Fatalf("delivery.Sign() error = %v", err)
	}
	entry, err := encodeRelayDelivery(relay.cfg.DataDir, delivery)
	if err != nil {
		t.Fatalf("encodeRelayDelivery() error = %v", err)
	}
	entry.EnqueuedAt = time.Now().UTC().Add(-relayQueueTTL - time.Minute)
	entry.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	relay.mu.Lock()
	relay.state.RelayQueues[recipientIdentity.PeerID] = []RelayQueueEntry{entry}
	relay.mu.Unlock()

	drainReq, err := signedDrainRequest(recipientIdentity, RoleRelay, "")
	if err != nil {
		t.Fatalf("signedDrainRequest() error = %v", err)
	}
	drained, transportErr := relay.peerRelayDrain(drainReq)
	if transportErr != nil {
		t.Fatalf("relay.peerRelayDrain() error = %+v", transportErr)
	}
	if len(drained) != 0 {
		t.Fatalf("drained expired queue length = %d want 0", len(drained))
	}
	relay.mu.RLock()
	remaining := len(relay.state.RelayQueues[recipientIdentity.PeerID])
	relay.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("remaining expired relay queue length = %d want 0", remaining)
	}
}

func TestPeerJoinSurfacesSaveFailuresAndRollsBackState(t *testing.T) {
	bootstrap, stopBootstrap := startService(t, Config{Role: RoleBootstrap, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	defer stopBootstrap()

	host, stopHost := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: time.Hour})
	defer stopHost()

	guest, stopGuest := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", BootstrapAddrs: []string{bootstrap.ListenAddress()}, DiscoveryInterval: time.Hour})
	defer stopGuest()

	server, err := host.CreateServer("rollback", "save failure")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	originalSaveStateFn := saveStateFn
	saveStateFn = func(string, persistedState) error { return errors.New("forced save failure") }
	defer func() { saveStateFn = originalSaveStateFn }()

	if _, err := guest.JoinByDeeplink(server.Invite); err == nil || !strings.Contains(err.Error(), "forced save failure") {
		t.Fatalf("JoinByDeeplink() error = %v, want forced save failure", err)
	}

	hostSnapshot := host.Snapshot()
	joinedServer, ok := serverByID(hostSnapshot.Servers, server.ID)
	if !ok {
		t.Fatal("expected server to remain present after failed join")
	}
	if len(joinedServer.Members) != 1 {
		t.Fatalf("server members after rollback = %d want 1", len(joinedServer.Members))
	}
	if hasPeer(hostSnapshot, guest.PeerID()) {
		t.Fatalf("guest peer %s should not persist after failed join", guest.PeerID())
	}
}

func TestRelayDrainRequiresSignedRequesterIdentity(t *testing.T) {
	relay, stopRelay := startService(t, Config{Role: RoleRelay, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: time.Hour})
	defer stopRelay()

	recipientIdentity, err := GenerateIdentity("recipient")
	if err != nil {
		t.Fatalf("GenerateIdentity(recipient) error = %v", err)
	}
	attackerIdentity, err := GenerateIdentity("attacker")
	if err != nil {
		t.Fatalf("GenerateIdentity(attacker) error = %v", err)
	}

	delivery := Delivery{
		ID:               randomID("msg"),
		Kind:             "channel_message",
		ScopeID:          "channel-1",
		ScopeType:        "channel",
		ServerID:         "server-1",
		SenderPeerID:     attackerIdentity.PeerID,
		SenderPublicKey:  attackerIdentity.PublicKey,
		RecipientPeerIDs: []string{recipientIdentity.PeerID},
		Body:             "queued for recipient",
		CreatedAt:        time.Now().UTC(),
	}
	if err := delivery.Sign(attackerIdentity); err != nil {
		t.Fatalf("delivery.Sign() error = %v", err)
	}
	if transportErr := relay.peerRelayStore(delivery); transportErr != nil {
		t.Fatalf("relay.peerRelayStore() error = %+v", transportErr)
	}

	forgedReq, err := signedDrainRequest(attackerIdentity, RoleClient, "")
	if err != nil {
		t.Fatalf("signedDrainRequest(attacker) error = %v", err)
	}
	forgedReq.Requester.PeerID = recipientIdentity.PeerID
	if drained, transportErr := relay.peerRelayDrain(forgedReq); transportErr == nil || !strings.Contains(transportErr.Message, "peer id mismatch") {
		t.Fatalf("forged drain transport error = %+v, want peer id mismatch", transportErr)
	} else if len(drained) != 0 {
		t.Fatalf("forged drain returned %d queue entries want 0", len(drained))
	}

	relay.mu.RLock()
	if got := len(relay.state.RelayQueues[recipientIdentity.PeerID]); got != 1 {
		relay.mu.RUnlock()
		t.Fatalf("recipient queue after forged drain = %d want 1", got)
	}
	relay.mu.RUnlock()

	recipientReq, err := signedDrainRequest(recipientIdentity, RoleClient, "")
	if err != nil {
		t.Fatalf("signedDrainRequest(recipient) error = %v", err)
	}
	drained, transportErr := relay.peerRelayDrain(recipientReq)
	if transportErr != nil {
		t.Fatalf("relay.peerRelayDrain(recipient) error = %+v", transportErr)
	}
	if len(drained) != 1 {
		t.Fatalf("recipient drain length = %d want 1", len(drained))
	}
	if drained[0].Body != "queued for recipient" {
		t.Fatalf("recipient drain body = %q want %q", drained[0].Body, "queued for recipient")
	}
	relay.mu.RLock()
	remaining := len(relay.state.RelayQueues[recipientIdentity.PeerID])
	relay.mu.RUnlock()
	if remaining != 0 {
		t.Fatalf("recipient queue after authorized drain = %d want 0", remaining)
	}
}

func TestSendChannelMessagePrunesPersistedServerHistoryToConfiguredLimit(t *testing.T) {
	service, stop := startService(t, Config{Role: RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 2})
	defer stop()

	server, err := service.CreateServer("retention", "rolling window")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}

	first, err := service.SendChannelMessage(general, "one")
	if err != nil {
		t.Fatalf("SendChannelMessage(first) error = %v", err)
	}
	second, err := service.SendChannelMessage(general, "two")
	if err != nil {
		t.Fatalf("SendChannelMessage(second) error = %v", err)
	}
	third, err := service.SendChannelMessage(general, "three")
	if err != nil {
		t.Fatalf("SendChannelMessage(third) error = %v", err)
	}

	snapshot := service.Snapshot()
	if hasMessage(snapshot, first.ID) {
		t.Fatalf("expected oldest message %s to be pruned", first.ID)
	}
	if !hasMessage(snapshot, second.ID) || !hasMessage(snapshot, third.ID) {
		t.Fatalf("expected recent messages to remain, snapshot=%+v", snapshot.Messages)
	}
	if got := countServerMessages(snapshot, server.ID); got != 2 {
		t.Fatalf("server message count = %d want 2", got)
	}
}

func TestStartPrunesExistingServerHistoryToConfiguredLimit(t *testing.T) {
	dataDir := t.TempDir()
	service, stop := startService(t, Config{Role: RoleClient, DataDir: dataDir, ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 5})

	server, err := service.CreateServer("restart-prune", "legacy oversized history")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	general := firstChannelID(server)
	if general == "" {
		t.Fatal("expected default channel")
	}

	var sent []MessageRecord
	for _, body := range []string{"one", "two", "three", "four"} {
		msg, err := service.SendChannelMessage(general, body)
		if err != nil {
			t.Fatalf("SendChannelMessage(%q) error = %v", body, err)
		}
		sent = append(sent, msg)
	}
	stop()

	restarted, stopRestarted := startService(t, Config{Role: RoleClient, DataDir: dataDir, ListenAddr: "127.0.0.1:0", DiscoveryInterval: 40 * time.Millisecond, HistoryLimit: 2})
	defer stopRestarted()

	snapshot := restarted.Snapshot()
	if got := countServerMessages(snapshot, server.ID); got != 2 {
		t.Fatalf("server message count after restart = %d want 2", got)
	}
	if hasMessage(snapshot, sent[0].ID) || hasMessage(snapshot, sent[1].ID) {
		t.Fatalf("expected oldest messages to be pruned after restart, snapshot=%+v", snapshot.Messages)
	}
	if !hasMessage(snapshot, sent[2].ID) || !hasMessage(snapshot, sent[3].ID) {
		t.Fatalf("expected newest messages to remain after restart, snapshot=%+v", snapshot.Messages)
	}
	if len(snapshot.Servers) != 1 {
		t.Fatalf("expected one server after restart, got %d", len(snapshot.Servers))
	}
	restartedServer := snapshot.Servers[0]
	if restartedServer.Manifest.HistoryRetentionMessages != 2 {
		t.Fatalf("manifest HistoryRetentionMessages = %d want 2", restartedServer.Manifest.HistoryRetentionMessages)
	}
	if restartedServer.Manifest.HistoryCoverage != HistoryCoverageLocalWindow {
		t.Fatalf("manifest HistoryCoverage = %q want %q", restartedServer.Manifest.HistoryCoverage, HistoryCoverageLocalWindow)
	}
	if restartedServer.Manifest.HistoryDurability != HistoryDurabilitySingleNode {
		t.Fatalf("manifest HistoryDurability = %q want %q", restartedServer.Manifest.HistoryDurability, HistoryDurabilitySingleNode)
	}
}

func countServerMessages(snapshot Snapshot, serverID string) int {
	count := 0
	for _, msg := range snapshot.Messages {
		if msg.ServerID == serverID {
			count++
		}
	}
	return count
}

func presenceByPeerID(records []PresenceRecord, peerID string) (PresenceRecord, bool) {
	for _, record := range records {
		if record.PeerID == peerID {
			return record, true
		}
	}
	return PresenceRecord{}, false
}

func hasChannelNamed(channels []ChannelRecord, name string) bool {
	for _, channel := range channels {
		if channel.Name == name {
			return true
		}
	}
	return false
}

func containsTelemetry(entries []string, fragment string) bool {
	for _, entry := range entries {
		if strings.Contains(entry, fragment) {
			return true
		}
	}
	return false
}
