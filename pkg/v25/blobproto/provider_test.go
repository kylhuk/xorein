package blobproto

import (
	"testing"
)

func TestProviderSequentialResume(t *testing.T) {
	provider := NewProvider([]string{"image/png"}, []string{"fixed"})
	manifest := Manifest{
		BlobID:     "blob-seq",
		Size:       5,
		ChunkSize:  2,
		MimeType:   "image/png",
		Profile:    "fixed",
		Visibility: VisibilityPublic,
	}
	resp, err := provider.PutManifest(PutManifestRequest{Manifest: manifest})
	if err != nil {
		t.Fatalf("manifest upload failed: %v", err)
	}
	if resp.NextChunkIndex != 0 {
		t.Fatalf("expected next chunk 0, got %d", resp.NextChunkIndex)
	}
	if resp.ResumeToken != "blob-seq:0" {
		t.Fatalf("expected resume token %q, got %q", "blob-seq:0", resp.ResumeToken)
	}

	chunk0 := []byte{0x01, 0x02}
	chunk1 := []byte{0x03, 0x04}
	chunk2 := []byte{0x05}

	if _, err := provider.PutBlobChunk(PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Data: chunk0}); err != nil {
		t.Fatalf("first chunk failed: %v", err)
	}
	if _, err := provider.PutBlobChunk(PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 1, Data: chunk1}); err != nil {
		t.Fatalf("second chunk failed: %v", err)
	}
	chunkResp, err := provider.PutBlobChunk(PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 2, Data: chunk2})
	if err != nil {
		t.Fatalf("final chunk failed: %v", err)
	}
	if chunkResp.NextChunkIndex != 3 {
		t.Fatalf("expected next chunk 3, got %d", chunkResp.NextChunkIndex)
	}

	manifestResp, err := provider.GetManifest(GetManifestRequest{BlobID: manifest.BlobID})
	if err != nil {
		t.Fatalf("manifest lookup failed: %v", err)
	}
	if !manifestResp.Completed {
		t.Fatalf("manifest should be complete after all chunks")
	}

	presence, err := provider.ChunkPresence(ChunkPresenceRequest{BlobID: manifest.BlobID, ChunkIndex: 2, Authorized: true})
	if err != nil {
		t.Fatalf("presence check failed: %v", err)
	}
	if !presence.Exists {
		t.Fatalf("expected chunk 2 to exist")
	}
	if presence.ResumeToken == "" {
		t.Fatalf("expected resume token for authorized presence check")
	}

	privateManifest := Manifest{
		BlobID:     "blob-private",
		Size:       2,
		ChunkSize:  2,
		MimeType:   "image/png",
		Profile:    "fixed",
		Visibility: VisibilityPrivate,
	}
	if _, err := provider.PutManifest(PutManifestRequest{Manifest: privateManifest}); err != nil {
		t.Fatalf("private manifest upload failed: %v", err)
	}
	// private presence without authorization must not confirm existence
	privatePresence, err := provider.ChunkPresence(ChunkPresenceRequest{BlobID: privateManifest.BlobID, ChunkIndex: 0, Authorized: false})
	if err != nil {
		t.Fatalf("private presence check failed: %v", err)
	}
	if privatePresence.Exists {
		t.Fatalf("unauthorized check should not reveal chunk presence")
	}
	if privatePresence.ResumeToken != "" {
		t.Fatalf("unauthorized presence must not expose resume token")
	}
}

func TestProviderRefusalMatrix(t *testing.T) {
	provider := NewProvider([]string{"image/png"}, []string{"fixed"})
	manifest := Manifest{
		BlobID:     "blob-refuse",
		Size:       4,
		ChunkSize:  2,
		MimeType:   "image/png",
		Profile:    "fixed",
		Visibility: VisibilityPublic,
	}
	if _, err := provider.PutManifest(PutManifestRequest{Manifest: manifest}); err != nil {
		t.Fatalf("setup manifest failed: %v", err)
	}
	if _, err := provider.PutBlobChunk(PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 1, Data: []byte{0x01, 0x02}}); err == nil {
		t.Fatalf("expected invalid chunk order refusal")
	} else {
		requireRefusal(t, err, RefusalInvalidChunkOrder)
	}
	if _, err := provider.PutBlobChunk(PutBlobChunkRequest{BlobID: manifest.BlobID, ChunkIndex: 0, Data: []byte{0x01}}); err == nil {
		t.Fatalf("expected chunk size mismatch refusal")
	} else {
		requireRefusal(t, err, RefusalChunkSizeMismatch)
	}

	_, err := provider.PutManifest(PutManifestRequest{Manifest: Manifest{BlobID: "blob-mime", ChunkSize: 1, Size: 1, MimeType: "text/plain", Profile: "fixed"}})
	requireRefusal(t, err, RefusalUnsupportedMime)

	_, err = provider.PutManifest(PutManifestRequest{Manifest: Manifest{BlobID: "blob-profile", ChunkSize: 1, Size: 1, MimeType: "image/png", Profile: "unknown"}})
	requireRefusal(t, err, RefusalUnsupportedProfile)
}

func requireRefusal(t *testing.T, err error, want RefusalReason) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected refusal %s", want)
	}
	re, ok := err.(*RefusalError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if re.Reason != want {
		t.Fatalf("expected refusal %s, got %s", want, re.Reason)
	}
}
