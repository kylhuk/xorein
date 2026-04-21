package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aether/code_aether/pkg/protocol"
)

const (
	notificationsReadThroughSetting      = "notifications_read_through"
	notificationsServerReadThroughPrefix = "notifications_read_through_server:"
	notificationsScopedReadThroughPrefix = "notifications_read_through_scope:"
)

var (
	livePeerRegistryMu sync.RWMutex
	livePeerRegistry   = map[string][]string{}
)

func registerLivePeer(peerID, addr string) {
	peerID = strings.TrimSpace(peerID)
	addr = strings.TrimSpace(addr)
	if peerID == "" || addr == "" {
		return
	}
	livePeerRegistryMu.Lock()
	livePeerRegistry[peerID] = dedupeSorted(append(append([]string(nil), livePeerRegistry[peerID]...), addr))
	livePeerRegistryMu.Unlock()
}

func unregisterLivePeer(peerID, addr string) {
	peerID = strings.TrimSpace(peerID)
	addr = strings.TrimSpace(addr)
	if peerID == "" || addr == "" {
		return
	}
	livePeerRegistryMu.Lock()
	defer livePeerRegistryMu.Unlock()
	addresses := append([]string(nil), livePeerRegistry[peerID]...)
	filtered := addresses[:0]
	for _, existing := range addresses {
		if existing != addr {
			filtered = append(filtered, existing)
		}
	}
	if len(filtered) == 0 {
		delete(livePeerRegistry, peerID)
		return
	}
	livePeerRegistry[peerID] = filtered
}

func livePeerAddresses(peerID string) []string {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return nil
	}
	livePeerRegistryMu.RLock()
	defer livePeerRegistryMu.RUnlock()
	return append([]string(nil), livePeerRegistry[peerID]...)
}

type Service struct {
	cfg Config

	mu    sync.RWMutex
	state persistedState

	listen          net.Listener
	controlListener net.Listener
	httpServer      *http.Server
	controlServer   *http.Server

	eventsMu    sync.RWMutex
	subscribers map[chan Event]struct{}

	closed chan struct{}
	once   sync.Once
	wg     sync.WaitGroup
}

func NewService(cfg Config) (*Service, error) {
	if !cfg.Role.Valid() {
		return nil, fmt.Errorf("invalid role %q", cfg.Role)
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		return nil, errors.New("data dir is required")
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:0"
	}
	if cfg.DiscoveryInterval <= 0 {
		cfg.DiscoveryInterval = 250 * time.Millisecond
	}
	if cfg.HistoryLimit <= 0 {
		cfg.HistoryLimit = 32
	}
	state, err := loadState(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	return &Service{
		cfg:         cfg,
		state:       state,
		subscribers: map[chan Event]struct{}{},
		closed:      make(chan struct{}),
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listen = ln
	s.httpServer = &http.Server{Handler: s.networkMux()}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = s.httpServer.Serve(ln)
	}()

	controlListener, controlEndpoint, err := createControlListener(s.cfg.ControlEndpoint, s.cfg.DataDir)
	if err != nil {
		_ = s.listen.Close()
		return err
	}
	s.controlListener = controlListener
	s.controlServer = &http.Server{Handler: s.controlMux()}
	s.mu.Lock()
	now := time.Now().UTC()
	s.state.Settings["control_endpoint"] = controlEndpoint
	s.state.Settings["listen_address"] = s.listen.Addr().String()
	s.upsertPeerLocked(PeerRecord{
		PeerID:     s.state.Identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{s.listen.Addr().String()},
		PublicKey:  s.state.Identity.PublicKey,
		Source:     "self",
		LastSeenAt: now,
	})
	refreshedManifests, refreshErr := s.refreshOwnedServerBindingsLocked(now)
	prunedHistory := s.pruneAllServerHistoryLocked()
	if refreshErr != nil {
		err = refreshErr
	} else {
		err = s.saveLocked()
	}
	token := s.state.ControlToken
	s.mu.Unlock()
	if prunedHistory {
		s.recordTelemetry("history.pruned startup=true")
	}
	if err == nil {
		err = ensureControlTokenFile(s.cfg.DataDir, token)
	}
	if err != nil {
		_ = s.listen.Close()
		_ = s.controlListener.Close()
		return err
	}
	registerLivePeer(s.PeerID(), s.ListenAddress())
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = s.controlServer.Serve(controlListener)
	}()
	for _, manifest := range refreshedManifests {
		_ = s.publishManifest(manifest)
	}

	s.recordTelemetry("runtime.started role=" + string(s.cfg.Role))
	s.emitEvent(Event{Type: "runtime.started", Time: time.Now().UTC(), Payload: map[string]any{"role": s.cfg.Role}})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.discoveryLoop()
	}()

	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()

	// Trigger initial discovery immediately.
	s.runDiscoveryPass()
	return nil
}

func (s *Service) Close() error {
	var retErr error
	s.once.Do(func() {
		peerID := s.PeerID()
		listenAddr := s.ListenAddress()
		unregisterLivePeer(peerID, listenAddr)
		close(s.closed)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if s.controlServer != nil {
			if err := s.controlServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				retErr = err
			}
		}
		if s.httpServer != nil {
			if err := s.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				retErr = err
			}
		}
		if s.controlListener != nil {
			_ = s.controlListener.Close()
		}
		if s.listen != nil {
			_ = s.listen.Close()
		}
		s.wg.Wait()
		s.eventsMu.Lock()
		for ch := range s.subscribers {
			close(ch)
			delete(s.subscribers, ch)
		}
		s.eventsMu.Unlock()
		s.mu.Lock()
		if err := s.saveLocked(); err != nil && retErr == nil {
			retErr = err
		}
		s.mu.Unlock()
	})
	return retErr
}

func (s *Service) ListenAddress() string {
	if s.listen == nil {
		return ""
	}
	return s.listen.Addr().String()
}

func (s *Service) ControlEndpoint() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Settings["control_endpoint"]
}

func (s *Service) ControlToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.ControlToken
}

func (s *Service) PeerID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Identity.PeerID
}

func (s *Service) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{
		Role:            s.cfg.Role,
		PeerID:          s.state.Identity.PeerID,
		ListenAddresses: []string{s.listenAddressLocked()},
		ControlEndpoint: s.state.Settings["control_endpoint"],
		Identity:        s.state.Identity,
		KnownPeers:      sortedPeers(s.state.KnownPeers),
		Servers:         sortedServers(s.state.Servers),
		DMs:             sortedDMs(s.state.DMs),
		Messages:        sortedMessages(s.state.Messages),
		VoiceSessions:   sortedVoice(s.state.Voice),
		Settings:        cloneStringMap(s.state.Settings),
		Telemetry:       append([]string(nil), s.state.Telemetry...),
	}
}

func (s *Service) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 32)
	s.eventsMu.Lock()
	s.subscribers[ch] = struct{}{}
	s.eventsMu.Unlock()
	cancel := func() {
		s.eventsMu.Lock()
		if _, ok := s.subscribers[ch]; ok {
			delete(s.subscribers, ch)
			close(ch)
		}
		s.eventsMu.Unlock()
	}
	return ch, cancel
}

func (s *Service) CreateIdentity(displayName, bio string) (Identity, error) {
	identity, err := GenerateIdentity(displayName)
	if err != nil {
		return Identity{}, err
	}
	identity.Profile.Bio = strings.TrimSpace(bio)
	s.mu.Lock()
	oldPeerID := s.state.Identity.PeerID
	listenAddr := s.listenAddressLocked()
	s.state.Identity = identity
	s.upsertPeerLocked(PeerRecord{
		PeerID:     identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{s.listenAddressLocked()},
		PublicKey:  identity.PublicKey,
		Source:     "self",
		LastSeenAt: time.Now().UTC(),
	})
	err = s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return Identity{}, err
	}
	unregisterLivePeer(oldPeerID, listenAddr)
	registerLivePeer(identity.PeerID, listenAddr)
	s.recordTelemetry("identity.created peer=" + identity.PeerID)
	s.emitEvent(Event{Type: "identity.created", Time: time.Now().UTC(), Payload: map[string]any{"peer_id": identity.PeerID}})
	return identity, nil
}

func (s *Service) BackupIdentity() ([]byte, error) {
	s.mu.RLock()
	identity := s.state.Identity
	s.mu.RUnlock()
	return identity.Backup()
}

func (s *Service) RestoreIdentity(raw []byte) (Identity, error) {
	identity, err := RestoreIdentity(raw)
	if err != nil {
		return Identity{}, err
	}
	s.mu.Lock()
	oldPeerID := s.state.Identity.PeerID
	listenAddr := s.listenAddressLocked()
	s.state.Identity = identity
	s.upsertPeerLocked(PeerRecord{
		PeerID:     identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{s.listenAddressLocked()},
		PublicKey:  identity.PublicKey,
		Source:     "self",
		LastSeenAt: time.Now().UTC(),
	})
	err = s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return Identity{}, err
	}
	unregisterLivePeer(oldPeerID, listenAddr)
	registerLivePeer(identity.PeerID, listenAddr)
	s.recordTelemetry("identity.restored peer=" + identity.PeerID)
	return identity, nil
}

func (s *Service) AddManualPeer(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New("manual peer address is required")
	}
	peer, err := s.fetchPeerInfo(address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	peer.Source = "manual"
	peer.LastSeenAt = time.Now().UTC()
	s.upsertPeerLocked(peer)
	err = s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return err
	}
	s.recordTelemetry("peer.manual.add peer=" + peer.PeerID + " addr=" + address)
	return nil
}

func (s *Service) RemoveManualPeer(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New("manual peer address is required")
	}
	s.mu.Lock()
	for id, peer := range s.state.KnownPeers {
		filtered := make([]string, 0, len(peer.Addresses))
		for _, addr := range peer.Addresses {
			if addr != address {
				filtered = append(filtered, addr)
			}
		}
		peer.Addresses = filtered
		if len(peer.Addresses) == 0 && peer.Source == "manual" {
			delete(s.state.KnownPeers, id)
			continue
		}
		s.state.KnownPeers[id] = peer
	}
	err := s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return err
	}
	s.recordTelemetry("peer.manual.remove addr=" + address)
	return nil
}

func (s *Service) refreshOwnedServerBindingsLocked(now time.Time) ([]Manifest, error) {
	listenAddr := s.listenAddressLocked()
	if listenAddr == "" {
		return nil, nil
	}
	identity := s.state.Identity
	refreshed := make([]Manifest, 0)
	for serverID, server := range s.state.Servers {
		if server.OwnerPeerID != identity.PeerID {
			continue
		}
		manifest := server.Manifest
		if manifest.ServerID == "" {
			manifest.ServerID = server.ID
		}
		if manifest.Name == "" {
			manifest.Name = server.Name
		}
		if manifest.Description == "" && server.Description != "" {
			manifest.Description = server.Description
		}
		if manifest.OwnerPeerID == "" {
			manifest.OwnerPeerID = identity.PeerID
		}
		if manifest.OwnerPublicKey == "" {
			manifest.OwnerPublicKey = identity.PublicKey
		}
		if manifest.IssuedAt.IsZero() {
			manifest.IssuedAt = server.CreatedAt
			if manifest.IssuedAt.IsZero() {
				manifest.IssuedAt = now
			}
		}

		desiredOwnerAddresses := dedupeSorted([]string{listenAddr})
		desiredBootstrapAddrs := dedupeSorted(append(append([]string(nil), manifest.BootstrapAddrs...), s.cfg.BootstrapAddrs...))
		if s.cfg.Role == RoleBootstrap {
			desiredBootstrapAddrs = dedupeSorted(append(desiredBootstrapAddrs, listenAddr))
		}
		desiredRelayAddrs := dedupeSorted(append(append([]string(nil), manifest.RelayAddrs...), s.cfg.RelayAddrs...))
		if s.cfg.Role == RoleRelay {
			desiredRelayAddrs = dedupeSorted(append(desiredRelayAddrs, listenAddr))
		}
		desiredCapabilities := advertisedCapabilities(s.cfg.Role)

		manifestChanged := manifest.ServerID != server.ID ||
			manifest.Name != server.Name ||
			manifest.Description != server.Description ||
			manifest.OwnerPeerID != identity.PeerID ||
			manifest.OwnerPublicKey != identity.PublicKey ||
			!sameStringSet(manifest.OwnerAddresses, desiredOwnerAddresses) ||
			!sameStringSet(manifest.BootstrapAddrs, desiredBootstrapAddrs) ||
			!sameStringSet(manifest.RelayAddrs, desiredRelayAddrs) ||
			!sameStringSet(manifest.Capabilities, desiredCapabilities) ||
			manifest.Signature == ""
		if manifestChanged {
			manifest.ServerID = server.ID
			manifest.Name = server.Name
			manifest.Description = server.Description
			manifest.OwnerPeerID = identity.PeerID
			manifest.OwnerPublicKey = identity.PublicKey
			manifest.OwnerAddresses = desiredOwnerAddresses
			manifest.BootstrapAddrs = desiredBootstrapAddrs
			manifest.RelayAddrs = desiredRelayAddrs
			manifest.Capabilities = desiredCapabilities
			manifest.UpdatedAt = now
			if err := manifest.Sign(identity); err != nil {
				return nil, err
			}
			server.Manifest = manifest
			server.UpdatedAt = now
			refreshed = append(refreshed, manifest)
		}

		inviteNeedsRefresh := false
		invite, err := ParseDeeplink(server.Invite)
		if err != nil {
			inviteNeedsRefresh = true
		} else if invite.OwnerPeerID != identity.PeerID || invite.OwnerPublicKey != identity.PublicKey || invite.ManifestHash != manifest.Hash() || invite.ExpiresAt.Before(now) || !sameStringSet(invite.ServerAddrs, desiredOwnerAddresses) || !sameStringSet(invite.BootstrapAddrs, manifest.BootstrapAddrs) || !sameStringSet(invite.RelayAddrs, manifest.RelayAddrs) {
			inviteNeedsRefresh = true
		}
		if inviteNeedsRefresh {
			invite = Invite{
				ServerID:       server.ID,
				OwnerPeerID:    identity.PeerID,
				OwnerPublicKey: identity.PublicKey,
				ServerAddrs:    desiredOwnerAddresses,
				BootstrapAddrs: append([]string(nil), manifest.BootstrapAddrs...),
				RelayAddrs:     append([]string(nil), manifest.RelayAddrs...),
				ManifestHash:   manifest.Hash(),
				ExpiresAt:      now.Add(7 * 24 * time.Hour),
			}
			if err := invite.Sign(identity); err != nil {
				return nil, err
			}
			deeplink, err := invite.Deeplink()
			if err != nil {
				return nil, err
			}
			server.Invite = deeplink
		}
		if manifestChanged || inviteNeedsRefresh {
			s.state.Servers[serverID] = server
		}
	}
	return refreshed, nil
}

func (s *Service) Presence() []PresenceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.presenceLocked(time.Now().UTC())
}

func (s *Service) presenceLocked(now time.Time) []PresenceRecord {
	peers := clonePeerMap(s.state.KnownPeers)
	if _, ok := peers[s.state.Identity.PeerID]; !ok {
		peers[s.state.Identity.PeerID] = PeerRecord{PeerID: s.state.Identity.PeerID, Role: s.cfg.Role, Addresses: []string{s.listenAddressLocked()}, PublicKey: s.state.Identity.PublicKey, Source: "self", LastSeenAt: now}
	}
	voiceByPeer := map[string][]string{}
	for channelID, session := range s.state.Voice {
		for peerID := range session.Participants {
			voiceByPeer[peerID] = append(voiceByPeer[peerID], channelID)
		}
	}
	sorted := sortedPeers(peers)
	out := make([]PresenceRecord, 0, len(sorted))
	for _, peer := range sorted {
		channels := dedupeSorted(append([]string(nil), voiceByPeer[peer.PeerID]...))
		out = append(out, PresenceRecord{PeerID: peer.PeerID, Role: peer.Role, Source: peer.Source, LastSeenAt: peer.LastSeenAt, Status: presenceStatusAt(now, peer.LastSeenAt, len(channels) > 0), ActiveVoiceChannels: channels})
	}
	return out
}

func (s *Service) CreateServer(name, description string) (ServerRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "New Server"
	}
	now := time.Now().UTC()
	serverID := randomID("server")
	general := ChannelRecord{ID: randomID("channel"), ServerID: serverID, Name: "general", Voice: false, CreatedAt: now}
	capabilities := advertisedCapabilities(s.cfg.Role)
	s.mu.Lock()
	manifest := Manifest{
		ServerID:                 serverID,
		Name:                     name,
		Description:              strings.TrimSpace(description),
		OwnerPeerID:              s.state.Identity.PeerID,
		OwnerPublicKey:           s.state.Identity.PublicKey,
		OwnerAddresses:           []string{s.listenAddressLocked()},
		BootstrapAddrs:           append([]string(nil), s.cfg.BootstrapAddrs...),
		RelayAddrs:               append([]string(nil), s.cfg.RelayAddrs...),
		Capabilities:             capabilities,
		HistoryRetentionMessages: s.cfg.HistoryLimit,
		HistoryCoverage:          HistoryCoverageLocalWindow,
		HistoryDurability:        HistoryDurabilitySingleNode,
		IssuedAt:                 now,
		UpdatedAt:                now,
	}
	if s.cfg.Role == RoleBootstrap {
		manifest.BootstrapAddrs = dedupeSorted(append(manifest.BootstrapAddrs, s.listenAddressLocked()))
	}
	if s.cfg.Role == RoleRelay {
		manifest.RelayAddrs = dedupeSorted(append(manifest.RelayAddrs, s.listenAddressLocked()))
	}
	if err := manifest.Sign(s.state.Identity); err != nil {
		s.mu.Unlock()
		return ServerRecord{}, err
	}
	invite := Invite{
		ServerID:       serverID,
		OwnerPeerID:    s.state.Identity.PeerID,
		OwnerPublicKey: s.state.Identity.PublicKey,
		ServerAddrs:    []string{s.listenAddressLocked()},
		BootstrapAddrs: append([]string(nil), manifest.BootstrapAddrs...),
		RelayAddrs:     append([]string(nil), manifest.RelayAddrs...),
		ManifestHash:   manifest.Hash(),
		ExpiresAt:      now.Add(7 * 24 * time.Hour),
	}
	if err := invite.Sign(s.state.Identity); err != nil {
		s.mu.Unlock()
		return ServerRecord{}, err
	}
	deeplink, err := invite.Deeplink()
	if err != nil {
		s.mu.Unlock()
		return ServerRecord{}, err
	}
	server := ServerRecord{
		ID:          serverID,
		Name:        name,
		Description: strings.TrimSpace(description),
		OwnerPeerID: s.state.Identity.PeerID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Members:     []string{s.state.Identity.PeerID},
		Channels:    map[string]ChannelRecord{general.ID: general},
		Manifest:    manifest,
		Invite:      deeplink,
	}
	s.state.Servers[serverID] = server
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return ServerRecord{}, err
	}
	s.mu.Unlock()
	_ = s.publishManifest(manifest)
	s.recordTelemetry("server.created id=" + serverID)
	s.emitEvent(Event{Type: "server.created", Time: now, Payload: map[string]any{"server_id": serverID}})
	return server, nil
}

func (s *Service) SearchMessages(req SearchMessagesRequest) ([]MessageRecord, error) {
	serverID, scopeType, scopeID, err := normalizeNotificationScope(req.ServerID, req.ScopeType, req.ScopeID)
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(req.Query))
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	s.mu.RLock()
	messages := make([]MessageRecord, 0, len(s.state.Messages))
	for _, msg := range s.state.Messages {
		if serverID != "" && msg.ServerID != serverID {
			continue
		}
		if scopeType != "" && msg.ScopeType != scopeType {
			continue
		}
		if scopeID != "" && msg.ScopeID != scopeID {
			continue
		}
		if !req.IncludeDeleted && msg.Deleted {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(msg.Body), query) {
			continue
		}
		messages = append(messages, msg)
	}
	s.mu.RUnlock()
	sort.Slice(messages, func(i, j int) bool {
		left := messageSortTime(messages[i])
		right := messageSortTime(messages[j])
		if !left.Equal(right) {
			return left.After(right)
		}
		return messages[i].ID > messages[j].ID
	})
	if len(messages) > limit {
		messages = messages[:limit]
	}
	return messages, nil
}

func (s *Service) SearchNotifications(req SearchNotificationsRequest) (SearchNotificationsResponse, error) {
	serverID, scopeType, scopeID, err := normalizeNotificationScope(req.ServerID, req.ScopeType, req.ScopeID)
	if err != nil {
		return SearchNotificationsResponse{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	s.mu.RLock()
	identity := s.state.Identity
	readThrough := notificationReadThroughForScopeLocked(s.state.Settings, serverID, scopeType, scopeID)
	notifications := make([]NotificationRecord, 0, len(s.state.Messages))
	unreadCount := 0
	for _, msg := range s.state.Messages {
		if serverID != "" && msg.ServerID != serverID {
			continue
		}
		if scopeType != "" && msg.ScopeType != scopeType {
			continue
		}
		if scopeID != "" && msg.ScopeID != scopeID {
			continue
		}
		if !req.IncludeDeleted && msg.Deleted {
			continue
		}
		note, ok := notificationRecordForMessage(identity, s.state.Settings, msg)
		if !ok {
			continue
		}
		note.ServerName, note.ScopeName = s.notificationSummaryLabelsLocked(msg)
		if msg.ScopeType == "dm" {
			note.ParticipantIDs = append(note.ParticipantIDs, s.notificationSummaryDirectParticipantIDsLocked(msg.ScopeID)...)
		}
		if note.Unread {
			unreadCount++
		}
		if req.UnreadOnly && !note.Unread {
			continue
		}
		notifications = append(notifications, note)
	}
	s.mu.RUnlock()
	sort.Slice(notifications, func(i, j int) bool {
		left := notifications[i].CreatedAt
		right := notifications[j].CreatedAt
		if !left.Equal(right) {
			return left.After(right)
		}
		return notifications[i].Message.ID > notifications[j].Message.ID
	})
	if len(notifications) > limit {
		notifications = notifications[:limit]
	}
	return SearchNotificationsResponse{Notifications: notifications, UnreadCount: unreadCount, ReadThrough: readThrough}, nil
}

func (s *Service) NotificationSummary() NotificationSummaryResponse {
	s.mu.RLock()
	identity := s.state.Identity
	readThrough := notificationsReadThroughLocked(s.state.Settings)
	bucketsByKey := make(map[string]NotificationSummaryBucket)
	serversByID := make(map[string]NotificationSummaryServerBucket)
	directsByID := make(map[string]NotificationSummaryDirectBucket)
	kindsByName := make(map[string]NotificationSummaryKindBucket)
	totalUnread := 0
	for _, msg := range s.state.Messages {
		if msg.Deleted {
			continue
		}
		note, ok := notificationRecordForMessage(identity, s.state.Settings, msg)
		if !ok || !note.Unread {
			continue
		}
		totalUnread++
		serverName, scopeName := s.notificationSummaryLabelsLocked(msg)
		key := fmt.Sprintf("%s|%s|%s", msg.ServerID, msg.ScopeType, msg.ScopeID)
		bucket := bucketsByKey[key]
		if bucket.ScopeID == "" {
			bucket = NotificationSummaryBucket{ServerID: msg.ServerID, ServerName: serverName, ScopeType: msg.ScopeType, ScopeID: msg.ScopeID, ScopeName: scopeName}
		}
		bucket.UnreadCount++
		if shouldAdvanceNotificationSummaryMessage(bucket.LatestAt, bucket.LatestMessageID, note.CreatedAt, msg.ID) {
			bucket.LatestAt = note.CreatedAt
			bucket.LatestMessageID = msg.ID
			bucket.LatestSenderPeerID = msg.SenderPeerID
		}
		bucketsByKey[key] = bucket
		kindBucket := kindsByName[note.Kind]
		if kindBucket.Kind == "" {
			kindBucket = NotificationSummaryKindBucket{Kind: note.Kind}
		}
		kindBucket.UnreadCount++
		if shouldAdvanceNotificationSummaryMessage(kindBucket.LatestAt, kindBucket.LatestMessageID, note.CreatedAt, msg.ID) {
			kindBucket.LatestAt = note.CreatedAt
			kindBucket.LatestMessageID = msg.ID
			kindBucket.LatestSenderPeerID = msg.SenderPeerID
			kindBucket.LatestServerID = msg.ServerID
			kindBucket.LatestServerName = serverName
			kindBucket.LatestScopeType = msg.ScopeType
			kindBucket.LatestScopeID = msg.ScopeID
			kindBucket.LatestScopeName = scopeName
			if msg.ScopeType == "dm" {
				kindBucket.LatestParticipantIDs = append([]string(nil), s.notificationSummaryDirectParticipantIDsLocked(msg.ScopeID)...)
			} else {
				kindBucket.LatestParticipantIDs = nil
			}
		}
		kindsByName[note.Kind] = kindBucket
		switch msg.ScopeType {
		case "channel":
			if msg.ServerID != "" {
				serverBucket := serversByID[msg.ServerID]
				if serverBucket.ServerID == "" {
					serverBucket = NotificationSummaryServerBucket{ServerID: msg.ServerID, ServerName: serverName}
				}
				serverBucket.UnreadCount++
				if shouldAdvanceNotificationSummaryMessage(serverBucket.LatestAt, serverBucket.LatestMessageID, note.CreatedAt, msg.ID) {
					serverBucket.LatestAt = note.CreatedAt
					serverBucket.LatestMessageID = msg.ID
					serverBucket.LatestSenderPeerID = msg.SenderPeerID
					serverBucket.LatestScopeType = msg.ScopeType
					serverBucket.LatestScopeID = msg.ScopeID
					serverBucket.LatestScopeName = scopeName
				}
				serversByID[msg.ServerID] = serverBucket
			}
		case "dm":
			directBucket := directsByID[msg.ScopeID]
			if directBucket.ScopeID == "" {
				directBucket = NotificationSummaryDirectBucket{ScopeID: msg.ScopeID, ScopeName: scopeName, ParticipantIDs: s.notificationSummaryDirectParticipantIDsLocked(msg.ScopeID)}
			}
			directBucket.UnreadCount++
			if shouldAdvanceNotificationSummaryMessage(directBucket.LatestAt, directBucket.LatestMessageID, note.CreatedAt, msg.ID) {
				directBucket.LatestAt = note.CreatedAt
				directBucket.LatestMessageID = msg.ID
				directBucket.LatestSenderPeerID = msg.SenderPeerID
			}
			directsByID[msg.ScopeID] = directBucket
		}
	}
	s.mu.RUnlock()
	buckets := make([]NotificationSummaryBucket, 0, len(bucketsByKey))
	for _, bucket := range bucketsByKey {
		buckets = append(buckets, bucket)
	}
	sort.Slice(buckets, func(i, j int) bool {
		if !buckets[i].LatestAt.Equal(buckets[j].LatestAt) {
			return buckets[i].LatestAt.After(buckets[j].LatestAt)
		}
		if buckets[i].ServerID != buckets[j].ServerID {
			return buckets[i].ServerID < buckets[j].ServerID
		}
		if buckets[i].ScopeType != buckets[j].ScopeType {
			return buckets[i].ScopeType < buckets[j].ScopeType
		}
		return buckets[i].ScopeID < buckets[j].ScopeID
	})
	servers := make([]NotificationSummaryServerBucket, 0, len(serversByID))
	for _, bucket := range serversByID {
		servers = append(servers, bucket)
	}
	sort.Slice(servers, func(i, j int) bool {
		if !servers[i].LatestAt.Equal(servers[j].LatestAt) {
			return servers[i].LatestAt.After(servers[j].LatestAt)
		}
		return servers[i].ServerID < servers[j].ServerID
	})
	directs := make([]NotificationSummaryDirectBucket, 0, len(directsByID))
	for _, bucket := range directsByID {
		directs = append(directs, bucket)
	}
	sort.Slice(directs, func(i, j int) bool {
		if !directs[i].LatestAt.Equal(directs[j].LatestAt) {
			return directs[i].LatestAt.After(directs[j].LatestAt)
		}
		return directs[i].ScopeID < directs[j].ScopeID
	})
	kinds := make([]NotificationSummaryKindBucket, 0, len(kindsByName))
	for _, bucket := range kindsByName {
		kinds = append(kinds, bucket)
	}
	sort.Slice(kinds, func(i, j int) bool {
		if !kinds[i].LatestAt.Equal(kinds[j].LatestAt) {
			return kinds[i].LatestAt.After(kinds[j].LatestAt)
		}
		return kinds[i].Kind < kinds[j].Kind
	})
	return NotificationSummaryResponse{TotalUnread: totalUnread, Buckets: buckets, Servers: servers, Directs: directs, Kinds: kinds, ReadThrough: readThrough}
}

func (s *Service) MarkNotificationsRead(through time.Time) (time.Time, error) {
	return s.MarkNotificationsReadScoped(MarkNotificationsReadRequest{Through: through})
}

func (s *Service) notificationsReadResponse(serverID, scopeType, scopeID string, readThrough time.Time) MarkNotificationsReadResponse {
	resp := MarkNotificationsReadResponse{
		ReadThrough: readThrough,
		ServerID:    serverID,
		ScopeType:   scopeType,
		ScopeID:     scopeID,
	}
	s.mu.RLock()
	serverName, scopeName := s.notificationSummaryLabelsLocked(MessageRecord{ServerID: serverID, ScopeType: scopeType, ScopeID: scopeID})
	if serverName != "" {
		resp.ServerName = serverName
	}
	if scopeName != "" {
		resp.ScopeName = scopeName
	}
	if scopeType == "dm" {
		resp.ParticipantIDs = append(resp.ParticipantIDs, s.notificationSummaryDirectParticipantIDsLocked(scopeID)...)
	}
	s.mu.RUnlock()
	return resp
}

func (s *Service) MarkNotificationsReadScoped(req MarkNotificationsReadRequest) (time.Time, error) {
	serverID, scopeType, scopeID, err := normalizeNotificationScope(req.ServerID, req.ScopeType, req.ScopeID)
	if err != nil {
		return time.Time{}, err
	}
	through := req.Through
	if through.IsZero() {
		through = time.Now().UTC()
	} else {
		through = through.UTC()
	}
	s.mu.Lock()
	current := notificationReadThroughForScopeLocked(s.state.Settings, serverID, scopeType, scopeID)
	if through.Before(current) {
		through = current
	}
	if s.state.Settings == nil {
		s.state.Settings = map[string]string{}
	}
	switch {
	case scopeType == "" && serverID == "":
		s.state.Settings[notificationsReadThroughSetting] = through.Format(time.RFC3339Nano)
	case scopeType == "":
		s.state.Settings[notificationsServerReadThroughKey(serverID)] = through.Format(time.RFC3339Nano)
	default:
		s.state.Settings[notificationsScopedReadThroughKey(serverID, scopeType, scopeID)] = through.Format(time.RFC3339Nano)
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return time.Time{}, err
	}
	serverName, scopeName := s.notificationSummaryLabelsLocked(MessageRecord{ServerID: serverID, ScopeType: scopeType, ScopeID: scopeID})
	participantIDs := []string(nil)
	if scopeType == "dm" {
		participantIDs = append(participantIDs, s.notificationSummaryDirectParticipantIDsLocked(scopeID)...)
	}
	s.mu.Unlock()
	telemetry := "notifications.read through=" + through.Format(time.RFC3339Nano)
	payload := map[string]any{"read_through": through}
	if serverID != "" {
		telemetry += " server_id=" + serverID
		payload["server_id"] = serverID
	}
	if scopeType != "" {
		telemetry += " scope_type=" + scopeType + " scope_id=" + scopeID
		payload["scope_type"] = scopeType
		payload["scope_id"] = scopeID
	}
	if serverName != "" {
		payload["server_name"] = serverName
	}
	if scopeName != "" {
		payload["scope_name"] = scopeName
	}
	if len(participantIDs) > 0 {
		payload["participant_ids"] = participantIDs
	}
	s.recordTelemetry(telemetry)
	s.emitEvent(Event{Type: "notifications.read", Time: time.Now().UTC(), Payload: payload})
	return through, nil
}

func (s *Service) messageSearchRecords(messages []MessageRecord) []MessageSearchRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := make([]MessageSearchRecord, 0, len(messages))
	for _, msg := range messages {
		serverName, scopeName := s.notificationSummaryLabelsLocked(msg)
		record := MessageSearchRecord{Message: msg, ServerName: serverName, ScopeName: scopeName}
		if msg.ScopeType == "dm" {
			record.ParticipantIDs = append(record.ParticipantIDs, s.notificationSummaryDirectParticipantIDsLocked(msg.ScopeID)...)
		}
		results = append(results, record)
	}
	return results
}

func (s *Service) SearchMentions(req SearchMentionsRequest) ([]MentionRecord, error) {
	serverID, scopeType, scopeID, err := normalizeNotificationScope(req.ServerID, req.ScopeType, req.ScopeID)
	if err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	s.mu.RLock()
	selfPeerID := s.state.Identity.PeerID
	tokens := selfMentionTokens(s.state.Identity)
	mentions := make([]MentionRecord, 0, len(s.state.Messages))
	for _, msg := range s.state.Messages {
		if msg.SenderPeerID == selfPeerID {
			continue
		}
		if serverID != "" && msg.ServerID != serverID {
			continue
		}
		if scopeType != "" && msg.ScopeType != scopeType {
			continue
		}
		if scopeID != "" && msg.ScopeID != scopeID {
			continue
		}
		if !req.IncludeDeleted && msg.Deleted {
			continue
		}
		found := findMentionTokens(msg.Body, tokens)
		if len(found) == 0 {
			continue
		}
		serverName, scopeName := s.notificationSummaryLabelsLocked(msg)
		mention := MentionRecord{Message: msg, Tokens: found, ServerName: serverName, ScopeName: scopeName}
		if msg.ScopeType == "dm" {
			mention.ParticipantIDs = append(mention.ParticipantIDs, s.notificationSummaryDirectParticipantIDsLocked(msg.ScopeID)...)
		}
		mentions = append(mentions, mention)
	}
	s.mu.RUnlock()
	sort.Slice(mentions, func(i, j int) bool {
		left := messageSortTime(mentions[i].Message)
		right := messageSortTime(mentions[j].Message)
		if !left.Equal(right) {
			return left.After(right)
		}
		return mentions[i].Message.ID > mentions[j].Message.ID
	})
	if len(mentions) > limit {
		mentions = mentions[:limit]
	}
	return mentions, nil
}

func (s *Service) PreviewDeeplink(raw string) (ServerPreview, error) {
	invite, err := ParseDeeplink(raw)
	if err != nil {
		return ServerPreview{}, err
	}
	previewInfo, previewErr := s.resolveServerPreview(invite.ServerID, invite.ServerAddrs)
	if previewErr == nil {
		if previewInfo.Manifest.Hash() != invite.ManifestHash {
			return ServerPreview{}, errors.New("invite manifest hash mismatch")
		}
		ownerRole := previewInfo.OwnerRole
		if ownerRole == "" {
			ownerRole = inferPeerRoleFromCapabilities("", previewInfo.Manifest.Capabilities)
		}
		labels := append([]string(nil), previewInfo.SafetyLabels...)
		if len(labels) == 0 {
			labels = safetyLabelsForManifest(previewInfo.Manifest)
		}
		return ServerPreview{
			Invite:       invite,
			Manifest:     previewInfo.Manifest,
			OwnerRole:    ownerRole,
			Channels:     previewInfo.Channels,
			MemberCount:  previewInfo.MemberCount,
			SafetyLabels: labels,
		}, nil
	}
	manifest, err := s.resolveManifest(invite.ServerID, invite.BootstrapAddrs, invite.ServerAddrs)
	if err != nil {
		return ServerPreview{}, err
	}
	if manifest.Hash() != invite.ManifestHash {
		return ServerPreview{}, errors.New("invite manifest hash mismatch")
	}
	return ServerPreview{Invite: invite, Manifest: manifest, OwnerRole: inferPeerRoleFromCapabilities("", manifest.Capabilities), SafetyLabels: safetyLabelsForManifest(manifest)}, nil
}

func (s *Service) JoinByDeeplink(raw string) (ServerRecord, error) {
	invite, err := ParseDeeplink(raw)
	if err != nil {
		return ServerRecord{}, err
	}
	manifest, err := s.resolveManifest(invite.ServerID, invite.BootstrapAddrs, invite.ServerAddrs)
	if err != nil {
		return ServerRecord{}, err
	}
	if manifest.Hash() != invite.ManifestHash {
		return ServerRecord{}, errors.New("invite manifest hash mismatch")
	}
	joinReq := JoinRequest{
		Invite: invite,
		Requester: PeerInfo{
			PeerID:    s.PeerID(),
			Role:      s.cfg.Role,
			Addresses: []string{s.ListenAddress()},
			PublicKey: s.Snapshot().Identity.PublicKey,
		},
	}
	joinReq.Capabilities = advertisedCapabilities(s.cfg.Role)
	addresses := dedupeSorted(append([]string{}, invite.ServerAddrs...))
	var joinResp JoinResponse
	var lastErr error
	for _, addr := range addresses {
		if err := s.postJSON(context.Background(), addr, "/_xorein/join", joinReq, &joinResp); err != nil {
			lastErr = err
			continue
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return ServerRecord{}, fmt.Errorf("join failed: %w", lastErr)
	}
	if err := joinResp.Manifest.Verify(); err != nil {
		return ServerRecord{}, err
	}
	s.mu.Lock()
	owner := PeerRecord{
		PeerID:     joinResp.Manifest.OwnerPeerID,
		Role:       inferPeerRoleFromCapabilities(RoleClient, joinResp.Manifest.Capabilities),
		Addresses:  append([]string(nil), joinResp.Manifest.OwnerAddresses...),
		PublicKey:  joinResp.Manifest.OwnerPublicKey,
		Source:     "manifest",
		LastSeenAt: time.Now().UTC(),
	}
	s.upsertPeerLocked(owner)
	server := joinResp.Server
	server.Manifest = joinResp.Manifest
	s.state.Servers[server.ID] = server
	for _, msg := range joinResp.History {
		s.state.Messages[msg.ID] = msg
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return ServerRecord{}, err
	}
	s.mu.Unlock()
	s.recordTelemetry("server.joined id=" + server.ID)
	s.emitEvent(Event{Type: "server.joined", Time: time.Now().UTC(), Payload: map[string]any{"server_id": server.ID}})
	return server, nil
}

func (s *Service) CreateChannel(serverID, name string, voice bool) (ChannelRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ChannelRecord{}, errors.New("channel name is required")
	}
	now := time.Now().UTC()
	s.mu.Lock()
	server, ok := s.state.Servers[serverID]
	if !ok {
		s.mu.Unlock()
		return ChannelRecord{}, errors.New("server not found")
	}
	channel := ChannelRecord{ID: randomID("channel"), ServerID: serverID, Name: name, Voice: voice, CreatedAt: now}
	if server.Channels == nil {
		server.Channels = map[string]ChannelRecord{}
	}
	server.Channels[channel.ID] = channel
	server.UpdatedAt = now
	s.state.Servers[serverID] = server
	recipients := make([]string, 0, len(server.Members))
	for _, member := range server.Members {
		if member != s.state.Identity.PeerID {
			recipients = append(recipients, member)
		}
	}
	identity := s.state.Identity
	delivery := Delivery{ID: channel.ID, Kind: "channel_create", ScopeID: channel.ID, ScopeType: "channel", ServerID: serverID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: dedupeSorted(recipients), Body: channel.Name, Muted: channel.Voice, CreatedAt: now}
	if err := delivery.Sign(identity); err != nil {
		s.mu.Unlock()
		return ChannelRecord{}, err
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return ChannelRecord{}, err
	}
	s.mu.Unlock()
	if len(delivery.RecipientPeerIDs) > 0 {
		if err := s.deliverToRecipients(delivery); err != nil {
			return ChannelRecord{}, err
		}
	}
	s.recordTelemetry("channel.created id=" + channel.ID)
	s.emitEvent(Event{Type: "channel.created", Time: now, Payload: map[string]any{"channel_id": channel.ID, "server_id": serverID}})
	return channel, nil
}

func (s *Service) CreateDM(peerID string) (DMRecord, error) {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return DMRecord{}, errors.New("peer id is required")
	}
	now := time.Now().UTC()
	participants := dedupeSorted([]string{s.PeerID(), peerID})
	dmID := strings.Join(participants, ":")
	s.mu.Lock()
	if dm, ok := s.state.DMs[dmID]; ok {
		s.mu.Unlock()
		return dm, nil
	}
	dm := DMRecord{ID: dmID, Participants: participants, CreatedAt: now}
	s.state.DMs[dm.ID] = dm
	identity := s.state.Identity
	delivery := Delivery{ID: dm.ID, Kind: "dm_create", ScopeID: dm.ID, ScopeType: "dm", SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: dedupeSorted([]string{peerID}), CreatedAt: now}
	if err := delivery.Sign(identity); err != nil {
		s.mu.Unlock()
		return DMRecord{}, err
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return DMRecord{}, err
	}
	s.mu.Unlock()
	if len(delivery.RecipientPeerIDs) > 0 {
		if err := s.deliverToRecipients(delivery); err != nil {
			return DMRecord{}, err
		}
	}
	s.recordTelemetry("dm.created id=" + dm.ID)
	s.emitEvent(Event{Type: "dm.created", Time: now, Payload: map[string]any{"dm_id": dm.ID, "participant_ids": append([]string(nil), dm.Participants...)}})
	return dm, nil
}

func (s *Service) SendChannelMessage(channelID, body string) (MessageRecord, error) {
	return s.sendMessage("channel", channelID, strings.TrimSpace(body))
}

func (s *Service) SendDMMessage(dmID, body string) (MessageRecord, error) {
	return s.sendMessage("dm", dmID, strings.TrimSpace(body))
}

func (s *Service) sendMessage(scopeType, scopeID, body string) (MessageRecord, error) {
	if body == "" {
		return MessageRecord{}, errors.New("message body is required")
	}
	now := time.Now().UTC()
	recipients, serverID, err := s.scopeRecipients(scopeType, scopeID)
	if err != nil {
		return MessageRecord{}, err
	}
	msg := MessageRecord{ID: randomID("msg"), ScopeType: scopeType, ScopeID: scopeID, ServerID: serverID, SenderPeerID: s.PeerID(), Body: body, CreatedAt: now}
	delivery := Delivery{
		ID:               msg.ID,
		Kind:             scopeType + "_message",
		ScopeID:          scopeID,
		ScopeType:        scopeType,
		ServerID:         serverID,
		SenderPeerID:     s.PeerID(),
		SenderPublicKey:  s.Snapshot().Identity.PublicKey,
		RecipientPeerIDs: recipients,
		Body:             body,
		CreatedAt:        now,
	}
	if err := delivery.Sign(s.Snapshot().Identity); err != nil {
		return MessageRecord{}, err
	}
	s.mu.Lock()
	s.state.Messages[msg.ID] = msg
	s.state.Deliveries[delivery.ID] = struct{}{}
	pruned := s.pruneServerHistoryLocked(serverID)
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	s.mu.Unlock()
	if pruned {
		s.recordTelemetry("history.pruned server=" + serverID)
	}
	if err := s.deliverToRecipients(delivery); err != nil {
		return MessageRecord{}, err
	}
	s.emitEvent(Event{Type: "message.created", Time: now, Payload: map[string]any{"message_id": msg.ID, "scope_id": scopeID}})
	return msg, nil
}

func (s *Service) EditMessage(messageID, body string) (MessageRecord, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return MessageRecord{}, errors.New("message body is required")
	}
	s.mu.Lock()
	msg, ok := s.state.Messages[messageID]
	if !ok {
		s.mu.Unlock()
		return MessageRecord{}, errors.New("message not found")
	}
	identity := s.state.Identity
	if msg.SenderPeerID != identity.PeerID {
		s.mu.Unlock()
		return MessageRecord{}, errors.New("message is owned by another peer")
	}
	msg.Body = body
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	scopeType, scopeID, serverID := msg.ScopeType, msg.ScopeID, msg.ServerID
	recipients, _, err := s.scopeRecipientsLocked(scopeType, scopeID)
	if err != nil {
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	delivery := Delivery{ID: messageID, Kind: "message_edit", ScopeID: scopeID, ScopeType: scopeType, ServerID: serverID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: recipients, Body: body, CreatedAt: msg.UpdatedAt}
	if err := delivery.Sign(identity); err != nil {
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	s.mu.Unlock()
	if err := s.deliverToRecipients(delivery); err != nil {
		return MessageRecord{}, err
	}
	s.emitEvent(Event{Type: "message.edited", Time: time.Now().UTC(), Payload: map[string]any{"message_id": msg.ID}})
	return msg, nil
}

func (s *Service) DeleteMessage(messageID string) error {
	s.mu.Lock()
	msg, ok := s.state.Messages[messageID]
	if !ok {
		s.mu.Unlock()
		return errors.New("message not found")
	}
	identity := s.state.Identity
	if msg.SenderPeerID != identity.PeerID {
		s.mu.Unlock()
		return errors.New("message is owned by another peer")
	}
	msg.Body = ""
	msg.Deleted = true
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	recipients, _, err := s.scopeRecipientsLocked(msg.ScopeType, msg.ScopeID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	delivery := Delivery{ID: messageID, Kind: "message_delete", ScopeID: msg.ScopeID, ScopeType: msg.ScopeType, ServerID: msg.ServerID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: recipients, CreatedAt: msg.UpdatedAt}
	if err := delivery.Sign(identity); err != nil {
		s.mu.Unlock()
		return err
	}
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()
	if err := s.deliverToRecipients(delivery); err != nil {
		return err
	}
	s.emitEvent(Event{Type: "message.deleted", Time: time.Now().UTC(), Payload: map[string]any{"message_id": messageID}})
	return nil
}

func (s *Service) JoinVoice(channelID string, muted bool) error {
	return s.sendVoiceControl(channelID, "voice_join", muted, nil)
}

func (s *Service) LeaveVoice(channelID string) error {
	return s.sendVoiceControl(channelID, "voice_leave", false, nil)
}

func (s *Service) SetVoiceMuted(channelID string, muted bool) error {
	return s.sendVoiceControl(channelID, "voice_mute", muted, nil)
}

func (s *Service) SendVoiceFrame(channelID string, data []byte) error {
	copied := append([]byte(nil), data...)
	return s.sendVoiceControl(channelID, "voice_frame", false, copied)
}

func (s *Service) sendVoiceControl(channelID, kind string, muted bool, data []byte) error {
	recipients, serverID, err := s.scopeRecipients("channel", channelID)
	if err != nil {
		return err
	}
	identity := s.Snapshot().Identity
	delivery := Delivery{ID: randomID("voice"), Kind: kind, ScopeID: channelID, ScopeType: "channel", ServerID: serverID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: recipients, Muted: muted, Data: data, CreatedAt: time.Now().UTC()}
	if err := delivery.Sign(identity); err != nil {
		return err
	}
	if err := s.applyDelivery(delivery); err != nil {
		return err
	}
	return s.deliverToRecipients(delivery)
}

func (s *Service) scopeRecipients(scopeType, scopeID string) ([]string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scopeRecipientsLocked(scopeType, scopeID)
}

func (s *Service) scopeRecipientsLocked(scopeType, scopeID string) ([]string, string, error) {
	scopeType = strings.TrimSpace(scopeType)
	scopeID = strings.TrimSpace(scopeID)
	switch scopeType {
	case "channel":
		for _, server := range s.state.Servers {
			if _, ok := server.Channels[scopeID]; ok {
				members := make([]string, 0, len(server.Members))
				for _, member := range server.Members {
					if member != s.state.Identity.PeerID {
						members = append(members, member)
					}
				}
				return dedupeSorted(members), server.ID, nil
			}
		}
		return nil, "", errors.New("channel not found")
	case "dm":
		dm, ok := s.state.DMs[scopeID]
		if !ok {
			return nil, "", errors.New("dm not found")
		}
		members := make([]string, 0, len(dm.Participants))
		for _, participant := range dm.Participants {
			if participant != s.state.Identity.PeerID {
				members = append(members, participant)
			}
		}
		return dedupeSorted(members), "", nil
	default:
		return nil, "", errors.New("unsupported scope")
	}
}

func (s *Service) runDiscoveryPass() {
	_ = s.fetchConfiguredManualPeers()
	_ = s.registerWithBootstraps()
	_ = s.fetchPeersFromBootstraps()
	_ = s.pingKnownPeers()
	_ = s.drainRelays()
}

func (s *Service) discoveryLoop() {
	ticker := time.NewTicker(s.cfg.DiscoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			s.runDiscoveryPass()
		}
	}
}

func (s *Service) fetchConfiguredManualPeers() error {
	var firstErr error
	for _, addr := range dedupeSorted(s.cfg.ManualPeers) {
		peer, err := s.fetchPeerInfo(addr)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.recordTelemetry("discovery.manual.failed addr=" + addr + " err=" + err.Error())
			continue
		}
		s.mu.Lock()
		peer.Source = "manual"
		peer.LastSeenAt = time.Now().UTC()
		s.upsertPeerLocked(peer)
		err = s.saveLocked()
		s.mu.Unlock()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.recordTelemetry("discovery.manual.added peer=" + peer.PeerID + " addr=" + addr)
	}
	return firstErr
}

func (s *Service) registerWithBootstraps() error {
	self := s.selfPeerInfo()
	var firstErr error
	for _, addr := range s.bootstrapTargets() {
		if err := s.postJSON(context.Background(), addr, "/_xorein/bootstrap/register", self, nil); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.recordTelemetry("discovery.bootstrap.register.failed addr=" + addr + " err=" + err.Error())
			continue
		}
		s.recordTelemetry("discovery.bootstrap.registered addr=" + addr)
	}
	return firstErr
}

func (s *Service) fetchPeersFromBootstraps() error {
	var firstErr error
	for _, addr := range s.bootstrapTargets() {
		var peers []PeerInfo
		if err := s.getJSON(context.Background(), addr, "/_xorein/bootstrap/peers", &peers); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.mu.Lock()
		for _, peer := range peers {
			if peer.PeerID == s.state.Identity.PeerID {
				continue
			}
			s.upsertPeerLocked(PeerRecord{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey, Source: "bootstrap", LastSeenAt: time.Now().UTC()})
		}
		_ = s.saveLocked()
		s.mu.Unlock()
	}
	return firstErr
}

func (s *Service) pingKnownPeers() error {
	var firstErr error
	peers := s.Snapshot().KnownPeers
	for _, peer := range peers {
		if peer.PeerID == s.PeerID() {
			continue
		}
		for _, addr := range dedupeSorted(peer.Addresses) {
			info, err := s.fetchPeerInfo(addr)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			s.mu.Lock()
			info.Source = peer.Source
			info.LastSeenAt = time.Now().UTC()
			s.upsertPeerLocked(info)
			_ = s.saveLocked()
			s.mu.Unlock()
			break
		}
	}
	return firstErr
}

func (s *Service) drainRelays() error {
	var firstErr error
	for _, addr := range s.relayTargets() {
		var deliveries []Delivery
		if err := s.postJSON(context.Background(), addr, "/_xorein/relay/drain", DrainRequest{PeerID: s.PeerID()}, &deliveries); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, delivery := range deliveries {
			if err := s.applyDelivery(delivery); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if len(deliveries) > 0 {
			s.recordTelemetry(fmt.Sprintf("relay.drain addr=%s count=%d", addr, len(deliveries)))
		}
	}
	return firstErr
}

func (s *Service) publishManifest(manifest Manifest) error {
	var firstErr error
	for _, addr := range s.bootstrapTargets() {
		if err := s.postJSON(context.Background(), addr, "/_xorein/manifests/publish", manifest, nil); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *Service) resolveManifest(serverID string, bootstrapAddrs, serverAddrs []string) (Manifest, error) {
	targets := dedupeSorted(append([]string{}, bootstrapAddrs...))
	for _, addr := range serverAddrs {
		targets = append(targets, addr)
	}
	targets = dedupeSorted(targets)
	var lastErr error
	for _, addr := range targets {
		var manifest Manifest
		path := "/_xorein/manifests/" + serverID
		if err := s.getJSON(context.Background(), addr, path, &manifest); err != nil {
			lastErr = err
			continue
		}
		if err := manifest.Verify(); err != nil {
			lastErr = err
			continue
		}
		return manifest, nil
	}
	if lastErr == nil {
		lastErr = errors.New("manifest not found")
	}
	return Manifest{}, lastErr
}

func (s *Service) resolveServerPreview(serverID string, serverAddrs []string) (ServerPreviewInfo, error) {
	targets := dedupeSorted(append([]string{}, serverAddrs...))
	var lastErr error
	for _, addr := range targets {
		var preview ServerPreviewInfo
		path := "/_xorein/preview/" + serverID
		if err := s.getJSON(context.Background(), addr, path, &preview); err != nil {
			lastErr = err
			continue
		}
		if err := preview.Manifest.Verify(); err != nil {
			lastErr = err
			continue
		}
		return preview, nil
	}
	if lastErr == nil {
		lastErr = errors.New("preview not found")
	}
	return ServerPreviewInfo{}, lastErr
}

func (s *Service) deliverToRecipients(delivery Delivery) error {
	var firstErr error
	for _, peerID := range dedupeSorted(delivery.RecipientPeerIDs) {
		if peerID == s.PeerID() {
			continue
		}
		if err := s.deliverToPeer(peerID, delivery); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *Service) deliverToPeer(peerID string, delivery Delivery) error {
	var peer PeerRecord
	var ok bool
	s.mu.RLock()
	peer, ok = s.state.KnownPeers[peerID]
	s.mu.RUnlock()
	if ok {
		addresses := dedupeSorted(peer.Addresses)
		for _, addr := range addresses {
			if err := s.postJSON(context.Background(), addr, "/_xorein/deliver", delivery, nil); err == nil {
				s.recordTelemetry("delivery.direct peer=" + peerID + " addr=" + addr)
				return nil
			} else {
				s.recordTelemetry("delivery.direct.failed peer=" + peerID + " addr=" + addr + " err=" + err.Error())
			}
		}
	}
	for _, addr := range livePeerAddresses(peerID) {
		if err := s.postJSON(context.Background(), addr, "/_xorein/deliver", delivery, nil); err == nil {
			s.recordTelemetry("delivery.local peer=" + peerID + " addr=" + addr)
			return nil
		} else {
			s.recordTelemetry("delivery.local.failed peer=" + peerID + " addr=" + addr + " err=" + err.Error())
		}
	}
	for _, addr := range s.relayTargets() {
		if err := s.postJSON(context.Background(), addr, "/_xorein/relay/store", delivery, nil); err == nil {
			s.recordTelemetry("delivery.relay peer=" + peerID + " addr=" + addr)
			return nil
		} else {
			s.recordTelemetry("delivery.relay.failed peer=" + peerID + " addr=" + addr + " err=" + err.Error())
		}
	}
	return errors.New("delivery failed on direct and relay paths")
}

func (s *Service) applyDelivery(delivery Delivery) error {
	if err := delivery.Verify(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, seen := s.state.Deliveries[delivery.ID]; seen && delivery.Kind != "message_edit" && delivery.Kind != "message_delete" {
		return nil
	}
	s.state.Deliveries[delivery.ID] = struct{}{}
	s.upsertPeerLocked(PeerRecord{PeerID: delivery.SenderPeerID, PublicKey: delivery.SenderPublicKey, Source: "delivery", LastSeenAt: time.Now().UTC()})
	now := time.Now().UTC()
	switch delivery.Kind {
	case "channel_create":
		server, ok := s.state.Servers[delivery.ServerID]
		if !ok {
			return errors.New("server not found for channel create")
		}
		if server.Channels == nil {
			server.Channels = map[string]ChannelRecord{}
		}
		channel := ChannelRecord{ID: delivery.ID, ServerID: delivery.ServerID, Name: delivery.Body, Voice: delivery.Muted, CreatedAt: delivery.CreatedAt}
		server.Channels[channel.ID] = channel
		if delivery.CreatedAt.After(server.UpdatedAt) {
			server.UpdatedAt = delivery.CreatedAt
		}
		s.state.Servers[server.ID] = server
		s.emitEventLocked(Event{Type: "channel.created", Time: now, Payload: map[string]any{"channel_id": channel.ID, "server_id": server.ID}})
	case "dm_create":
		if delivery.ScopeType != "dm" {
			return errors.New("dm create scope type mismatch")
		}
		if delivery.ID != delivery.ScopeID {
			return errors.New("dm create id does not match scope")
		}
		participants := dedupeSorted([]string{s.state.Identity.PeerID, delivery.SenderPeerID})
		expectedDMID := strings.Join(participants, ":")
		if delivery.ScopeID != expectedDMID {
			return errors.New("dm scope id does not match participants")
		}
		if _, ok := s.state.DMs[delivery.ScopeID]; !ok {
			s.state.DMs[delivery.ScopeID] = DMRecord{ID: delivery.ScopeID, Participants: participants, CreatedAt: delivery.CreatedAt}
			s.emitEventLocked(Event{Type: "dm.created", Time: now, Payload: map[string]any{"dm_id": delivery.ScopeID, "participant_ids": append([]string(nil), participants...)}})
		}
	case "channel_message", "dm_message":
		if delivery.ScopeType == "dm" {
			participants := dedupeSorted([]string{s.state.Identity.PeerID, delivery.SenderPeerID})
			expectedDMID := strings.Join(participants, ":")
			if delivery.ScopeID != expectedDMID {
				return errors.New("dm scope id does not match participants")
			}
			if _, ok := s.state.DMs[delivery.ScopeID]; !ok {
				s.state.DMs[delivery.ScopeID] = DMRecord{ID: delivery.ScopeID, Participants: participants, CreatedAt: delivery.CreatedAt}
			}
		}
		msg := MessageRecord{ID: delivery.ID, ScopeType: delivery.ScopeType, ScopeID: delivery.ScopeID, ServerID: delivery.ServerID, SenderPeerID: delivery.SenderPeerID, Body: delivery.Body, CreatedAt: delivery.CreatedAt}
		s.state.Messages[msg.ID] = msg
		s.pruneServerHistoryLocked(msg.ServerID)
		s.emitEventLocked(Event{Type: "message.created", Time: now, Payload: map[string]any{"message_id": msg.ID, "scope_id": msg.ScopeID}})
		if note, ok := notificationRecordForMessage(s.state.Identity, s.state.Settings, msg); ok {
			s.emitNotificationCreatedLocked(note, now)
		}
	case "message_edit":
		msg, ok := s.state.Messages[delivery.ID]
		if !ok {
			return errors.New("message not found for edit")
		}
		if msg.SenderPeerID != delivery.SenderPeerID {
			return errors.New("message edit sender mismatch")
		}
		before, hadBefore := notificationRecordForMessage(s.state.Identity, s.state.Settings, msg)
		msg.Body = delivery.Body
		msg.UpdatedAt = delivery.CreatedAt
		msg.Deleted = false
		s.state.Messages[msg.ID] = msg
		s.pruneServerHistoryLocked(msg.ServerID)
		s.emitEventLocked(Event{Type: "message.edited", Time: now, Payload: map[string]any{"message_id": msg.ID}})
		if note, ok := notificationRecordForMessage(s.state.Identity, s.state.Settings, msg); ok && (!hadBefore || !sameStringSet(before.Tokens, note.Tokens)) {
			s.emitNotificationCreatedLocked(note, now)
		}
	case "message_delete":
		msg, ok := s.state.Messages[delivery.ID]
		if !ok {
			return errors.New("message not found for delete")
		}
		if msg.SenderPeerID != delivery.SenderPeerID {
			return errors.New("message delete sender mismatch")
		}
		msg.Deleted = true
		msg.Body = ""
		msg.UpdatedAt = delivery.CreatedAt
		s.state.Messages[msg.ID] = msg
		s.pruneServerHistoryLocked(msg.ServerID)
		s.emitEventLocked(Event{Type: "message.deleted", Time: now, Payload: map[string]any{"message_id": msg.ID}})
	case "voice_join":
		session := s.voiceSessionLocked(delivery.ScopeID)
		session.Participants[delivery.SenderPeerID] = VoiceParticipant{PeerID: delivery.SenderPeerID, Muted: delivery.Muted, JoinedAt: delivery.CreatedAt}
		s.state.Voice[delivery.ScopeID] = session
		s.emitEventLocked(Event{Type: "voice.join", Time: now, Payload: map[string]any{"channel_id": delivery.ScopeID, "peer_id": delivery.SenderPeerID}})
	case "voice_leave":
		session := s.voiceSessionLocked(delivery.ScopeID)
		delete(session.Participants, delivery.SenderPeerID)
		s.state.Voice[delivery.ScopeID] = session
		s.emitEventLocked(Event{Type: "voice.leave", Time: now, Payload: map[string]any{"channel_id": delivery.ScopeID, "peer_id": delivery.SenderPeerID}})
	case "voice_mute":
		session := s.voiceSessionLocked(delivery.ScopeID)
		participant := session.Participants[delivery.SenderPeerID]
		participant.PeerID = delivery.SenderPeerID
		if participant.JoinedAt.IsZero() {
			participant.JoinedAt = delivery.CreatedAt
		}
		participant.Muted = delivery.Muted
		session.Participants[delivery.SenderPeerID] = participant
		s.state.Voice[delivery.ScopeID] = session
		s.emitEventLocked(Event{Type: "voice.mute", Time: now, Payload: map[string]any{"channel_id": delivery.ScopeID, "peer_id": delivery.SenderPeerID, "muted": delivery.Muted}})
	case "voice_frame":
		session := s.voiceSessionLocked(delivery.ScopeID)
		participant := session.Participants[delivery.SenderPeerID]
		participant.PeerID = delivery.SenderPeerID
		if participant.JoinedAt.IsZero() {
			participant.JoinedAt = delivery.CreatedAt
		}
		participant.LastFrameAt = delivery.CreatedAt
		session.Participants[delivery.SenderPeerID] = participant
		if session.LastFrameBy == nil {
			session.LastFrameBy = map[string]time.Time{}
		}
		session.LastFrameBy[delivery.SenderPeerID] = delivery.CreatedAt
		s.state.Voice[delivery.ScopeID] = session
		s.emitEventLocked(Event{Type: "voice.frame", Time: now, Payload: map[string]any{"channel_id": delivery.ScopeID, "peer_id": delivery.SenderPeerID, "bytes": len(delivery.Data)}})
	default:
		return fmt.Errorf("unsupported delivery kind %q", delivery.Kind)
	}
	return s.saveLocked()
}

func (s *Service) voiceSessionLocked(channelID string) VoiceSession {
	session, ok := s.state.Voice[channelID]
	if !ok {
		session = VoiceSession{ChannelID: channelID, Participants: map[string]VoiceParticipant{}, LastFrameBy: map[string]time.Time{}}
	}
	if session.Participants == nil {
		session.Participants = map[string]VoiceParticipant{}
	}
	if session.LastFrameBy == nil {
		session.LastFrameBy = map[string]time.Time{}
	}
	return session
}

func (s *Service) historyForServer(serverID string) []MessageRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.historyForServerLocked(serverID)
}

func (s *Service) historyForServerLocked(serverID string) []MessageRecord {
	messages := make([]MessageRecord, 0, s.cfg.HistoryLimit)
	for _, msg := range sortedMessages(s.state.Messages) {
		if msg.ServerID == serverID {
			messages = append(messages, msg)
		}
	}
	if len(messages) > s.cfg.HistoryLimit {
		messages = messages[len(messages)-s.cfg.HistoryLimit:]
	}
	return messages
}

func (s *Service) selfPeerInfo() PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return PeerInfo{PeerID: s.state.Identity.PeerID, Role: s.cfg.Role, Addresses: []string{s.listenAddressLocked()}, PublicKey: s.state.Identity.PublicKey}
}

func (s *Service) bootstrapTargets() []string {
	targets := append([]string{}, s.cfg.BootstrapAddrs...)
	s.mu.RLock()
	for _, peer := range s.state.KnownPeers {
		if peer.Role == RoleBootstrap {
			targets = append(targets, peer.Addresses...)
		}
	}
	s.mu.RUnlock()
	return dedupeSorted(targets)
}

func (s *Service) relayTargets() []string {
	targets := append([]string{}, s.cfg.RelayAddrs...)
	s.mu.RLock()
	for _, peer := range s.state.KnownPeers {
		if peer.Role == RoleRelay || peer.Role == RoleBootstrap {
			targets = append(targets, peer.Addresses...)
		}
	}
	s.mu.RUnlock()
	return dedupeSorted(targets)
}

func (s *Service) fetchPeerInfo(address string) (PeerRecord, error) {
	var info PeerInfo
	if err := s.getJSON(context.Background(), address, "/_xorein/meta", &info); err != nil {
		return PeerRecord{}, err
	}
	return PeerRecord{PeerID: info.PeerID, Role: info.Role, Addresses: info.Addresses, PublicKey: info.PublicKey, LastSeenAt: time.Now().UTC()}, nil
}

func (s *Service) networkMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/_xorein/meta", s.handleMeta)
	mux.HandleFunc("/_xorein/bootstrap/register", s.handleBootstrapRegister)
	mux.HandleFunc("/_xorein/bootstrap/peers", s.handleBootstrapPeers)
	mux.HandleFunc("/_xorein/manifests/publish", s.handleManifestPublish)
	mux.HandleFunc("/_xorein/manifests/", s.handleManifestGet)
	mux.HandleFunc("/_xorein/preview/", s.handlePreviewGet)
	mux.HandleFunc("/_xorein/join", s.handleJoin)
	mux.HandleFunc("/_xorein/deliver", s.handleDeliver)
	mux.HandleFunc("/_xorein/relay/store", s.handleRelayStore)
	mux.HandleFunc("/_xorein/relay/drain", s.handleRelayDrain)
	return mux
}

func (s *Service) handleMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.selfPeerInfo())
}

func (s *Service) handleBootstrapRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var peer PeerInfo
	if err := decodeJSON(r.Body, &peer); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	s.mu.Lock()
	s.upsertPeerLocked(PeerRecord{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey, Source: "bootstrap-register", LastSeenAt: time.Now().UTC()})
	_ = s.saveLocked()
	s.mu.Unlock()
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Service) handleBootstrapPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	peers := []PeerInfo{s.selfPeerInfo()}
	s.mu.RLock()
	for _, peer := range sortedPeers(s.state.KnownPeers) {
		if peer.PeerID == s.state.Identity.PeerID {
			continue
		}
		peers = append(peers, PeerInfo{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey})
	}
	s.mu.RUnlock()
	writeJSON(w, http.StatusOK, peers)
}

func (s *Service) handleManifestPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var manifest Manifest
	if err := decodeJSON(r.Body, &manifest); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	if err := manifest.Verify(); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_signature", Message: err.Error()})
		return
	}
	s.mu.Lock()
	existing, ok := s.state.Servers[manifest.ServerID]
	if ok && existing.Manifest.UpdatedAt.After(manifest.UpdatedAt) {
		s.mu.Unlock()
		writeJSON(w, http.StatusNoContent, nil)
		return
	}
	server := existing
	if server.ID == "" {
		server = ServerRecord{ID: manifest.ServerID, Name: manifest.Name, Description: manifest.Description, OwnerPeerID: manifest.OwnerPeerID, CreatedAt: manifest.IssuedAt, UpdatedAt: manifest.UpdatedAt, Members: []string{manifest.OwnerPeerID}, Channels: map[string]ChannelRecord{}}
	}
	server.Name = manifest.Name
	server.Description = manifest.Description
	server.OwnerPeerID = manifest.OwnerPeerID
	server.UpdatedAt = manifest.UpdatedAt
	server.Manifest = manifest
	s.state.Servers[server.ID] = server
	s.upsertPeerLocked(PeerRecord{PeerID: manifest.OwnerPeerID, Role: inferPeerRoleFromCapabilities(RoleClient, manifest.Capabilities), Addresses: manifest.OwnerAddresses, PublicKey: manifest.OwnerPublicKey, Source: "manifest", LastSeenAt: time.Now().UTC()})
	_ = s.saveLocked()
	s.mu.Unlock()
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Service) handleManifestGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	serverID := strings.TrimPrefix(r.URL.Path, "/_xorein/manifests/")
	s.mu.RLock()
	server, ok := s.state.Servers[serverID]
	s.mu.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "manifest not found"})
		return
	}
	writeJSON(w, http.StatusOK, server.Manifest)
}

func (s *Service) handlePreviewGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	serverID := strings.TrimPrefix(r.URL.Path, "/_xorein/preview/")
	s.mu.RLock()
	server, ok := s.state.Servers[serverID]
	if !ok {
		s.mu.RUnlock()
		writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "server preview not found"})
		return
	}
	channels := make([]ChannelRecord, 0, len(server.Channels))
	for _, channel := range server.Channels {
		channels = append(channels, channel)
	}
	memberCount := len(server.Members)
	s.mu.RUnlock()
	sort.Slice(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	writeJSON(w, http.StatusOK, ServerPreviewInfo{Manifest: server.Manifest, OwnerRole: inferPeerRoleFromCapabilities("", server.Manifest.Capabilities), Channels: channels, MemberCount: memberCount, SafetyLabels: safetyLabelsForManifest(server.Manifest)})
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req JoinRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	if err := req.Invite.Verify(); err != nil {
		code := "invalid_signature"
		if strings.Contains(err.Error(), "expired") {
			code = "expired_invite"
		}
		writeError(w, http.StatusBadRequest, APIError{Code: code, Message: err.Error()})
		return
	}
	s.mu.Lock()
	server, ok := s.state.Servers[req.Invite.ServerID]
	if !ok {
		s.mu.Unlock()
		writeError(w, http.StatusNotFound, APIError{Code: "not_found", Message: "server not found"})
		return
	}
	if server.Manifest.Hash() != req.Invite.ManifestHash {
		s.mu.Unlock()
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_manifest", Message: "invite/manifest mismatch"})
		return
	}
	if !contains(server.Members, req.Requester.PeerID) {
		server.Members = append(server.Members, req.Requester.PeerID)
		sort.Strings(server.Members)
	}
	server.UpdatedAt = time.Now().UTC()
	s.state.Servers[server.ID] = server
	s.upsertPeerLocked(PeerRecord{PeerID: req.Requester.PeerID, Role: req.Requester.Role, Addresses: req.Requester.Addresses, PublicKey: req.Requester.PublicKey, Source: "join", LastSeenAt: time.Now().UTC()})
	_ = s.saveLocked()
	history := s.historyForServerLocked(server.ID)
	s.mu.Unlock()
	channels := make([]ChannelRecord, 0, len(server.Channels))
	for _, channel := range server.Channels {
		channels = append(channels, channel)
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	resp := JoinResponse{Manifest: server.Manifest, Server: server, Channels: channels, History: history}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Service) handleDeliver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var delivery Delivery
	if err := decodeJSON(r.Body, &delivery); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	if err := s.applyDelivery(delivery); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_signature", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Service) handleRelayStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var delivery Delivery
	if err := decodeJSON(r.Body, &delivery); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	if err := delivery.Verify(); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_signature", Message: err.Error()})
		return
	}
	s.mu.Lock()
	for _, peerID := range dedupeSorted(delivery.RecipientPeerIDs) {
		if peerID == s.state.Identity.PeerID {
			continue
		}
		queue := append([]Delivery(nil), s.state.RelayQueues[peerID]...)
		queue = append(queue, delivery)
		s.state.RelayQueues[peerID] = queue
	}
	_ = s.saveLocked()
	s.mu.Unlock()
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Service) handleRelayDrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, APIError{Code: "method_not_allowed", Message: "method not allowed"})
		return
	}
	var req DrainRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeError(w, http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()})
		return
	}
	s.mu.Lock()
	queue := append([]Delivery(nil), s.state.RelayQueues[req.PeerID]...)
	delete(s.state.RelayQueues, req.PeerID)
	_ = s.saveLocked()
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, queue)
}

func (s *Service) getJSON(ctx context.Context, address, path string, out any) error {
	url := s.normalizedURL(address) + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return decodeJSON(resp.Body, out)
}

func (s *Service) postJSON(ctx context.Context, address, path string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := s.normalizedURL(address) + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return decodeJSON(resp.Body, out)
}

func (s *Service) httpClient() *http.Client {
	return &http.Client{Timeout: 1200 * time.Millisecond}
}

func (s *Service) normalizedURL(address string) string {
	address = strings.TrimSpace(address)
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return strings.TrimRight(address, "/")
	}
	return "http://" + strings.TrimRight(address, "/")
}

func (s *Service) saveLocked() error {
	return saveState(s.cfg.DataDir, s.state)
}

func (s *Service) listenAddressLocked() string {
	if s.listen == nil {
		if addr := strings.TrimSpace(s.cfg.ListenAddr); addr != "" {
			return addr
		}
		return ""
	}
	return s.listen.Addr().String()
}

func (s *Service) upsertPeerLocked(peer PeerRecord) {
	if peer.PeerID == "" {
		return
	}
	existing, ok := s.state.KnownPeers[peer.PeerID]
	if !ok {
		existing = PeerRecord{PeerID: peer.PeerID}
	}
	if peer.Role.Valid() {
		existing.Role = peer.Role
	}
	if peer.PublicKey != "" {
		existing.PublicKey = peer.PublicKey
	}
	existing.Source = peer.Source
	existing.LastSeenAt = peer.LastSeenAt
	existing.Addresses = dedupeSorted(append(existing.Addresses, peer.Addresses...))
	s.state.KnownPeers[peer.PeerID] = existing
}

func (s *Service) recordTelemetry(entry string) {
	s.mu.Lock()
	s.state.Telemetry = append(s.state.Telemetry, fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339Nano), entry))
	if len(s.state.Telemetry) > 256 {
		s.state.Telemetry = append([]string(nil), s.state.Telemetry[len(s.state.Telemetry)-256:]...)
	}
	_ = s.saveLocked()
	s.mu.Unlock()
}

func (s *Service) emitEvent(event Event) {
	s.eventsMu.RLock()
	defer s.eventsMu.RUnlock()
	for ch := range s.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *Service) emitEventLocked(event Event) {
	go s.emitEvent(event)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	if payload == nil {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err APIError) {
	writeJSON(w, status, err)
}

func decodeJSON(r io.Reader, out any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func parseAPIError(resp *http.Response) error {
	var apiErr APIError
	if err := decodeJSON(resp.Body, &apiErr); err == nil && apiErr.Code != "" {
		return fmt.Errorf("%s: %s", apiErr.Code, apiErr.Message)
	}
	return fmt.Errorf("http %d", resp.StatusCode)
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}

func advertisedCapabilities(role Role) []string {
	capabilities := make([]string, 0, len(protocol.DefaultFeatureFlags())+1)
	for _, flag := range protocol.DefaultFeatureFlags() {
		capabilities = append(capabilities, string(flag))
	}
	if role == RoleArchivist {
		capabilities = append(capabilities, string(protocol.FeatureArchivist))
	}
	return dedupeSorted(capabilities)
}

func notificationRecordForMessage(identity Identity, settings map[string]string, msg MessageRecord) (NotificationRecord, bool) {
	if msg.SenderPeerID == identity.PeerID {
		return NotificationRecord{}, false
	}
	found := findMentionTokens(msg.Body, selfMentionTokens(identity))
	if len(found) == 0 {
		return NotificationRecord{}, false
	}
	createdAt := messageSortTime(msg)
	readThrough := notificationReadThroughForMessageLocked(settings, msg)
	unread := readThrough.IsZero() || createdAt.After(readThrough)
	return NotificationRecord{Kind: "mention", Message: msg, Tokens: found, Unread: unread, CreatedAt: createdAt}, true
}

func (s *Service) emitNotificationCreatedLocked(note NotificationRecord, now time.Time) {
	serverName, scopeName := s.notificationSummaryLabelsLocked(note.Message)
	payload := map[string]any{
		"kind":           note.Kind,
		"message_id":     note.Message.ID,
		"sender_peer_id": note.Message.SenderPeerID,
		"scope_type":     note.Message.ScopeType,
		"scope_id":       note.Message.ScopeID,
		"server_id":      note.Message.ServerID,
		"tokens":         append([]string(nil), note.Tokens...),
		"unread":         note.Unread,
	}
	if serverName != "" {
		payload["server_name"] = serverName
	}
	if scopeName != "" {
		payload["scope_name"] = scopeName
	}
	if note.Message.ScopeType == "dm" {
		if participantIDs := s.notificationSummaryDirectParticipantIDsLocked(note.Message.ScopeID); len(participantIDs) > 0 {
			payload["participant_ids"] = participantIDs
		}
	}
	s.emitEventLocked(Event{Type: "notification.created", Time: now, Payload: payload})
}

func shouldAdvanceNotificationSummaryMessage(currentAt time.Time, currentID string, candidateAt time.Time, candidateID string) bool {
	if currentID == "" {
		return true
	}
	if candidateAt.After(currentAt) {
		return true
	}
	if candidateAt.Equal(currentAt) && candidateID > currentID {
		return true
	}
	return false
}

func (s *Service) notificationSummaryDirectParticipantIDsLocked(scopeID string) []string {
	if dm, ok := s.state.DMs[scopeID]; ok {
		others := make([]string, 0, len(dm.Participants))
		for _, peerID := range dm.Participants {
			if peerID == s.state.Identity.PeerID {
				continue
			}
			others = append(others, peerID)
		}
		sort.Strings(others)
		return others
	}
	return nil
}

func (s *Service) notificationSummaryLabelsLocked(msg MessageRecord) (serverName, scopeName string) {
	if msg.ServerID != "" {
		if server, ok := s.state.Servers[msg.ServerID]; ok {
			serverName = strings.TrimSpace(server.Name)
			if msg.ScopeType == "channel" {
				if channel, ok := server.Channels[msg.ScopeID]; ok {
					scopeName = strings.TrimSpace(channel.Name)
				}
			}
		}
	}
	if msg.ScopeType == "dm" {
		if dm, ok := s.state.DMs[msg.ScopeID]; ok {
			others := make([]string, 0, len(dm.Participants))
			for _, peerID := range dm.Participants {
				if peerID == s.state.Identity.PeerID {
					continue
				}
				others = append(others, peerID)
			}
			sort.Strings(others)
			scopeName = strings.Join(others, ", ")
		}
	}
	return serverName, scopeName
}

func notificationReadThroughForMessageLocked(settings map[string]string, msg MessageRecord) time.Time {
	return notificationReadThroughForScopeLocked(settings, msg.ServerID, msg.ScopeType, msg.ScopeID)
}

func notificationReadThroughForScopeLocked(settings map[string]string, serverID, scopeType, scopeID string) time.Time {
	through := notificationsReadThroughLocked(settings)
	if serverID != "" {
		serverThrough := notificationsServerReadThroughLocked(settings, serverID)
		if serverThrough.After(through) {
			through = serverThrough
		}
	}
	if scopeType == "" || scopeID == "" {
		return through
	}
	scoped := notificationsScopedReadThroughLocked(settings, serverID, scopeType, scopeID)
	if scoped.After(through) {
		through = scoped
	}
	return through
}

func notificationsServerReadThroughLocked(settings map[string]string, serverID string) time.Time {
	return notificationsReadThroughValue(strings.TrimSpace(settings[notificationsServerReadThroughKey(serverID)]))
}

func notificationsServerReadThroughKey(serverID string) string {
	return notificationsServerReadThroughPrefix + strings.TrimSpace(serverID)
}

func notificationsScopedReadThroughLocked(settings map[string]string, serverID, scopeType, scopeID string) time.Time {
	return notificationsReadThroughValue(strings.TrimSpace(settings[notificationsScopedReadThroughKey(serverID, scopeType, scopeID)]))
}

func notificationsScopedReadThroughKey(serverID, scopeType, scopeID string) string {
	return notificationsScopedReadThroughPrefix + strings.TrimSpace(serverID) + "|" + strings.TrimSpace(scopeType) + "|" + strings.TrimSpace(scopeID)
}

func normalizeNotificationScope(serverID, scopeType, scopeID string) (string, string, string, error) {
	serverID = strings.TrimSpace(serverID)
	scopeType = strings.TrimSpace(scopeType)
	scopeID = strings.TrimSpace(scopeID)
	if scopeType != "" && scopeType != "channel" && scopeType != "dm" {
		return "", "", "", errors.New("unsupported scope type")
	}
	if scopeID != "" && scopeType == "" {
		return "", "", "", errors.New("scope type is required when scope id is set")
	}
	if scopeType == "dm" && serverID != "" {
		return "", "", "", errors.New("server id is not valid for dm scope")
	}
	if scopeType != "" && scopeID == "" {
		return "", "", "", errors.New("scope id is required when scope type is set")
	}
	return serverID, scopeType, scopeID, nil
}

func notificationsReadThroughLocked(settings map[string]string) time.Time {
	if settings == nil {
		return time.Time{}
	}
	return notificationsReadThroughValue(strings.TrimSpace(settings[notificationsReadThroughSetting]))
}

func notificationsReadThroughValue(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func selfMentionTokens(identity Identity) []string {
	tokens := []string{"@" + strings.ToLower(identity.PeerID)}
	displayName := strings.TrimSpace(identity.Profile.DisplayName)
	if displayName != "" && !strings.ContainsAny(displayName, " \t\r\n") {
		tokens = append(tokens, "@"+strings.ToLower(displayName))
	}
	return dedupeSorted(tokens)
}

func findMentionTokens(body string, tokens []string) []string {
	if len(tokens) == 0 {
		return nil
	}
	lower := strings.ToLower(body)
	found := make([]string, 0, len(tokens))
	for _, token := range dedupeSorted(tokens) {
		if token != "" && strings.Contains(lower, token) {
			found = append(found, token)
		}
	}
	return found
}

func presenceStatusAt(now time.Time, lastSeen time.Time, inVoice bool) string {
	if inVoice {
		return "online"
	}
	if lastSeen.IsZero() {
		return "offline"
	}
	age := now.Sub(lastSeen)
	if age <= 2*time.Minute {
		return "online"
	}
	if age <= 15*time.Minute {
		return "recent"
	}
	return "offline"
}

func messageSortTime(msg MessageRecord) time.Time {
	if !msg.UpdatedAt.IsZero() {
		return msg.UpdatedAt
	}
	return msg.CreatedAt
}

func safetyLabelsForManifest(manifest Manifest) []string {
	labels := map[string]struct{}{}
	if manifest.HistoryCoverage == HistoryCoverageLocalWindow {
		labels["history-local-window"] = struct{}{}
	}
	if manifest.HistoryDurability == HistoryDurabilitySingleNode {
		labels["history-single-node"] = struct{}{}
	}
	for _, capability := range manifest.Capabilities {
		if strings.TrimSpace(capability) == string(protocol.FeatureArchivist) {
			labels["owner-archivist"] = struct{}{}
			break
		}
	}
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(labels))
	for label := range labels {
		out = append(out, label)
	}
	sort.Strings(out)
	return out
}

func inferPeerRoleFromCapabilities(fallback Role, capabilities []string) Role {
	for _, capability := range capabilities {
		if strings.TrimSpace(capability) == string(protocol.FeatureArchivist) {
			return RoleArchivist
		}
	}
	return fallback
}

func (s *Service) pruneAllServerHistoryLocked() bool {
	pruned := false
	for serverID := range s.state.Servers {
		if s.pruneServerHistoryLocked(serverID) {
			pruned = true
		}
	}
	return pruned
}

func (s *Service) pruneServerHistoryLocked(serverID string) bool {
	if serverID == "" || s.cfg.HistoryLimit <= 0 {
		return false
	}
	messages := make([]MessageRecord, 0, s.cfg.HistoryLimit+1)
	for _, msg := range s.state.Messages {
		if msg.ServerID == serverID {
			messages = append(messages, msg)
		}
	}
	if len(messages) <= s.cfg.HistoryLimit {
		return false
	}
	sort.Slice(messages, func(i, j int) bool {
		if !messages[i].CreatedAt.Equal(messages[j].CreatedAt) {
			return messages[i].CreatedAt.Before(messages[j].CreatedAt)
		}
		return messages[i].ID < messages[j].ID
	})
	for _, msg := range messages[:len(messages)-s.cfg.HistoryLimit] {
		delete(s.state.Messages, msg.ID)
	}
	return true
}

func sameStringSet(left, right []string) bool {
	left = dedupeSorted(left)
	right = dedupeSorted(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func controlTokenPath(dataDir string) string {
	return filepath.Join(dataDir, "control.token")
}

func ensureControlTokenFile(dataDir, token string) error {
	if err := os.WriteFile(controlTokenPath(dataDir), []byte(token), 0o600); err != nil {
		return fmt.Errorf("write control token: %w", err)
	}
	return nil
}
