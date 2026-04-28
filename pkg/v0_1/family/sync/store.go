package sync

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Message is a server-indexed ciphertext delivery.
type Message struct {
	ID        string
	ServerID  string
	Sequence  int64
	CreatedAt time.Time
	Body      []byte
	Signature []byte
}

// Range represents a gap in sequence coverage.
type Range struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// CoverageResult is the output of a coverage query.
type CoverageResult struct {
	AvailableFrom int64
	AvailableTo   int64
	MessageHashes []string
	SnapshotRoot  string
	GapRanges     []Range
}

// Store is the in-memory ciphertext archive for all servers.
type Store struct {
	mu      sync.RWMutex
	servers map[string]*serverStore // serverID → serverStore
}

func NewStore() *Store {
	return &Store{servers: make(map[string]*serverStore)}
}

type serverStore struct {
	messages map[int64]*Message       // seq → Message
	byID     map[string]*Message      // messageID → Message (for fetch by ID)
	members  map[string]struct{}      // memberPeerIDs
	maxSeq   int64
	minSeq   int64
	hasData  bool
}

func newServerStore() *serverStore {
	return &serverStore{
		messages: make(map[int64]*Message),
		byID:     make(map[string]*Message),
		members:  make(map[string]struct{}),
	}
}

func (ss *serverStore) push(msg *Message) (added bool) {
	if _, exists := ss.byID[msg.ID]; exists {
		return false // dedup
	}
	seq := msg.Sequence
	ss.messages[seq] = msg
	ss.byID[msg.ID] = msg
	if !ss.hasData {
		ss.minSeq = seq
		ss.maxSeq = seq
		ss.hasData = true
	} else {
		if seq < ss.minSeq {
			ss.minSeq = seq
		}
		if seq > ss.maxSeq {
			ss.maxSeq = seq
		}
	}
	return true
}

func (ss *serverStore) coverage(fromSeq, toSeq int64) CoverageResult {
	if !ss.hasData {
		return CoverageResult{}
	}
	lo, hi := ss.minSeq, ss.maxSeq
	if fromSeq > 0 && fromSeq > lo {
		lo = fromSeq
	}
	if toSeq > 0 && toSeq < hi {
		hi = toSeq
	}
	// Collect ordered sequences in [lo, hi].
	var seqs []int64
	for seq := range ss.messages {
		if seq >= lo && seq <= hi {
			seqs = append(seqs, seq)
		}
	}
	sort.Slice(seqs, func(i, j int) bool { return seqs[i] < seqs[j] })

	hashes := make([]string, 0, len(seqs))
	for _, seq := range seqs {
		m := ss.messages[seq]
		hashes = append(hashes, messageHash(m))
	}

	var gaps []Range
	prev := lo
	for _, seq := range seqs {
		if seq > prev {
			gaps = append(gaps, Range{From: prev, To: seq - 1})
		}
		prev = seq + 1
	}
	if prev <= hi {
		gaps = append(gaps, Range{From: prev, To: hi})
	}

	root := snapshotRoot(hashes)
	return CoverageResult{
		AvailableFrom: lo,
		AvailableTo:   hi,
		MessageHashes: hashes,
		SnapshotRoot:  root,
		GapRanges:     gaps,
	}
}

// messageHash returns base64url-no-pad(SHA-256(Delivery.ID || Delivery.created_at.RFC3339Nano)).
func messageHash(m *Message) string {
	h := sha256.New()
	h.Write([]byte(m.ID))
	h.Write([]byte(m.CreatedAt.UTC().Format(time.RFC3339Nano)))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// snapshotRoot returns base64url-no-pad(SHA-256(h0 || h1 || ... || hn)).
func snapshotRoot(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	if len(hashes) == 1 {
		return hashes[0]
	}
	h := sha256.New()
	for _, hash := range hashes {
		h.Write([]byte(hash))
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// --- Store API ---

func (s *Store) getOrCreate(serverID string) *serverStore {
	ss := s.servers[serverID]
	if ss == nil {
		ss = newServerStore()
		s.servers[serverID] = ss
	}
	return ss
}

// AddMember registers a peer as a server member (allows sync.fetch).
func (s *Store) AddMember(serverID, peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getOrCreate(serverID).members[peerID] = struct{}{}
}

// IsMember checks membership.
func (s *Store) IsMember(serverID, peerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ss := s.servers[serverID]
	if ss == nil {
		return false
	}
	_, ok := ss.members[peerID]
	return ok
}

// HasServer checks if the server is known.
func (s *Store) HasServer(serverID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.servers[serverID] != nil
}

// Push stores a message; returns (added, err). Duplicate IDs silently no-op.
func (s *Store) Push(msg *Message) (bool, error) {
	if msg.ID == "" || msg.ServerID == "" || msg.Sequence <= 0 {
		return false, fmt.Errorf("message missing required fields")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.getOrCreate(msg.ServerID)
	added := ss.push(msg)
	return added, nil
}

// Coverage returns coverage metadata for a server and sequence range.
func (s *Store) Coverage(serverID string, fromSeq, toSeq int64) CoverageResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ss := s.servers[serverID]
	if ss == nil {
		return CoverageResult{}
	}
	return ss.coverage(fromSeq, toSeq)
}

// FetchByIDs returns messages for known IDs, and the list of not-found IDs.
func (s *Store) FetchByIDs(serverID string, ids []string) ([]*Message, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ss := s.servers[serverID]
	var found []*Message
	var notFound []string
	for _, id := range ids {
		if ss != nil {
			if m := ss.byID[id]; m != nil {
				found = append(found, m)
				continue
			}
		}
		notFound = append(notFound, id)
	}
	return found, notFound
}
