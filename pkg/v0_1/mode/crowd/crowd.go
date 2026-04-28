// Package crowd implements Crowd mode: epoch-based sender-keyed broadcast E2EE.
// Source: docs/spec/v0.1/13-mode-crowd.md
package crowd

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

const (
	MaxEpochMessages = 1000
	EpochTTL         = 7 * 24 * time.Hour
	LegacyWindowSize = 2 // spec 13 §3.3: exactly 2 previous epochs valid
)

var (
	ErrEpochExpired  = errors.New("crowd: epoch expired (not in legacy window)")
	ErrDecryptFailed = errors.New("crowd: AEAD decrypt failed")
	ErrUnknownSender = errors.New("crowd: unknown sender key")
)

// EpochState holds key material for a single crowd epoch.
type EpochState struct {
	EpochID      uint64            `json:"epoch_id"`
	EpochRoot    []byte            `json:"epoch_root"`  // 32 bytes
	SenderKeys   map[string][]byte `json:"sender_keys"` // peerID → 32-byte sender key
	MessageCount uint32            `json:"msg_count"`
	StartedAt    int64             `json:"started_at"` // ms
}

// GroupState is the mutable state of a Crowd mode scope.
type GroupState struct {
	ScopeID      string        `json:"scope_id"`
	CurrentEpoch *EpochState   `json:"current_epoch"`
	PrevEpochs   []*EpochState `json:"prev_epochs"` // up to LegacyWindowSize
}

// Ciphertext is a Crowd mode encrypted message.
type Ciphertext struct {
	EpochID  uint64 `json:"epoch_id"`
	SenderID string `json:"sender_id"`
	Nonce    []byte `json:"nonce"` // 12 bytes
	CT       []byte `json:"ct"`    // ChaCha20-Poly1305 ciphertext + tag
}

// NewGroup creates a new Crowd mode scope with a random initial epoch root.
func NewGroup(scopeID string) (*GroupState, error) {
	root, err := v0crypto.RandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("crowd: new group: %w", err)
	}
	epoch, err := epochFromRoot(root, 0)
	if err != nil {
		return nil, err
	}
	return &GroupState{
		ScopeID:      scopeID,
		CurrentEpoch: epoch,
	}, nil
}

// DeriveEpochRoot derives the next epoch root from the current one.
func DeriveEpochRoot(current []byte) ([]byte, error) {
	return v0crypto.DeriveKey(current, nil, v0crypto.LabelCrowdEpochRoot, 32)
}

// DeriveSenderKey derives a per-sender key from the epoch root.
func DeriveSenderKey(epochRoot []byte, peerID string) ([]byte, error) {
	info := v0crypto.LabelCrowdSenderKey + peerID
	return v0crypto.DeriveKey(epochRoot, nil, info, 32)
}

// AddSender derives and registers the sender key for a peer in the current epoch.
func AddSender(g *GroupState, peerID string) error {
	sk, err := DeriveSenderKey(g.CurrentEpoch.EpochRoot, peerID)
	if err != nil {
		return fmt.Errorf("crowd: add sender: %w", err)
	}
	if g.CurrentEpoch.SenderKeys == nil {
		g.CurrentEpoch.SenderKeys = make(map[string][]byte)
	}
	g.CurrentEpoch.SenderKeys[peerID] = sk
	return nil
}

// Encrypt encrypts a message with the sender's key (ChaCha20-Poly1305).
// Rotates epoch if limit/TTL reached.
func Encrypt(g *GroupState, senderID string, plaintext []byte) (*Ciphertext, error) {
	if needsRotation(g.CurrentEpoch) {
		if err := rotateEpoch(g); err != nil {
			return nil, err
		}
	}
	sk := g.CurrentEpoch.SenderKeys[senderID]
	if sk == nil {
		// Derive on-demand.
		var err error
		sk, err = DeriveSenderKey(g.CurrentEpoch.EpochRoot, senderID)
		if err != nil {
			return nil, fmt.Errorf("crowd: encrypt derive: %w", err)
		}
		if g.CurrentEpoch.SenderKeys == nil {
			g.CurrentEpoch.SenderKeys = make(map[string][]byte)
		}
		g.CurrentEpoch.SenderKeys[senderID] = sk
	}
	nonce, err := v0crypto.RandomBytes(12)
	if err != nil {
		return nil, fmt.Errorf("crowd: nonce: %w", err)
	}
	var sk32 [32]byte
	copy(sk32[:], sk)
	var nonce12 [12]byte
	copy(nonce12[:], nonce)
	aad := epochAAD(g.CurrentEpoch.EpochID, senderID)
	ct, err := v0crypto.SealChaCha20Poly1305(sk32, nonce12, plaintext, aad)
	if err != nil {
		return nil, fmt.Errorf("crowd: encrypt: %w", err)
	}
	g.CurrentEpoch.MessageCount++
	return &Ciphertext{
		EpochID:  g.CurrentEpoch.EpochID,
		SenderID: senderID,
		Nonce:    nonce,
		CT:       ct,
	}, nil
}

// Decrypt decrypts a crowd ciphertext using the current or legacy epoch sender key.
func Decrypt(g *GroupState, ct *Ciphertext) ([]byte, error) {
	epoch := findEpoch(g, ct.EpochID)
	if epoch == nil {
		return nil, ErrEpochExpired
	}
	sk := epoch.SenderKeys[ct.SenderID]
	if sk == nil {
		var err error
		sk, err = DeriveSenderKey(epoch.EpochRoot, ct.SenderID)
		if err != nil {
			return nil, fmt.Errorf("crowd: decrypt derive: %w", err)
		}
	}
	var sk32 [32]byte
	copy(sk32[:], sk)
	var nonce12 [12]byte
	copy(nonce12[:], ct.Nonce)
	aad := epochAAD(ct.EpochID, ct.SenderID)
	pt, err := v0crypto.OpenChaCha20Poly1305(sk32, nonce12, ct.CT, aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return pt, nil
}

// rotateEpoch rotates to a new epoch, chaining from the current root.
// For time/message-limit rotations, the next root is derived via the epoch chain.
// Use RotateEpochMembership when a membership change occurs (fresh random root).
func rotateEpoch(g *GroupState) error {
	nextRoot, err := DeriveEpochRoot(g.CurrentEpoch.EpochRoot)
	if err != nil {
		return fmt.Errorf("crowd: rotate epoch root: %w", err)
	}
	nextID := g.CurrentEpoch.EpochID + 1
	nextEpoch, err := epochFromRoot(nextRoot, nextID)
	if err != nil {
		return err
	}
	prev := append([]*EpochState{g.CurrentEpoch}, g.PrevEpochs...)
	if len(prev) > LegacyWindowSize {
		prev = prev[:LegacyWindowSize]
	}
	g.PrevEpochs = prev
	g.CurrentEpoch = nextEpoch
	return nil
}

// RotateEpochMembership rotates the epoch using a fresh random root per spec 13 §6.2.
// MUST be called on any membership change (member join or leave).
// A fresh random epoch_root_secret is used (no chaining) to exclude removed members.
func RotateEpochMembership(g *GroupState) error {
	freshRoot, err := v0crypto.RandomBytes(32)
	if err != nil {
		return fmt.Errorf("crowd: rotate membership: generate fresh root: %w", err)
	}
	nextID := g.CurrentEpoch.EpochID + 1
	nextEpoch, err := epochFromRoot(freshRoot, nextID)
	if err != nil {
		return err
	}
	prev := append([]*EpochState{g.CurrentEpoch}, g.PrevEpochs...)
	if len(prev) > LegacyWindowSize {
		prev = prev[:LegacyWindowSize]
	}
	g.PrevEpochs = prev
	g.CurrentEpoch = nextEpoch
	return nil
}

func epochFromRoot(root []byte, epochID uint64) (*EpochState, error) {
	return &EpochState{
		EpochID:    epochID,
		EpochRoot:  root,
		SenderKeys: make(map[string][]byte),
		StartedAt:  time.Now().UnixMilli(),
	}, nil
}

func needsRotation(e *EpochState) bool {
	if e.MessageCount >= MaxEpochMessages {
		return true
	}
	return time.Now().UnixMilli()-e.StartedAt > EpochTTL.Milliseconds()
}

func findEpoch(g *GroupState, epochID uint64) *EpochState {
	if g.CurrentEpoch.EpochID == epochID {
		return g.CurrentEpoch
	}
	for _, e := range g.PrevEpochs {
		if e.EpochID == epochID {
			return e
		}
	}
	return nil
}

func epochAAD(epochID uint64, senderID string) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], epochID)
	return append(buf[:], []byte(senderID)...)
}

// MarshalState serializes the group state to JSON.
func MarshalState(g *GroupState) ([]byte, error) { return json.Marshal(g) }

// UnmarshalState deserializes a group state from JSON.
func UnmarshalState(data []byte) (*GroupState, error) {
	var g GroupState
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}
