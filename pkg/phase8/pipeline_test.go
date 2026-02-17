package phase8

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	apb "github.com/aether/code_aether/gen/go/proto"
)

type testCapture struct {
	startErr    error
	stopErr     error
	startCalled int
	stopCalled  int
}

func (t *testCapture) Start(_ context.Context, _ *apb.VoiceCodecProfile) error {
	t.startCalled++
	return t.startErr
}

func (t *testCapture) Stop(_ context.Context) error {
	t.stopCalled++
	return t.stopErr
}

type testTransport struct {
	connectErr       error
	disconnectErr    error
	connectCalled    int
	disconnectCalled int
}

func (t *testTransport) Connect(_ context.Context, _ string, _ *apb.VoiceCodecProfile, _ *apb.VoiceTransportProfile) error {
	t.connectCalled++
	return t.connectErr
}

func (t *testTransport) Disconnect(_ context.Context, _ string) error {
	t.disconnectCalled++
	return t.disconnectErr
}

type testPeerConnection struct {
	id               string
	connectErr       error
	disconnectErr    error
	connectCalled    int
	disconnectCalled int
}

func (p *testPeerConnection) ID() string { return p.id }

func (p *testPeerConnection) Connect(_ context.Context, _ string, _ *apb.VoiceCodecProfile, _ *apb.VoiceTransportProfile) error {
	p.connectCalled++
	return p.connectErr
}

func (p *testPeerConnection) Disconnect(_ context.Context, _ string) error {
	p.disconnectCalled++
	return p.disconnectErr
}

type blockingPeerConnection struct {
	id               string
	connectStarted   chan struct{}
	releaseCh        chan struct{}
	disconnectCalled int
}

func newBlockingPeerConnection(id string) *blockingPeerConnection {
	return &blockingPeerConnection{
		id:             id,
		connectStarted: make(chan struct{}, 1),
		releaseCh:      make(chan struct{}),
	}
}

func (p *blockingPeerConnection) ID() string { return p.id }

func (p *blockingPeerConnection) Connect(_ context.Context, _ string, _ *apb.VoiceCodecProfile, _ *apb.VoiceTransportProfile) error {
	select {
	case p.connectStarted <- struct{}{}:
	default:
	}
	select {
	case <-p.releaseCh:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("connect release timed out")
	}
}

func (p *blockingPeerConnection) Disconnect(_ context.Context, _ string) error {
	p.disconnectCalled++
	return nil
}

func TestNewVoicePipelineBaselineDefaults(t *testing.T) {
	baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
	if err != nil {
		t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
	}
	if baseline.SessionId != "session-a" {
		t.Fatalf("SessionId = %q, want %q", baseline.SessionId, "session-a")
	}
	if baseline.CodecProfile.GetCodec() != apb.VoiceCodec_VOICE_CODEC_OPUS {
		t.Fatalf("codec = %v, want OPUS", baseline.CodecProfile.GetCodec())
	}
	if baseline.CodecProfile.GetFrameDurationMs() != defaultFrameDurationMS {
		t.Fatalf("frame_duration_ms = %d, want %d", baseline.CodecProfile.GetFrameDurationMs(), defaultFrameDurationMS)
	}
	if !baseline.TransportProfile.GetRtcpMux() {
		t.Fatalf("rtcp_mux = false, want true")
	}
	if !baseline.TransportProfile.GetIceRestartEnabled() {
		t.Fatalf("ice_restart_enabled = false, want true")
	}
	if baseline.ReconnectPolicy.GetMaxAttempts() != defaultReconnectAttempts {
		t.Fatalf("max_attempts = %d, want %d", baseline.ReconnectPolicy.GetMaxAttempts(), defaultReconnectAttempts)
	}
	if got := len(baseline.ReconnectPolicy.GetRetryableClasses()); got == 0 {
		t.Fatalf("retryable_classes length = %d, want > 0", got)
	}
}

func TestNewVoicePipelineBaselineValidation(t *testing.T) {
	if _, err := NewVoicePipelineBaseline("", 1700000000); err == nil {
		t.Fatalf("expected error for empty session id")
	}
}

func TestAppendLifecycleHook(t *testing.T) {
	baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
	if err != nil {
		t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
	}
	err = AppendLifecycleHook(baseline, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_READY, "peer-1", 1700000001, "connected")
	if err != nil {
		t.Fatalf("AppendLifecycleHook() error = %v", err)
	}
	if len(baseline.LifecycleHooks) != 1 {
		t.Fatalf("lifecycle_hooks length = %d, want 1", len(baseline.LifecycleHooks))
	}
	hook := baseline.LifecycleHooks[0]
	if hook.GetPeerId() != "peer-1" {
		t.Fatalf("peer_id = %q, want %q", hook.GetPeerId(), "peer-1")
	}
	if hook.GetState() != apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_READY {
		t.Fatalf("state = %v, want READY", hook.GetState())
	}

	if err := AppendLifecycleHook(nil, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_READY, "peer-1", 1700000001, "connected"); err == nil {
		t.Fatalf("expected error for nil baseline")
	}
	if err := AppendLifecycleHook(baseline, apb.VoiceLifecycleState_VOICE_LIFECYCLE_STATE_READY, "", 1700000001, "connected"); err == nil {
		t.Fatalf("expected error for empty peer id")
	}
}

func TestReconnectBackoffMS(t *testing.T) {
	policy := DefaultReconnectPolicy()

	cases := []struct {
		attempt uint32
		want    uint32
	}{
		{attempt: 1, want: 250},
		{attempt: 2, want: 500},
		{attempt: 3, want: 1000},
		{attempt: 4, want: 2000},
		{attempt: 5, want: 4000},
	}

	for _, tc := range cases {
		got, err := ReconnectBackoffMS(policy, tc.attempt)
		if err != nil {
			t.Fatalf("ReconnectBackoffMS(attempt=%d) error = %v", tc.attempt, err)
		}
		if got != tc.want {
			t.Fatalf("ReconnectBackoffMS(attempt=%d) = %d, want %d", tc.attempt, got, tc.want)
		}
	}

	if _, err := ReconnectBackoffMS(policy, 0); err == nil {
		t.Fatalf("expected error for attempt 0")
	}
	if _, err := ReconnectBackoffMS(policy, policy.MaxAttempts+1); err == nil {
		t.Fatalf("expected error for attempt > max")
	}
	if _, err := ReconnectBackoffMS(nil, 1); err == nil {
		t.Fatalf("expected error for nil policy")
	}
}

func TestIsRetryableReconnectClass(t *testing.T) {
	policy := DefaultReconnectPolicy()

	if !IsRetryableReconnectClass(policy, apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_ICE_DISCONNECTED) {
		t.Fatalf("expected ICE_DISCONNECTED to be retryable")
	}
	if IsRetryableReconnectClass(policy, apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_AUTHENTICATION_FAILURE) {
		t.Fatalf("expected AUTHENTICATION_FAILURE to be non-retryable")
	}
	if IsRetryableReconnectClass(nil, apb.VoiceReconnectClass_VOICE_RECONNECT_CLASS_ICE_DISCONNECTED) {
		t.Fatalf("expected nil policy to be non-retryable")
	}
}

func TestVoiceEngineStartStopSuccess(t *testing.T) {
	baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
	if err != nil {
		t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
	}

	capture := &testCapture{}
	transport := &testTransport{}
	engine, err := NewVoiceEngine(baseline, EngineDeps{Capture: capture, Transport: transport})
	if err != nil {
		t.Fatalf("NewVoiceEngine() error = %v", err)
	}

	if err := engine.Start(context.Background(), "peer-1", 1700000001); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if got := engine.State(); got != EngineStateReady {
		t.Fatalf("State() after Start = %v, want %v", got, EngineStateReady)
	}
	if capture.startCalled != 1 {
		t.Fatalf("capture Start() calls = %d, want 1", capture.startCalled)
	}
	if transport.connectCalled != 1 {
		t.Fatalf("transport Connect() calls = %d, want 1", transport.connectCalled)
	}

	if err := engine.Stop(context.Background(), "peer-1", 1700000002); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if got := engine.State(); got != EngineStateStopped {
		t.Fatalf("State() after Stop = %v, want %v", got, EngineStateStopped)
	}
	if capture.stopCalled != 1 {
		t.Fatalf("capture Stop() calls = %d, want 1", capture.stopCalled)
	}
	if transport.disconnectCalled != 1 {
		t.Fatalf("transport Disconnect() calls = %d, want 1", transport.disconnectCalled)
	}
	if len(baseline.GetLifecycleHooks()) != 3 {
		t.Fatalf("lifecycle_hooks length = %d, want 3", len(baseline.GetLifecycleHooks()))
	}
}

func TestVoiceEngineStartFailurePaths(t *testing.T) {
	t.Run("capture start failure", func(t *testing.T) {
		baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
		if err != nil {
			t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
		}
		capture := &testCapture{startErr: errors.New("capture failed")}
		transport := &testTransport{}
		engine, err := NewVoiceEngine(baseline, EngineDeps{Capture: capture, Transport: transport})
		if err != nil {
			t.Fatalf("NewVoiceEngine() error = %v", err)
		}

		err = engine.Start(context.Background(), "peer-1", 1700000001)
		if err == nil {
			t.Fatalf("expected Start() error")
		}
		if got := engine.State(); got != EngineStateStopped {
			t.Fatalf("State() = %v, want %v", got, EngineStateStopped)
		}
		if transport.connectCalled != 0 {
			t.Fatalf("transport Connect() calls = %d, want 0", transport.connectCalled)
		}
	})

	t.Run("transport connect failure", func(t *testing.T) {
		baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
		if err != nil {
			t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
		}
		capture := &testCapture{}
		transport := &testTransport{connectErr: errors.New("connect failed")}
		engine, err := NewVoiceEngine(baseline, EngineDeps{Capture: capture, Transport: transport})
		if err != nil {
			t.Fatalf("NewVoiceEngine() error = %v", err)
		}

		err = engine.Start(context.Background(), "peer-1", 1700000001)
		if err == nil {
			t.Fatalf("expected Start() error")
		}
		if capture.stopCalled != 1 {
			t.Fatalf("capture Stop() calls = %d, want 1", capture.stopCalled)
		}
		if got := engine.State(); got != EngineStateStopped {
			t.Fatalf("State() = %v, want %v", got, EngineStateStopped)
		}
	})
}

func TestVoiceEngineStateGuards(t *testing.T) {
	baseline, err := NewVoicePipelineBaseline("session-a", 1700000000)
	if err != nil {
		t.Fatalf("NewVoicePipelineBaseline() error = %v", err)
	}
	capture := &testCapture{}
	transport := &testTransport{}
	engine, err := NewVoiceEngine(baseline, EngineDeps{Capture: capture, Transport: transport})
	if err != nil {
		t.Fatalf("NewVoiceEngine() error = %v", err)
	}

	if err := engine.Stop(context.Background(), "peer-1", 1700000001); !errors.Is(err, ErrEngineNotReady) {
		t.Fatalf("Stop() error = %v, want ErrEngineNotReady", err)
	}
	if err := engine.Start(context.Background(), "peer-1", 1700000002); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := engine.Start(context.Background(), "peer-1", 1700000003); !errors.Is(err, ErrEngineStarted) {
		t.Fatalf("second Start() error = %v, want ErrEngineStarted", err)
	}
	if err := engine.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if got := engine.State(); got != EngineStateClosed {
		t.Fatalf("State() after Close = %v, want %v", got, EngineStateClosed)
	}
	if err := engine.Start(context.Background(), "peer-1", 1700000004); !errors.Is(err, ErrEngineClosed) {
		t.Fatalf("Start() after Close error = %v, want ErrEngineClosed", err)
	}
}

func TestVoiceSessionManagerJoinLeaveAndCap(t *testing.T) {
	fallbackCalls := 0
	manager, err := NewVoiceSessionManager(2, func() error {
		fallbackCalls++
		return nil
	})
	if err != nil {
		t.Fatalf("NewVoiceSessionManager() error = %v", err)
	}

	codec := DefaultCodecProfile()
	transport := DefaultTransportProfile()
	ctx := context.Background()

	peer1 := &testPeerConnection{id: "peer-1"}
	peer2 := &testPeerConnection{id: "peer-2"}
	peer3 := &testPeerConnection{id: "peer-3"}

	if joinErr := manager.Join(ctx, "session-a", codec, transport, peer1); joinErr != nil {
		t.Fatalf("Join(peer-1) error = %v", joinErr)
	}
	if joinErr := manager.Join(ctx, "session-a", codec, transport, peer2); joinErr != nil {
		t.Fatalf("Join(peer-2) error = %v", joinErr)
	}
	if got := manager.ParticipantCount(); got != 2 {
		t.Fatalf("ParticipantCount() = %d, want 2", got)
	}

	err = manager.Join(ctx, "session-a", codec, transport, peer3)
	if !errors.Is(err, ErrPeerLimit) {
		t.Fatalf("Join(peer-3) error = %v, want ErrPeerLimit", err)
	}
	if fallbackCalls != 1 {
		t.Fatalf("fallback calls = %d, want 1", fallbackCalls)
	}
	if !manager.IsCongested() {
		t.Fatalf("IsCongested() = false, want true")
	}

	if err := manager.Leave(ctx, "session-a", "peer-1"); err != nil {
		t.Fatalf("Leave(peer-1) error = %v", err)
	}
	if peer1.disconnectCalled != 1 {
		t.Fatalf("peer-1 disconnect calls = %d, want 1", peer1.disconnectCalled)
	}
	if manager.IsCongested() {
		t.Fatalf("IsCongested() = true, want false")
	}

	if err := manager.Join(ctx, "session-a", codec, transport, peer3); err != nil {
		t.Fatalf("Join(peer-3) after leave error = %v", err)
	}
	if peer3.connectCalled != 1 {
		t.Fatalf("peer-3 connect calls = %d, want 1", peer3.connectCalled)
	}
}

func TestVoiceSessionManagerJoinEnforcesMaxPeersConcurrently(t *testing.T) {
	manager, err := NewVoiceSessionManager(1, nil)
	if err != nil {
		t.Fatalf("NewVoiceSessionManager() error = %v", err)
	}

	codec := DefaultCodecProfile()
	transport := DefaultTransportProfile()
	ctx := context.Background()
	firstPeer := newBlockingPeerConnection("peer-1")
	var firstJoinErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		firstJoinErr = manager.Join(ctx, "session-a", codec, transport, firstPeer)
	}()

	select {
	case <-firstPeer.connectStarted:
	case <-time.After(2 * time.Second):
		t.Fatalf("first join did not reach Connect")
	}

	if err := manager.Join(ctx, "session-a", codec, transport, &testPeerConnection{id: "peer-2"}); !errors.Is(err, ErrPeerLimit) {
		t.Fatalf("second Join() error = %v, want ErrPeerLimit", err)
	}

	close(firstPeer.releaseCh)
	wg.Wait()

	if firstJoinErr != nil {
		t.Fatalf("first Join() unexpected error = %v", firstJoinErr)
	}
	if got := manager.ParticipantCount(); got != 1 {
		t.Fatalf("ParticipantCount() = %d, want 1", got)
	}
}

func TestVoiceSessionManagerCongestionRemainsDuringPendingJoin(t *testing.T) {
	manager, err := NewVoiceSessionManager(1, nil)
	if err != nil {
		t.Fatalf("NewVoiceSessionManager() error = %v", err)
	}

	codec := DefaultCodecProfile()
	transport := DefaultTransportProfile()
	ctx := context.Background()
	firstPeer := newBlockingPeerConnection("peer-1")

	var wg sync.WaitGroup
	var firstErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		firstErr = manager.Join(ctx, "session-a", codec, transport, firstPeer)
	}()

	select {
	case <-firstPeer.connectStarted:
	case <-time.After(2 * time.Second):
		t.Fatalf("first join did not reach Connect")
	}

	secondErr := manager.Join(ctx, "session-a", codec, transport, &testPeerConnection{id: "peer-2"})
	if !errors.Is(secondErr, ErrPeerLimit) {
		t.Fatalf("second Join() error = %v, want ErrPeerLimit", secondErr)
	}
	if !manager.IsCongested() {
		t.Fatalf("IsCongested() = false, want true after concurrent refused join")
	}

	close(firstPeer.releaseCh)
	wg.Wait()

	if firstErr != nil {
		t.Fatalf("first Join() unexpected error = %v", firstErr)
	}
	if !manager.IsCongested() {
		t.Fatalf("IsCongested() = false, want true when max peers are connected")
	}
	if got := manager.ParticipantCount(); got != 1 {
		t.Fatalf("ParticipantCount() = %d, want 1", got)
	}
}

func TestVoiceSessionManagerTeardownRejectsStalePendingJoin(t *testing.T) {
	manager, err := NewVoiceSessionManager(1, nil)
	if err != nil {
		t.Fatalf("NewVoiceSessionManager() error = %v", err)
	}

	codec := DefaultCodecProfile()
	transport := DefaultTransportProfile()
	ctx := context.Background()
	firstPeer := newBlockingPeerConnection("peer-1")

	var wg sync.WaitGroup
	var firstErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		firstErr = manager.Join(ctx, "session-a", codec, transport, firstPeer)
	}()

	select {
	case <-firstPeer.connectStarted:
	case <-time.After(2 * time.Second):
		t.Fatalf("first join did not reach Connect")
	}

	if err := manager.Teardown(ctx, "session-a"); err != nil {
		t.Fatalf("Teardown() error = %v", err)
	}

	if err := manager.Join(ctx, "session-a", codec, transport, &testPeerConnection{id: "peer-2"}); err != nil {
		t.Fatalf("Join(peer-2) error after Teardown = %v", err)
	}
	if got := manager.ParticipantCount(); got != 1 {
		t.Fatalf("ParticipantCount() = %d, want 1", got)
	}

	close(firstPeer.releaseCh)
	wg.Wait()

	if !errors.Is(firstErr, ErrJoinCanceled) {
		t.Fatalf("first Join() error = %v, want ErrJoinCanceled", firstErr)
	}
	if firstPeer.disconnectCalled != 1 {
		t.Fatalf("first peer disconnect calls = %d, want 1", firstPeer.disconnectCalled)
	}
	if got := manager.ParticipantCount(); got != 1 {
		t.Fatalf("ParticipantCount() after stale join completion = %d, want 1", got)
	}
	if err := manager.Teardown(ctx, "session-a"); err != nil {
		t.Fatalf("Teardown() cleanup error = %v", err)
	}
}

func TestVoiceSessionManagerValidationAndTeardown(t *testing.T) {
	t.Run("constructor validation", func(t *testing.T) {
		if _, err := NewVoiceSessionManager(0, nil); err == nil {
			t.Fatalf("expected error for zero max peers")
		}
		if _, err := NewVoiceSessionManager(9, nil); err == nil {
			t.Fatalf("expected error for max peers > 8")
		}
	})

	t.Run("join/leave validation", func(t *testing.T) {
		manager, err := NewVoiceSessionManager(2, nil)
		if err != nil {
			t.Fatalf("NewVoiceSessionManager() error = %v", err)
		}
		codec := DefaultCodecProfile()
		transport := DefaultTransportProfile()
		ctx := context.Background()

		if err := manager.Join(ctx, "session-a", codec, transport, nil); !errors.Is(err, ErrPeerRequired) {
			t.Fatalf("Join(nil) error = %v, want ErrPeerRequired", err)
		}
		emptyPeer := &testPeerConnection{id: ""}
		if err := manager.Join(ctx, "session-a", codec, transport, emptyPeer); !errors.Is(err, ErrEmptyPeerID) {
			t.Fatalf("Join(empty-id) error = %v, want ErrEmptyPeerID", err)
		}

		peer := &testPeerConnection{id: "peer-1"}
		if err := manager.Join(ctx, "session-a", codec, transport, peer); err != nil {
			t.Fatalf("Join(peer-1) error = %v", err)
		}
		if err := manager.Join(ctx, "session-a", codec, transport, peer); !errors.Is(err, ErrPeerExists) {
			t.Fatalf("Join(duplicate peer-1) error = %v, want ErrPeerExists", err)
		}
		if err := manager.Leave(ctx, "session-a", ""); !errors.Is(err, ErrEmptyPeerID) {
			t.Fatalf("Leave(empty-id) error = %v, want ErrEmptyPeerID", err)
		}
		if err := manager.Leave(ctx, "session-a", "missing"); !errors.Is(err, ErrPeerNotFound) {
			t.Fatalf("Leave(missing) error = %v, want ErrPeerNotFound", err)
		}
	})

	t.Run("teardown disconnects all participants", func(t *testing.T) {
		manager, err := NewVoiceSessionManager(3, nil)
		if err != nil {
			t.Fatalf("NewVoiceSessionManager() error = %v", err)
		}
		codec := DefaultCodecProfile()
		transport := DefaultTransportProfile()
		ctx := context.Background()

		peerA := &testPeerConnection{id: "peer-a"}
		peerB := &testPeerConnection{id: "peer-b"}
		if err := manager.Join(ctx, "session-a", codec, transport, peerA); err != nil {
			t.Fatalf("Join(peer-a) error = %v", err)
		}
		if err := manager.Join(ctx, "session-a", codec, transport, peerB); err != nil {
			t.Fatalf("Join(peer-b) error = %v", err)
		}

		if err := manager.Teardown(ctx, "session-a"); err != nil {
			t.Fatalf("Teardown() error = %v", err)
		}
		if peerA.disconnectCalled != 1 || peerB.disconnectCalled != 1 {
			t.Fatalf("disconnect calls peer-a/peer-b = %d/%d, want 1/1", peerA.disconnectCalled, peerB.disconnectCalled)
		}
		if got := manager.ParticipantCount(); got != 0 {
			t.Fatalf("ParticipantCount() after teardown = %d, want 0", got)
		}
	})
}
