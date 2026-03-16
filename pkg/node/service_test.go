package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	migratedRaw, err := os.ReadFile(filepath.Join(legacyDir, "state.json"))
	if err != nil {
		t.Fatalf("Read migrated state error = %v", err)
	}
	if !bytes.Contains(migratedRaw, []byte(`"schema_version": 2`)) {
		t.Fatalf("expected migrated schema version, got %s", migratedRaw)
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

func startService(t *testing.T, cfg Config) (*Service, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	service, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
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

func hasPeer(snapshot Snapshot, peerID string) bool {
	for _, peer := range snapshot.KnownPeers {
		if peer.PeerID == peerID {
			return true
		}
	}
	return false
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
