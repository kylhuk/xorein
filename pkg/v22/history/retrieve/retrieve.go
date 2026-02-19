package retrieve

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/aether/code_aether/pkg/v22/history/integrity"
)

var (
	ErrRetrievalFailure     = errors.New("HISTORY_RETRIEVAL_FAILURE")
	ErrRetrievalRateLimited = errors.New("RETRIEVAL_RATE_LIMITED")
)

type RetrievalRequest struct {
	SpaceID   string
	ChannelID string
	Key       string
}

type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	counts map[string]int
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{limit: limit, counts: make(map[string]int)}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.limit <= 0 {
		return true
	}

	count := r.counts[key]
	if count >= r.limit {
		return false
	}

	r.counts[key] = count + 1
	return true
}

func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.counts, key)
}

type RetrievalStore struct {
	mu             sync.RWMutex
	heads          map[string]integrity.HistoryHead
	manifests      map[string]integrity.HistorySegmentManifest
	segments       map[string]map[string][]byte
	membershipKeys map[string]string
	limiter        *RateLimiter
}

func NewRetrievalStore(limit int) *RetrievalStore {
	return &RetrievalStore{
		heads:          make(map[string]integrity.HistoryHead),
		manifests:      make(map[string]integrity.HistorySegmentManifest),
		segments:       make(map[string]map[string][]byte),
		membershipKeys: make(map[string]string),
		limiter:        NewRateLimiter(limit),
	}
}

func (s *RetrievalStore) RegisterMembership(spaceID, secret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.membershipKeys[spaceID] = secret
}

func (s *RetrievalStore) StoreHead(head integrity.HistoryHead) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.storageKey(head.SpaceID, head.ChannelID)
	s.heads[key] = head
}

func (s *RetrievalStore) StoreManifest(manifest integrity.HistorySegmentManifest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.storageKey(manifest.SpaceID, manifest.ChannelID)
	s.manifests[key] = manifest
}

func (s *RetrievalStore) StoreSegment(spaceID, channelID, segmentID string, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.storageKey(spaceID, channelID)
	bucket, ok := s.segments[key]
	if !ok {
		bucket = make(map[string][]byte)
		s.segments[key] = bucket
	}
	bucket[segmentID] = append([]byte(nil), payload...)
}

func (s *RetrievalStore) RetrieveHead(req RetrievalRequest) (*integrity.HistoryHead, error) {
	if err := s.guardRequest(req); err != nil {
		return nil, err
	}

	s.mu.RLock()
	head, ok := s.heads[s.storageKey(req.SpaceID, req.ChannelID)]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrRetrievalFailure
	}

	return &head, nil
}

func (s *RetrievalStore) RetrieveManifest(req RetrievalRequest) (*integrity.HistorySegmentManifest, error) {
	if err := s.guardRequest(req); err != nil {
		return nil, err
	}

	s.mu.RLock()
	manifest, ok := s.manifests[s.storageKey(req.SpaceID, req.ChannelID)]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrRetrievalFailure
	}

	return &manifest, nil
}

func (s *RetrievalStore) RetrieveSegment(req RetrievalRequest, segmentID string) ([]byte, error) {
	if err := s.guardRequest(req); err != nil {
		return nil, err
	}

	s.mu.RLock()
	bucket, ok := s.segments[s.storageKey(req.SpaceID, req.ChannelID)]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrRetrievalFailure
	}

	payload, ok := bucket[segmentID]
	if !ok {
		return nil, integrity.ErrSegmentNotFound
	}

	return append([]byte(nil), payload...), nil
}

func (s *RetrievalStore) guardRequest(req RetrievalRequest) error {
	if !s.verifyKey(req.SpaceID, req.Key) {
		return ErrRetrievalFailure
	}

	if !s.limiter.Allow(req.Key) {
		return ErrRetrievalRateLimited
	}

	return nil
}

func (s *RetrievalStore) verifyKey(spaceID, provided string) bool {
	s.mu.RLock()
	secret, ok := s.membershipKeys[spaceID]
	s.mu.RUnlock()
	if !ok {
		return false
	}

	expected := DeriveRetrievalKey(spaceID, secret)
	return expected == provided
}

func DeriveRetrievalKey(spaceID, membershipSecret string) string {
	sum := sha256.Sum256([]byte(spaceID + "|" + membershipSecret))
	return hex.EncodeToString(sum[:])
}

func (s *RetrievalStore) storageKey(spaceID, channelID string) string {
	return spaceID + ":" + channelID
}
