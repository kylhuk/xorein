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
	s.state.Settings["control_endpoint"] = controlEndpoint
	s.state.Settings["listen_address"] = s.listen.Addr().String()
	s.upsertPeerLocked(PeerRecord{
		PeerID:     s.state.Identity.PeerID,
		Role:       s.cfg.Role,
		Addresses:  []string{s.listen.Addr().String()},
		PublicKey:  s.state.Identity.PublicKey,
		Source:     "self",
		LastSeenAt: time.Now().UTC(),
	})
	err = s.saveLocked()
	token := s.state.ControlToken
	s.mu.Unlock()
	if err == nil {
		err = ensureControlTokenFile(s.cfg.DataDir, token)
	}
	if err != nil {
		_ = s.listen.Close()
		_ = s.controlListener.Close()
		return err
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = s.controlServer.Serve(controlListener)
	}()

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

func (s *Service) CreateServer(name, description string) (ServerRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "New Server"
	}
	now := time.Now().UTC()
	serverID := randomID("server")
	general := ChannelRecord{ID: randomID("channel"), ServerID: serverID, Name: "general", Voice: false, CreatedAt: now}
	capabilities := make([]string, 0, len(protocol.DefaultFeatureFlags()))
	for _, flag := range protocol.DefaultFeatureFlags() {
		capabilities = append(capabilities, string(flag))
	}
	s.mu.Lock()
	manifest := Manifest{
		ServerID:       serverID,
		Name:           name,
		Description:    strings.TrimSpace(description),
		OwnerPeerID:    s.state.Identity.PeerID,
		OwnerPublicKey: s.state.Identity.PublicKey,
		OwnerAddresses: []string{s.listenAddressLocked()},
		BootstrapAddrs: append([]string(nil), s.cfg.BootstrapAddrs...),
		RelayAddrs:     append([]string(nil), s.cfg.RelayAddrs...),
		Capabilities:   capabilities,
		IssuedAt:       now,
		UpdatedAt:      now,
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
	for _, flag := range protocol.DefaultFeatureFlags() {
		joinReq.Capabilities = append(joinReq.Capabilities, string(flag))
	}
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
		Role:       RoleClient,
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
	s.mu.Lock()
	server, ok := s.state.Servers[serverID]
	if !ok {
		s.mu.Unlock()
		return ChannelRecord{}, errors.New("server not found")
	}
	channel := ChannelRecord{ID: randomID("channel"), ServerID: serverID, Name: name, Voice: voice, CreatedAt: time.Now().UTC()}
	if server.Channels == nil {
		server.Channels = map[string]ChannelRecord{}
	}
	server.Channels[channel.ID] = channel
	server.UpdatedAt = time.Now().UTC()
	s.state.Servers[serverID] = server
	err := s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return ChannelRecord{}, err
	}
	s.recordTelemetry("channel.created id=" + channel.ID)
	s.emitEvent(Event{Type: "channel.created", Time: time.Now().UTC(), Payload: map[string]any{"channel_id": channel.ID, "server_id": serverID}})
	return channel, nil
}

func (s *Service) CreateDM(peerID string) (DMRecord, error) {
	peerID = strings.TrimSpace(peerID)
	if peerID == "" {
		return DMRecord{}, errors.New("peer id is required")
	}
	participants := dedupeSorted([]string{s.PeerID(), peerID})
	dmID := strings.Join(participants, ":")
	s.mu.Lock()
	if dm, ok := s.state.DMs[dmID]; ok {
		s.mu.Unlock()
		return dm, nil
	}
	dm := DMRecord{ID: dmID, Participants: participants, CreatedAt: time.Now().UTC()}
	s.state.DMs[dm.ID] = dm
	err := s.saveLocked()
	s.mu.Unlock()
	if err != nil {
		return DMRecord{}, err
	}
	s.recordTelemetry("dm.created id=" + dm.ID)
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
	if err := s.saveLocked(); err != nil {
		s.mu.Unlock()
		return MessageRecord{}, err
	}
	s.mu.Unlock()
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
	msg.Body = body
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	identity := s.state.Identity
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
	msg.Body = ""
	msg.Deleted = true
	msg.UpdatedAt = time.Now().UTC()
	s.state.Messages[msg.ID] = msg
	identity := s.state.Identity
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
	case "channel_message", "dm_message":
		msg := MessageRecord{ID: delivery.ID, ScopeType: delivery.ScopeType, ScopeID: delivery.ScopeID, ServerID: delivery.ServerID, SenderPeerID: delivery.SenderPeerID, Body: delivery.Body, CreatedAt: delivery.CreatedAt}
		s.state.Messages[msg.ID] = msg
		s.emitEventLocked(Event{Type: "message.created", Time: now, Payload: map[string]any{"message_id": msg.ID, "scope_id": msg.ScopeID}})
	case "message_edit":
		msg := s.state.Messages[delivery.ID]
		msg.Body = delivery.Body
		msg.UpdatedAt = delivery.CreatedAt
		msg.Deleted = false
		if msg.ID == "" {
			msg = MessageRecord{ID: delivery.ID, ScopeType: delivery.ScopeType, ScopeID: delivery.ScopeID, ServerID: delivery.ServerID, SenderPeerID: delivery.SenderPeerID, Body: delivery.Body, CreatedAt: delivery.CreatedAt, UpdatedAt: delivery.CreatedAt}
		}
		s.state.Messages[msg.ID] = msg
		s.emitEventLocked(Event{Type: "message.edited", Time: now, Payload: map[string]any{"message_id": msg.ID}})
	case "message_delete":
		msg := s.state.Messages[delivery.ID]
		msg.ID = delivery.ID
		msg.ScopeID = delivery.ScopeID
		msg.ScopeType = delivery.ScopeType
		msg.ServerID = delivery.ServerID
		msg.SenderPeerID = delivery.SenderPeerID
		msg.Deleted = true
		msg.Body = ""
		msg.UpdatedAt = delivery.CreatedAt
		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = delivery.CreatedAt
		}
		s.state.Messages[msg.ID] = msg
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
	s.upsertPeerLocked(PeerRecord{PeerID: manifest.OwnerPeerID, Role: RoleClient, Addresses: manifest.OwnerAddresses, PublicKey: manifest.OwnerPublicKey, Source: "manifest", LastSeenAt: time.Now().UTC()})
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

func controlTokenPath(dataDir string) string {
	return filepath.Join(dataDir, "control.token")
}

func ensureControlTokenFile(dataDir, token string) error {
	if err := os.WriteFile(controlTokenPath(dataDir), []byte(token), 0o600); err != nil {
		return fmt.Errorf("write control token: %w", err)
	}
	return nil
}
