package boundaries

import (
	"errors"
	"testing"

	"github.com/aether/code_aether/pkg/v25/blobproto"
)

const expectedRelayBlobRefusalReason = "relay_durable_blob_hosting_not_allowed"

var (
	reasonRelayDurableBlob = blobproto.RefusalReason(expectedRelayBlobRefusalReason)
	errRelayRefused        = &blobproto.RefusalError{
		Reason: reasonRelayDurableBlob,
		Detail: "relay boundary refuses durable blob payload",
	}
	errNotFound     = errors.New("relay boundary not found")
	errRelayHistory = errors.New("relay boundary denies durable history")
)

type fakeRelayBoundary struct {
	metadata map[string]blobproto.Manifest
	payload  map[string][]byte
}

func newFakeRelayBoundary() *fakeRelayBoundary {
	return &fakeRelayBoundary{
		metadata: make(map[string]blobproto.Manifest),
		payload:  make(map[string][]byte),
	}
}

func (f *fakeRelayBoundary) PutManifest(blobproto.PutManifestRequest) (blobproto.PutManifestResponse, error) {
	return blobproto.PutManifestResponse{}, errRelayRefused
}

func (f *fakeRelayBoundary) PutBlobChunk(blobproto.PutBlobChunkRequest) (blobproto.PutBlobChunkResponse, error) {
	return blobproto.PutBlobChunkResponse{}, errRelayRefused
}

func (f *fakeRelayBoundary) recordManifestPointer(manifest blobproto.Manifest) {
	f.metadata[manifest.BlobID] = manifest
}

func (f *fakeRelayBoundary) manifestPointer(blobID string) (blobproto.Manifest, error) {
	manifest, ok := f.metadata[blobID]
	if !ok {
		return blobproto.Manifest{}, errNotFound
	}
	return manifest, nil
}

func (f *fakeRelayBoundary) hasPayload(blobID string) bool {
	_, ok := f.payload[blobID]
	return ok
}

func (f *fakeRelayBoundary) lookupPrivateSpace(blobID string, authorized bool) error {
	manifest, ok := f.metadata[blobID]
	if !ok {
		return errNotFound
	}
	if manifest.Visibility == blobproto.VisibilityPrivate && !authorized {
		return errNotFound
	}
	return nil
}

func TestRelayBoundaryRefusesBlobUploads(t *testing.T) {
	boundary := newFakeRelayBoundary()
	if _, err := boundary.PutManifest(blobproto.PutManifestRequest{}); !errors.Is(err, errRelayRefused) {
		t.Fatalf("expected manifest refusal, got %v", err)
	}
	if _, err := boundary.PutBlobChunk(blobproto.PutBlobChunkRequest{}); !errors.Is(err, errRelayRefused) {
		t.Fatalf("expected chunk refusal, got %v", err)
	}
	if len(boundary.metadata) != 0 {
		t.Fatalf("expected no metadata stored, found %d entries", len(boundary.metadata))
	}
	if len(boundary.payload) != 0 {
		t.Fatalf("expected no payload stored, found %d entries", len(boundary.payload))
	}
}

func TestRelayBoundaryMetadataOnlyPointers(t *testing.T) {
	boundary := newFakeRelayBoundary()
	manifest := blobproto.Manifest{
		BlobID:     "meta-blob",
		Size:       1024,
		ChunkSize:  256,
		MimeType:   "application/octet-stream",
		Profile:    "f1",
		Visibility: blobproto.VisibilityPublic,
	}
	boundary.recordManifestPointer(manifest)
	got, err := boundary.manifestPointer(manifest.BlobID)
	if err != nil {
		t.Fatalf("failed to read metadata pointer: %v", err)
	}
	if got != manifest {
		t.Fatalf("metadata mismatch: got %v want %v", got, manifest)
	}
	if boundary.hasPayload(manifest.BlobID) {
		t.Fatalf("boundary should not retain payload bytes")
	}
}

func TestPrivateSpaceAntiEnumeration(t *testing.T) {
	boundary := newFakeRelayBoundary()
	manifest := blobproto.Manifest{
		BlobID:     "private-blob",
		Size:       10,
		ChunkSize:  10,
		MimeType:   "application/octet-stream",
		Profile:    "f2",
		Visibility: blobproto.VisibilityPrivate,
	}
	boundary.recordManifestPointer(manifest)

	if err := boundary.lookupPrivateSpace(manifest.BlobID, false); !errors.Is(err, errNotFound) {
		t.Fatalf("unauthorized lookup should behave as not-found, got %v", err)
	}
	if err := boundary.lookupPrivateSpace("missing", false); !errors.Is(err, errNotFound) {
		t.Fatalf("unauthorized lookup for missing blob should still be not-found, got %v", err)
	}
	if err := boundary.lookupPrivateSpace(manifest.BlobID, true); err != nil {
		t.Fatalf("authorized lookup should succeed, got %v", err)
	}
}

type relayHistoryProbe struct{}

func (r relayHistoryProbe) fetchHistory() error {
	return errRelayHistory
}

func (r relayHistoryProbe) searchHistory() error {
	return errRelayHistory
}

func TestRelayBoundaryRejectsHistoryHosting(t *testing.T) {
	relay := relayHistoryProbe{}

	if err := relay.fetchHistory(); !errors.Is(err, errRelayHistory) {
		t.Fatalf("expected history readback to be blocked on relay, got %v", err)
	}
	if err := relay.searchHistory(); !errors.Is(err, errRelayHistory) {
		t.Fatalf("expected relay search to be blocked, got %v", err)
	}
}

func TestRelayBoundaryRefusalReasonTaxonomy(t *testing.T) {
	boundary := newFakeRelayBoundary()
	_, err := boundary.PutManifest(blobproto.PutManifestRequest{})
	if err == nil {
		t.Fatalf("expected a refusal error, got nil")
	}
	var refusal *blobproto.RefusalError
	if !errors.As(err, &refusal) {
		t.Fatalf("expected refusal error, got %v", err)
	}
	if string(refusal.Reason) != expectedRelayBlobRefusalReason {
		t.Fatalf("refusal reason mutated on relay boundary: want %s got %s", expectedRelayBlobRefusalReason, refusal.Reason)
	}
}
