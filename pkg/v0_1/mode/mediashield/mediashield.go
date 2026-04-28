// Package mediashield implements MediaShield mode: SFrame-based E2EE for media frames.
// Source: docs/spec/v0.1/15-mode-mediashield.md
package mediashield

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

const (
	// MaxFrameCounter is 2^48 - 1; after this, rekey required (spec 15 §4).
	MaxFrameCounter = uint64(1<<48 - 1)
	// NonceSize is 12 bytes (AES-128-GCM / ChaCha20 nonce).
	NonceSize = 12
	// KIDSize is 8 bytes (first 8 bytes of SHA-256(peer_id)).
	KIDSize = 8
	// SFrameHeaderSize is KIDSize + 8 (CTR field, 8 bytes).
	SFrameHeaderSize = KIDSize + 8
)

var (
	ErrCounterOverflow = errors.New("mediashield: frame counter overflow (rekey required)")
	ErrDecryptFailed   = errors.New("mediashield: AEAD decrypt failed")
	ErrKIDMismatch     = errors.New("mediashield: KID in SFrame header does not match peer identity")
	ErrReplay          = errors.New("mediashield: frame counter replay detected")
)

// PeerKey holds a peer's MediaShield key and frame counter state.
type PeerKey struct {
	PeerID       string
	Key          []byte // 32 bytes (AES-128 uses first 16)
	FrameCounter uint64 // for EncryptFrame: next frame counter to use
	// MaxDecryptedCounter tracks the highest frame counter successfully decrypted.
	// Used for replay protection in DecryptFrame.
	// A value of 0 means no frames have been decrypted yet.
	MaxDecryptedCounter uint64
	// HasDecrypted is true after the first successful DecryptFrame call.
	HasDecrypted bool
}

// DeriveNonce derives a per-frame nonce from the mediashield key and frame counter.
// Implements spec 15 §3.1: HKDF-SHA-256(key, frame_counter_be8, "xorein/mediashield/v1/nonce", 12)
func DeriveNonce(key []byte, frameCounter uint64) ([]byte, error) {
	var salt [8]byte
	binary.BigEndian.PutUint64(salt[:], frameCounter)
	return v0crypto.DeriveKey(key, salt[:], v0crypto.LabelMediaShield+"/nonce", NonceSize)
}

// PeerKID returns the first 8 bytes of SHA-256(peerID), used as the SFrame Key ID.
func PeerKID(peerID string) []byte {
	h := sha256.Sum256([]byte(peerID))
	return h[:KIDSize]
}

// BuildSFrameHeader builds the SFrame header: KID (8 bytes) || CTR_BE8 (8 bytes).
// Simplified from RFC 9605 §4 for Xorein's fixed-width KID and CTR fields.
func BuildSFrameHeader(peerID string, frameCounter uint64) []byte {
	hdr := make([]byte, SFrameHeaderSize)
	copy(hdr[:KIDSize], PeerKID(peerID))
	binary.BigEndian.PutUint64(hdr[KIDSize:], frameCounter)
	return hdr
}

// EncryptFrame encrypts a media frame per spec 15 §3.3.
// Returns (sframe_header, ciphertext_with_tag, error).
func EncryptFrame(pk *PeerKey, rtpHeader, plaintext []byte) (sframeHeader []byte, ct []byte, err error) {
	if pk.FrameCounter > MaxFrameCounter {
		return nil, nil, ErrCounterOverflow
	}
	nonce, err := DeriveNonce(pk.Key, pk.FrameCounter)
	if err != nil {
		return nil, nil, fmt.Errorf("mediashield: derive nonce: %w", err)
	}
	sframeHeader = BuildSFrameHeader(pk.PeerID, pk.FrameCounter)
	aad := append(sframeHeader, rtpHeader...)

	var key16 [16]byte
	copy(key16[:], pk.Key[:16])
	var nonce12 [12]byte
	copy(nonce12[:], nonce)

	ct, err = v0crypto.SealAES128GCM(key16, nonce12, plaintext, aad)
	if err != nil {
		return nil, nil, fmt.Errorf("mediashield: encrypt: %w", err)
	}
	pk.FrameCounter++
	return sframeHeader, ct, nil
}

// DecryptFrame decrypts an SFrame ciphertext.
// sframeHeader must be KIDSize+8 bytes.
//
// Validates that the KID embedded in the SFrame header matches pk.PeerID.
// Returns ErrKIDMismatch if the KID does not match.
//
// Enforces replay protection: if the frame counter in the header is less than
// or equal to the highest previously decrypted frame counter, returns ErrReplay.
func DecryptFrame(pk *PeerKey, rtpHeader, sframeHeader, ctWithTag []byte) ([]byte, error) {
	if len(sframeHeader) != SFrameHeaderSize {
		return nil, fmt.Errorf("mediashield: bad header size %d", len(sframeHeader))
	}

	// Validate KID: first KIDSize bytes of sframeHeader must match PeerKID(pk.PeerID).
	expectedKID := PeerKID(pk.PeerID)
	for i := 0; i < KIDSize; i++ {
		if sframeHeader[i] != expectedKID[i] {
			return nil, ErrKIDMismatch
		}
	}

	frameCounter := binary.BigEndian.Uint64(sframeHeader[KIDSize:])

	// Replay protection: reject any frame with counter <= max already decrypted.
	if pk.HasDecrypted && frameCounter <= pk.MaxDecryptedCounter {
		return nil, ErrReplay
	}

	nonce, err := DeriveNonce(pk.Key, frameCounter)
	if err != nil {
		return nil, fmt.Errorf("mediashield: derive nonce: %w", err)
	}
	aad := append(sframeHeader, rtpHeader...)

	var key16 [16]byte
	copy(key16[:], pk.Key[:16])
	var nonce12 [12]byte
	copy(nonce12[:], nonce)

	pt, err := v0crypto.OpenAES128GCM(key16, nonce12, ctWithTag, aad)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	// Update replay state.
	pk.MaxDecryptedCounter = frameCounter
	pk.HasDecrypted = true

	return pt, nil
}

// NeedsRekey returns true when the frame counter is about to overflow.
func NeedsRekey(pk *PeerKey) bool {
	return pk.FrameCounter >= MaxFrameCounter
}
