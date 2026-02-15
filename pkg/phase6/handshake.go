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

func (h *HandshakeMachine) ensureState(serverID string) *MembershipState {
	h.mu.Lock()
	defer h.mu.Unlock()
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
	st := h.ensureState(link.ServerID)

	if st.Status == MembershipStatusActive {
		return st, &HandshakeError{Kind: HandshakeErrAlreadyJoined, Err: errors.New("already joined")}
	}

	st.Status = MembershipStatusJoining
	st.LastError = ""
	st.RetryCount++

	manifest, err := h.store.Resolve(link.ServerID)
	if err != nil {
		st.Status = MembershipStatusFailed
		st.LastError = err.Error()
		st.LastHandshake = time.Now().UTC()
		return st, &HandshakeError{Kind: HandshakeErrManifestMissing, Err: err}
	}

	if !manifest.ValidateSignature(h.identity) {
		st.Status = MembershipStatusFailed
		st.LastError = "invalid signature"
		st.LastHandshake = time.Now().UTC()
		return st, &HandshakeError{Kind: HandshakeErrInvalidSignature, Err: errors.New("signature mismatch")}
	}

	st.Manifest = manifest
	st.Status = MembershipStatusActive
	st.ChatReady = manifest.Capabilities.Chat
	st.VoiceReady = manifest.Capabilities.Voice
	st.LastHandshake = time.Now().UTC()

	return st, nil
}
