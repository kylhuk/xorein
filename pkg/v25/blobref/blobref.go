package blobref

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RefusalCode represents a deterministic refusal reason for blob references and manifests.
type RefusalCode string

const (
	RefusalCodeMissingField           RefusalCode = "missing_field"
	RefusalCodeUnsupportedAlgorithm   RefusalCode = "unsupported_algorithm"
	RefusalCodeInvalidSize            RefusalCode = "invalid_size"
	RefusalCodeInvalidChunkSize       RefusalCode = "invalid_chunk_size"
	RefusalCodeChunkSizeMismatch      RefusalCode = "chunk_size_mismatch"
	RefusalCodeInvalidChunk           RefusalCode = "invalid_chunk"
	RefusalCodeChunkSumMismatch       RefusalCode = "chunk_sum_mismatch"
	RefusalCodeInvalidMetadataPointer RefusalCode = "invalid_metadata_pointer"
)

// RefusalError wraps a RefusalCode with a human-readable reason.
type RefusalError struct {
	Code   RefusalCode
	Reason string
}

func (e RefusalError) Error() string {
	if e.Code == "" && e.Reason == "" {
		return "refusal"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

func newRefusalError(code RefusalCode, reason string) *RefusalError {
	return &RefusalError{Code: code, Reason: reason}
}

var supportedHashAlgorithms = map[string]struct{}{
	"BLAKE3":  {},
	"SHA-256": {},
}

// BlobRef is the canonical metadata descriptor for a blob payload.
type BlobRef struct {
	HashAlgorithm            string             `json:"hashAlgorithm"`
	ContentHash              string             `json:"contentHash"`
	Size                     int64              `json:"size"`
	MimeType                 string             `json:"mimeType"`
	ChunkSize                int64              `json:"chunkSize"`
	ChunkProfile             string             `json:"chunkProfile"`
	EncryptedMetadataPointer *string            `json:"encryptedMetadataPointer,omitempty"`
	Metadata                 MetadataExtensions `json:"metadata,omitempty"`
}

// ChunkDescriptor names a chunk within a manifest.
type ChunkDescriptor struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// Manifest describes how to reconstruct a chunked blob payload.
type Manifest struct {
	ContentHash  string             `json:"contentHash"`
	TotalSize    int64              `json:"totalSize"`
	ChunkSize    int64              `json:"chunkSize"`
	ChunkProfile string             `json:"chunkProfile"`
	Chunks       []ChunkDescriptor  `json:"chunks"`
	Metadata     MetadataExtensions `json:"metadata,omitempty"`
}

// MetadataExtensions stores extension metadata in a forward-compatible way.
type MetadataExtensions map[string]json.RawMessage

// Set marshals the provided value and stores it under the given key.
func (me *MetadataExtensions) Set(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("extension key required")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if *me == nil {
		*me = MetadataExtensions{}
	}
	(*me)[key] = raw
	return nil
}

// Get decodes the stored value at key into dest.
func (me MetadataExtensions) Get(key string, dest interface{}) error {
	if me == nil {
		return fmt.Errorf("extension %q missing", key)
	}
	raw, ok := me[key]
	if !ok {
		return fmt.Errorf("extension %q missing", key)
	}
	return json.Unmarshal(raw, dest)
}

// ValidateBlobRef ensures the blob reference carries deterministic, reusable metadata.
func ValidateBlobRef(ref BlobRef) error {
	if ref.HashAlgorithm == "" {
		return newRefusalError(RefusalCodeMissingField, "hash algorithm is required")
	}
	if _, ok := supportedHashAlgorithms[strings.ToUpper(ref.HashAlgorithm)]; !ok {
		return newRefusalError(RefusalCodeUnsupportedAlgorithm, fmt.Sprintf("hash algorithm %q not allowed", ref.HashAlgorithm))
	}
	if ref.ContentHash == "" {
		return newRefusalError(RefusalCodeMissingField, "content hash is required")
	}
	if ref.Size < 0 {
		return newRefusalError(RefusalCodeInvalidSize, "size must be non-negative")
	}
	if ref.MimeType == "" {
		return newRefusalError(RefusalCodeMissingField, "mime type is required")
	}
	if ref.ChunkSize <= 0 {
		return newRefusalError(RefusalCodeInvalidChunkSize, "chunk size must be positive")
	}
	if ref.ChunkProfile == "" {
		return newRefusalError(RefusalCodeMissingField, "chunk profile is required")
	}
	if ref.EncryptedMetadataPointer != nil && strings.TrimSpace(*ref.EncryptedMetadataPointer) == "" {
		return newRefusalError(RefusalCodeInvalidMetadataPointer, "encrypted metadata pointer must not be empty")
	}
	return nil
}

// ValidateManifest ensures chunk correctness and deterministic size verification.
func ValidateManifest(man Manifest) error {
	if man.ContentHash == "" {
		return newRefusalError(RefusalCodeMissingField, "manifest content hash is required")
	}
	if man.TotalSize < 0 {
		return newRefusalError(RefusalCodeInvalidSize, "manifest total size must be non-negative")
	}
	if man.ChunkSize <= 0 {
		return newRefusalError(RefusalCodeInvalidChunkSize, "manifest chunk size must be positive")
	}
	if man.ChunkProfile == "" {
		return newRefusalError(RefusalCodeMissingField, "manifest chunk profile is required")
	}
	if len(man.Chunks) == 0 {
		return newRefusalError(RefusalCodeInvalidChunk, "manifest must reference at least one chunk")
	}

	var total int64
	for idx, chunk := range man.Chunks {
		if chunk.Hash == "" {
			return newRefusalError(RefusalCodeInvalidChunk, fmt.Sprintf("chunk %d missing hash", idx))
		}
		if chunk.Size <= 0 {
			return newRefusalError(RefusalCodeInvalidChunk, fmt.Sprintf("chunk %d size must be positive", idx))
		}
		if chunk.Size > man.ChunkSize {
			return newRefusalError(RefusalCodeChunkSizeMismatch, fmt.Sprintf("chunk %d exceeds manifest chunk size", idx))
		}
		total += chunk.Size
	}
	if total != man.TotalSize {
		return newRefusalError(RefusalCodeChunkSumMismatch, fmt.Sprintf("sum of chunk sizes %d does not match manifest size %d", total, man.TotalSize))
	}
	return nil
}
