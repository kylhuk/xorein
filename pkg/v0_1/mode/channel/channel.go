// Package channel implements Channel mode: single-writer broadcast E2EE.
// Source: docs/spec/v0.1/14-mode-channel.md
package channel

import (
	"crypto/ed25519"
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
	LegacyWindowSize = 2
)

var (
	ErrNotWriter      = errors.New("channel: only the channel writer can encrypt")
	ErrEpochExpired   = errors.New("channel: epoch expired (not in legacy window)")
	ErrDecryptFailed  = errors.New("channel: AEAD decrypt failed")
	ErrBadMsgSig      = errors.New("channel: message signature verification failed")
)

// msgSignCanonicalPrefix is the domain-separation prefix per spec 14 §2.2.
const msgSignCanonicalPrefix = "xorein/channel/v1/msg-sign"

// EpochState holds key material for a single channel epoch.
type EpochState struct {
	EpochID      uint64 `json:"epoch_id"`
	EpochRoot    []byte `json:"epoch_root"`   // 32 bytes
	SenderKey    []byte `json:"sender_key"`   // 32 bytes (writer's key)
	MessageCount uint32 `json:"msg_count"`
	StartedAt    int64  `json:"started_at"`
}

// ChannelState is the mutable state of a Channel mode scope.
type ChannelState struct {
	ScopeID      string        `json:"scope_id"`
	WriterID     string        `json:"writer_id"` // only peer who can encrypt
	CurrentEpoch *EpochState   `json:"current_epoch"`
	PrevEpochs   []*EpochState `json:"prev_epochs"`
	// WriterEdPriv is the writer's Ed25519 private key for signing messages.
	// Not serialized (caller must set on load).
	WriterEdPriv ed25519.PrivateKey `json:"-"`
	// WriterMLDSAPriv is the writer's ML-DSA-65 private key for signing messages.
	// Not serialized (caller must set on load).
	WriterMLDSAPriv []byte `json:"-"`
	// WriterEdPub is the writer's Ed25519 public key for verifying messages.
	WriterEdPub ed25519.PublicKey `json:"writer_ed_pub,omitempty"`
	// WriterMLDSAPub is the writer's ML-DSA-65 public key for verifying messages.
	WriterMLDSAPub []byte `json:"writer_mldsa_pub,omitempty"`
}

// Ciphertext is a Channel mode encrypted message.
type Ciphertext struct {
	EpochID uint64 `json:"epoch_id"`
	ScopeID string `json:"scope_id,omitempty"` // channel/server scope for msg signature
	Nonce   []byte `json:"nonce"`              // 12 bytes
	CT      []byte `json:"ct"`                 // ChaCha20-Poly1305 ciphertext + tag
	// MsgSig is the hybrid signature over the canonical message form per spec 14 §2.2.
	// Empty for legacy messages (verification skipped for backward compatibility).
	MsgSig []byte `json:"msg_sig,omitempty"`
}

// NewChannel creates a new Channel mode scope with a random initial epoch.
func NewChannel(scopeID, writerID string) (*ChannelState, error) {
	root, err := v0crypto.RandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("channel: new: %w", err)
	}
	epoch, err := epochFromRoot(root, 0, writerID)
	if err != nil {
		return nil, err
	}
	return &ChannelState{
		ScopeID:      scopeID,
		WriterID:     writerID,
		CurrentEpoch: epoch,
	}, nil
}

// DeriveSenderKey derives the writer's sender key for a channel epoch.
func DeriveSenderKey(epochRoot []byte, writerID string) ([]byte, error) {
	info := v0crypto.LabelChannelSenderKey + writerID
	return v0crypto.DeriveKey(epochRoot, nil, info, 32)
}

// Encrypt encrypts a message. Only the channel writer can call this.
// If the ChannelState has WriterEdPriv and WriterMLDSAPriv set, the ciphertext
// will carry a hybrid message signature (MsgSig) per spec 14 §2.2.
func Encrypt(c *ChannelState, senderID string, plaintext []byte) (*Ciphertext, error) {
	if senderID != c.WriterID {
		return nil, ErrNotWriter
	}
	if needsRotation(c.CurrentEpoch) {
		if err := rotateEpoch(c); err != nil {
			return nil, err
		}
	}
	nonce, err := v0crypto.RandomBytes(12)
	if err != nil {
		return nil, fmt.Errorf("channel: nonce: %w", err)
	}
	var sk32 [32]byte
	copy(sk32[:], c.CurrentEpoch.SenderKey)
	var nonce12 [12]byte
	copy(nonce12[:], nonce)
	aad := epochAAD(c.CurrentEpoch.EpochID)
	ct, err := v0crypto.SealChaCha20Poly1305(sk32, nonce12, plaintext, aad)
	if err != nil {
		return nil, fmt.Errorf("channel: encrypt: %w", err)
	}
	c.CurrentEpoch.MessageCount++

	ciphertext := &Ciphertext{
		EpochID: c.CurrentEpoch.EpochID,
		ScopeID: c.ScopeID,
		Nonce:   nonce,
		CT:      ct,
	}

	// Sign the message if identity keys are available per spec 14 §2.2.
	if len(c.WriterEdPriv) > 0 && len(c.WriterMLDSAPriv) > 0 {
		canonical := msgSignCanonical(c.CurrentEpoch.EpochID, c.ScopeID, ct)
		sig, sigErr := v0crypto.HybridSign(c.WriterEdPriv, c.WriterMLDSAPriv, canonical)
		if sigErr != nil {
			return nil, fmt.Errorf("channel: sign message: %w", sigErr)
		}
		ciphertext.MsgSig = sig
	}

	return ciphertext, nil
}

// Decrypt decrypts a channel ciphertext.
// If ct.MsgSig is non-empty and ChannelState has WriterEdPub/WriterMLDSAPub set,
// the hybrid signature is verified before returning plaintext.
// Legacy messages (MsgSig == nil) skip signature verification for backward compatibility.
func Decrypt(c *ChannelState, ct *Ciphertext) ([]byte, error) {
	epoch := findEpoch(c, ct.EpochID)
	if epoch == nil {
		return nil, ErrEpochExpired
	}
	var sk32 [32]byte
	copy(sk32[:], epoch.SenderKey)
	var nonce12 [12]byte
	copy(nonce12[:], ct.Nonce)
	aad := epochAAD(ct.EpochID)
	pt, err := v0crypto.OpenChaCha20Poly1305(sk32, nonce12, ct.CT, aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	// Verify MsgSig if present per spec 14 §2.2.
	// Skip verification for legacy messages (no MsgSig) for backward compatibility.
	if len(ct.MsgSig) > 0 && len(c.WriterEdPub) > 0 && len(c.WriterMLDSAPub) > 0 {
		canonical := msgSignCanonical(ct.EpochID, ct.ScopeID, ct.CT)
		if verErr := v0crypto.HybridVerify(c.WriterEdPub, c.WriterMLDSAPub, canonical, ct.MsgSig); verErr != nil {
			return nil, ErrBadMsgSig
		}
	}

	return pt, nil
}

// msgSignCanonical builds the canonical byte sequence for channel message signing per spec 14 §2.2:
//
//	canonical = "xorein/channel/v1/msg-sign" || epoch_id (8 B BE) || scope_id (len-prefixed UTF-8) || body_ciphertext
func msgSignCanonical(epochID uint64, scopeID string, bodyCiphertext []byte) []byte {
	var buf []byte
	buf = append(buf, []byte(msgSignCanonicalPrefix)...)
	var epochBuf [8]byte
	binary.BigEndian.PutUint64(epochBuf[:], epochID)
	buf = append(buf, epochBuf[:]...)
	// scope_id as 2-byte len-prefixed UTF-8
	var scopeLen [2]byte
	binary.BigEndian.PutUint16(scopeLen[:], uint16(len(scopeID)))
	buf = append(buf, scopeLen[:]...)
	buf = append(buf, []byte(scopeID)...)
	buf = append(buf, bodyCiphertext...)
	return buf
}

func rotateEpoch(c *ChannelState) error {
	nextRoot, err := v0crypto.DeriveKey(c.CurrentEpoch.EpochRoot, nil, v0crypto.LabelCrowdEpochRoot, 32)
	if err != nil {
		return fmt.Errorf("channel: rotate root: %w", err)
	}
	nextID := c.CurrentEpoch.EpochID + 1
	nextEpoch, err := epochFromRoot(nextRoot, nextID, c.WriterID)
	if err != nil {
		return err
	}
	prev := append([]*EpochState{c.CurrentEpoch}, c.PrevEpochs...)
	if len(prev) > LegacyWindowSize {
		prev = prev[:LegacyWindowSize]
	}
	c.PrevEpochs = prev
	c.CurrentEpoch = nextEpoch
	return nil
}

func epochFromRoot(root []byte, epochID uint64, writerID string) (*EpochState, error) {
	sk, err := DeriveSenderKey(root, writerID)
	if err != nil {
		return nil, err
	}
	return &EpochState{
		EpochID:   epochID,
		EpochRoot: root,
		SenderKey: sk,
		StartedAt: time.Now().UnixMilli(),
	}, nil
}

func needsRotation(e *EpochState) bool {
	if e.MessageCount >= MaxEpochMessages {
		return true
	}
	return time.Now().UnixMilli()-e.StartedAt > EpochTTL.Milliseconds()
}

func findEpoch(c *ChannelState, epochID uint64) *EpochState {
	if c.CurrentEpoch.EpochID == epochID {
		return c.CurrentEpoch
	}
	for _, e := range c.PrevEpochs {
		if e.EpochID == epochID {
			return e
		}
	}
	return nil
}

func epochAAD(epochID uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], epochID)
	return buf[:]
}

// MarshalState serializes the channel state to JSON.
func MarshalState(c *ChannelState) ([]byte, error) { return json.Marshal(c) }

// UnmarshalState deserializes a channel state from JSON.
func UnmarshalState(data []byte) (*ChannelState, error) {
	var c ChannelState
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
