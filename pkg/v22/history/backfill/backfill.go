package backfill

import (
	"errors"
	"fmt"
	"sync"
)

type FailureReason string

const (
	ReasonDenied       FailureReason = "BACKFILL_DENIED"
	ReasonIncomplete   FailureReason = "BACKFILL_INCOMPLETE"
	ReasonVerifyFailed FailureReason = "BACKFILL_VERIFY_FAILED"
	ReasonInvalidRange FailureReason = "BACKFILL_INVALID_RANGE"
	ReasonApplyFailed  FailureReason = "BACKFILL_APPLY_FAILED"
)

var (
	ErrBackfillDenied       = errors.New(string(ReasonDenied))
	ErrBackfillIncomplete   = errors.New(string(ReasonIncomplete))
	ErrBackfillVerifyFailed = errors.New(string(ReasonVerifyFailed))
	ErrBackfillInvalidRange = errors.New(string(ReasonInvalidRange))
	ErrBackfillApplyFailed  = errors.New(string(ReasonApplyFailed))
)

type TimeRange struct {
	Start int64
	End   int64
}

type BackfillRequest struct {
	SpaceID     string
	ChannelID   string
	Range       TimeRange
	MaxSegments int
}

type Segment struct {
	ID   string
	Data []byte
}

type progressEntry struct {
	Applied   int
	Total     int
	Completed bool
	Reason    FailureReason
}

type ProgressReport struct {
	Applied   int
	Total     int
	Completed bool
	Reason    FailureReason
}

type Manager struct {
	maxAttempts int
	mu          sync.Mutex
	attempts    map[string]int
	progress    map[string]progressEntry
}

func NewManager(maxAttempts int) *Manager {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return &Manager{
		maxAttempts: maxAttempts,
		attempts:    make(map[string]int),
		progress:    make(map[string]progressEntry),
	}
}

func (m *Manager) requestKey(req BackfillRequest) string {
	return fmt.Sprintf("%s:%s:%d-%d", req.SpaceID, req.ChannelID, req.Range.Start, req.Range.End)
}

func (m *Manager) Progress(req BackfillRequest) ProgressReport {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.progress[m.requestKey(req)]
	return ProgressReport{
		Applied:   entry.Applied,
		Total:     entry.Total,
		Completed: entry.Completed,
		Reason:    entry.Reason,
	}
}

func (m *Manager) Backfill(
	req BackfillRequest,
	fetch func() ([]Segment, error),
	apply func(Segment) error,
) error {
	key := m.requestKey(req)
	if !validRange(req.Range) {
		m.setProgress(key, progressEntry{Reason: ReasonInvalidRange})
		return ErrBackfillInvalidRange
	}

	m.mu.Lock()
	entry := m.progress[key]
	if m.attempts[key] >= m.maxAttempts {
		m.mu.Unlock()
		return ErrBackfillDenied
	}
	if entry.Completed {
		m.mu.Unlock()
		return nil
	}
	m.attempts[key]++
	m.mu.Unlock()

	segments, err := fetch()
	if err != nil {
		m.setProgress(key, progressEntry{Reason: ReasonIncomplete})
		return ErrBackfillIncomplete
	}
	if len(segments) == 0 {
		m.setProgress(key, progressEntry{Reason: ReasonVerifyFailed})
		return ErrBackfillVerifyFailed
	}

	m.setProgress(key, progressEntry{Total: len(segments)})
	for i, segment := range segments {
		if err := apply(segment); err != nil {
			m.setProgress(key, progressEntry{Applied: i, Total: len(segments), Reason: ReasonApplyFailed})
			return ErrBackfillApplyFailed
		}
		m.setProgress(key, progressEntry{Applied: i + 1, Total: len(segments)})
	}
	m.setProgress(key, progressEntry{Applied: len(segments), Total: len(segments), Completed: true})
	return nil
}

func validRange(r TimeRange) bool {
	return r.End > r.Start
}

func (m *Manager) setProgress(key string, entry progressEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progress[key] = entry
}
