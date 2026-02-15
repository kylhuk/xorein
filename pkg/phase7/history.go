package phase7

import (
	"errors"
	"sort"
	"sync"
)

var ErrWindowExceeded = errors.New("phase7: history window exceeded")

type HistoryEntry struct {
	Sequence uint64
	Payload  []byte
}

type HistoryWindow struct {
	entries []HistoryEntry
	limit   int
}

type HistoryManager struct {
	mu       sync.Mutex
	bySender map[ParticipantID]*HistoryWindow
	window   int
}

func NewHistoryManager(window int) *HistoryManager {
	return &HistoryManager{
		bySender: make(map[ParticipantID]*HistoryWindow),
		window:   window,
	}
}

func (h *HistoryManager) Apply(sender ParticipantID, incoming []HistoryEntry) ([]HistoryEntry, error) {
	if len(incoming) == 0 {
		return nil, nil
	}
	if len(incoming) > h.window {
		return nil, ErrWindowExceeded
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	window := h.bySender[sender]
	if window == nil {
		window = &HistoryWindow{limit: h.window}
		h.bySender[sender] = window
	}

	merged := mergeEntries(window.entries, incoming)
	window.entries = truncate(merged, h.window)
	return window.entries, nil
}

func (h *HistoryManager) Fetch(sender ParticipantID, fallback func() ([]HistoryEntry, error)) ([]HistoryEntry, error) {
	h.mu.Lock()
	window := h.bySender[sender]
	h.mu.Unlock()
	if window != nil && len(window.entries) > 0 {
		return copyEntries(window.entries), nil
	}
	if fallback != nil {
		entries, err := fallback()
		if err != nil {
			return nil, err
		}
		return entries, nil
	}
	return nil, nil
}

func mergeEntries(existing, incoming []HistoryEntry) []HistoryEntry {
	combined := append([]HistoryEntry{}, existing...)
	combined = append(combined, incoming...)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Sequence < combined[j].Sequence
	})
	unique := make([]HistoryEntry, 0, len(combined))
	seen := make(map[uint64]struct{})
	for _, entry := range combined {
		if _, ok := seen[entry.Sequence]; ok {
			continue
		}
		seen[entry.Sequence] = struct{}{}
		unique = append(unique, entry)
	}
	return unique
}

func truncate(entries []HistoryEntry, limit int) []HistoryEntry {
	if len(entries) <= limit {
		return entries
	}
	return entries[len(entries)-limit:]
}

func copyEntries(entries []HistoryEntry) []HistoryEntry {
	copy := make([]HistoryEntry, len(entries))
	for i, entry := range entries {
		copy[i] = HistoryEntry{Sequence: entry.Sequence, Payload: append([]byte(nil), entry.Payload...)}
	}
	return copy
}
