// Package tree implements Tree mode: epoch-based group E2EE (hybrid MLS ciphersuite 0xFF01).
// Source: docs/spec/v0.1/12-mode-tree.md
package tree

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

const (
	MaxMembers           = 50
	MaxEpochMessages     = 1000
	EpochTTL             = 7 * 24 * time.Hour
	LegacyWindowSize     = 2
	CiphersuiteMagic     = 0xFF01 // spec 12 §3 ciphersuite ID
)

var (
	ErrGroupFull       = errors.New("tree: group full (50 members max)")
	ErrNotMember       = errors.New("tree: peer is not a member")
	ErrAlreadyMember   = errors.New("tree: peer is already a member")
	ErrEpochExpired    = errors.New("tree: epoch expired (not in legacy window)")
	ErrDecryptFailed   = errors.New("tree: AEAD decrypt failed")
	ErrGroupDisbanded  = errors.New("tree: group disbanded (0 members)")
)

// Member represents a group participant.
type Member struct {
	PeerID    string `json:"peer_id"`
	EdPub     []byte `json:"ed_pub"`    // Ed25519 public key
	MLDSAPub  []byte `json:"mldsa_pub"` // ML-DSA-65 public key (may be nil for basic ops)
	JoinedAt  int64  `json:"joined_at"` // epoch when joined
}

// EpochState holds key material for a single epoch.
type EpochState struct {
	EpochID      uint64 `json:"epoch_id"`
	EpochKey     []byte `json:"epoch_key"`    // 32 bytes, AES-128 uses first 16
	SenderKey    []byte `json:"sender_key"`   // 32 bytes per sender (simplified)
	MessageCount uint32 `json:"msg_count"`
	StartedAt    int64  `json:"started_at"` // Unix milliseconds
}

// GroupState is the full mutable state of a Tree mode group.
type GroupState struct {
	GroupID      string        `json:"group_id"`
	CurrentEpoch *EpochState   `json:"current_epoch"`
	PrevEpochs   []*EpochState `json:"prev_epochs"` // up to LegacyWindowSize
	Members      []Member      `json:"members"`
	RootKey      []byte        `json:"root_key"` // 32 bytes, evolves with each epoch
	// MLSEpoch holds the MLS key schedule secrets for this epoch.
	// Set by ApplyWelcome or DeriveEpochSecrets on commit. May be nil for
	// groups that have not yet performed an MLS-style epoch derivation.
	MLSEpoch *EpochSecrets `json:"mls_epoch,omitempty"`
}

// Ciphertext is a Tree mode encrypted message.
type Ciphertext struct {
	EpochID    uint64 `json:"epoch_id"`
	SenderID   string `json:"sender_id"`
	Nonce      []byte `json:"nonce"`   // 12 bytes
	CT         []byte `json:"ct"`      // AES-128-GCM ciphertext + 16-byte tag
}

// Commit is the result of an epoch rotation (add/remove/rekey).
type Commit struct {
	EpochID    uint64   `json:"epoch_id"`
	AddedPeers []string `json:"added,omitempty"`
	RemovedPeers []string `json:"removed,omitempty"`
	// In full MLS this would carry TreeKEM welcome data; here we carry the encrypted epoch key
	// per-member (simplified: encrypted with sender's key agreement key).
	EncryptedEpochKeys map[string][]byte `json:"encrypted_epoch_keys,omitempty"`
}

// NewGroup creates a new group with a single creator member.
func NewGroup(groupID string, creator Member) (*GroupState, error) {
	rootKey, err := v0crypto.RandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("tree: new group root key: %w", err)
	}
	epoch, err := deriveEpoch(rootKey, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("tree: derive epoch 0: %w", err)
	}
	creator.JoinedAt = 0
	return &GroupState{
		GroupID:      groupID,
		CurrentEpoch: epoch,
		Members:      []Member{creator},
		RootKey:      rootKey,
	}, nil
}

// AddMember adds a peer to the group and rotates the epoch.
func AddMember(g *GroupState, member Member) (*Commit, error) {
	if len(g.Members) >= MaxMembers {
		return nil, ErrGroupFull
	}
	for _, m := range g.Members {
		if m.PeerID == member.PeerID {
			return nil, ErrAlreadyMember
		}
	}
	nextEpoch := g.CurrentEpoch.EpochID + 1
	member.JoinedAt = int64(nextEpoch)
	g.Members = append(g.Members, member)
	return rotateEpoch(g, nextEpoch, []string{member.PeerID}, nil)
}

// RemoveMember removes a peer and rotates the epoch. Removed peer cannot decrypt future messages.
func RemoveMember(g *GroupState, peerID string) (*Commit, error) {
	found := false
	newMembers := g.Members[:0]
	for _, m := range g.Members {
		if m.PeerID == peerID {
			found = true
		} else {
			newMembers = append(newMembers, m)
		}
	}
	if !found {
		return nil, ErrNotMember
	}
	g.Members = newMembers
	if len(g.Members) == 0 {
		return nil, ErrGroupDisbanded
	}
	nextEpoch := g.CurrentEpoch.EpochID + 1
	return rotateEpoch(g, nextEpoch, nil, []string{peerID})
}

// Encrypt encrypts plaintext with the current epoch key (AES-128-GCM).
// Rotates epoch if message limit or TTL is reached.
func Encrypt(g *GroupState, senderID string, plaintext []byte) (*Ciphertext, *Commit, error) {
	var commit *Commit
	// Check if epoch rotation is needed.
	if needsRotation(g.CurrentEpoch) {
		nextEpoch := g.CurrentEpoch.EpochID + 1
		var err error
		commit, err = rotateEpoch(g, nextEpoch, nil, nil)
		if err != nil {
			return nil, nil, err
		}
	}

	nonce, err := v0crypto.RandomBytes(12)
	if err != nil {
		return nil, nil, fmt.Errorf("tree: nonce: %w", err)
	}
	var key16 [16]byte
	copy(key16[:], g.CurrentEpoch.EpochKey[:16])
	var nonce12 [12]byte
	copy(nonce12[:], nonce)
	aad := epochAAD(g.CurrentEpoch.EpochID, senderID)
	ct, err := v0crypto.SealAES128GCM(key16, nonce12, plaintext, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("tree: encrypt: %w", err)
	}
	g.CurrentEpoch.MessageCount++

	return &Ciphertext{
		EpochID:  g.CurrentEpoch.EpochID,
		SenderID: senderID,
		Nonce:    nonce,
		CT:       ct,
	}, commit, nil
}

// Decrypt decrypts a ciphertext, checking the current and legacy epochs.
func Decrypt(g *GroupState, ct *Ciphertext) ([]byte, error) {
	epoch := findEpoch(g, ct.EpochID)
	if epoch == nil {
		return nil, ErrEpochExpired
	}
	var key16 [16]byte
	copy(key16[:], epoch.EpochKey[:16])
	var nonce12 [12]byte
	copy(nonce12[:], ct.Nonce)
	aad := epochAAD(ct.EpochID, ct.SenderID)
	pt, err := v0crypto.OpenAES128GCM(key16, nonce12, ct.CT, aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return pt, nil
}

// IsMember checks whether peerID is currently a member.
func IsMember(g *GroupState, peerID string) bool {
	for _, m := range g.Members {
		if m.PeerID == peerID {
			return true
		}
	}
	return false
}

// --- helpers ---

func needsRotation(e *EpochState) bool {
	if e.MessageCount >= MaxEpochMessages {
		return true
	}
	age := time.Now().UnixMilli() - e.StartedAt
	return age > EpochTTL.Milliseconds()
}

func rotateEpoch(g *GroupState, newEpochID uint64, added, removed []string) (*Commit, error) {
	// Evolve root key: HKDF(root_key, epoch_nonce, "xorein/tree/v1/epoch-root", 32)
	var epochNonce [8]byte
	binary.BigEndian.PutUint64(epochNonce[:], newEpochID)
	newRoot, err := v0crypto.DeriveKey(g.RootKey, epochNonce[:], v0crypto.LabelTreeEpochRoot, 32)
	if err != nil {
		return nil, fmt.Errorf("tree: derive root: %w", err)
	}

	newEpoch, err := deriveEpoch(newRoot, newEpochID, epochNonce[:])
	if err != nil {
		return nil, fmt.Errorf("tree: derive epoch: %w", err)
	}

	// Keep legacy window.
	prev := append([]*EpochState{g.CurrentEpoch}, g.PrevEpochs...)
	if len(prev) > LegacyWindowSize {
		prev = prev[:LegacyWindowSize]
	}
	g.PrevEpochs = prev
	g.CurrentEpoch = newEpoch
	copy(g.RootKey, newRoot)

	return &Commit{
		EpochID:      newEpochID,
		AddedPeers:   added,
		RemovedPeers: removed,
	}, nil
}

func deriveEpoch(rootKey []byte, epochID uint64, salt []byte) (*EpochState, error) {
	epochKey, err := v0crypto.DeriveKey(rootKey, salt, v0crypto.LabelTreeExporter, 32)
	if err != nil {
		return nil, err
	}
	return &EpochState{
		EpochID:   epochID,
		EpochKey:  epochKey,
		StartedAt: time.Now().UnixMilli(),
	}, nil
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
func MarshalState(g *GroupState) ([]byte, error) {
	return json.Marshal(g)
}

// UnmarshalState deserializes a group state from JSON.
func UnmarshalState(data []byte) (*GroupState, error) {
	var g GroupState
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}
