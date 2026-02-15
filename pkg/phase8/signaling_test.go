package phase8

import (
	"bytes"
	"errors"
	"testing"

	apb "github.com/aether/code_aether/gen/go/proto"
	"google.golang.org/protobuf/proto"
)

type testSignalCipher struct {
	encryptErr error
	decryptErr error
}

func (c testSignalCipher) Encrypt(plaintext []byte) ([]byte, error) {
	if c.encryptErr != nil {
		return nil, c.encryptErr
	}
	out := append([]byte("enc:"), plaintext...)
	return out, nil
}

func (c testSignalCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if c.decryptErr != nil {
		return nil, c.decryptErr
	}
	if !bytes.HasPrefix(ciphertext, []byte("enc:")) {
		return nil, errors.New("invalid ciphertext prefix")
	}
	return append([]byte(nil), ciphertext[len("enc:"):]...), nil
}

type testSignalPublisher struct {
	failCount int
	called    int
	lastTopic string
	last      []byte
}

func (p *testSignalPublisher) Publish(topic string, payload []byte) error {
	p.called++
	p.lastTopic = topic
	p.last = append([]byte(nil), payload...)
	if p.called <= p.failCount {
		return errors.New("forced publish failure")
	}
	return nil
}

func TestNewVoiceSignalFrameValidationAndDefaults(t *testing.T) {
	sessionRef := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-1", VoiceSessionId: "voice-1"}

	frame, err := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, 1, []byte("enc-offer"), 100, 0, nil)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	if frame.GetSignalId() != "sig-1-1" {
		t.Fatalf("signal_id = %q, want %q", frame.GetSignalId(), "sig-1-1")
	}
	if frame.GetExpiresAt() != 115 {
		t.Fatalf("expires_at = %d, want %d", frame.GetExpiresAt(), uint64(115))
	}

	if _, err := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_UNSPECIFIED, 1, []byte("x"), 100, 1, nil); !errors.Is(err, ErrSignalTypeInvalid) {
		t.Fatalf("expected ErrSignalTypeInvalid, got %v", err)
	}
	if _, err := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, 1, nil, 100, 1, nil); !errors.Is(err, ErrSignalPayloadRequired) {
		t.Fatalf("expected ErrSignalPayloadRequired, got %v", err)
	}
}

func TestSignalingSessionMachineRetryAndTiming(t *testing.T) {
	now := uint64(100)
	clock := func() uint64 { return now }
	sessionRef := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-1", VoiceSessionId: "voice-1"}
	retry := &apb.VoiceSignalRetryPolicy{
		MaxOfferAttempts:     2,
		OfferRetryBackoffMs:  200,
		MaxAnswerAttempts:    2,
		AnswerRetryBackoffMs: 150,
		MaxIceUpdates:        2,
		IceUpdateTimeoutMs:   3000,
	}

	machine, err := NewSignalingSessionMachine(sessionRef, retry, clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}

	offerAttempt, offerDelay, err := machine.NextOfferAttempt()
	if err != nil {
		t.Fatalf("NextOfferAttempt() error = %v", err)
	}
	if offerAttempt != 1 || offerDelay != 200 {
		t.Fatalf("offer attempt/delay = %d/%d, want 1/200", offerAttempt, offerDelay)
	}

	answerAttempt, answerDelay, err := machine.NextAnswerAttempt()
	if err != nil {
		t.Fatalf("NextAnswerAttempt() error = %v", err)
	}
	if answerAttempt != 1 || answerDelay != 150 {
		t.Fatalf("answer attempt/delay = %d/%d, want 1/150", answerAttempt, answerDelay)
	}

	if err := machine.RegisterICEUpdate(100); err != nil {
		t.Fatalf("RegisterICEUpdate() error = %v", err)
	}
	now = 102
	if machine.ICEUpdateTimedOut(100) {
		t.Fatalf("ICEUpdateTimedOut() = true, want false")
	}
	now = 104
	if !machine.ICEUpdateTimedOut(100) {
		t.Fatalf("ICEUpdateTimedOut() = false, want true")
	}

	_, _, _ = machine.NextOfferAttempt()
	if _, _, err := machine.NextOfferAttempt(); !errors.Is(err, ErrSignalRetryLimit) {
		t.Fatalf("expected ErrSignalRetryLimit, got %v", err)
	}
}

func TestSignalingSessionMachineRejectsStaleExpiredAndMismatchedFrames(t *testing.T) {
	now := uint64(50)
	clock := func() uint64 { return now }
	sessionRef := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-1", VoiceSessionId: "voice-1"}
	machine, err := NewSignalingSessionMachine(sessionRef, DefaultVoiceSignalRetryPolicy(), clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}

	good, err := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, 1, []byte("enc"), 49, 10, nil)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	if err := machine.Advance(good); err != nil {
		t.Fatalf("Advance(good) error = %v", err)
	}

	stale, _ := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 1, []byte("enc-ice"), 50, 10, nil)
	if err := machine.Advance(stale); !errors.Is(err, ErrSignalSequenceStale) {
		t.Fatalf("expected ErrSignalSequenceStale, got %v", err)
	}

	otherRef := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-2", VoiceSessionId: "voice-1"}
	mismatch, _ := NewVoiceSignalFrame(otherRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 2, []byte("enc-ice"), 50, 10, nil)
	if err := machine.Advance(mismatch); !errors.Is(err, ErrSignalSessionMismatch) {
		t.Fatalf("expected ErrSignalSessionMismatch, got %v", err)
	}

	now = 70
	expired, _ := NewVoiceSignalFrame(sessionRef, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 2, []byte("enc-ice"), 50, 5, nil)
	if err := machine.Advance(expired); !errors.Is(err, ErrSignalFrameExpired) {
		t.Fatalf("expected ErrSignalFrameExpired, got %v", err)
	}
}

func TestGossipSignalingRuntimePublishAndReceiveRoundTrip(t *testing.T) {
	now := uint64(100)
	clock := func() uint64 { return now }
	ref := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-rt", VoiceSessionId: "voice-rt"}
	policy := &apb.VoiceSignalRetryPolicy{
		MaxOfferAttempts:     3,
		OfferRetryBackoffMs:  10,
		MaxAnswerAttempts:    3,
		AnswerRetryBackoffMs: 10,
		MaxIceUpdates:        4,
		IceUpdateTimeoutMs:   0,
	}
	machine, err := NewSignalingSessionMachine(ref, policy, clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}
	publisher := &testSignalPublisher{}
	runtime, err := NewGossipSignalingRuntime(machine, testSignalCipher{}, publisher, "aether/v0.1/voice/signal")
	if err != nil {
		t.Fatalf("NewGossipSignalingRuntime() error = %v", err)
	}

	frame, err := runtime.PublishEncrypted(apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, []byte("sdp-offer"), now, 10)
	if err != nil {
		t.Fatalf("PublishEncrypted() error = %v", err)
	}
	if frame.GetSequence() != 1 {
		t.Fatalf("published sequence = %d, want 1", frame.GetSequence())
	}
	if publisher.called != 1 {
		t.Fatalf("publisher called = %d, want 1", publisher.called)
	}
	if publisher.lastTopic != "aether/v0.1/voice/signal" {
		t.Fatalf("publisher topic = %q, want %q", publisher.lastTopic, "aether/v0.1/voice/signal")
	}

	rcvFrame, decrypted, err := runtime.ReceiveAndDispatch(publisher.last)
	if err != nil {
		t.Fatalf("ReceiveAndDispatch() error = %v", err)
	}
	if rcvFrame.GetSignalType() != apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER {
		t.Fatalf("received signal type = %v, want OFFER", rcvFrame.GetSignalType())
	}
	if string(decrypted) != "sdp-offer" {
		t.Fatalf("decrypted payload = %q, want %q", string(decrypted), "sdp-offer")
	}
}

func TestGossipSignalingRuntimePublishRetryAndLimit(t *testing.T) {
	now := uint64(100)
	clock := func() uint64 { return now }
	ref := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-retry", VoiceSessionId: "voice-retry"}
	policy := &apb.VoiceSignalRetryPolicy{
		MaxOfferAttempts:     2,
		OfferRetryBackoffMs:  10,
		MaxAnswerAttempts:    1,
		AnswerRetryBackoffMs: 10,
		MaxIceUpdates:        2,
		IceUpdateTimeoutMs:   1000,
	}
	machine, err := NewSignalingSessionMachine(ref, policy, clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}

	t.Run("publish succeeds after retry", func(t *testing.T) {
		publisher := &testSignalPublisher{failCount: 1}
		runtime, err := NewGossipSignalingRuntime(machine, testSignalCipher{}, publisher, "topic")
		if err != nil {
			t.Fatalf("NewGossipSignalingRuntime() error = %v", err)
		}
		if _, err := runtime.PublishEncrypted(apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, []byte("x"), now, 10); err != nil {
			t.Fatalf("PublishEncrypted() error = %v", err)
		}
		if publisher.called != 2 {
			t.Fatalf("publisher called = %d, want 2", publisher.called)
		}
	})

	t.Run("publish fails when retry limit exceeded", func(t *testing.T) {
		machine2, err := NewSignalingSessionMachine(ref, policy, clock)
		if err != nil {
			t.Fatalf("NewSignalingSessionMachine() error = %v", err)
		}
		publisher := &testSignalPublisher{failCount: 5}
		runtime, err := NewGossipSignalingRuntime(machine2, testSignalCipher{}, publisher, "topic")
		if err != nil {
			t.Fatalf("NewGossipSignalingRuntime() error = %v", err)
		}
		if _, err := runtime.PublishEncrypted(apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER, []byte("x"), now, 10); !errors.Is(err, ErrSignalPublishFailed) {
			t.Fatalf("expected ErrSignalPublishFailed, got %v", err)
		}
	})
}

func TestGossipSignalingRuntimeReceiveValidationAndTimeout(t *testing.T) {
	now := uint64(100)
	clock := func() uint64 { return now }
	ref := &apb.VoiceSignalSessionRef{SignalSessionId: "sig-rx", VoiceSessionId: "voice-rx"}
	policy := &apb.VoiceSignalRetryPolicy{
		MaxOfferAttempts:     2,
		OfferRetryBackoffMs:  10,
		MaxAnswerAttempts:    2,
		AnswerRetryBackoffMs: 10,
		MaxIceUpdates:        2,
		IceUpdateTimeoutMs:   1000,
	}
	machine, err := NewSignalingSessionMachine(ref, policy, clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}
	runtime, err := NewGossipSignalingRuntime(machine, testSignalCipher{}, &testSignalPublisher{}, "topic")
	if err != nil {
		t.Fatalf("NewGossipSignalingRuntime() error = %v", err)
	}

	iceFrame, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 1, []byte("enc:ice1"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	encoded, err := proto.Marshal(iceFrame)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	if _, _, err := runtime.ReceiveAndDispatch(encoded); err != nil {
		t.Fatalf("ReceiveAndDispatch(ice1) error = %v", err)
	}

	now = 102
	stale, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 1, []byte("enc:ice2"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	staleEncoded, err := proto.Marshal(stale)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	if _, _, err := runtime.ReceiveAndDispatch(staleEncoded); !errors.Is(err, ErrSignalSequenceStale) {
		t.Fatalf("expected ErrSignalSequenceStale, got %v", err)
	}

	now = 103
	iceFrame2, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 2, []byte("enc:ice3"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	encoded2, err := proto.Marshal(iceFrame2)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	now = 104
	if _, _, err := runtime.ReceiveAndDispatch(encoded2); err != nil {
		t.Fatalf("ReceiveAndDispatch(ice2) error = %v", err)
	}

	now = 107
	iceFrame3, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 3, []byte("enc:ice4"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	encoded3, err := proto.Marshal(iceFrame3)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	now = 109
	if _, _, err := runtime.ReceiveAndDispatch(encoded3); !errors.Is(err, ErrSignalICEUpdateLimit) {
		t.Fatalf("expected ErrSignalICEUpdateLimit, got %v", err)
	}

	machine2, err := NewSignalingSessionMachine(ref, policy, clock)
	if err != nil {
		t.Fatalf("NewSignalingSessionMachine() error = %v", err)
	}
	runtime2, err := NewGossipSignalingRuntime(machine2, testSignalCipher{}, &testSignalPublisher{}, "topic")
	if err != nil {
		t.Fatalf("NewGossipSignalingRuntime() error = %v", err)
	}
	now = 100
	iceFresh, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE, 1, []byte("enc:ice-fresh"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	iceFreshEnc, err := proto.Marshal(iceFresh)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	if _, _, err := runtime2.ReceiveAndDispatch(iceFreshEnc); err != nil {
		t.Fatalf("ReceiveAndDispatch(iceFresh) error = %v", err)
	}
	now = 103
	answer, err := NewVoiceSignalFrame(ref, apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ANSWER, 2, []byte("enc:answer"), now, 5, policy)
	if err != nil {
		t.Fatalf("NewVoiceSignalFrame() error = %v", err)
	}
	answerEnc, err := proto.Marshal(answer)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	now = 106
	if _, _, err := runtime2.ReceiveAndDispatch(answerEnc); !errors.Is(err, ErrSignalICEUpdateTimeout) {
		t.Fatalf("expected ErrSignalICEUpdateTimeout, got %v", err)
	}
}
