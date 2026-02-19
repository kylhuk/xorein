package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ManifestSegment represents a discrete history segment entry in a manifest.
type ManifestSegment struct {
	ID    string
	Start time.Time
	End   time.Time
	Hash  string
}

// Manifest holds metadata about a history segment bundle and a deterministic hash.
type Manifest struct {
	SpaceID      string
	ChannelID    string
	Segments     []ManifestSegment
	ManifestHash string
}

// ManifestValidationCode enumerates all deterministic failure codes for manifest validation.
type ManifestValidationCode string

const (
	ManifestValidationCodeMissingSpaceID     ManifestValidationCode = "MANIFEST_MISSING_SPACE_ID"
	ManifestValidationCodeMissingChannelID   ManifestValidationCode = "MANIFEST_MISSING_CHANNEL_ID"
	ManifestValidationCodeNoSegments         ManifestValidationCode = "MANIFEST_NO_SEGMENTS"
	ManifestValidationCodeSegmentIDEmpty     ManifestValidationCode = "MANIFEST_SEGMENT_ID_EMPTY"
	ManifestValidationCodeSegmentHashMissing ManifestValidationCode = "MANIFEST_SEGMENT_HASH_EMPTY"
	ManifestValidationCodeSegmentTimeOrder   ManifestValidationCode = "MANIFEST_SEGMENT_TIME_ORDER_INVALID"
	ManifestValidationCodeHashMissing        ManifestValidationCode = "MANIFEST_HASH_MISSING"
	ManifestValidationCodeHashMismatch       ManifestValidationCode = "MANIFEST_HASH_MISMATCH"
)

// ManifestValidationError exposes deterministic codes for manifest validation failures.
type ManifestValidationError struct {
	Code ManifestValidationCode
	Err  error
}

func (e *ManifestValidationError) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Err.Error())
}

func (e *ManifestValidationError) Unwrap() error {
	return e.Err
}

func newManifestValidationError(code ManifestValidationCode, err error) error {
	if err == nil {
		err = errors.New(string(code))
	}
	return &ManifestValidationError{Code: code, Err: err}
}

// RegenerateHash recomputes and stores the deterministic manifest hash.
func (m *Manifest) RegenerateHash() {
	m.ManifestHash = m.computeDeterministicHash()
}

func (m Manifest) computeDeterministicHash() string {
	segments := make([]ManifestSegment, len(m.Segments))
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

// Validate ensures the manifest is well-formed and the deterministic hash matches.
func (m Manifest) Validate() error {
	if strings.TrimSpace(m.SpaceID) == "" {
		return newManifestValidationError(ManifestValidationCodeMissingSpaceID, errors.New("space identifier is required"))
	}

	if strings.TrimSpace(m.ChannelID) == "" {
		return newManifestValidationError(ManifestValidationCodeMissingChannelID, errors.New("channel identifier is required"))
	}

	if len(m.Segments) == 0 {
		return newManifestValidationError(ManifestValidationCodeNoSegments, errors.New("at least one segment is required"))
	}

	for _, segment := range m.Segments {
		if strings.TrimSpace(segment.ID) == "" {
			return newManifestValidationError(ManifestValidationCodeSegmentIDEmpty, errors.New("segment identifier is required"))
		}
		if strings.TrimSpace(segment.Hash) == "" {
			return newManifestValidationError(ManifestValidationCodeSegmentHashMissing, errors.New("segment hash is required"))
		}
		if !segment.Start.Before(segment.End) {
			return newManifestValidationError(ManifestValidationCodeSegmentTimeOrder, errors.New("segment start must precede end"))
		}
	}

	if strings.TrimSpace(m.ManifestHash) == "" {
		return newManifestValidationError(ManifestValidationCodeHashMissing, errors.New("manifest hash is required"))
	}

	if m.computeDeterministicHash() != m.ManifestHash {
		return newManifestValidationError(ManifestValidationCodeHashMismatch, errors.New("manifest hash mismatch"))
	}

	return nil
}

// Head represents the manifest plus a deterministic signature signed with a membership key.
type Head struct {
	SpaceID   string
	ChannelID string
	Manifest  Manifest
	Signature string
}

// HeadValidationCode enumerates deterministic failure codes for head validation.
type HeadValidationCode string

const (
	HeadValidationCodeMissingMembershipKey HeadValidationCode = "HEAD_MEMBERSHIP_KEY_MISSING"
	HeadValidationCodeMissingSignature     HeadValidationCode = "HEAD_SIGNATURE_MISSING"
	HeadValidationCodeSignatureMismatch    HeadValidationCode = "HEAD_SIGNATURE_MISMATCH"
	HeadValidationCodeInvalidManifest      HeadValidationCode = "HEAD_INVALID_MANIFEST"
	HeadValidationCodeSpaceMismatch        HeadValidationCode = "HEAD_SPACE_MISMATCH"
	HeadValidationCodeChannelMismatch      HeadValidationCode = "HEAD_CHANNEL_MISMATCH"
)

// HeadValidationError exposes deterministic codes for head validation failures.
type HeadValidationError struct {
	Code HeadValidationCode
	Err  error
}

func (e *HeadValidationError) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Err.Error())
}

func (e *HeadValidationError) Unwrap() error {
	return e.Err
}

func newHeadValidationError(code HeadValidationCode, err error) error {
	if err == nil {
		err = errors.New(string(code))
	}
	return &HeadValidationError{Code: code, Err: err}
}

// DeriveHeadSignature derives the deterministic signature for a manifest/head.
func DeriveHeadSignature(manifestHash, membershipKey string) string {
	sum := sha256.Sum256([]byte(manifestHash + "|" + membershipKey))
	return hex.EncodeToString(sum[:])
}

// Validate ensures the head metadata, manifest, and signature are well-formed.
func (h Head) Validate(membershipKey string) error {
	if strings.TrimSpace(membershipKey) == "" {
		return newHeadValidationError(HeadValidationCodeMissingMembershipKey, errors.New("membership key is required"))
	}

	if strings.TrimSpace(h.Signature) == "" {
		return newHeadValidationError(HeadValidationCodeMissingSignature, errors.New("head signature is required"))
	}

	if err := h.Manifest.Validate(); err != nil {
		return newHeadValidationError(HeadValidationCodeInvalidManifest, err)
	}

	if h.SpaceID != h.Manifest.SpaceID {
		return newHeadValidationError(HeadValidationCodeSpaceMismatch, errors.New("head space metadata must match manifest"))
	}

	if h.ChannelID != h.Manifest.ChannelID {
		return newHeadValidationError(HeadValidationCodeChannelMismatch, errors.New("head channel metadata must match manifest"))
	}

	expected := DeriveHeadSignature(h.Manifest.ManifestHash, membershipKey)
	if expected != h.Signature {
		return newHeadValidationError(HeadValidationCodeSignatureMismatch, errors.New("head signature mismatch"))
	}

	return nil
}
