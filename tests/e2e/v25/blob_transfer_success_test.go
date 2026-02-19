package v25

import (
	"testing"

	"github.com/aether/code_aether/pkg/v25/blobproto"
)

func TestBlobTransferSuccess(t *testing.T) {
	provider := blobproto.NewProvider([]string{"application/octet-stream"}, []string{"fixed"})
	manifest := blobproto.Manifest{
		BlobID:     "blob-e2e",
		Size:       5,
		ChunkSize:  2,
		MimeType:   "application/octet-stream",
		Profile:    "fixed",
		Visibility: blobproto.VisibilityPublic,
	}
	if _, err := provider.PutManifest(blobproto.PutManifestRequest{Manifest: manifest}); err != nil {
		t.Fatalf("manifest upload failed: %v", err)
	}
	if _, err := provider.PutBlobChunk(blobproto.PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Data: []byte{0x10, 0x11}}); err != nil {
		t.Fatalf("chunk 0 failed: %v", err)
	}
	presence, err := provider.ChunkPresence(blobproto.ChunkPresenceRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Authorized: true})
	if err != nil {
		t.Fatalf("presence query failed: %v", err)
	}
	if !presence.Exists {
		t.Fatalf("expected first chunk to exist after upload")
	}
	if _, err := provider.PutBlobChunk(blobproto.PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 1, Data: []byte{0x12, 0x13}}); err != nil {
		t.Fatalf("chunk 1 failed: %v", err)
	}
	if _, err := provider.PutBlobChunk(blobproto.PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 2, Data: []byte{0x14}}); err != nil {
		t.Fatalf("chunk 2 failed: %v", err)
	}
	completed, err := provider.GetManifest(blobproto.GetManifestRequest{BlobID: manifest.BlobID})
	if err != nil {
		t.Fatalf("get manifest failed: %v", err)
	}
	if !completed.Completed {
		t.Fatalf("manifest should signal completion")
	}
}
