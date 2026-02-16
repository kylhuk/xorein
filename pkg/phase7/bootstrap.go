package phase7

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"sync"
)

type ParticipantID string

type KeyState struct {
	MLSSecret    []byte
	SenderKey    []byte
	Signer       ed25519.PrivateKey
	Verifier     ed25519.PublicKey
	Rotation     uint64
	LegacySender [][]byte
}

type Bootstrapper struct {
	mu     sync.Mutex
	states map[ParticipantID]*KeyState
}

func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		states: make(map[ParticipantID]*KeyState),
	}
}

func (b *Bootstrapper) Bootstrap(id ParticipantID) (*KeyState, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if state, ok := b.states[id]; ok {
		return state, nil
	}
	state, err := newKeyState(1)
	if err != nil {
		return nil, err
	}
	b.states[id] = state
	return state, nil
}

func (b *Bootstrapper) Rotate(id ParticipantID) (*KeyState, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	prev, ok := b.states[id]
	if !ok {
		state, err := newKeyState(1)
		if err != nil {
			return nil, err
		}
		b.states[id] = state
		return state, nil
	}
	next, err := newKeyState(prev.Rotation + 1)
	if err != nil {
		return nil, err
	}
	next.LegacySender = append(next.LegacySender, cloneKey(prev.SenderKey))
	for _, legacy := range prev.LegacySender {
		next.LegacySender = append(next.LegacySender, cloneKey(legacy))
	}
	if len(next.LegacySender) > 2 {
		next.LegacySender = next.LegacySender[:2]
	}
	b.states[id] = next
	return next, nil
}

func (b *Bootstrapper) SenderCompatible(id ParticipantID, candidate []byte) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	state, ok := b.states[id]
	if !ok || len(candidate) == 0 {
		return false
	}
	if bytes.Equal(state.SenderKey, candidate) {
		return true
	}
	for _, legacy := range state.LegacySender {
		if bytes.Equal(legacy, candidate) {
			return true
		}
	}
	return false
}

func (b *Bootstrapper) RekeyOnMismatch(id ParticipantID, candidate []byte) (*KeyState, error) {
	if len(candidate) == 0 {
		return b.Rotate(id)
	}
	b.mu.Lock()
	state, ok := b.states[id]
	b.mu.Unlock()
	if !ok || !bytes.Equal(state.SenderKey, candidate) {
		return b.Rotate(id)
	}
	return state, nil
}

func newKeyState(rotation uint64) (*KeyState, error) {
	mls, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	sender, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyState{
		MLSSecret:    mls,
		SenderKey:    sender,
		Signer:       priv,
		Verifier:     pub,
		Rotation:     rotation,
		LegacySender: nil,
	}, nil
}

func randomBytes(length int) ([]byte, error) {
	out := make([]byte, length)
	if _, err := rand.Read(out); err != nil {
		return nil, err
	}
	return out, nil
}

func cloneKey(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
