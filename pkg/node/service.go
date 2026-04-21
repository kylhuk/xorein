package node

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
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

	"github.com/aether/code_aether/pkg/network"
	"github.com/aether/code_aether/pkg/protocol"
	"github.com/aether/code_aether/pkg/storage"
)

const (
	notificationsReadThroughSetting      = "notifications_read_through"
	notificationsServerReadThroughPrefix = "notifications_read_through_server:"
	notificationsScopedReadThroughPrefix = "notifications_read_through_scope:"
	relayQueueTTL                        = 24 * time.Hour
	relayQueueLimit                      = 256
)

var (
	livePeerRegistryMu sync.RWMutex
	livePeerRegistry   = map[string][]string{}
)

func registerLivePeer(peerID, addr string) {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return
	}
	normalized := normalizePeerAddresses([]string{addr})
	if len(normalized) == 0 {
		return
	}
	addr = normalized[0]
	livePeerRegistryMu.Lock()
	livePeerRegistry[peerID] = dedupeSorted(append(append([]string(nil), livePeerRegistry[peerID]...), addr))
	livePeerRegistryMu.Unlock()
}

func unregisterLivePeer(peerID, addr string) {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return
	}
	normalized := normalizePeerAddresses([]string{addr})
	if len(normalized) == 0 {
		return
	}
	addr = normalized[0]
	livePeerRegistryMu.Lock()
	defer livePeerRegistryMu.Unlock()
	current := append([]string(nil), livePeerRegistry[peerID]...)
	filtered := current[:0]
	for _, existing := range current {
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

	discoveryMu      sync.Mutex
	discoveryBackoff map[string]discoveryBackoffState

	controlListener net.Listener
	controlServer   *http.Server
	peerRuntime     network.Runtime

	eventsMu    sync.RWMutex
	subscribers map[chan Event]struct{}

	closed chan struct{}
	once   sync.Once
	wg     sync.WaitGroup
}

type discoveryBackoffState struct {
	Failures    int
	NextAttempt time.Time
}

type Option func(*Service)

func WithPeerRuntime(runtime network.Runtime) Option {
	return func(s *Service) {
		s.peerRuntime = runtime
	}
}

func NewService(cfg Config, opts ...Option) (*Service, error) {
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
	service := &Service{
		cfg:              cfg,
		state:            state,
		discoveryBackoff: map[string]discoveryBackoffState{},
		subscribers:      map[chan Event]struct{}{},
		closed:           make(chan struct{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(service)
		}
	}
	return service, nil
}

func (s *Service) Start(ctx context.Context) error {
	if s.peerRuntime == nil {
		return errors.New("peer runtime is required")
	}
	if err := s.peerRuntime.Start(ctx); err != nil {
		return err
	}
	listenAddr := s.peerRuntime.ListenAddress()

	controlListener, controlEndpoint, err := createControlListener(s.cfg.ControlEndpoint, s.cfg.DataDir)
	if err != nil {
		_ = s.peerRuntime.Close()
		return err
	}
	s.controlListener = controlListener
	s.controlServer = &http.Server{Handler: s.controlMux()}
	s.mu.Lock()
	now := time.Now().UTC()
	s.state.Settings["control_endpoint"] = controlEndpoint
	s.state.Settings["listen_address"] = listenAddr
	s.upsertPeerLocked(PeerRecord{
		PeerID:     s.state.Identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{listenAddr},
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
		_ = s.peerRuntime.Close()
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
	if err := s.runDiscoveryPass(); err != nil {
		s.recordTelemetry("discovery.pass error=" + err.Error())
	}
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
		if s.peerRuntime != nil {
			if err := s.peerRuntime.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				retErr = err
			}
		}
		if s.controlListener != nil {
			_ = s.controlListener.Close()
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
	if s.peerRuntime != nil {
		return s.peerRuntime.ListenAddress()
	}
	return strings.TrimSpace(s.cfg.ListenAddr)
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
	now := time.Now().UTC()
	s.mu.Lock()
	oldPeerID := s.state.Identity.PeerID
	listenAddr := s.listenAddressLocked()
	refreshedManifests, err := s.replaceIdentityLocked(identity, now)
	s.mu.Unlock()
	if err != nil {
		return Identity{}, err
	}
	unregisterLivePeer(oldPeerID, listenAddr)
	registerLivePeer(identity.PeerID, listenAddr)
	for _, manifest := range refreshedManifests {
		_ = s.publishManifest(manifest)
	}
	s.recordTelemetry("identity.created peer=" + identity.PeerID)
	s.emitEvent(Event{Type: "identity.created", Time: now, Payload: map[string]any{"peer_id": identity.PeerID}})
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
	now := time.Now().UTC()
	s.mu.Lock()
	oldPeerID := s.state.Identity.PeerID
	listenAddr := s.listenAddressLocked()
	refreshedManifests, err := s.replaceIdentityLocked(identity, now)
	s.mu.Unlock()
	if err != nil {
		return Identity{}, err
	}
	unregisterLivePeer(oldPeerID, listenAddr)
	registerLivePeer(identity.PeerID, listenAddr)
	for _, manifest := range refreshedManifests {
		_ = s.publishManifest(manifest)
	}
	s.recordTelemetry("identity.restored peer=" + identity.PeerID)
	s.emitEvent(Event{Type: "identity.restored", Time: now, Payload: map[string]any{"peer_id": identity.PeerID}})
	return identity, nil
}

func (s *Service) replaceIdentityLocked(identity Identity, now time.Time) ([]Manifest, error) {
	prior := clonePersistedState(s.state)
	oldPeerID := s.state.Identity.PeerID
	s.state.Identity = identity
	if oldPeerID != "" && oldPeerID != identity.PeerID {
		delete(s.state.KnownPeers, oldPeerID)
	}
	s.rebindLocalStateLocked(oldPeerID, identity.PeerID)
	s.upsertPeerLocked(PeerRecord{
		PeerID:     identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{s.listenAddressLocked()},
		PublicKey:  identity.PublicKey,
		Source:     "self",
		LastSeenAt: now,
	})
	refreshedManifests, err := s.refreshOwnedServerBindingsLocked(now)
	if err != nil {
		s.state = prior
		return nil, err
	}
	if err := s.saveLockedWithRollback(prior); err != nil {
		return nil, err
	}
	return refreshedManifests, nil
}

func (s *Service) rebindLocalStateLocked(oldPeerID, newPeerID string) {
	if strings.TrimSpace(oldPeerID) == "" || strings.TrimSpace(newPeerID) == "" || oldPeerID == newPeerID {
		return
	}
	for serverID, server := range s.state.Servers {
		changed := false
		if server.OwnerPeerID == oldPeerID {
			server.OwnerPeerID = newPeerID
			changed = true
		}
		updatedMembers, membersChanged := replacePeerID(server.Members, oldPeerID, newPeerID)
		if membersChanged {
			server.Members = updatedMembers
			changed = true
		}
		if changed {
			s.state.Servers[serverID] = server
		}
	}
	for dmID, dm := range s.state.DMs {
		updatedParticipants, changed := replacePeerID(dm.Participants, oldPeerID, newPeerID)
		if !changed {
			continue
		}
		dm.Participants = updatedParticipants
		s.state.DMs[dmID] = dm
	}
	for channelID, session := range s.state.Voice {
		changed := false
		if participant, ok := session.Participants[oldPeerID]; ok {
			delete(session.Participants, oldPeerID)
			participant.PeerID = newPeerID
			session.Participants[newPeerID] = participant
			changed = true
		}
		if frameAt, ok := session.LastFrameBy[oldPeerID]; ok {
			delete(session.LastFrameBy, oldPeerID)
			session.LastFrameBy[newPeerID] = frameAt
			changed = true
		}
		if changed {
			s.state.Voice[channelID] = session
		}
	}
}

func replacePeerID(values []string, oldPeerID, newPeerID string) ([]string, bool) {
	if oldPeerID == newPeerID {
		return append([]string(nil), values...), false
	}
	updated := make([]string, 0, len(values))
	changed := false
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == oldPeerID {
			value = newPeerID
			changed = true
		}
		if value == "" || contains(updated, value) {
			continue
		}
		updated = append(updated, value)
	}
	return updated, changed
}

func (s *Service) AddManualPeer(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New("manual peer address is required")
	}
	peer, err := s.fetchPeerInfo("manual", address)
	if err != nil {
		return err
	}
	s.mu.Lock()
	prior := clonePersistedState(s.state)
	peer.Source = "manual"
	peer.LastSeenAt = time.Now().UTC()
	s.upsertPeerLocked(peer)
	err = s.saveLockedWithRollback(prior)
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
	prior := clonePersistedState(s.state)
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
	err := s.saveLockedWithRollback(prior)
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
		desiredHistoryRetention := s.cfg.HistoryLimit
		desiredHistoryCoverage := HistoryCoverageLocalWindow
		desiredHistoryDurability := HistoryDurabilitySingleNode

		manifestChanged := manifest.ServerID != server.ID ||
			manifest.Name != server.Name ||
			manifest.Description != server.Description ||
			manifest.OwnerPeerID != identity.PeerID ||
			manifest.OwnerPublicKey != identity.PublicKey ||
			!sameStringSet(manifest.OwnerAddresses, desiredOwnerAddresses) ||
			!sameStringSet(manifest.BootstrapAddrs, desiredBootstrapAddrs) ||
			!sameStringSet(manifest.RelayAddrs, desiredRelayAddrs) ||
			!sameStringSet(manifest.Capabilities, desiredCapabilities) ||
			manifest.HistoryRetentionMessages != desiredHistoryRetention ||
			manifest.HistoryCoverage != desiredHistoryCoverage ||
			manifest.HistoryDurability != desiredHistoryDurability ||
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
			manifest.HistoryRetentionMessages = desiredHistoryRetention
			manifest.HistoryCoverage = desiredHistoryCoverage
			manifest.HistoryDurability = desiredHistoryDurability
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
	prior := clonePersistedState(s.state)
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
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
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
	if err := s.saveLockedWithRollback(prior); err != nil {
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
		if err := validateInviteManifest(invite, previewInfo.Manifest); err != nil {
			return ServerPreview{}, err
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
	if err := validateInviteManifest(invite, manifest); err != nil {
		return ServerPreview{}, err
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
	if err := validateInviteManifest(invite, manifest); err != nil {
		return ServerRecord{}, err
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
	addresses := normalizePeerAddresses(invite.ServerAddrs)
	var joinResp JoinResponse
	var lastErr error
	for _, addr := range addresses {
		if err := s.peerCall(context.Background(), addr, network.OperationJoin, joinReq, &joinResp); err != nil {
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
	if err := validateInviteManifest(invite, joinResp.Manifest); err != nil {
		return ServerRecord{}, err
	}
	if joinResp.Server.ID != invite.ServerID {
		return ServerRecord{}, errors.New("join response server mismatch")
	}
	s.mu.Lock()
	prior := clonePersistedState(s.state)
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
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
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
		s.state = prior
		s.mu.Unlock()
		return ChannelRecord{}, err
	}
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
	if dm, ok := s.state.DMs[dmID]; ok {
		s.mu.Unlock()
		return dm, nil
	}
	dm := DMRecord{ID: dmID, Participants: participants, CreatedAt: now}
	s.state.DMs[dm.ID] = dm
	identity := s.state.Identity
	delivery := Delivery{ID: dm.ID, Kind: "dm_create", ScopeID: dm.ID, ScopeType: "dm", SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: dedupeSorted([]string{peerID}), CreatedAt: now}
	if err := delivery.Sign(identity); err != nil {
		s.state = prior
		s.mu.Unlock()
		return DMRecord{}, err
	}
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
	s.state.Messages[msg.ID] = msg
	s.state.Deliveries[delivery.ID] = struct{}{}
	pruned := s.pruneServerHistoryLocked(serverID)
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
	msg.Body = body
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	scopeType, scopeID, serverID := msg.ScopeType, msg.ScopeID, msg.ServerID
	recipients, _, err := s.scopeRecipientsLocked(scopeType, scopeID)
	if err != nil {
		s.state = prior
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	delivery := Delivery{ID: messageID, Kind: "message_edit", ScopeID: scopeID, ScopeType: scopeType, ServerID: serverID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: recipients, Body: body, CreatedAt: msg.UpdatedAt}
	if err := delivery.Sign(identity); err != nil {
		s.state = prior
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	if err := s.saveLockedWithRollback(prior); err != nil {
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
	prior := clonePersistedState(s.state)
	msg.Body = ""
	msg.Deleted = true
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	recipients, _, err := s.scopeRecipientsLocked(msg.ScopeType, msg.ScopeID)
	if err != nil {
		s.state = prior
		s.mu.Unlock()
		return err
	}
	delivery := Delivery{ID: messageID, Kind: "message_delete", ScopeID: msg.ScopeID, ScopeType: msg.ScopeType, ServerID: msg.ServerID, SenderPeerID: identity.PeerID, SenderPublicKey: identity.PublicKey, RecipientPeerIDs: recipients, CreatedAt: msg.UpdatedAt}
	if err := delivery.Sign(identity); err != nil {
		s.state = prior
		s.mu.Unlock()
		return err
	}
	if err := s.saveLockedWithRollback(prior); err != nil {
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

func (s *Service) discoveryLoop() {
	ticker := time.NewTicker(s.cfg.DiscoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			if err := s.runDiscoveryPass(); err != nil {
				s.recordTelemetry("discovery.pass error=" + err.Error())
			}
		}
	}
}

func (s *Service) runDiscoveryPass() error {
	var errs []error
	if err := s.discoverFromKnownPeerCache(); err != nil {
		errs = append(errs, err)
	}
	if err := s.discoverLANPeers(); err != nil {
		errs = append(errs, err)
	}
	if err := s.registerWithBootstraps(); err != nil {
		errs = append(errs, err)
	}
	if err := s.fetchPeersFromBootstraps(); err != nil {
		errs = append(errs, err)
	}
	if err := s.walkDiscoveryGraph(); err != nil {
		errs = append(errs, err)
	}
	if err := s.discoverRendezvousPeers(); err != nil {
		errs = append(errs, err)
	}
	if err := s.exchangePeers(); err != nil {
		errs = append(errs, err)
	}
	if err := s.fetchConfiguredManualPeers(); err != nil {
		errs = append(errs, err)
	}
	if err := s.drainRelays(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *Service) discoverFromKnownPeerCache() error {
	return s.pingKnownPeers("cache", s.Snapshot().KnownPeers)
}

func (s *Service) discoverLANPeers() error {
	local := []PeerRecord{}
	for peerID, addresses := range livePeerRegistrySnapshot() {
		if peerID == s.PeerID() {
			continue
		}
		local = append(local, PeerRecord{PeerID: peerID, Addresses: normalizePeerAddresses(addresses), Source: "lan"})
	}
	return s.pingKnownPeers("lan", local)
}

func (s *Service) fetchConfiguredManualPeers() error {
	var firstErr error
	for _, addr := range normalizePeerAddresses(s.cfg.ManualPeers) {
		peer, err := s.fetchPeerInfo("manual", addr)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.recordTelemetry("discovery.manual.failed addr=" + addr + " err=" + err.Error())
			continue
		}
		s.mu.Lock()
		priorPeers := clonePeerMap(s.state.KnownPeers)
		peer.Source = "manual"
		peer.LastSeenAt = time.Now().UTC()
		s.upsertPeerLocked(peer)
		err = s.saveKnownPeersLocked(priorPeers)
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
		if !s.shouldAttemptDiscovery("bootstrap-register", addr) {
			continue
		}
		if err := s.peerCall(context.Background(), addr, network.OperationBootstrapRegister, self, nil); err != nil {
			s.markDiscoveryFailure("bootstrap-register", addr, err)
			if firstErr == nil {
				firstErr = err
			}
			s.recordTelemetry("discovery.bootstrap.register.failed addr=" + addr + " err=" + err.Error())
			continue
		}
		s.markDiscoverySuccess("bootstrap-register", addr)
		s.recordTelemetry("discovery.bootstrap.registered addr=" + addr)
	}
	return firstErr
}

func (s *Service) fetchPeersFromBootstraps() error {
	var firstErr error
	for _, addr := range s.bootstrapTargets() {
		if !s.shouldAttemptDiscovery("bootstrap-peers", addr) {
			continue
		}
		var peers []PeerInfo
		if err := s.peerCall(context.Background(), addr, network.OperationBootstrapPeers, nil, &peers); err != nil {
			s.markDiscoveryFailure("bootstrap-peers", addr, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.markDiscoverySuccess("bootstrap-peers", addr)
		selfPeerID := s.PeerID()
		for _, peer := range peers {
			if peer.PeerID == selfPeerID {
				continue
			}
			s.mu.Lock()
			priorPeers := clonePeerMap(s.state.KnownPeers)
			s.upsertPeerLocked(PeerRecord{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey, Source: "bootstrap", LastSeenAt: time.Now().UTC()})
			if err := s.saveKnownPeersLocked(priorPeers); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				s.mu.Unlock()
				continue
			}
			s.mu.Unlock()
		}
	}
	return firstErr
}

func (s *Service) walkDiscoveryGraph() error {
	seedTargets := append([]PeerRecord(nil), s.Snapshot().KnownPeers...)
	for _, addr := range s.bootstrapTargets() {
		seedTargets = append(seedTargets, PeerRecord{Addresses: []string{addr}, Source: "dht"})
	}
	return s.exchangePeersWithTargets("dht", seedTargets, PeerExchangeRequest{Limit: 32})
}

func (s *Service) discoverRendezvousPeers() error {
	serverIDs := s.knownServerIDs()
	if len(serverIDs) == 0 {
		return nil
	}
	return s.exchangePeersWithTargets("rendezvous", s.Snapshot().KnownPeers, PeerExchangeRequest{ServerIDs: serverIDs, Limit: 32})
}

func (s *Service) exchangePeers() error {
	knownPeerIDs := []string{s.PeerID()}
	for _, peer := range s.Snapshot().KnownPeers {
		knownPeerIDs = append(knownPeerIDs, peer.PeerID)
	}
	return s.exchangePeersWithTargets("pex", s.Snapshot().KnownPeers, PeerExchangeRequest{KnownPeerIDs: dedupeSorted(knownPeerIDs), Limit: 64})
}

func (s *Service) pingKnownPeers(source string, peers []PeerRecord) error {
	var firstErr error
	for _, peer := range peers {
		if peer.PeerID == s.PeerID() {
			continue
		}
		for _, addr := range dedupeSorted(peer.Addresses) {
			if !s.shouldAttemptDiscovery("peer-info", addr) {
				continue
			}
			info, err := s.fetchPeerInfo(source, addr)
			if err != nil {
				s.markDiscoveryFailure("peer-info", addr, err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			s.markDiscoverySuccess("peer-info", addr)
			s.mu.Lock()
			priorPeers := clonePeerMap(s.state.KnownPeers)
			if info.Source == "" {
				info.Source = peer.Source
			}
			if info.Source == "" {
				info.Source = source
			}
			info.LastSeenAt = time.Now().UTC()
			s.upsertPeerLocked(info)
			if err := s.saveKnownPeersLocked(priorPeers); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				s.mu.Unlock()
				continue
			}
			s.mu.Unlock()
			break
		}
	}
	return firstErr
}

func (s *Service) exchangePeersWithTargets(layer string, peers []PeerRecord, request PeerExchangeRequest) error {
	var firstErr error
	for _, peer := range peers {
		if peer.PeerID == s.PeerID() {
			continue
		}
		for _, addr := range dedupeSorted(peer.Addresses) {
			if !s.shouldAttemptDiscovery(layer, addr) {
				continue
			}
			var discovered []PeerInfo
			if err := s.peerCall(context.Background(), addr, network.OperationPeerExchange, request, &discovered); err != nil {
				s.markDiscoveryFailure(layer, addr, err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			s.markDiscoverySuccess(layer, addr)
			if err := s.mergeDiscoveredPeers(layer, discovered); err != nil && firstErr == nil {
				firstErr = err
			}
			break
		}
	}
	return firstErr
}

func (s *Service) mergeDiscoveredPeers(source string, peers []PeerInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	priorPeers := clonePeerMap(s.state.KnownPeers)
	updated := false
	for _, peer := range peers {
		if peer.PeerID == "" || peer.PeerID == s.state.Identity.PeerID {
			continue
		}
		s.upsertPeerLocked(PeerRecord{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey, Source: source, LastSeenAt: time.Now().UTC()})
		updated = true
	}
	if !updated {
		return nil
	}
	return s.saveKnownPeersLocked(priorPeers)
}

func (s *Service) drainRelays() error {
	var firstErr error
	req, err := signedDrainRequest(s.Snapshot().Identity, s.cfg.Role, s.ListenAddress())
	if err != nil {
		return err
	}
	for _, addr := range s.relayTargets() {
		var queued []Delivery
		if err := s.peerCall(context.Background(), addr, network.OperationRelayDrain, req, &queued); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		for _, delivery := range queued {
			if err := s.applyDelivery(delivery); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if len(queued) > 0 {
			s.recordTelemetry(fmt.Sprintf("relay.drain addr=%s count=%d", addr, len(queued)))
		}
	}
	return firstErr
}

func (s *Service) publishManifest(manifest Manifest) error {
	var firstErr error
	for _, addr := range s.bootstrapTargets() {
		if err := s.peerCall(context.Background(), addr, network.OperationManifestPublish, manifest, nil); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (s *Service) resolveManifest(serverID string, bootstrapAddrs, serverAddrs []string) (Manifest, error) {
	targets := normalizePeerAddresses(bootstrapAddrs)
	targets = append(targets, normalizePeerAddresses(serverAddrs)...)
	targets = dedupeSorted(targets)
	var lastErr error
	for _, addr := range targets {
		var manifest Manifest
		if err := s.peerCall(context.Background(), addr, network.OperationManifestFetch, map[string]string{"server_id": serverID}, &manifest); err != nil {
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
	targets := normalizePeerAddresses(serverAddrs)
	var lastErr error
	for _, addr := range targets {
		var preview ServerPreviewInfo
		if err := s.peerCall(context.Background(), addr, network.OperationPreviewFetch, map[string]string{"server_id": serverID}, &preview); err != nil {
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
		addresses := normalizePeerAddresses(peer.Addresses)
		for _, addr := range addresses {
			if err := s.peerCall(context.Background(), addr, network.OperationDeliver, delivery, nil); err == nil {
				s.recordTelemetry("delivery.direct peer=" + peerID + " addr=" + addr)
				return nil
			} else {
				s.recordTelemetry("delivery.direct.failed peer=" + peerID + " addr=" + addr + " err=" + err.Error())
			}
		}
	}
	for _, addr := range normalizePeerAddresses(livePeerAddresses(peerID)) {
		if err := s.peerCall(context.Background(), addr, network.OperationDeliver, delivery, nil); err == nil {
			s.recordTelemetry("delivery.local peer=" + peerID + " addr=" + addr)
			return nil
		} else {
			s.recordTelemetry("delivery.local.failed peer=" + peerID + " addr=" + addr + " err=" + err.Error())
		}
	}
	for _, addr := range s.relayTargets() {
		if err := s.peerCall(context.Background(), addr, network.OperationRelayStore, delivery, nil); err == nil {
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
	if err := s.validateDeliveryLocked(delivery); err != nil {
		return err
	}
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
	targets := normalizePeerAddresses(s.cfg.BootstrapAddrs)
	s.mu.RLock()
	for _, peer := range s.state.KnownPeers {
		if peer.Role == RoleBootstrap {
			targets = append(targets, normalizePeerAddresses(peer.Addresses)...)
		}
	}
	s.mu.RUnlock()
	return dedupeSorted(targets)
}

func (s *Service) relayTargets() []string {
	targets := normalizePeerAddresses(s.cfg.RelayAddrs)
	s.mu.RLock()
	for _, peer := range s.state.KnownPeers {
		if peer.Role == RoleRelay || peer.Role == RoleBootstrap {
			targets = append(targets, normalizePeerAddresses(peer.Addresses)...)
		}
	}
	s.mu.RUnlock()
	return dedupeSorted(targets)
}

func (s *Service) peerClient() network.Client {
	return network.NewClient(1200 * time.Millisecond)
}

func (s *Service) peerCall(ctx context.Context, address string, operation network.Operation, body any, out any) error {
	_, err := s.peerClient().Call(ctx, address, operation, body, out)
	return err
}

func internalNetworkError(err error) *network.Error {
	if err == nil {
		return nil
	}
	return &network.Error{Code: "internal_error", Message: err.Error()}
}

func drainRequestPayload(requester PeerInfo) []byte {
	return []byte(requester.PeerID + "\n" + requester.PublicKey)
}

func signedDrainRequest(identity Identity, role Role, listenAddr string) (DrainRequest, error) {
	requester := PeerInfo{
		PeerID:    identity.PeerID,
		Role:      role,
		Addresses: dedupeSorted([]string{strings.TrimSpace(listenAddr)}),
		PublicKey: identity.PublicKey,
	}
	signature, err := identity.SignCanonical(drainRequestPayload(requester))
	if err != nil {
		return DrainRequest{}, err
	}
	return DrainRequest{Requester: requester, Signature: signature}, nil
}

func validateDrainRequest(req DrainRequest) error {
	requester := req.Requester
	if strings.TrimSpace(requester.PeerID) == "" || strings.TrimSpace(requester.PublicKey) == "" {
		return errors.New("requester peer info is incomplete")
	}
	if derived := derivePeerIDFromPublicKey(requester.PublicKey); derived != requester.PeerID {
		return fmt.Errorf("requester peer id mismatch: got %s want %s", requester.PeerID, derived)
	}
	if err := VerifyCanonical(requester.PublicKey, drainRequestPayload(requester), req.Signature); err != nil {
		return err
	}
	return nil
}

func derivePeerIDFromPublicKey(publicKey string) string {
	pub, err := base64.RawURLEncoding.DecodeString(publicKey)
	if err != nil {
		return ""
	}
	return derivePeerID(ed25519.PublicKey(pub))
}

func marshalPeerPayload(payload any) ([]byte, *network.Error) {
	raw, err := network.MarshalPayload(payload)
	if err != nil {
		return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
	}
	return raw, nil
}

func (s *Service) HandlePeerOperation(_ context.Context, operation network.Operation, payload []byte) ([]byte, *network.Error) {
	switch operation {
	case network.OperationPeerInfo:
		return marshalPeerPayload(s.selfPeerInfo())
	case network.OperationPeerExchange:
		var request PeerExchangeRequest
		if err := network.UnmarshalPayload(payload, &request); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		return marshalPeerPayload(s.peerExchange(request))
	case network.OperationBootstrapRegister:
		var peer PeerInfo
		if err := network.UnmarshalPayload(payload, &peer); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		if transportErr := s.bootstrapRegisterPeer(peer); transportErr != nil {
			return nil, transportErr
		}
		return nil, nil
	case network.OperationBootstrapPeers:
		return marshalPeerPayload(s.bootstrapPeers())
	case network.OperationManifestPublish:
		var manifest Manifest
		if err := network.UnmarshalPayload(payload, &manifest); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		if transportErr := s.publishPeerManifest(manifest); transportErr != nil {
			return nil, transportErr
		}
		return nil, nil
	case network.OperationManifestFetch:
		var request struct {
			ServerID string `json:"server_id"`
		}
		if err := network.UnmarshalPayload(payload, &request); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		manifest, transportErr := s.peerManifest(request.ServerID)
		if transportErr != nil {
			return nil, transportErr
		}
		return marshalPeerPayload(manifest)
	case network.OperationPreviewFetch:
		var request struct {
			ServerID string `json:"server_id"`
		}
		if err := network.UnmarshalPayload(payload, &request); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		preview, transportErr := s.peerPreview(request.ServerID)
		if transportErr != nil {
			return nil, transportErr
		}
		return marshalPeerPayload(preview)
	case network.OperationJoin:
		var request JoinRequest
		if err := network.UnmarshalPayload(payload, &request); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		response, transportErr := s.peerJoin(request)
		if transportErr != nil {
			return nil, transportErr
		}
		return marshalPeerPayload(response)
	case network.OperationDeliver:
		var delivery Delivery
		if err := network.UnmarshalPayload(payload, &delivery); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		if err := s.applyDelivery(delivery); err != nil {
			return nil, &network.Error{Code: "invalid_signature", Message: err.Error()}
		}
		return nil, nil
	case network.OperationRelayStore:
		var delivery Delivery
		if err := network.UnmarshalPayload(payload, &delivery); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		if transportErr := s.peerRelayStore(delivery); transportErr != nil {
			return nil, transportErr
		}
		return nil, nil
	case network.OperationRelayDrain:
		var request DrainRequest
		if err := network.UnmarshalPayload(payload, &request); err != nil {
			return nil, &network.Error{Code: "invalid_request", Message: err.Error()}
		}
		deliveries, transportErr := s.peerRelayDrain(request)
		if transportErr != nil {
			return nil, transportErr
		}
		return marshalPeerPayload(deliveries)
	default:
		return nil, &network.Error{Code: "unsupported_operation", Message: "unsupported peer operation"}
	}
}

func (s *Service) bootstrapRegisterPeer(peer PeerInfo) *network.Error {
	s.mu.Lock()
	prior, hadPrior := s.state.KnownPeers[peer.PeerID]
	s.upsertPeerLocked(PeerRecord{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey, Source: "bootstrap-register", LastSeenAt: time.Now().UTC()})
	if err := s.saveLocked(); err != nil {
		if hadPrior {
			s.state.KnownPeers[peer.PeerID] = prior
		} else {
			delete(s.state.KnownPeers, peer.PeerID)
		}
		s.mu.Unlock()
		return internalNetworkError(err)
	}
	s.mu.Unlock()
	return nil
}

func (s *Service) bootstrapPeers() []PeerInfo {
	peers := []PeerInfo{s.selfPeerInfo()}
	s.mu.RLock()
	for _, peer := range sortedPeers(s.state.KnownPeers) {
		if peer.PeerID == s.state.Identity.PeerID {
			continue
		}
		peers = append(peers, PeerInfo{PeerID: peer.PeerID, Role: peer.Role, Addresses: peer.Addresses, PublicKey: peer.PublicKey})
	}
	s.mu.RUnlock()
	return peers
}

func (s *Service) peerExchange(request PeerExchangeRequest) []PeerInfo {
	selfPeerID := s.PeerID()
	known := make(map[string]struct{}, len(request.KnownPeerIDs)+1)
	for _, peerID := range request.KnownPeerIDs {
		peerID = strings.TrimSpace(peerID)
		if peerID != "" {
			known[peerID] = struct{}{}
		}
	}
	limit := request.Limit
	if limit <= 0 || limit > 64 {
		limit = 32
	}
	out := []PeerInfo{}
	appendPeer := func(peer PeerRecord) {
		if len(out) >= limit || peer.PeerID == "" || peer.PeerID == selfPeerID {
			return
		}
		if _, skip := known[peer.PeerID]; skip {
			return
		}
		out = append(out, PeerInfo{PeerID: peer.PeerID, Role: peer.Role, Addresses: append([]string(nil), peer.Addresses...), PublicKey: peer.PublicKey})
		known[peer.PeerID] = struct{}{}
	}

	s.mu.RLock()
	for _, serverID := range dedupeSorted(request.ServerIDs) {
		server, ok := s.state.Servers[serverID]
		if !ok {
			continue
		}
		owner, ok := s.state.KnownPeers[server.OwnerPeerID]
		if ok {
			appendPeer(owner)
		}
		for _, memberID := range server.Members {
			peer, ok := s.state.KnownPeers[memberID]
			if ok {
				appendPeer(peer)
			}
		}
	}
	for _, peer := range sortedPeers(s.state.KnownPeers) {
		appendPeer(peer)
	}
	s.mu.RUnlock()
	return out
}

func (s *Service) publishPeerManifest(manifest Manifest) *network.Error {
	if err := manifest.Verify(); err != nil {
		return &network.Error{Code: "invalid_signature", Message: err.Error()}
	}
	s.mu.Lock()
	existing, ok := s.state.Servers[manifest.ServerID]
	if ok && existing.Manifest.UpdatedAt.After(manifest.UpdatedAt) {
		s.mu.Unlock()
		return nil
	}
	hadServer := ok
	priorServer := existing
	priorPeer, hadPriorPeer := s.state.KnownPeers[manifest.OwnerPeerID]
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
	if err := s.saveLocked(); err != nil {
		if hadServer {
			s.state.Servers[server.ID] = priorServer
		} else {
			delete(s.state.Servers, server.ID)
		}
		if hadPriorPeer {
			s.state.KnownPeers[manifest.OwnerPeerID] = priorPeer
		} else {
			delete(s.state.KnownPeers, manifest.OwnerPeerID)
		}
		s.mu.Unlock()
		return internalNetworkError(err)
	}
	s.mu.Unlock()
	return nil
}

func (s *Service) peerManifest(serverID string) (Manifest, *network.Error) {
	serverID = strings.TrimSpace(serverID)
	s.mu.RLock()
	server, ok := s.state.Servers[serverID]
	s.mu.RUnlock()
	if !ok {
		return Manifest{}, &network.Error{Code: "not_found", Message: "manifest not found"}
	}
	return server.Manifest, nil
}

func (s *Service) peerPreview(serverID string) (ServerPreviewInfo, *network.Error) {
	serverID = strings.TrimSpace(serverID)
	s.mu.RLock()
	server, ok := s.state.Servers[serverID]
	if !ok {
		s.mu.RUnlock()
		return ServerPreviewInfo{}, &network.Error{Code: "not_found", Message: "server preview not found"}
	}
	channels := make([]ChannelRecord, 0, len(server.Channels))
	for _, channel := range server.Channels {
		channels = append(channels, channel)
	}
	memberCount := len(server.Members)
	s.mu.RUnlock()
	sort.Slice(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	return ServerPreviewInfo{Manifest: server.Manifest, OwnerRole: inferPeerRoleFromCapabilities("", server.Manifest.Capabilities), Channels: channels, MemberCount: memberCount, SafetyLabels: safetyLabelsForManifest(server.Manifest)}, nil
}

func (s *Service) peerJoin(req JoinRequest) (JoinResponse, *network.Error) {
	if err := req.Invite.Verify(); err != nil {
		code := "invalid_signature"
		if strings.Contains(err.Error(), "expired") {
			code = "expired_invite"
		}
		return JoinResponse{}, &network.Error{Code: code, Message: err.Error()}
	}
	s.mu.Lock()
	server, ok := s.state.Servers[req.Invite.ServerID]
	if !ok {
		s.mu.Unlock()
		return JoinResponse{}, &network.Error{Code: "not_found", Message: "server not found"}
	}
	if server.Manifest.Hash() != req.Invite.ManifestHash {
		s.mu.Unlock()
		return JoinResponse{}, &network.Error{Code: "invalid_manifest", Message: "invite/manifest mismatch"}
	}
	if err := validateInviteManifest(req.Invite, server.Manifest); err != nil {
		s.mu.Unlock()
		return JoinResponse{}, &network.Error{Code: "invalid_manifest", Message: err.Error()}
	}
	if !contains(server.Members, req.Requester.PeerID) {
		server.Members = append(server.Members, req.Requester.PeerID)
		sort.Strings(server.Members)
	}
	server.UpdatedAt = time.Now().UTC()
	priorServer := s.state.Servers[server.ID]
	priorPeer, hadPriorPeer := s.state.KnownPeers[req.Requester.PeerID]
	s.state.Servers[server.ID] = server
	s.upsertPeerLocked(PeerRecord{PeerID: req.Requester.PeerID, Role: req.Requester.Role, Addresses: req.Requester.Addresses, PublicKey: req.Requester.PublicKey, Source: "join", LastSeenAt: time.Now().UTC()})
	if err := s.saveLocked(); err != nil {
		s.state.Servers[server.ID] = priorServer
		if hadPriorPeer {
			s.state.KnownPeers[req.Requester.PeerID] = priorPeer
		} else {
			delete(s.state.KnownPeers, req.Requester.PeerID)
		}
		s.mu.Unlock()
		return JoinResponse{}, internalNetworkError(err)
	}
	history := s.historyForServerLocked(server.ID)
	s.mu.Unlock()
	channels := make([]ChannelRecord, 0, len(server.Channels))
	for _, channel := range server.Channels {
		channels = append(channels, channel)
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })
	return JoinResponse{Manifest: server.Manifest, Server: server, Channels: channels, History: history}, nil
}

func (s *Service) peerRelayStore(delivery Delivery) *network.Error {
	if !s.supportsRelayStore() {
		return &network.Error{Code: "unsupported_operation", Message: "relay storage is not enabled for this role"}
	}
	if err := delivery.Verify(); err != nil {
		return &network.Error{Code: "invalid_signature", Message: err.Error()}
	}
	if err := validateRelayDelivery(delivery); err != nil {
		return &network.Error{Code: "invalid_request", Message: err.Error()}
	}
	entry, err := encodeRelayDelivery(s.cfg.DataDir, delivery)
	if err != nil {
		return internalNetworkError(err)
	}
	s.mu.Lock()
	now := time.Now().UTC()
	priorQueues := make(map[string][]RelayQueueEntry)
	for _, peerID := range dedupeSorted(delivery.RecipientPeerIDs) {
		if peerID == s.state.Identity.PeerID {
			continue
		}
		if _, ok := priorQueues[peerID]; !ok {
			priorQueues[peerID] = append([]RelayQueueEntry(nil), s.state.RelayQueues[peerID]...)
		}
		queue := append([]RelayQueueEntry(nil), s.state.RelayQueues[peerID]...)
		queue, _ = pruneExpiredRelayQueue(queue, now)
		queue = upsertRelayQueue(queue, entry, now)
		s.state.RelayQueues[peerID] = queue
	}
	if err := s.saveLocked(); err != nil {
		for peerID, queue := range priorQueues {
			if len(queue) == 0 {
				delete(s.state.RelayQueues, peerID)
				continue
			}
			s.state.RelayQueues[peerID] = queue
		}
		s.mu.Unlock()
		return internalNetworkError(err)
	}
	s.mu.Unlock()
	return nil
}

func (s *Service) peerRelayDrain(req DrainRequest) ([]Delivery, *network.Error) {
	if !s.supportsRelayStore() {
		return nil, &network.Error{Code: "unsupported_operation", Message: "relay drain is not enabled for this role"}
	}
	if err := validateDrainRequest(req); err != nil {
		code := "invalid_request"
		if strings.Contains(err.Error(), "signature") {
			code = "invalid_signature"
		}
		return nil, &network.Error{Code: code, Message: err.Error()}
	}
	s.mu.Lock()
	now := time.Now().UTC()
	originalQueue := append([]RelayQueueEntry(nil), s.state.RelayQueues[req.Requester.PeerID]...)
	queue := append([]RelayQueueEntry(nil), originalQueue...)
	queue, pruned := pruneExpiredRelayQueue(queue, now)
	if len(queue) == 0 {
		if pruned {
			delete(s.state.RelayQueues, req.Requester.PeerID)
			if err := s.saveLocked(); err != nil {
				s.state.RelayQueues[req.Requester.PeerID] = originalQueue
				s.mu.Unlock()
				return nil, internalNetworkError(err)
			}
		}
		s.mu.Unlock()
		return nil, nil
	}
	deliveries := make([]Delivery, 0, len(queue))
	for _, entry := range queue {
		delivery, err := decodeRelayDelivery(s.cfg.DataDir, entry)
		if err != nil {
			s.state.RelayQueues[req.Requester.PeerID] = originalQueue
			s.mu.Unlock()
			return nil, internalNetworkError(err)
		}
		deliveries = append(deliveries, delivery)
	}
	priorQueue := append([]RelayQueueEntry(nil), originalQueue...)
	delete(s.state.RelayQueues, req.Requester.PeerID)
	if err := s.saveLocked(); err != nil {
		s.state.RelayQueues[req.Requester.PeerID] = priorQueue
		s.mu.Unlock()
		return nil, internalNetworkError(err)
	}
	s.mu.Unlock()
	return deliveries, nil
}

func validateInviteManifest(invite Invite, manifest Manifest) error {
	if manifest.ServerID != invite.ServerID {
		return errors.New("invite server id mismatch")
	}
	if manifest.Hash() != invite.ManifestHash {
		return errors.New("invite manifest hash mismatch")
	}
	if invite.OwnerPeerID != manifest.OwnerPeerID {
		return errors.New("invite owner peer mismatch")
	}
	if invite.OwnerPublicKey != manifest.OwnerPublicKey {
		return errors.New("invite owner public key mismatch")
	}
	if !sameStringSet(invite.ServerAddrs, manifest.OwnerAddresses) {
		return errors.New("invite server addresses mismatch")
	}
	if !sameStringSet(invite.BootstrapAddrs, manifest.BootstrapAddrs) {
		return errors.New("invite bootstrap addresses mismatch")
	}
	if !sameStringSet(invite.RelayAddrs, manifest.RelayAddrs) {
		return errors.New("invite relay addresses mismatch")
	}
	return nil
}

func (s *Service) validateDeliveryLocked(delivery Delivery) error {
	if delivery.SenderPeerID == s.state.Identity.PeerID && delivery.SenderPublicKey != s.state.Identity.PublicKey {
		return errors.New("delivery sender public key mismatch")
	}
	if peer, ok := s.state.KnownPeers[delivery.SenderPeerID]; ok && peer.PublicKey != "" && peer.PublicKey != delivery.SenderPublicKey {
		return errors.New("delivery sender public key mismatch")
	}
	return nil
}

func (s *Service) fetchPeerInfo(source, address string) (PeerRecord, error) {
	var info PeerInfo
	if err := s.peerCall(context.Background(), address, network.OperationPeerInfo, nil, &info); err != nil {
		return PeerRecord{}, err
	}
	return PeerRecord{PeerID: info.PeerID, Role: info.Role, Addresses: info.Addresses, PublicKey: info.PublicKey, Source: source, LastSeenAt: time.Now().UTC()}, nil
}

func livePeerRegistrySnapshot() map[string][]string {
	livePeerRegistryMu.RLock()
	defer livePeerRegistryMu.RUnlock()
	out := make(map[string][]string, len(livePeerRegistry))
	for peerID, addresses := range livePeerRegistry {
		out[peerID] = append([]string(nil), addresses...)
	}
	return out
}

func (s *Service) knownServerIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.state.Servers))
	for serverID := range s.state.Servers {
		ids = append(ids, serverID)
	}
	sort.Strings(ids)
	return ids
}

func (s *Service) shouldAttemptDiscovery(layer, address string) bool {
	address = strings.TrimSpace(address)
	if address == "" {
		return false
	}
	key := layer + "|" + address
	now := time.Now().UTC()
	s.discoveryMu.Lock()
	defer s.discoveryMu.Unlock()
	state, ok := s.discoveryBackoff[key]
	if !ok || state.NextAttempt.IsZero() || !state.NextAttempt.After(now) {
		return true
	}
	return false
}

func (s *Service) markDiscoverySuccess(layer, address string) {
	address = strings.TrimSpace(address)
	if address == "" {
		return
	}
	key := layer + "|" + address
	s.discoveryMu.Lock()
	delete(s.discoveryBackoff, key)
	s.discoveryMu.Unlock()
}

func (s *Service) markDiscoveryFailure(layer, address string, err error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return
	}
	key := layer + "|" + address
	now := time.Now().UTC()
	s.discoveryMu.Lock()
	state := s.discoveryBackoff[key]
	state.Failures++
	if state.Failures > 6 {
		state.Failures = 6
	}
	backoff := time.Duration(1<<uint(state.Failures-1)) * s.cfg.DiscoveryInterval
	if backoff < s.cfg.DiscoveryInterval {
		backoff = s.cfg.DiscoveryInterval
	}
	state.NextAttempt = now.Add(backoff)
	s.discoveryBackoff[key] = state
	s.discoveryMu.Unlock()
	if err != nil {
		s.recordTelemetry(fmt.Sprintf("discovery.backoff layer=%s addr=%s retry_in=%s err=%s", layer, address, backoff, err.Error()))
	}
}

func (s *Service) saveLocked() error {
	return saveStateFn(s.cfg.DataDir, s.state)
}

func (s *Service) saveLockedWithRollback(prior persistedState) error {
	if err := s.saveLocked(); err != nil {
		s.state = prior
		return err
	}
	return nil
}

func (s *Service) saveKnownPeersLocked(prior map[string]PeerRecord) error {
	if err := s.saveLocked(); err != nil {
		s.state.KnownPeers = prior
		return err
	}
	return nil
}

func encodeRelayDelivery(dataDir string, delivery Delivery) (RelayQueueEntry, error) {
	raw, err := json.Marshal(delivery)
	if err != nil {
		return RelayQueueEntry{}, err
	}
	encrypted, err := storage.SealPayload(dataDir, raw)
	if err != nil {
		return RelayQueueEntry{}, err
	}
	now := time.Now().UTC()
	return RelayQueueEntry{Key: relayDeliveryKey(delivery), Payload: encrypted, EnqueuedAt: now, ExpiresAt: now.Add(relayQueueTTL)}, nil
}

func decodeRelayDelivery(dataDir string, entry RelayQueueEntry) (Delivery, error) {
	raw, err := storage.OpenPayload(dataDir, entry.Payload)
	if err != nil {
		return Delivery{}, err
	}
	var delivery Delivery
	if err := json.Unmarshal(raw, &delivery); err != nil {
		return Delivery{}, err
	}
	return delivery, nil
}

func (s *Service) listenAddressLocked() string {
	if s.peerRuntime != nil {
		if addr := strings.TrimSpace(s.peerRuntime.ListenAddress()); addr != "" {
			return addr
		}
	}
	if addr := strings.TrimSpace(s.cfg.ListenAddr); addr != "" {
		return addr
	}
	return ""
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
	existing.Addresses = normalizePeerAddresses(append(existing.Addresses, peer.Addresses...))
	s.state.KnownPeers[peer.PeerID] = existing
}

func normalizePeerAddresses(addresses []string) []string {
	normalized := make([]string, 0, len(addresses))
	for _, address := range addresses {
		addr, err := network.NormalizePeerAddress(address)
		if err != nil {
			continue
		}
		normalized = append(normalized, addr)
	}
	return dedupeSorted(normalized)
}

func (s *Service) recordTelemetry(entry string) {
	var prior persistedState
	s.mu.Lock()
	prior = clonePersistedState(s.state)
	s.state.Telemetry = append(s.state.Telemetry, fmt.Sprintf("%s %s", time.Now().UTC().Format(time.RFC3339Nano), entry))
	if len(s.state.Telemetry) > 256 {
		s.state.Telemetry = append([]string(nil), s.state.Telemetry[len(s.state.Telemetry)-256:]...)
	}
	if err := s.saveLocked(); err != nil {
		s.state = prior
		s.mu.Unlock()
		_, _ = fmt.Fprintf(os.Stderr, "telemetry persistence failed: %v\n", err)
		return
	}
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

func (s *Service) supportsRelayStore() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return supportsRelayRole(s.cfg.Role)
}

func supportsRelayRole(role Role) bool {
	return role == RoleRelay || role == RoleBootstrap
}

func validateRelayDelivery(delivery Delivery) error {
	if strings.TrimSpace(delivery.ID) == "" {
		return errors.New("delivery id is required")
	}
	if strings.TrimSpace(delivery.Kind) == "" {
		return errors.New("delivery kind is required")
	}
	recipients := dedupeSorted(delivery.RecipientPeerIDs)
	if len(recipients) == 0 {
		return errors.New("relay delivery must target at least one recipient")
	}
	return nil
}

func pruneExpiredRelayQueue(queue []RelayQueueEntry, now time.Time) ([]RelayQueueEntry, bool) {
	filtered := queue[:0]
	pruned := false
	for _, existing := range queue {
		if !existing.ExpiresAt.IsZero() && !existing.ExpiresAt.After(now) {
			pruned = true
			continue
		}
		filtered = append(filtered, existing)
	}
	if !pruned {
		return append([]RelayQueueEntry(nil), queue...), false
	}
	return append([]RelayQueueEntry(nil), filtered...), true
}

func upsertRelayQueue(queue []RelayQueueEntry, delivery RelayQueueEntry, now time.Time) []RelayQueueEntry {
	key := strings.TrimSpace(delivery.Key)
	prunedQueue, _ := pruneExpiredRelayQueue(queue, now)
	filtered := prunedQueue[:0]
	for _, existing := range prunedQueue {
		if strings.TrimSpace(existing.Key) != key {
			filtered = append(filtered, existing)
		}
	}
	filtered = append(filtered, delivery)
	if len(filtered) > relayQueueLimit {
		filtered = append([]RelayQueueEntry(nil), filtered[len(filtered)-relayQueueLimit:]...)
		return filtered
	}
	return append([]RelayQueueEntry(nil), filtered...)
}

func relayDeliveryKey(delivery Delivery) string {
	return strings.TrimSpace(delivery.Kind) + ":" + strings.TrimSpace(delivery.ID)
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
