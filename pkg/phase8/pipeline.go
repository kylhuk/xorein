package phase8

import (
	"context"
	"errors"
	"fmt"
	"sync"

	apb "github.com/aether/code_aether/gen/go/proto"
)

const (
	defaultFrameDurationMS uint32 = 20
	defaultSampleRateHz    uint32 = 48000
	defaultChannels        uint32 = 2
	defaultMaxBitrateBPS   uint32 = 64000

	defaultReconnectAttempts  uint32 = 5
	defaultInitialBackoffMS   uint32 = 250
	defaultMaxBackoffMS       uint32 = 4000
	defaultVoiceTransportName        = "webrtc"
)

var (
	ErrEmptySessionID = errors.New("phase8: session id required")
	ErrEmptyPeerID    = errors.New("phase8: peer id required")
	ErrEngineClosed   = errors.New("phase8: voice engine closed")
	ErrEngineStarted  = errors.New("phase8: voice engine already started")
	ErrEngineNotReady = errors.New("phase8: voice engine not ready")
	ErrPeerRequired   = errors.New("phase8: peer connection required")
	ErrPeerLimit      = errors.New("phase8: peer limit reached")
	ErrPeerExists     = errors.New("phase8: peer already connected")
	ErrPeerNotFound   = errors.New("phase8: peer not connected")
	ErrJoinCanceled   = errors.New("phase8: join canceled by teardown")
)

type MediaCapture interface {
	Start(ctx context.Context, codec *apb.VoiceCodecProfile) error
	Stop(ctx context.Context) error
}

type PeerTransport interface {
	Connect(ctx context.Context, sessionID string, codec *apb.VoiceCodecProfile, transport *apb.VoiceTransportProfile) error
	Disconnect(ctx context.Context, sessionID string) error
}

type EngineDeps struct {
	Capture   MediaCapture
	Transport PeerTransport
}

type PeerConnection interface {
	PeerTransport
	ID() string
}

type VoiceSessionManager struct {
	mux        sync.Mutex
	peers      map[string]PeerConnection
	pending    map[string]struct{}
	maxPeers   int
	generation uint64
	congested  bool
	fallbackFn func() error
}

func (m *VoiceSessionManager) updateCongestionLocked() {
	m.congested = len(m.peers)+len(m.pending) >= m.maxPeers
}

func NewVoiceSessionManager(maxPeers int, fallbackFn func() error) (*VoiceSessionManager, error) {
	if maxPeers <= 0 {
		return nil, fmt.Errorf("phase8: max peers must be > 0")
	}
	if maxPeers > 8 {
		return nil, fmt.Errorf("phase8: max peers exceeds supported cap: %d", maxPeers)
	}
	return &VoiceSessionManager{
		peers:      make(map[string]PeerConnection, maxPeers),
		pending:    make(map[string]struct{}, maxPeers),
		maxPeers:   maxPeers,
		fallbackFn: fallbackFn,
	}, nil
}

func (m *VoiceSessionManager) Join(ctx context.Context, sessionID string, codec *apb.VoiceCodecProfile, transport *apb.VoiceTransportProfile, peer PeerConnection) error {
	if peer == nil {
		return ErrPeerRequired
	}
	peerID := peer.ID()
	if peerID == "" {
		return ErrEmptyPeerID
	}

	m.mux.Lock()
	if _, ok := m.peers[peerID]; ok {
		m.mux.Unlock()
		return ErrPeerExists
	}
	if _, ok := m.pending[peerID]; ok {
		m.mux.Unlock()
		return ErrPeerExists
	}
	if len(m.peers)+len(m.pending) >= m.maxPeers {
		m.congested = true
		fallback := m.fallbackFn
		m.mux.Unlock()
		if fallback != nil {
			if err := fallback(); err != nil {
				return fmt.Errorf("phase8: congestion fallback: %w", err)
			}
		}
		return ErrPeerLimit
	}
	joinGeneration := m.generation
	m.pending[peerID] = struct{}{}
	m.mux.Unlock()

	if err := peer.Connect(ctx, sessionID, codec, transport); err != nil {
		m.mux.Lock()
		delete(m.pending, peerID)
		m.updateCongestionLocked()
		m.mux.Unlock()
		return fmt.Errorf("phase8: connect peer %s: %w", peerID, err)
	}

	m.mux.Lock()
	delete(m.pending, peerID)
	if joinGeneration != m.generation {
		m.updateCongestionLocked()
		m.mux.Unlock()
		if err := peer.Disconnect(ctx, sessionID); err != nil {
			return fmt.Errorf("phase8: disconnect stale peer %s: %w", peerID, err)
		}
		return ErrJoinCanceled
	}
	m.peers[peerID] = peer
	m.updateCongestionLocked()
	m.mux.Unlock()
	return nil
}

func (m *VoiceSessionManager) Leave(ctx context.Context, sessionID, peerID string) error {
	if peerID == "" {
		return ErrEmptyPeerID
	}

	m.mux.Lock()
	peer, ok := m.peers[peerID]
	if !ok {
		m.mux.Unlock()
		return ErrPeerNotFound
	}
	delete(m.peers, peerID)
	m.updateCongestionLocked()
	m.mux.Unlock()

	if err := peer.Disconnect(ctx, sessionID); err != nil {
		return fmt.Errorf("phase8: disconnect peer %s: %w", peerID, err)
	}

	return nil
}

func (m *VoiceSessionManager) Teardown(ctx context.Context, sessionID string) error {
	m.mux.Lock()
	peers := make([]PeerConnection, 0, len(m.peers))
	for _, peer := range m.peers {
		peers = append(peers, peer)
	}
	m.generation++
	m.peers = make(map[string]PeerConnection, m.maxPeers)
	m.pending = make(map[string]struct{}, m.maxPeers)
	m.updateCongestionLocked()
	m.mux.Unlock()

	var errs []error
	for _, peer := range peers {
		if err := peer.Disconnect(ctx, sessionID); err != nil {
			errs = append(errs, fmt.Errorf("phase8: disconnect peer %s: %w", peer.ID(), err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (m *VoiceSessionManager) ParticipantIDs() []string {
	m.mux.Lock()
	defer m.mux.Unlock()
	ids := make([]string, 0, len(m.peers))
	for id := range m.peers {
		ids = append(ids, id)
	}
	return ids
}

func (m *VoiceSessionManager) ParticipantCount() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return len(m.peers)
}

func (m *VoiceSessionManager) IsCongested() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.congested
}

type EngineState uint8

const (
	EngineStateStopped EngineState = iota
	EngineStateStarting
	EngineStateReady
	EngineStateStopping
	EngineStateClosed
)

type VoiceEngine struct {
	mu       sync.Mutex
	state    EngineState
	baseline *apb.VoicePipelineBaseline
	deps     EngineDeps
}

func NewVoiceEngine(baseline *apb.VoicePipelineBaseline, deps EngineDeps) (*VoiceEngine, error) {
	if baseline == nil {
		return nil, errors.New("phase8: baseline required")
	}
	if baseline.GetSessionId() == "" {
		return nil, ErrEmptySessionID
	}
	if deps.Capture == nil {
		return nil, errors.New("phase8: media capture required")
	}
	if deps.Transport == nil {
		return nil, errors.New("phase8: peer transport required")
	}

	return &VoiceEngine{
		state:    EngineStateStopped,
		baseline: baseline,
		deps:     deps,
	}, nil
}

func (e *VoiceEngine) Start(ctx context.Context, peerID string, occurredAt uint64) error {
	e.mu.Lock()
	if e.state == EngineStateClosed {
		e.mu.Unlock()
		return ErrEngineClosed
	}
	if e.state == EngineStateReady || e.state == EngineStateStarting {
		e.mu.Unlock()
		return ErrEngineStarted
	}
	e.state = EngineStateStarting
	e.mu.Unlock()

	if err := AppendLifecycleHook(e.baseline, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_INITIALIZING, peerID, occurredAt, "engine_start"); err != nil {
		e.failToStopped()
		return err
	}
	if err := e.deps.Capture.Start(ctx, e.baseline.GetCodecProfile()); err != nil {
		e.failToStopped()
		return fmt.Errorf("phase8: start capture: %w", err)
	}
	if err := e.deps.Transport.Connect(ctx, e.baseline.GetSessionId(), e.baseline.GetCodecProfile(), e.baseline.GetTransportProfile()); err != nil {
		_ = e.deps.Capture.Stop(ctx)
		e.failToStopped()
		return fmt.Errorf("phase8: connect transport: %w", err)
	}
	if err := AppendLifecycleHook(e.baseline, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_READY, peerID, occurredAt, "stream_established"); err != nil {
		_ = e.deps.Transport.Disconnect(ctx, e.baseline.GetSessionId())
		_ = e.deps.Capture.Stop(ctx)
		e.failToStopped()
		return err
	}

	e.mu.Lock()
	e.state = EngineStateReady
	e.mu.Unlock()
	return nil
}

func (e *VoiceEngine) Stop(ctx context.Context, peerID string, occurredAt uint64) error {
	e.mu.Lock()
	if e.state == EngineStateClosed {
		e.mu.Unlock()
		return ErrEngineClosed
	}
	if e.state != EngineStateReady {
		e.mu.Unlock()
		return ErrEngineNotReady
	}
	e.state = EngineStateStopping
	e.mu.Unlock()

	errTransport := e.deps.Transport.Disconnect(ctx, e.baseline.GetSessionId())
	errCapture := e.deps.Capture.Stop(ctx)

	err := errors.Join(errTransport, errCapture)
	if err != nil {
		e.failToStopped()
		return fmt.Errorf("phase8: stop pipeline: %w", err)
	}
	if err := AppendLifecycleHook(e.baseline, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_CLOSED, peerID, occurredAt, "engine_stop"); err != nil {
		e.failToStopped()
		return err
	}

	e.mu.Lock()
	e.state = EngineStateStopped
	e.mu.Unlock()
	return nil
}

func (e *VoiceEngine) Close(ctx context.Context) error {
	e.mu.Lock()
	if e.state == EngineStateClosed {
		e.mu.Unlock()
		return nil
	}
	state := e.state
	e.mu.Unlock()

	if state == EngineStateReady {
		if err := e.Stop(ctx, "local", 0); err != nil {
			return err
		}
	}

	e.mu.Lock()
	e.state = EngineStateClosed
	e.mu.Unlock()
	return nil
}

func (e *VoiceEngine) State() EngineState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

func (e *VoiceEngine) failToStopped() {
	e.mu.Lock()
	if e.state != EngineStateClosed {
		e.state = EngineStateStopped
	}
	e.mu.Unlock()
}

func DefaultCodecProfile() *apb.VoiceCodecProfile {
	return &apb.VoiceCodecProfile{
		Codec:           apb.VoiceCodec_VOICE_CODEC_OPUS,
		FrameDurationMs: defaultFrameDurationMS,
		SampleRateHz:    defaultSampleRateHz,
		Channels:        defaultChannels,
		DtxEnabled:      true,
		FecEnabled:      true,
		MaxBitrateBps:   defaultMaxBitrateBPS,
	}
}

func DefaultTransportProfile() *apb.VoiceTransportProfile {
	return &apb.VoiceTransportProfile{
		Transport:          defaultVoiceTransportName,
		RtcpMux:            true,
		IceRestartEnabled:  true,
		ContinualGathering: true,
	}
}

func DefaultReconnectPolicy() *apb.VoiceReconnectPolicy {
	return &apb.VoiceReconnectPolicy{
		MaxAttempts:      defaultReconnectAttempts,
		InitialBackoffMs: defaultInitialBackoffMS,
		MaxBackoffMs:     defaultMaxBackoffMS,
		RetryableClasses: []apb.VoiceReconnectClass{
			apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_ICE_DISCONNECTED,
			apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_TRANSPORT_FAILURE,
			apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_REMOTE_RESTART,
			apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_LOCAL_DEVICE_CHANGE,
		},
	}
}

func NewVoicePipelineBaseline(sessionID string, configuredAt uint64) (*apb.VoicePipelineBaseline, error) {
	if sessionID == "" {
		return nil, ErrEmptySessionID
	}
	return &apb.VoicePipelineBaseline{
		SessionId:        sessionID,
		CodecProfile:     DefaultCodecProfile(),
		TransportProfile: DefaultTransportProfile(),
		ReconnectPolicy:  DefaultReconnectPolicy(),
		ConfiguredAt:     configuredAt,
	}, nil
}

func AppendLifecycleHook(baseline *apb.VoicePipelineBaseline, state apb.VoiceLifecycleState, peerID string, occurredAt uint64, detail string) error {
	if baseline == nil {
		return errors.New("phase8: baseline required")
	}
	if peerID == "" {
		return ErrEmptyPeerID
	}
	baseline.LifecycleHooks = append(baseline.LifecycleHooks, &apb.VoiceLifecycleHook{
		State:      state,
		PeerId:     peerID,
		OccurredAt: occurredAt,
		Detail:     detail,
	})
	return nil
}

func ReconnectBackoffMS(policy *apb.VoiceReconnectPolicy, attempt uint32) (uint32, error) {
	if policy == nil {
		return 0, errors.New("phase8: reconnect policy required")
	}
	if policy.InitialBackoffMs == 0 || policy.MaxBackoffMs == 0 || policy.MaxAttempts == 0 {
		return 0, fmt.Errorf("phase8: reconnect policy invalid")
	}
	if attempt == 0 || attempt > policy.MaxAttempts {
		return 0, fmt.Errorf("phase8: reconnect attempt out of range: %d", attempt)
	}

	backoff := policy.InitialBackoffMs
	for i := uint32(1); i < attempt; i++ {
		if backoff >= policy.MaxBackoffMs {
			return policy.MaxBackoffMs, nil
		}
		next := backoff * 2
		if next < backoff || next > policy.MaxBackoffMs {
			return policy.MaxBackoffMs, nil
		}
		backoff = next
	}
	return backoff, nil
}

func IsRetryableReconnectClass(policy *apb.VoiceReconnectPolicy, class apb.VoiceReconnectClass) bool {
	if policy == nil {
		return false
	}
	for _, item := range policy.RetryableClasses {
		if item == class {
			return true
		}
	}
	return false
}
