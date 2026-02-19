package blobproto

import (
	"fmt"
	"sync"
)

// SpaceVisibility describes whether a blob belongs to a public or private space.
type SpaceVisibility string

const (
	VisibilityPublic  SpaceVisibility = "public"
	VisibilityPrivate SpaceVisibility = "private"
)

// Manifest describes the metadata the provider uses to validate chunk uploads.
type Manifest struct {
	BlobID     string
	Size       int64
	ChunkSize  int64
	MimeType   string
	Profile    string
	Visibility SpaceVisibility
}

// RefusalReason enumerates deterministic refusal codes returned by provider endpoints.
type RefusalReason string

const (
	RefusalInvalidChunkOrder  RefusalReason = "invalid_chunk_order"
	RefusalChunkSizeMismatch  RefusalReason = "chunk_size_mismatch"
	RefusalUnsupportedMime    RefusalReason = "unsupported_mime"
	RefusalUnsupportedProfile RefusalReason = "unsupported_profile"
)

// RefusalError wraps a refusal reason with context for deterministic diagnostics.
type RefusalError struct {
	Reason RefusalReason
	Detail string
}

func (e *RefusalError) Error() string {
	return fmt.Sprintf("provider refusal(%s): %s", e.Reason, e.Detail)
}

type manifestState struct {
	manifest   Manifest
	chunks     map[int][]byte
	nextChunk  int
	chunkCount int
	completed  bool
	mu         sync.Mutex
}

// Provider hosts an in-memory blob transfer service with deterministic ordering and refusals.
type Provider struct {
	mu                sync.Mutex
	manifests         map[string]*manifestState
	supportedMimes    map[string]struct{}
	supportedProfiles map[string]struct{}
}

// NewProvider builds a provider instance that only accepts the supplied mime/profile sets.
func NewProvider(mimes, profiles []string) *Provider {
	mimeSet := make(map[string]struct{})
	for _, m := range mimes {
		mimeSet[m] = struct{}{}
	}
	profileSet := make(map[string]struct{})
	for _, p := range profiles {
		profileSet[p] = struct{}{}
	}
	return &Provider{
		manifests:         make(map[string]*manifestState),
		supportedMimes:    mimeSet,
		supportedProfiles: profileSet,
	}
}

// PutManifestRequest sets the manifest metadata before chunk upload begins.
type PutManifestRequest struct {
	Manifest Manifest
}

// PutManifestResponse seeds the resume token and next expected chunk index.
type PutManifestResponse struct {
	ResumeToken    string
	NextChunkIndex int
}

// PutBlobChunkRequest uploads a chunk for a manifest.
type PutBlobChunkRequest struct {
	BlobID     string
	ChunkIndex int
	Data       []byte
}

// PutBlobChunkResponse reports the next expected chunk index and resume token.
type PutBlobChunkResponse struct {
	NextChunkIndex int
	ResumeToken    string
}

// GetManifestRequest and response expose manifest state.
type GetManifestRequest struct {
	BlobID string
}

// GetManifestResponse returns manifest state plus resume info.
type GetManifestResponse struct {
	Manifest    Manifest
	NextChunk   int
	ResumeToken string
	Completed   bool
}

// ChunkPresenceRequest queries for chunk presence without completing a transfer.
type ChunkPresenceRequest struct {
	BlobID     string
	ChunkIndex int
	Authorized bool
}

// ChunkPresenceResponse reports whether a chunk is already committed.
type ChunkPresenceResponse struct {
	Exists      bool
	ResumeToken string
}

// QuotaRequest/QuotaResponse are optional helpers for quota tooling.
type QuotaRequest struct {
	BlobID string
}

type QuotaResponse struct {
	UsedBytes  int64
	LimitBytes int64
}

func (p *Provider) PutManifest(req PutManifestRequest) (PutManifestResponse, error) {
	if req.Manifest.ChunkSize <= 0 {
		return PutManifestResponse{}, fmt.Errorf("chunk size must be positive")
	}
	if _, ok := p.supportedMimes[req.Manifest.MimeType]; !ok {
		return PutManifestResponse{}, &RefusalError{Reason: RefusalUnsupportedMime, Detail: req.Manifest.MimeType}
	}
	if _, ok := p.supportedProfiles[req.Manifest.Profile]; !ok {
		return PutManifestResponse{}, &RefusalError{Reason: RefusalUnsupportedProfile, Detail: req.Manifest.Profile}
	}
	totalChunks := calculateChunks(req.Manifest.Size, req.Manifest.ChunkSize)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.manifests[req.Manifest.BlobID] = &manifestState{
		manifest:   req.Manifest,
		chunks:     make(map[int][]byte, totalChunks),
		nextChunk:  0,
		chunkCount: totalChunks,
	}
	return PutManifestResponse{
		ResumeToken:    resumeToken(req.Manifest.BlobID, 0),
		NextChunkIndex: 0,
	}, nil
}

func (p *Provider) PutBlobChunk(req PutBlobChunkRequest) (PutBlobChunkResponse, error) {
	entry, err := p.lookupManifest(req.BlobID)
	if err != nil {
		return PutBlobChunkResponse{}, err
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if req.ChunkIndex != entry.nextChunk {
		return PutBlobChunkResponse{}, &RefusalError{Reason: RefusalInvalidChunkOrder, Detail: fmt.Sprintf("next chunk %d", entry.nextChunk)}
	}
	expectedSize := entry.manifest.ChunkSize
	if entry.nextChunk == entry.chunkCount-1 {
		remaining := int(entry.manifest.Size - int64(entry.nextChunk)*entry.manifest.ChunkSize)
		if len(req.Data) != remaining {
			return PutBlobChunkResponse{}, &RefusalError{Reason: RefusalChunkSizeMismatch, Detail: fmt.Sprintf("final chunk expected %d", remaining)}
		}
	} else {
		if len(req.Data) != int(expectedSize) {
			return PutBlobChunkResponse{}, &RefusalError{Reason: RefusalChunkSizeMismatch, Detail: fmt.Sprintf("chunk %d expected %d", req.ChunkIndex, expectedSize)}
		}
	}
	entry.chunks[req.ChunkIndex] = append([]byte(nil), req.Data...)
	entry.nextChunk++
	if entry.nextChunk == entry.chunkCount {
		entry.completed = true
	}
	return PutBlobChunkResponse{
		NextChunkIndex: entry.nextChunk,
		ResumeToken:    resumeToken(req.BlobID, entry.nextChunk),
	}, nil
}

func (p *Provider) GetManifest(req GetManifestRequest) (GetManifestResponse, error) {
	entry, err := p.lookupManifest(req.BlobID)
	if err != nil {
		return GetManifestResponse{}, err
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	return GetManifestResponse{
		Manifest:    entry.manifest,
		NextChunk:   entry.nextChunk,
		ResumeToken: resumeToken(req.BlobID, entry.nextChunk),
		Completed:   entry.completed,
	}, nil
}

func (p *Provider) ChunkPresence(req ChunkPresenceRequest) (ChunkPresenceResponse, error) {
	entry, err := p.lookupManifest(req.BlobID)
	if err != nil {
		return ChunkPresenceResponse{}, err
	}
	if entry.manifest.Visibility == VisibilityPrivate && !req.Authorized {
		return ChunkPresenceResponse{Exists: false}, nil
	}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	exists := req.ChunkIndex < entry.nextChunk
	resume := resumeToken(req.BlobID, entry.nextChunk)
	return ChunkPresenceResponse{
		Exists:      exists,
		ResumeToken: resume,
	}, nil
}

func (p *Provider) QueryQuota(_ QuotaRequest) (QuotaResponse, error) {
	return QuotaResponse{}, nil
}

func (p *Provider) lookupManifest(blobID string) (*manifestState, error) {
	p.mu.Lock()
	entry, ok := p.manifests[blobID]
	p.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("manifest %s not found", blobID)
	}
	return entry, nil
}

func resumeToken(blobID string, next int) string {
	return fmt.Sprintf("%s:%d", blobID, next)
}

func calculateChunks(size, chunkSize int64) int {
	if chunkSize <= 0 {
		return 0
	}
	count := (size + chunkSize - 1) / chunkSize
	if count <= 0 {
		return 1
	}
	return int(count)
}
