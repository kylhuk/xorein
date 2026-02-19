package tombstone

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrCorruptedEntry = errors.New("CORRUPTED_ENTRY")

type RemovalHook func(id string)

type Option func(*Store)

type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
	hook    RemovalHook
	now     func() time.Time
}

type Entry struct {
	ID            string
	Metadata      map[string]string
	Reason        string
	AuditPointer  string
	TombstonedAt  time.Time
	ContentExists bool
}

type EntryState struct {
	ID            string
	Metadata      map[string]string
	Reason        string
	AuditPointer  string
	TombstonedAt  time.Time
	ContentExists bool
}

func NewStore(hook RemovalHook, opts ...Option) *Store {
	s := &Store{
		entries: make(map[string]*Entry),
		hook:    hook,
		now:     time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithTimeSource(now func() time.Time) Option {
	return func(s *Store) {
		if now != nil {
			s.now = now
		}
	}
}

func (s *Store) Apply(ctx context.Context, id string, metadata map[string]string, reason, audit string) (*Entry, error) {
	if id == "" {
		return nil, ErrCorruptedEntry
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if !ok {
		entry = &Entry{
			ID:            id,
			Metadata:      cloneMetadata(metadata),
			Reason:        reason,
			AuditPointer:  audit,
			ContentExists: true,
		}
		s.entries[id] = entry
	} else {
		if metadata != nil {
			entry.Metadata = cloneMetadata(metadata)
		}
		if reason != "" {
			entry.Reason = reason
		}
		if audit != "" {
			entry.AuditPointer = audit
		}
	}
	if entry.ContentExists {
		entry.TombstonedAt = s.now()
		entry.ContentExists = false
		if s.hook != nil {
			s.hook(id)
		}
	}
	return entry, nil
}

func (s *Store) Snapshot() []EntryState {
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]EntryState, 0, len(s.entries))
	for _, entry := range s.entries {
		states = append(states, entry.toState())
	}
	sort.Slice(states, func(i, j int) bool {
		return states[i].ID < states[j].ID
	})
	return states
}

func (s *Store) Restore(states []EntryState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]*Entry, len(states))
	for _, state := range states {
		e := &Entry{
			ID:            state.ID,
			Metadata:      cloneMetadata(state.Metadata),
			Reason:        state.Reason,
			AuditPointer:  state.AuditPointer,
			TombstonedAt:  state.TombstonedAt,
			ContentExists: state.ContentExists,
		}
		s.entries[state.ID] = e
	}
}

func (s *Store) Validate(states []EntryState) error {
	for _, state := range states {
		if state.ID == "" {
			return ErrCorruptedEntry
		}
	}
	return nil
}

func (s *Store) Repair(states []EntryState) []EntryState {
	clean := make([]EntryState, 0, len(states))
	for _, state := range states {
		if state.ID == "" {
			continue
		}
		clean = append(clean, state)
	}
	return clean
}

func (s *Store) PruneBefore(cutoff time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, entry := range s.entries {
		if entry.ContentExists {
			continue
		}
		if entry.TombstonedAt.IsZero() {
			continue
		}
		if entry.TombstonedAt.Before(cutoff) {
			delete(s.entries, id)
		}
	}
}

func cloneMetadata(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (e *Entry) toState() EntryState {
	return EntryState{
		ID:            e.ID,
		Metadata:      cloneMetadata(e.Metadata),
		Reason:        e.Reason,
		AuditPointer:  e.AuditPointer,
		TombstonedAt:  e.TombstonedAt,
		ContentExists: e.ContentExists,
	}
}
