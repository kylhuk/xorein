package blobref

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateBlobRef(t *testing.T) {
	ok := BlobRef{
		HashAlgorithm: "BLAKE3",
		ContentHash:   "deadbeef",
		Size:          1024,
		MimeType:      "application/octet-stream",
		ChunkSize:     512,
		ChunkProfile:  "fixed",
	}
	if err := ValidateBlobRef(ok); err != nil {
		t.Fatalf("valid reference rejected: %v", err)
	}

	tests := []struct {
		name string
		ref  BlobRef
		want RefusalCode
	}{
		{"missing algorithm", BlobRef{}, RefusalCodeMissingField},
		{"unsupported algorithm", BlobRef{HashAlgorithm: "MD5"}, RefusalCodeUnsupportedAlgorithm},
		{"missing hash", BlobRef{HashAlgorithm: "BLAKE3"}, RefusalCodeMissingField},
		{"negative size", BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "foo", MimeType: "a", ChunkSize: 1, ChunkProfile: "p", Size: -1}, RefusalCodeInvalidSize},
		{"missing mime", BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "foo", Size: 1, ChunkSize: 1, ChunkProfile: "p"}, RefusalCodeMissingField},
		{"invalid chunk size", BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "foo", Size: 1, MimeType: "a", ChunkProfile: "p", ChunkSize: 0}, RefusalCodeInvalidChunkSize},
		{"missing profile", BlobRef{HashAlgorithm: "BLAKE3", ContentHash: "foo", Size: 1, MimeType: "a", ChunkSize: 1}, RefusalCodeMissingField},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlobRef(tt.ref)
			if err == nil {
				t.Fatalf("expected refusal for %s", tt.name)
			}
			re, ok := err.(*RefusalError)
			if !ok {
				t.Fatalf("unexpected error type: %T", err)
			}
			if re.Code != tt.want {
				t.Fatalf("got %s want %s", re.Code, tt.want)
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	manifest := Manifest{
		ContentHash:  "beadface",
		TotalSize:    1024,
		ChunkSize:    512,
		ChunkProfile: "fixed",
		Chunks: []ChunkDescriptor{
			{Hash: "aaa", Size: 512},
			{Hash: "bbb", Size: 512},
		},
	}
	if err := ValidateManifest(manifest); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}

	tests := []struct {
		name string
		man  Manifest
		want RefusalCode
	}{
		{"missing content", Manifest{}, RefusalCodeMissingField},
		{"negative size", Manifest{ContentHash: "a", ChunkSize: 1, ChunkProfile: "p", TotalSize: -1, Chunks: []ChunkDescriptor{{Hash: "a", Size: 1}}}, RefusalCodeInvalidSize},
		{"invalid chunk size", Manifest{ContentHash: "a", TotalSize: 1, ChunkSize: 0, ChunkProfile: "p", Chunks: []ChunkDescriptor{{Hash: "a", Size: 1}}}, RefusalCodeInvalidChunkSize},
		{"missing profile", Manifest{ContentHash: "a", TotalSize: 1, ChunkSize: 1, Chunks: []ChunkDescriptor{{Hash: "a", Size: 1}}}, RefusalCodeMissingField},
		{"empty chunk list", Manifest{ContentHash: "a", TotalSize: 0, ChunkSize: 1, ChunkProfile: "p"}, RefusalCodeInvalidChunk},
		{"chunk without hash", Manifest{ContentHash: "a", TotalSize: 1, ChunkSize: 1, ChunkProfile: "p", Chunks: []ChunkDescriptor{{Hash: "", Size: 1}}}, RefusalCodeInvalidChunk},
		{"chunk exceeds size", Manifest{ContentHash: "a", TotalSize: 2, ChunkSize: 1, ChunkProfile: "p", Chunks: []ChunkDescriptor{{Hash: "a", Size: 2}}}, RefusalCodeChunkSizeMismatch},
		{"sum mismatch", Manifest{ContentHash: "a", TotalSize: 2, ChunkSize: 2, ChunkProfile: "p", Chunks: []ChunkDescriptor{{Hash: "a", Size: 1}}}, RefusalCodeChunkSumMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.man)
			if err == nil {
				t.Fatalf("expected refusal for %s", tt.name)
			}
			re, ok := err.(*RefusalError)
			if !ok {
				t.Fatalf("unexpected error type: %T", err)
			}
			if re.Code != tt.want {
				t.Fatalf("got %s want %s", re.Code, tt.want)
			}
		})
	}
}

func TestMetadataExtensionsRoundTrip(t *testing.T) {
	var ext MetadataExtensions
	if err := ext.Set("x", map[string]string{"foo": "bar"}); err != nil {
		t.Fatalf("set extension: %v", err)
	}
	manifest := Manifest{
		ContentHash:  "feedface",
		TotalSize:    1,
		ChunkSize:    1,
		ChunkProfile: "fixed",
		Chunks:       []ChunkDescriptor{{Hash: "c", Size: 1}},
		Metadata:     ext,
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var decoded Manifest
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if err := decoded.Metadata.Get("x", &map[string]string{}); err != nil {
		t.Fatalf("read extension: %v", err)
	}
	var payload map[string]string
	if err := decoded.Metadata.Get("x", &payload); err != nil {
		t.Fatalf("unmarshal extension payload: %v", err)
	}
	if payload["foo"] != "bar" {
		t.Fatalf("unexpected extension payload: %v", payload)
	}
	if err := ValidateManifest(decoded); err != nil {
		t.Fatalf("manifest validation should ignore metadata: %v", err)
	}
}

func TestMetadataExtensionsMissingKey(t *testing.T) {
	var ext MetadataExtensions
	err := ext.Get("absent", &map[string]string{})
	if err == nil {
		t.Fatalf("expected error for missing key")
	}
	if got := err.Error(); !strings.Contains(got, "absent") {
		t.Fatalf("unexpected error message: %s", got)
	}
}
