package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/blobref"
)

func requireRefusalCode(t *testing.T, err error, want blobref.RefusalCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected refusal")
	}
	re, ok := err.(*blobref.RefusalError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if re.Code != want {
		t.Fatalf("unexpected refusal code %s", re.Code)
	}
}

func TestBlobRefManifestValidation(t *testing.T) {
	manifest := blobref.Manifest{
		ContentHash:  "digest",
		TotalSize:    2,
		ChunkSize:    1,
		ChunkProfile: "fixed",
		Chunks: []blobref.ChunkDescriptor{
			{Hash: "a", Size: 1},
			{Hash: "b", Size: 1},
		},
	}
	if err := blobref.ValidateManifest(manifest); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

func TestBlobRefManifestRejectsChunkSumMismatch(t *testing.T) {
	manifest := blobref.Manifest{
		ContentHash:  "digest",
		TotalSize:    3,
		ChunkSize:    2,
		ChunkProfile: "fixed",
		Chunks: []blobref.ChunkDescriptor{
			{Hash: "a", Size: 1},
		},
	}
	err := blobref.ValidateManifest(manifest)
	requireRefusalCode(t, err, blobref.RefusalCodeChunkSumMismatch)
}
