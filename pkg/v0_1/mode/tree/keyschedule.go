// keyschedule.go implements the MLS epoch key schedule for the Xorein hybrid ciphersuite.
// Source: docs/spec/v0.1/12-mode-tree.md §4 and RFC 9420 §8
package tree

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	v0crypto "github.com/aether/code_aether/pkg/v0_1/crypto"
)

// Key schedule labels per spec 12 §4.
const (
	labelMLSHandshake    = "xorein/tree/v1/mls-handshake-key"
	labelMLSApplication  = "xorein/tree/v1/mls-application-key"
	labelMLSExporter     = "xorein/tree/v1/mls-exporter-secret"
	labelMLSMemberKey    = "xorein/tree/v1/mls-member-key"

	// MediaShield exporter label per spec 12 §3.
	LabelMediaShieldExporter = "xorein/mediashield/v1"
)

// EpochSecrets holds the per-epoch derived keys per spec 12 §4.
// All keys are 32 bytes.
type EpochSecrets struct {
	EpochID        uint64
	GroupID        string
	HandshakeKey   [32]byte // for MLS handshake (Commit/Welcome/Proposal framing)
	ApplicationKey [32]byte // for MLS ApplicationMessage AEAD
	ExporterSecret [32]byte // for MLS-Exporter / MediaShield binding (spec 12 §3)
}

// DeriveEpochSecrets derives the epoch key schedule from the commit secret and epoch metadata.
// Uses HKDF-SHA-256 with spec labels.
//
// commitSecret is the shared secret produced by TreeKEM UpdatePath / ProcessUpdatePath.
// epochID and groupID are bound into the derivation to domain-separate epochs.
func DeriveEpochSecrets(commitSecret []byte, epochID uint64, groupID string) (*EpochSecrets, error) {
	// Epoch context: SHA-256(groupID_bytes || epoch_id_be8)
	ctx := epochContext(epochID, groupID)

	hk, err := v0crypto.DeriveKey(commitSecret, ctx, labelMLSHandshake, 32)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: derive handshake key: %w", err)
	}
	ak, err := v0crypto.DeriveKey(commitSecret, ctx, labelMLSApplication, 32)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: derive application key: %w", err)
	}
	ek, err := v0crypto.DeriveKey(commitSecret, ctx, labelMLSExporter, 32)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: derive exporter secret: %w", err)
	}

	s := &EpochSecrets{
		EpochID: epochID,
		GroupID: groupID,
	}
	copy(s.HandshakeKey[:], hk)
	copy(s.ApplicationKey[:], ak)
	copy(s.ExporterSecret[:], ek)
	return s, nil
}

// MLSExporter derives an application secret from the epoch's exporter_secret per
// RFC 9420 §8.2 and spec 12 §3 (MLS-Exporter for MediaShield).
//
//	output = HKDF-SHA-256(
//	    IKM  = exporter_secret,
//	    salt = SHA-256(context),
//	    info = "xorein/tree/v1/mls-exporter-secret" || label,
//	    L    = length,
//	)
func (s *EpochSecrets) MLSExporter(label string, context []byte, length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("tree/mls: MLSExporter: length must be > 0, got %d", length)
	}
	// Per RFC 9420 §8.2: salt = H(context).
	ctxHash := sha256.Sum256(context)
	info := labelMLSExporter + label
	out, err := v0crypto.DeriveKey(s.ExporterSecret[:], ctxHash[:], info, length)
	if err != nil {
		return nil, fmt.Errorf("tree/mls: MLSExporter: %w", err)
	}
	return out, nil
}

// DeriveMediaShieldKey derives the MediaShield key from the MLS epoch per spec 12 §3:
//
//	mediashield_key = MLS-Exporter("xorein/mediashield/v1", b"", 32)
func (s *EpochSecrets) DeriveMediaShieldKey() ([]byte, error) {
	return s.MLSExporter(LabelMediaShieldExporter, []byte{}, 32)
}

// epochContext returns a fixed-size context for epoch secret derivation:
// SHA-256(group_id_utf8 || epoch_id_be8)
func epochContext(epochID uint64, groupID string) []byte {
	var epochBuf [8]byte
	binary.BigEndian.PutUint64(epochBuf[:], epochID)
	h := sha256.New()
	h.Write([]byte(groupID))
	h.Write(epochBuf[:])
	return h.Sum(nil)
}
