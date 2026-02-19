package ingest

import "github.com/aether/code_aether/pkg/v21/store"

// ApplyResult captures how an event insertion went.
type ApplyResult struct {
	Success bool
	Stored  bool
	Reason  store.StoreFailureReason
}

// Ingestor models the deterministic ingestion pipeline.
type Ingestor struct {
	locked            bool
	corrupted         bool
	migrationRequired bool
	capacity          int
	events            map[string]store.TimelineEvent
	order             []string
}

// NewIngestor creates an ingestion pipeline with the provided capacity.
func NewIngestor(capacity int) *Ingestor {
	if capacity <= 0 {
		capacity = 1024
	}
	return &Ingestor{
		capacity: capacity,
		events:   make(map[string]store.TimelineEvent, capacity),
		order:    make([]string, 0, capacity),
	}
}

// Lock marks the ingestion pipeline as unavailable for writes.
func (i *Ingestor) Lock() {
	i.locked = true
}

// Unlock clears the write lock.
func (i *Ingestor) Unlock() {
	i.locked = false
}

// RequireMigration indicates the store schema must be migrated before writes.
func (i *Ingestor) RequireMigration(required bool) {
	i.migrationRequired = required
}

// MarkCorrupt flags the store as corrupt until cleared.
func (i *Ingestor) MarkCorrupt(corrupt bool) {
	i.corrupted = corrupt
}

// Apply attempts to persist the provided event with canonical dedupe.
func (i *Ingestor) Apply(event store.TimelineEvent) ApplyResult {
	if i.locked {
		return ApplyResult{Success: false, Reason: store.StoreLocked}
	}
	if i.corrupted {
		return ApplyResult{Success: false, Reason: store.StoreCorrupt}
	}
	if i.migrationRequired {
		return ApplyResult{Success: false, Reason: store.StoreMigrationRequired}
	}
	if i.capacity > 0 && len(i.order) >= i.capacity {
		return ApplyResult{Success: false, Reason: store.StoreQuotaExceeded}
	}
	key := event.Canonical.Key()
	if _, exists := i.events[key]; exists {
		return ApplyResult{Success: true, Stored: false}
	}
	i.events[key] = event
	i.order = append(i.order, key)
	return ApplyResult{Success: true, Stored: true}
}

// StoredEvents returns a deterministic slice of stored timeline events.
func (i *Ingestor) StoredEvents() []store.TimelineEvent {
	result := make([]store.TimelineEvent, 0, len(i.order))
	for _, key := range i.order {
		result = append(result, i.events[key])
	}
	return result
}
