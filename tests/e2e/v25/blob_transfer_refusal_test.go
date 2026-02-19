package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/blobproto"
)

func TestBlobTransferRefusals(t *testing.T) {
	provider := blobproto.NewProvider([]string{"application/octet-stream"}, []string{"fixed"})
	manifest := blobproto.Manifest{
		BlobID:     "blob-e2e-refuse",
		Size:       2,
		ChunkSize:  2,
		MimeType:   "application/octet-stream",
		Profile:    "fixed",
		Visibility: blobproto.VisibilityPrivate,
	}
	if _, err := provider.PutManifest(blobproto.PutManifestRequest{Manifest: manifest}); err != nil {
		t.Fatalf("manifest upload failed: %v", err)
	}
	if _, err := provider.ChunkPresence(blobproto.ChunkPresenceRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Authorized: false}); err != nil {
		t.Fatalf("presence query failed: %v", err)
	}
	if _, err := provider.PutBlobChunk(blobproto.PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 1, Data: []byte{0}}); err == nil {
		t.Fatalf("expected invalid chunk order refusal")
	}
	if _, err := provider.PutBlobChunk(blobproto.PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Data: []byte{0x1}}); err == nil {
		t.Fatalf("expected chunk size mismatch refusal")
	}
}
