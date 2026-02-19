package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrHistoryHeadInvalidSignature = errors.New("HISTORY_HEAD_INVALID_SIGNATURE")
	ErrManifestHashMismatch        = errors.New("MANIFEST_HASH_MISMATCH")
	ErrSegmentNotFound             = errors.New("SEGMENT_NOT_FOUND")
)

type HistorySegment struct {
	ID    string
	Start time.Time
	End   time.Time
	Hash  string
}

type HistorySegmentManifest struct {
	SpaceID      string
	ChannelID    string
	Segments     []HistorySegment
	ManifestHash string
}

type HistoryHead struct {
	SpaceID   string
	ChannelID string
	Manifest  HistorySegmentManifest
	Signature string
}

func (m HistorySegmentManifest) computeDeterministicHash() string {
	segments := make([]HistorySegment, len(m.Segments))
	copy(segments, m.Segments)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].ID < segments[j].ID
	})

	var builder strings.Builder
	builder.WriteString(m.SpaceID)
	builder.WriteString("|")
	builder.WriteString(m.ChannelID)
	builder.WriteString("|")
	for _, segment := range segments {
		builder.WriteString(segment.ID)
		builder.WriteString(":")
		builder.WriteString(segment.Hash)
		builder.WriteString(":")
		builder.WriteString(segment.Start.UTC().Format(time.RFC3339Nano))
		builder.WriteString("-")
		builder.WriteString(segment.End.UTC().Format(time.RFC3339Nano))
		builder.WriteString("|")
	}

	summary := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(summary[:])
}

func (m *HistorySegmentManifest) RegenerateHash() {
	m.ManifestHash = m.computeDeterministicHash()
}

func (m HistorySegmentManifest) VerifyHash() error {
	if len(m.ManifestHash) == 0 {
		return ErrManifestHashMismatch
	}

	if m.computeDeterministicHash() != m.ManifestHash {
		return ErrManifestHashMismatch
	}

	return nil
}

func (m HistorySegmentManifest) SegmentByID(id string) (HistorySegment, error) {
	for _, segment := range m.Segments {
		if segment.ID == id {
			return segment, nil
		}
	}

	return HistorySegment{}, ErrSegmentNotFound
}

func DeriveHeadSignature(manifestHash, membershipKey string) string {
	sum := sha256.Sum256([]byte(manifestHash + "|" + membershipKey))
	return hex.EncodeToString(sum[:])
}

func (h HistoryHead) VerifySignature(membershipKey string) error {
	if err := h.Manifest.VerifyHash(); err != nil {
		return err
	}

	expected := DeriveHeadSignature(h.Manifest.ManifestHash, membershipKey)
	if expected != h.Signature {
		return ErrHistoryHeadInvalidSignature
	}

	return nil
}
