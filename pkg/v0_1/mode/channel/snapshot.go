// snapshot.go implements Channel mode history snapshots per spec 14 §3.
// Source: docs/spec/v0.1/14-mode-channel.md §3
package channel

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

var (
	// ErrSnapshotSig is returned when a history snapshot hybrid signature fails.
	ErrSnapshotSig = errors.New("channel: snapshot signature verification failed")
	// ErrSnapshotMerkle is returned when the snapshot Merkle root does not match the message hashes.
	ErrSnapshotMerkle = errors.New("channel: snapshot Merkle root mismatch")
)

// HashedMessage is a message ID paired with its SHA-256 hash.
type HashedMessage struct {
	ID   string `json:"id"`
	Hash []byte `json:"hash"` // SHA-256 of the raw ciphertext bytes
}

// HistorySnapshot is a signed snapshot of a channel's message history per spec 14 §3.1.
type HistorySnapshot struct {
	ScopeID      string          `json:"scope_id"`
	EpochID      uint64          `json:"epoch_id"`
	FromSeq      uint64          `json:"from_seq"`
	ToSeq        uint64          `json:"to_seq"`
	MessageCount uint64          `json:"message_count"`
	Messages     []HashedMessage `json:"message_hashes"`
	// MerkleRoot is the root of the Merkle tree built from Messages[*].Hash.
	MerkleRoot []byte `json:"snapshot_root"`
	CreatedAt  string `json:"created_at"` // RFC3339Nano
	// Signature is the hybrid signature (Ed25519 || ML-DSA-65) over the canonical form.
	Signature []byte `json:"signature"`
}

// BuildHistorySnapshot builds a HistorySnapshot for the given messages (raw ciphertext bytes)
// and signs it with the owner/archivist's hybrid identity keys.
//
// messages is ordered: messages[0] corresponds to fromSeq, messages[len-1] to toSeq.
// scopeID identifies the channel. epochID is the current epoch.
// signerEdPriv is the Ed25519 private key; signerMLDSAPriv is the ML-DSA-65 private key.
func BuildHistorySnapshot(
	scopeID string,
	epochID uint64,
	fromSeq uint64,
	messages [][]byte,
	signerEdPriv ed25519.PrivateKey,
	signerMLDSAPriv []byte,
) (*HistorySnapshot, error) {
	// Build HashedMessage list.
	hashed := make([]HashedMessage, len(messages))
	for i, msg := range messages {
		h := sha256.Sum256(msg)
		hashed[i] = HashedMessage{
			ID:   fmt.Sprintf("%s/seq/%d", scopeID, fromSeq+uint64(i)),
			Hash: h[:],
		}
	}

	// Compute Merkle root.
	merkleRoot := buildMerkleRoot(hashed)

	toSeq := fromSeq
	if len(messages) > 0 {
		toSeq = fromSeq + uint64(len(messages)) - 1
	}

	snap := &HistorySnapshot{
		ScopeID:      scopeID,
		EpochID:      epochID,
		FromSeq:      fromSeq,
		ToSeq:        toSeq,
		MessageCount: uint64(len(messages)),
		Messages:     hashed,
		MerkleRoot:   merkleRoot,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339Nano),
	}

	canonical, err := snapshotSigningBytes(snap)
	if err != nil {
		return nil, fmt.Errorf("channel: snapshot sign: %w", err)
	}
	sig, err := v0crypto.HybridSign(signerEdPriv, signerMLDSAPriv, canonical)
	if err != nil {
		return nil, fmt.Errorf("channel: snapshot sign: %w", err)
	}
	snap.Signature = sig
	return snap, nil
}

// VerifyHistorySnapshot verifies the hybrid signature and Merkle root of a HistorySnapshot.
// Returns ErrSnapshotSig on signature failure or ErrSnapshotMerkle on Merkle mismatch.
func VerifyHistorySnapshot(snap *HistorySnapshot, signerEdPub ed25519.PublicKey, signerMLDSAPub []byte) error {
	// Verify Merkle root.
	recomputed := buildMerkleRoot(snap.Messages)
	if !bytesEqualSlice(recomputed, snap.MerkleRoot) {
		return ErrSnapshotMerkle
	}

	// Verify hybrid signature.
	canonical, err := snapshotSigningBytes(snap)
	if err != nil {
		return fmt.Errorf("channel: snapshot verify: %w", err)
	}
	if err := v0crypto.HybridVerify(signerEdPub, signerMLDSAPub, canonical, snap.Signature); err != nil {
		return ErrSnapshotSig
	}
	return nil
}

// MarshalSnapshot serializes a HistorySnapshot to JSON.
func MarshalSnapshot(s *HistorySnapshot) ([]byte, error) { return json.Marshal(s) }

// UnmarshalSnapshot deserializes a HistorySnapshot from JSON.
func UnmarshalSnapshot(data []byte) (*HistorySnapshot, error) {
	var s HistorySnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// --- helpers ---

// buildMerkleRoot builds a simple binary Merkle tree over the SHA-256 hashes of each message
// and returns the root hash. Uses SHA-256(left || right) for internal nodes.
// An empty set returns a zero-filled 32-byte slice.
func buildMerkleRoot(msgs []HashedMessage) []byte {
	if len(msgs) == 0 {
		return make([]byte, 32)
	}
	// Leaf layer: each leaf is the SHA-256 hash of the message.
	layer := make([][]byte, len(msgs))
	for i, m := range msgs {
		if len(m.Hash) == 32 {
			layer[i] = append([]byte(nil), m.Hash...)
		} else {
			h := sha256.Sum256(m.Hash)
			layer[i] = h[:]
		}
	}
	// Iteratively combine until one root remains.
	for len(layer) > 1 {
		var next [][]byte
		for i := 0; i+1 < len(layer); i += 2 {
			combined := append(append([]byte(nil), layer[i]...), layer[i+1]...)
			h := sha256.Sum256(combined)
			next = append(next, h[:])
		}
		if len(layer)%2 == 1 {
			// Odd node: duplicate the last leaf.
			last := layer[len(layer)-1]
			combined := append(append([]byte(nil), last...), last...)
			h := sha256.Sum256(combined)
			next = append(next, h[:])
		}
		layer = next
	}
	return layer[0]
}

// snapshotSigningBytes returns the canonical byte form of a HistorySnapshot for signing.
// All fields except Signature are included.
func snapshotSigningBytes(s *HistorySnapshot) ([]byte, error) {
	var buf []byte
	// scope_id
	buf = appendSnapshotField(buf, []byte(s.ScopeID))
	// epoch_id (8 BE)
	var epochBuf [8]byte
	binary.BigEndian.PutUint64(epochBuf[:], s.EpochID)
	buf = append(buf, epochBuf[:]...)
	// from_seq, to_seq, message_count (8 BE each)
	var u64Buf [8]byte
	binary.BigEndian.PutUint64(u64Buf[:], s.FromSeq)
	buf = append(buf, u64Buf[:]...)
	binary.BigEndian.PutUint64(u64Buf[:], s.ToSeq)
	buf = append(buf, u64Buf[:]...)
	binary.BigEndian.PutUint64(u64Buf[:], s.MessageCount)
	buf = append(buf, u64Buf[:]...)
	// merkle_root (32 bytes)
	buf = append(buf, s.MerkleRoot...)
	// message_hashes: each hash as hex string (for human-readability in the signed form)
	for _, m := range s.Messages {
		buf = appendSnapshotField(buf, []byte(m.ID))
		buf = append(buf, []byte(hex.EncodeToString(m.Hash))...)
		buf = append(buf, 0) // NUL separator
	}
	// created_at
	buf = appendSnapshotField(buf, []byte(s.CreatedAt))
	return buf, nil
}

// appendSnapshotField appends a 2-byte big-endian length-prefixed field.
func appendSnapshotField(buf, data []byte) []byte {
	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(data)))
	buf = append(buf, lenBuf[:]...)
	return append(buf, data...)
}

// bytesEqualSlice returns true if a and b are byte-identical.
func bytesEqualSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
