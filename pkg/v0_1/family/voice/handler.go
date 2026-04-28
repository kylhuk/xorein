// Package voice implements the /aether/voice/0.1.0 protocol family.
// Source: docs/spec/v0.1/52-family-voice.md
package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"

	ms "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

const protocolID = "/aether/voice/0.1.0"

// MaxFrameBytes is the maximum SFrame payload size per spec 52.
const MaxFrameBytes = 65_535

// Handler implements transport.FamilyHandler for the Voice family.
type Handler struct {
	mu          sync.RWMutex
	LocalPeerID string
	RelayMode   bool // when true, voice.frame is rejected with RELAY_OPACITY_VIOLATION
	sessions    map[string]*Session

	signals  *SignalState
	frames   *FrameCounterState
	mutes    *MuteState
	ice      *ICEState
}

var _ transport.FamilyHandler = (*Handler)(nil)

func (h *Handler) init() {
	if h.sessions == nil {
		h.sessions = make(map[string]*Session)
	}
	if h.signals == nil {
		h.signals = newSignalState()
	}
	if h.frames == nil {
		h.frames = newFrameCounterState()
	}
	if h.mutes == nil {
		h.mutes = newMuteState()
	}
	if h.ice == nil {
		h.ice = newICEState()
	}
}

func (h *Handler) ProtocolID() libp2pprotocol.ID { return libp2pprotocol.ID(protocolID) }

// Session represents an active voice session with MediaShield encryption.
type Session struct {
	ID           string
	Participants []string
	Keys         map[string]*ms.PeerKey // peerID → key material
	CreatedAt    time.Time
}

// ParticipantRecord tracks frame stats for a participant.
type ParticipantRecord struct {
	PeerID      string
	FramesRx    uint64
	FramesTx    uint64
	LastFrameAt time.Time
}

func (h *Handler) HandleStream(ctx context.Context, req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	required, known := proto.RequiredCapsFor(req.Operation)
	if !known {
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "unknown: "+req.Operation, nil)
	}
	remote := toFeatureFlags(req.AdvertisedCaps)
	result := proto.NegotiateCapabilities(h.localCaps(), remote, required)
	if len(result.MissingRequired) > 0 {
		return errorResp(req.RequestID, proto.CodeMissingRequiredCapability, "missing caps", result.MissingRequired)
	}

	h.mu.Lock()
	h.init()
	h.mu.Unlock()

	switch req.Operation {
	case "voice.join":
		return h.handleJoin(req)
	case "voice.leave":
		return h.handleLeave(req)
	case "voice.mute":
		return h.handleMute(req)
	case "voice.offer":
		return h.handleOffer(req)
	case "voice.answer":
		return h.handleAnswer(req)
	case "voice.ice":
		return h.handleICE(req)
	case "voice.ice_complete":
		return h.handleICEComplete(req)
	case "voice.restart":
		return h.handleRestart(req)
	case "voice.terminate":
		return h.handleTerminate(req)
	case "voice.frame":
		return h.handleFrame(req)
	default:
		return errorResp(req.RequestID, proto.CodeUnsupportedOperation, "not implemented: "+req.Operation, nil)
	}
}

func (h *Handler) localCaps() []proto.FeatureFlag {
	return []proto.FeatureFlag{"cap.voice", "mode.mediashield"}
}

// CreateSession creates a new voice session with MediaShield key material.
func (h *Handler) CreateSession(sessionID string, participants []string, keys map[string]*ms.PeerKey) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.init()
	h.sessions[sessionID] = &Session{
		ID:           sessionID,
		Participants: participants,
		Keys:         keys,
		CreatedAt:    time.Now(),
	}
	return nil
}

// EncryptFrame encrypts a media frame for a participant.
func (h *Handler) EncryptFrame(sessionID, senderID string, rtpHeader, plaintext []byte) ([]byte, []byte, error) {
	h.mu.RLock()
	s := h.sessions[sessionID]
	h.mu.RUnlock()
	if s == nil {
		return nil, nil, fmt.Errorf("voice: session %q not found", sessionID)
	}
	pk := s.Keys[senderID]
	if pk == nil {
		return nil, nil, fmt.Errorf("voice: no key for sender %q", senderID)
	}
	hdr, ct, err := ms.EncryptFrame(pk, rtpHeader, plaintext)
	if err != nil {
		return nil, nil, fmt.Errorf("voice: encrypt: %w", err)
	}
	return hdr, ct, nil
}

// LeaveSession removes a peer from a voice session (control API path).
func (h *Handler) LeaveSession(sessionID, peerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s := h.sessions[sessionID]; s != nil {
		newP := s.Participants[:0]
		for _, pid := range s.Participants {
			if pid != peerID {
				newP = append(newP, pid)
			}
		}
		s.Participants = newP
		delete(s.Keys, peerID)
		if len(s.Participants) == 0 {
			delete(h.sessions, sessionID)
			if h.frames != nil {
				h.frames.Remove(sessionID)
			}
			if h.ice != nil {
				h.ice.Remove(sessionID)
			}
		}
	}
}

// SetMuted sets the mute state for a peer in a session (control API path).
func (h *Handler) SetMuted(sessionID, peerID string, muted bool) {
	h.mu.Lock()
	h.init()
	h.mu.Unlock()
	h.mutes.SetMuted(sessionID, peerID, muted)
}

// IsMuted returns the mute state for a peer in a session.
func (h *Handler) IsMuted(sessionID, peerID string) bool {
	h.mu.RLock()
	if h.mutes == nil {
		h.mu.RUnlock()
		return false
	}
	h.mu.RUnlock()
	return h.mutes.IsMuted(sessionID, peerID)
}

// Sessions returns a snapshot of all active voice sessions.
func (h *Handler) Sessions() []*Session {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]*Session, 0, len(h.sessions))
	for _, s := range h.sessions {
		cp := &Session{
			ID:           s.ID,
			Participants: append([]string(nil), s.Participants...),
			CreatedAt:    s.CreatedAt,
		}
		out = append(out, cp)
	}
	return out
}

// DecryptFrame decrypts a media frame from a participant.
func (h *Handler) DecryptFrame(sessionID, senderID string, rtpHeader, sframeHeader, ct []byte) ([]byte, error) {
	h.mu.RLock()
	s := h.sessions[sessionID]
	h.mu.RUnlock()
	if s == nil {
		return nil, fmt.Errorf("voice: session %q not found", sessionID)
	}
	pk := s.Keys[senderID]
	if pk == nil {
		return nil, fmt.Errorf("voice: no key for sender %q", senderID)
	}
	return ms.DecryptFrame(pk, rtpHeader, sframeHeader, ct)
}

// --- wire handlers ---

type joinPayload struct {
	SessionID    string `json:"session_id"`
	PeerID       string `json:"peer_id"`
	SecurityMode string `json:"security_mode,omitempty"` // expected: "mediashield" or empty
}

func (h *Handler) handleJoin(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p joinPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	// Security mode check: if caller specifies a mode that is not mediashield, reject.
	if p.SecurityMode != "" && p.SecurityMode != "mediashield" {
		return errorResp(req.RequestID, proto.CodeVoiceNotAuthorized, "security mode not supported: "+p.SecurityMode, nil)
	}
	h.mu.Lock()
	// MediaShield key presence required for join (spec 52); if no key entry exists at all,
	// treat as MEDIASHIELD_KEY_UNAVAILABLE. Sessions created via CreateSession have keys.
	if _, exists := h.sessions[p.SessionID]; !exists {
		// Auto-create session without keys (no MediaShield key available).
		h.sessions[p.SessionID] = &Session{ID: p.SessionID, CreatedAt: time.Now(), Keys: make(map[string]*ms.PeerKey)}
	}
	s := h.sessions[p.SessionID]
	if len(s.Keys) == 0 && p.PeerID != "" {
		// No keys established means MediaShield parent scope not set up.
		h.mu.Unlock()
		return errorResp(req.RequestID, proto.CodeMediaShieldKeyUnavailable, "no MediaShield key for session", nil)
	}
	s.Participants = append(s.Participants, p.PeerID)
	h.mu.Unlock()
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "joined": p.PeerID})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type leavePayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
}

func (h *Handler) handleLeave(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p leavePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	if s := h.sessions[p.SessionID]; s != nil {
		newP := s.Participants[:0]
		for _, pid := range s.Participants {
			if pid != p.PeerID {
				newP = append(newP, pid)
			}
		}
		s.Participants = newP
		delete(s.Keys, p.PeerID)
	}
	h.mu.Unlock()
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "left": p.PeerID})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type mutePayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	Muted     bool   `json:"muted"`
	Sequence  uint64 `json:"sequence"`
}

func (h *Handler) handleMute(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p mutePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if !h.signals.Check(p.SessionID, p.PeerID, p.Sequence) {
		return errorResp(req.RequestID, proto.CodeVoiceSignalReplay, "signal replay detected", nil)
	}
	h.mutes.SetMuted(p.SessionID, p.PeerID, p.Muted)
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "peer_id": p.PeerID, "muted": p.Muted})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type offerPayload struct {
	SessionID  string `json:"session_id"`
	PeerID     string `json:"peer_id"`
	SDP        string `json:"sdp"`
	Sequence   uint64 `json:"sequence"`
	ExpiresAt  int64  `json:"expires_at"` // unix seconds
}

func (h *Handler) handleOffer(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p offerPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if p.ExpiresAt > 0 && time.Now().Unix() > p.ExpiresAt {
		return errorResp(req.RequestID, proto.CodeVoiceSignalExpired, "offer expired", nil)
	}
	if !h.signals.Check(p.SessionID, p.PeerID, p.Sequence) {
		return errorResp(req.RequestID, proto.CodeVoiceSignalReplay, "signal replay detected", nil)
	}
	if !strings.Contains(p.SDP, "opus/") {
		return errorResp(req.RequestID, proto.CodeVoiceCodecUnsupported, "Opus codec required", nil)
	}
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "accepted": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type answerPayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	SDP       string `json:"sdp"`
	Sequence  uint64 `json:"sequence"`
}

func (h *Handler) handleAnswer(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p answerPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if !h.signals.Check(p.SessionID, p.PeerID, p.Sequence) {
		return errorResp(req.RequestID, proto.CodeVoiceSignalReplay, "signal replay detected", nil)
	}
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "accepted": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type icePayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	Candidate string `json:"candidate"`
	Sequence  uint64 `json:"sequence"`
}

func (h *Handler) handleICE(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p icePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if !h.signals.Check(p.SessionID, p.PeerID, p.Sequence) {
		return errorResp(req.RequestID, proto.CodeVoiceSignalReplay, "signal replay detected", nil)
	}
	h.ice.Add(p.SessionID, p.Candidate)
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "accepted": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type iceCompletePayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	Sequence  uint64 `json:"sequence"`
}

func (h *Handler) handleICEComplete(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p iceCompletePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	if !h.signals.Check(p.SessionID, p.PeerID, p.Sequence) {
		return errorResp(req.RequestID, proto.CodeVoiceSignalReplay, "signal replay detected", nil)
	}
	h.ice.Complete(p.SessionID)
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "gathering_complete": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type restartPayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
	Sequence  uint64 `json:"sequence"`
}

func (h *Handler) handleRestart(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p restartPayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	// Reset signal and ICE state to allow fresh offer/answer (spec 52 §3.9).
	h.signals.Reset(p.SessionID, p.PeerID)
	h.ice.Remove(p.SessionID)
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "restarted": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type terminatePayload struct {
	SessionID string `json:"session_id"`
	PeerID    string `json:"peer_id"`
}

func (h *Handler) handleTerminate(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	var p terminatePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.Lock()
	if h.sessions[p.SessionID] == nil {
		h.mu.Unlock()
		return errorResp(req.RequestID, proto.CodeVoiceSessionNotFound, "session not found: "+p.SessionID, nil)
	}
	delete(h.sessions, p.SessionID)
	h.mu.Unlock()
	h.frames.Remove(p.SessionID)
	h.ice.Remove(p.SessionID)
	resp, _ := json.Marshal(map[string]any{"session_id": p.SessionID, "terminated": true})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

type framePayload struct {
	SessionID    string `json:"session_id"`
	SenderID     string `json:"sender_id"`
	RTPHeader    []byte `json:"rtp_header"`
	SFrameHeader []byte `json:"sframe_header"`
	Ciphertext   []byte `json:"ciphertext"`
	Counter      uint64 `json:"counter"`
}

func (h *Handler) handleFrame(req *proto.PeerStreamRequest) *proto.PeerStreamResponse {
	// Relay mode must reject frames to uphold relay opacity (spec 52).
	if h.RelayMode {
		return errorResp(req.RequestID, proto.CodeRelayOpacityViolation, "relay nodes must not process voice frames", nil)
	}
	var p framePayload
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errorResp(req.RequestID, proto.CodeOperationFailed, "invalid payload", nil)
	}
	h.mu.RLock()
	sessionExists := h.sessions[p.SessionID] != nil
	h.mu.RUnlock()
	if !sessionExists {
		return errorResp(req.RequestID, proto.CodeVoiceSessionNotFound, "session not found: "+p.SessionID, nil)
	}
	if len(p.Ciphertext) > MaxFrameBytes {
		return errorResp(req.RequestID, proto.CodeVoiceFrameTooLarge, fmt.Sprintf("frame exceeds %d bytes", MaxFrameBytes), nil)
	}
	// Mute check.
	if h.mutes.IsMuted(p.SessionID, p.SenderID) {
		return errorResp(req.RequestID, proto.CodeVoiceNotAuthorized, "sender is muted", nil)
	}
	// Frame counter replay check.
	if !h.frames.Check(p.SessionID, p.SenderID, p.Counter) {
		return errorResp(req.RequestID, proto.CodeReplayDetected, "frame counter replay", nil)
	}
	// Forward the ciphertext opaquely — the server never decrypts (E2EE property).
	resp, _ := json.Marshal(map[string]any{
		"session_id":      p.SessionID,
		"ciphertext_len":  len(p.Ciphertext),
		"ciphertext":      p.Ciphertext,
	})
	return &proto.PeerStreamResponse{NegotiatedProtocol: protocolID, Payload: resp}
}

func errorResp(requestID, code, msg string, missing []string) *proto.PeerStreamResponse {
	return &proto.PeerStreamResponse{
		RequestID: requestID,
		Error:     proto.NewPeerStreamError(code, msg, missing),
	}
}

func toFeatureFlags(ss []string) []proto.FeatureFlag {
	out := make([]proto.FeatureFlag, len(ss))
	for i, s := range ss {
		out[i] = proto.FeatureFlag(s)
	}
	return out
}
