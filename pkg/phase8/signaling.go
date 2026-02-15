package phase8

import (
	"errors"
	"fmt"
	"sync"

	apb "github.com/aether/code_aether/gen/go/proto"
	"google.golang.org/protobuf/proto"
)

const (
	defaultMaxOfferAttempts     uint32 = 3
	defaultOfferRetryBackoffMS  uint32 = 250
	defaultMaxAnswerAttempts    uint32 = 3
	defaultAnswerRetryBackoffMS uint32 = 250
	defaultMaxICEUpdates        uint32 = 64
	defaultICEUpdateTimeoutMS   uint32 = 4000
	defaultSignalFrameTTLSec    uint64 = 15
)

var (
	ErrSignalingSessionRefRequired = errors.New("phase8: signaling session ref required")
	ErrSignalSessionIDRequired     = errors.New("phase8: signaling session id required")
	ErrSignalFrameRequired         = errors.New("phase8: signaling frame required")
	ErrSignalTypeInvalid           = errors.New("phase8: signaling type invalid")
	ErrSignalPayloadRequired       = errors.New("phase8: signaling encrypted payload required")
	ErrSignalSequenceStale         = errors.New("phase8: signaling sequence stale")
	ErrSignalFrameExpired          = errors.New("phase8: signaling frame expired")
	ErrSignalSessionMismatch       = errors.New("phase8: signaling session mismatch")
	ErrSignalRetryLimit            = errors.New("phase8: signaling retry limit exceeded")
	ErrSignalICEUpdateLimit        = errors.New("phase8: signaling ice update limit exceeded")
	ErrSignalICEUpdateTimeout      = errors.New("phase8: signaling ice update timeout")
	ErrSignalTransitionInvalid     = errors.New("phase8: signaling transition invalid")
	ErrSignalEncodeFailed          = errors.New("phase8: signaling frame encode failed")
	ErrSignalDecodeFailed          = errors.New("phase8: signaling frame decode failed")
	ErrSignalEncryptFailed         = errors.New("phase8: signaling payload encryption failed")
	ErrSignalDecryptFailed         = errors.New("phase8: signaling payload decryption failed")
	ErrSignalPublishFailed         = errors.New("phase8: signaling publish failed")
	ErrSignalPublisherRequired     = errors.New("phase8: signaling publisher required")
	ErrSignalCipherRequired        = errors.New("phase8: signaling cipher required")
	ErrSignalTopicRequired         = errors.New("phase8: signaling topic required")
)

type SignalCipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type SignalPublisher interface {
	Publish(topic string, payload []byte) error
}

type GossipSignalingRuntime struct {
	mu            sync.Mutex
	machine       *SignalingSessionMachine
	cipher        SignalCipher
	publisher     SignalPublisher
	topic         string
	nextSequence  uint64
	retryPolicy   *apb.VoiceSignalRetryPolicy
	lastICEUpdate uint64
}

func NewGossipSignalingRuntime(machine *SignalingSessionMachine, cipher SignalCipher, publisher SignalPublisher, topic string) (*GossipSignalingRuntime, error) {
	if machine == nil {
		return nil, ErrSignalingSessionRefRequired
	}
	if cipher == nil {
		return nil, ErrSignalCipherRequired
	}
	if publisher == nil {
		return nil, ErrSignalPublisherRequired
	}
	if topic == "" {
		return nil, ErrSignalTopicRequired
	}
	return &GossipSignalingRuntime{
		machine:      machine,
		cipher:       cipher,
		publisher:    publisher,
		topic:        topic,
		nextSequence: 1,
		retryPolicy:  machine.retryPolicy,
	}, nil
}

func (r *GossipSignalingRuntime) PublishEncrypted(signalType apb.VoiceSignalType, payload []byte, sentAt uint64, ttlSec uint64) (*apb.VoiceSignalFrame, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	attempt := 0
	for {
		attempt++
		if err := r.nextRetryAttempt(signalType); err != nil {
			return nil, err
		}

		encryptedPayload := payload
		if requiresPayload(signalType) {
			ciphertext, err := r.cipher.Encrypt(payload)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrSignalEncryptFailed, err)
			}
			encryptedPayload = ciphertext
		}

		frame, err := NewVoiceSignalFrame(r.machine.state.GetSessionRef(), signalType, r.nextSequence, encryptedPayload, sentAt, ttlSec, r.retryPolicy)
		if err != nil {
			return nil, err
		}
		encoded, err := proto.Marshal(frame)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSignalEncodeFailed, err)
		}
		if err := r.publisher.Publish(r.topic, encoded); err == nil {
			r.nextSequence++
			return frame, nil
		}

		if !r.canRetry(signalType, attempt) {
			return nil, fmt.Errorf("%w after %d attempt(s)", ErrSignalPublishFailed, attempt)
		}
	}
}

func (r *GossipSignalingRuntime) ReceiveAndDispatch(encoded []byte) (*apb.VoiceSignalFrame, []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	frame := &apb.VoiceSignalFrame{}
	if err := proto.Unmarshal(encoded, frame); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrSignalDecodeFailed, err)
	}
	if err := r.machine.Advance(frame); err != nil {
		return nil, nil, err
	}

	var decrypted []byte
	if requiresPayload(frame.GetSignalType()) {
		plaintext, err := r.cipher.Decrypt(frame.GetEncryptedPayload())
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %v", ErrSignalDecryptFailed, err)
		}
		decrypted = plaintext
	}

	if frame.GetSignalType() == apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE {
		r.lastICEUpdate = frame.GetSentAt()
	}
	if r.machine.FailOnICETimeout(r.lastICEUpdate) {
		return nil, nil, ErrSignalICEUpdateTimeout
	}

	return frame, decrypted, nil
}

func (r *GossipSignalingRuntime) nextRetryAttempt(signalType apb.VoiceSignalType) error {
	switch signalType {
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER:
		_, _, err := r.machine.NextOfferAttempt()
		return err
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ANSWER:
		_, _, err := r.machine.NextAnswerAttempt()
		return err
	default:
		return nil
	}
}

func (r *GossipSignalingRuntime) canRetry(signalType apb.VoiceSignalType, attempt int) bool {
	switch signalType {
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER:
		return uint32(attempt) < r.retryPolicy.GetMaxOfferAttempts()
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ANSWER:
		return uint32(attempt) < r.retryPolicy.GetMaxAnswerAttempts()
	default:
		return false
	}
}

type SignalingSessionMachine struct {
	mu           sync.Mutex
	state        *apb.VoiceSignalSessionState
	retryPolicy  *apb.VoiceSignalRetryPolicy
	lastSequence uint64
	nowUnix      func() uint64
}

func DefaultVoiceSignalRetryPolicy() *apb.VoiceSignalRetryPolicy {
	return &apb.VoiceSignalRetryPolicy{
		MaxOfferAttempts:     defaultMaxOfferAttempts,
		OfferRetryBackoffMs:  defaultOfferRetryBackoffMS,
		MaxAnswerAttempts:    defaultMaxAnswerAttempts,
		AnswerRetryBackoffMs: defaultAnswerRetryBackoffMS,
		MaxIceUpdates:        defaultMaxICEUpdates,
		IceUpdateTimeoutMs:   defaultICEUpdateTimeoutMS,
	}
}

func NewSignalingSessionMachine(sessionRef *apb.VoiceSignalSessionRef, retryPolicy *apb.VoiceSignalRetryPolicy, nowUnix func() uint64) (*SignalingSessionMachine, error) {
	if sessionRef == nil {
		return nil, ErrSignalingSessionRefRequired
	}
	if sessionRef.GetSignalSessionId() == "" {
		return nil, ErrSignalSessionIDRequired
	}
	if retryPolicy == nil {
		retryPolicy = DefaultVoiceSignalRetryPolicy()
	}
	if nowUnix == nil {
		nowUnix = func() uint64 { return 0 }
	}
	now := nowUnix()
	return &SignalingSessionMachine{
		state: &apb.VoiceSignalSessionState{
			SessionRef:       cloneSessionRef(sessionRef),
			Status:           apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_IDLE,
			OfferAttempt:     0,
			AnswerAttempt:    0,
			IceUpdateCount:   0,
			LastTransitionAt: now,
			TerminalReason:   "",
		},
		retryPolicy: retryPolicy,
		nowUnix:     nowUnix,
	}, nil
}

func NewVoiceSignalFrame(sessionRef *apb.VoiceSignalSessionRef, signalType apb.VoiceSignalType, sequence uint64, encryptedPayload []byte, sentAt uint64, ttlSec uint64, retryPolicy *apb.VoiceSignalRetryPolicy) (*apb.VoiceSignalFrame, error) {
	if sessionRef == nil {
		return nil, ErrSignalingSessionRefRequired
	}
	if sessionRef.GetSignalSessionId() == "" {
		return nil, ErrSignalSessionIDRequired
	}
	if signalType == apb.VoiceSignalType_VOICE_SIGNAL_TYPE_UNSPECIFIED {
		return nil, ErrSignalTypeInvalid
	}
	if requiresPayload(signalType) && len(encryptedPayload) == 0 {
		return nil, ErrSignalPayloadRequired
	}
	if ttlSec == 0 {
		ttlSec = defaultSignalFrameTTLSec
	}
	return &apb.VoiceSignalFrame{
		SignalId:         fmt.Sprintf("%s-%d", sessionRef.GetSignalSessionId(), sequence),
		SessionRef:       cloneSessionRef(sessionRef),
		SignalType:       signalType,
		Sequence:         sequence,
		EncryptedPayload: append([]byte(nil), encryptedPayload...),
		SentAt:           sentAt,
		ExpiresAt:        sentAt + ttlSec,
		RetryPolicy:      retryPolicy,
	}, nil
}

func (m *SignalingSessionMachine) State() *apb.VoiceSignalSessionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := *m.state
	copy.SessionRef = cloneSessionRef(m.state.GetSessionRef())
	return &copy
}

func (m *SignalingSessionMachine) NextOfferAttempt() (uint32, uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.GetOfferAttempt() >= m.retryPolicy.GetMaxOfferAttempts() {
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED
		m.state.TerminalReason = ErrSignalRetryLimit.Error()
		m.state.LastTransitionAt = m.nowUnix()
		return 0, 0, ErrSignalRetryLimit
	}
	m.state.OfferAttempt++
	m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_OFFER_SENT
	m.state.LastTransitionAt = m.nowUnix()
	return m.state.OfferAttempt, m.state.OfferAttempt * m.retryPolicy.GetOfferRetryBackoffMs(), nil
}

func (m *SignalingSessionMachine) NextAnswerAttempt() (uint32, uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.GetAnswerAttempt() >= m.retryPolicy.GetMaxAnswerAttempts() {
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED
		m.state.TerminalReason = ErrSignalRetryLimit.Error()
		m.state.LastTransitionAt = m.nowUnix()
		return 0, 0, ErrSignalRetryLimit
	}
	m.state.AnswerAttempt++
	m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ANSWER_SENT
	m.state.LastTransitionAt = m.nowUnix()
	return m.state.AnswerAttempt, m.state.AnswerAttempt * m.retryPolicy.GetAnswerRetryBackoffMs(), nil
}

func (m *SignalingSessionMachine) RegisterICEUpdate(lastUpdateAt uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.GetIceUpdateCount() >= m.retryPolicy.GetMaxIceUpdates() {
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED
		m.state.TerminalReason = ErrSignalICEUpdateLimit.Error()
		m.state.LastTransitionAt = m.nowUnix()
		return ErrSignalICEUpdateLimit
	}
	m.state.IceUpdateCount++
	m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHING
	m.state.LastTransitionAt = lastUpdateAt
	return nil
}

func (m *SignalingSessionMachine) ICEUpdateTimedOut(lastUpdateAt uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.retryPolicy.GetIceUpdateTimeoutMs() == 0 {
		return false
	}
	now := m.nowUnix()
	if now <= lastUpdateAt {
		return false
	}
	return (now-lastUpdateAt)*1000 > uint64(m.retryPolicy.GetIceUpdateTimeoutMs())
}

func (m *SignalingSessionMachine) FailOnICETimeout(lastUpdateAt uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.retryPolicy.GetIceUpdateTimeoutMs() == 0 {
		return false
	}
	now := m.nowUnix()
	if now <= lastUpdateAt {
		return false
	}
	if (now-lastUpdateAt)*1000 <= uint64(m.retryPolicy.GetIceUpdateTimeoutMs()) {
		return false
	}
	m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED
	m.state.TerminalReason = ErrSignalICEUpdateTimeout.Error()
	m.state.LastTransitionAt = now
	return true
}

func (m *SignalingSessionMachine) Advance(frame *apb.VoiceSignalFrame) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.nowUnix()
	if err := validateSignalFrame(frame, m.state.GetSessionRef(), m.lastSequence, now); err != nil {
		return err
	}

	status := m.state.GetStatus()
	switch frame.GetSignalType() {
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER:
		if status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_TERMINATED || status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED {
			return ErrSignalTransitionInvalid
		}
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_OFFER_RECEIVED
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ANSWER:
		if status != apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_OFFER_SENT && status != apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_OFFER_RECEIVED && status != apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHING {
			return ErrSignalTransitionInvalid
		}
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHED
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE:
		if status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_TERMINATED || status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED {
			return ErrSignalTransitionInvalid
		}
		if m.state.GetIceUpdateCount() >= m.retryPolicy.GetMaxIceUpdates() {
			m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED
			m.state.TerminalReason = ErrSignalICEUpdateLimit.Error()
			return ErrSignalICEUpdateLimit
		}
		m.state.IceUpdateCount++
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHING
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_COMPLETE:
		if status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_TERMINATED || status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_FAILED {
			return ErrSignalTransitionInvalid
		}
		if status != apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHED {
			m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_ESTABLISHING
		}
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_RESTART:
		if status == apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_TERMINATED {
			return ErrSignalTransitionInvalid
		}
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_RESTARTING
		m.state.OfferAttempt = 0
		m.state.AnswerAttempt = 0
		m.state.IceUpdateCount = 0
		m.state.TerminalReason = ""
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_TERMINATE:
		m.state.Status = apb.VoiceSignalSessionStatus_VOICE_SIGNAL_SESSION_STATUS_TERMINATED
	default:
		return ErrSignalTypeInvalid
	}

	m.lastSequence = frame.GetSequence()
	m.state.LastTransitionAt = now
	return nil
}

func validateSignalFrame(frame *apb.VoiceSignalFrame, sessionRef *apb.VoiceSignalSessionRef, lastSequence uint64, now uint64) error {
	if frame == nil {
		return ErrSignalFrameRequired
	}
	if frame.GetSignalType() == apb.VoiceSignalType_VOICE_SIGNAL_TYPE_UNSPECIFIED {
		return ErrSignalTypeInvalid
	}
	if frame.GetSessionRef() == nil || frame.GetSessionRef().GetSignalSessionId() == "" {
		return ErrSignalSessionIDRequired
	}
	if sessionRef != nil && frame.GetSessionRef().GetSignalSessionId() != sessionRef.GetSignalSessionId() {
		return ErrSignalSessionMismatch
	}
	if frame.GetSequence() <= lastSequence {
		return ErrSignalSequenceStale
	}
	if frame.GetExpiresAt() > 0 && frame.GetExpiresAt() < now {
		return ErrSignalFrameExpired
	}
	if requiresPayload(frame.GetSignalType()) && len(frame.GetEncryptedPayload()) == 0 {
		return ErrSignalPayloadRequired
	}
	return nil
}

func requiresPayload(signalType apb.VoiceSignalType) bool {
	switch signalType {
	case apb.VoiceSignalType_VOICE_SIGNAL_TYPE_OFFER,
		apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ANSWER,
		apb.VoiceSignalType_VOICE_SIGNAL_TYPE_ICE_CANDIDATE,
		apb.VoiceSignalType_VOICE_SIGNAL_TYPE_RESTART:
		return true
	default:
		return false
	}
}

func cloneSessionRef(in *apb.VoiceSignalSessionRef) *apb.VoiceSignalSessionRef {
	if in == nil {
		return nil
	}
	return &apb.VoiceSignalSessionRef{
		SignalSessionId:     in.GetSignalSessionId(),
		VoiceSessionId:      in.GetVoiceSessionId(),
		ServerId:            in.GetServerId(),
		ChannelId:           in.GetChannelId(),
		LocalParticipantId:  in.GetLocalParticipantId(),
		RemoteParticipantId: in.GetRemoteParticipantId(),
	}
}
