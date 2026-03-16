package phase6

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type MembershipStatus string

const (
	MembershipStatusNone    MembershipStatus = "none"
	MembershipStatusJoining MembershipStatus = "joining"
	MembershipStatusActive  MembershipStatus = "active"
	MembershipStatusFailed  MembershipStatus = "failed"
)

type MembershipState struct {
	ServerID      string
	Status        MembershipStatus
	LastError     string
	LastHandshake time.Time
	RetryCount    int
	Manifest      *Manifest
	ChatReady     bool
	VoiceReady    bool
}

func (s *MembershipState) ChatFlowEnabled() bool {
	return s.ChatReady && s.Status == MembershipStatusActive
}

func (s *MembershipState) VoiceFlowEnabled() bool {
	return s.VoiceReady && s.Status == MembershipStatusActive
}

func (s *MembershipState) Clone() *MembershipState {
	if s == nil {
		return nil
	}
	clone := *s
	clone.Manifest = s.Manifest.Clone()
	return &clone
}

type HandshakeErrorKind string

const (
	HandshakeErrManifestMissing  HandshakeErrorKind = "manifest_missing"
	HandshakeErrInvalidSignature HandshakeErrorKind = "invalid_signature"
	HandshakeErrAlreadyJoined    HandshakeErrorKind = "already_joined"
)

type HandshakeError struct {
	Kind HandshakeErrorKind
	Err  error
}

func (e *HandshakeError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("handshake[%s]: %v", e.Kind, e.Err)
}

func (e *HandshakeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type HandshakeMachine struct {
	mu       sync.Mutex
	store    *ManifestStore
	identity string
	state    map[string]*MembershipState
}

func NewHandshakeMachine(store *ManifestStore, identity string) *HandshakeMachine {
	return &HandshakeMachine{store: store, identity: identity, state: map[string]*MembershipState{}}
}

func (h *HandshakeMachine) ensureStateLocked(serverID string) *MembershipState {
	if st, ok := h.state[serverID]; ok {
		return st
	}
	st := &MembershipState{ServerID: serverID, Status: MembershipStatusNone}
	h.state[serverID] = st
	return st
}

func (h *HandshakeMachine) Join(link *DeepLink) (*MembershipState, error) {
	if h == nil || h.store == nil || link == nil {
		return nil, errors.New("handshake machine not ready")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	st := h.ensureStateLocked(link.ServerID)
	if st.Status == MembershipStatusActive {
		return st.Clone(), &HandshakeError{Kind: HandshakeErrAlreadyJoined, Err: errors.New("already joined")}
	}

	st.Status = MembershipStatusJoining
	st.LastError = ""
	st.RetryCount++

	manifest, err := h.store.Resolve(link.ServerID)
	if err != nil {
		st.Status = MembershipStatusFailed
		st.LastError = err.Error()
		st.LastHandshake = time.Now().UTC()
		return st.Clone(), &HandshakeError{Kind: HandshakeErrManifestMissing, Err: err}
	}

	if err := manifest.ValidateStoredSignature(); err != nil {
		st.Status = MembershipStatusFailed
		st.LastError = "invalid signature"
		st.LastHandshake = time.Now().UTC()
		return st.Clone(), &HandshakeError{Kind: HandshakeErrInvalidSignature, Err: err}
	}

	st.Manifest = manifest.Clone()
	st.Status = MembershipStatusActive
	st.ChatReady = manifest.Capabilities.Chat
	st.VoiceReady = manifest.Capabilities.Voice
	st.LastHandshake = time.Now().UTC()

	return st.Clone(), nil
}
