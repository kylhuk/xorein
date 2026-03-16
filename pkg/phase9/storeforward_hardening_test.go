package phase9

import (
	"sync"
	"testing"
	"time"
)

func TestStoreServiceNormalizesZeroTime(t *testing.T) {
	svc, err := NewStoreService(StoreConfig{
		RetentionTTL:    10 * time.Minute,
		MaxMessages:     4,
		MaxBytes:        64,
		MaxPayloadBytes: 16,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}

	result := svc.Store(time.Time{}, "alice", []byte("abcd"))
	if !result.Stored {
		t.Fatalf("Store(zero time) = %+v, want Stored=true", result)
	}
	if len(svc.items) != 1 {
		t.Fatalf("expected one stored envelope, got %d", len(svc.items))
	}
	if svc.items[0].storedAt.IsZero() {
		t.Fatalf("storedAt should be normalized to current UTC time")
	}
	if svc.items[0].expiresAt.IsZero() {
		t.Fatalf("expiresAt should be normalized to current UTC time")
	}
	if !svc.items[0].expiresAt.After(svc.items[0].storedAt) {
		t.Fatalf("expiresAt should be after storedAt")
	}
	if svc.items[0].storedAt.Location() != time.UTC || svc.items[0].expiresAt.Location() != time.UTC {
		t.Fatalf("timestamps should be stored in UTC")
	}
}

func TestStoreServiceConcurrentStoresRespectQuota(t *testing.T) {
	svc, err := NewStoreService(StoreConfig{
		RetentionTTL:    10 * time.Minute,
		MaxMessages:     8,
		MaxBytes:        32,
		MaxPayloadBytes: 4,
	})
	if err != nil {
		t.Fatalf("NewStoreService() unexpected error: %v", err)
	}

	const writers = 64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			svc.Store(time.Date(2026, time.January, 1, 12, 0, 0, i, time.UTC), "alice", []byte("data"))
		}(i)
	}
	close(start)
	wg.Wait()

	snap := svc.Snapshot(time.Date(2026, time.January, 1, 12, 1, 0, 0, time.UTC))
	if snap.QueuedMessages != 8 {
		t.Fatalf("Snapshot().QueuedMessages = %d, want 8", snap.QueuedMessages)
	}
	if snap.QueuedBytes != 32 {
		t.Fatalf("Snapshot().QueuedBytes = %d, want 32", snap.QueuedBytes)
	}
	if snap.DroppedByCap != writers-8 {
		t.Fatalf("Snapshot().DroppedByCap = %d, want %d", snap.DroppedByCap, writers-8)
	}
	if snap.RejectedWrites != 0 {
		t.Fatalf("Snapshot().RejectedWrites = %d, want 0", snap.RejectedWrites)
	}
}
