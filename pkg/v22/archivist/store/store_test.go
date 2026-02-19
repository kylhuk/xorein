package store

import (
	"testing"
	"time"
)

func TestStorePutQuotaExceeded(t *testing.T) {
	now := time.Now()
	s := NewStore(Config{
		QuotaPerSpace: map[SpaceID]int64{"space": 100},
		Now:           func() time.Time { return now },
	})

	err := s.Put("space", "channel", "seg1", 120)
	if err == nil {
		t.Fatalf("expected quota exceeded error")
	}
	se, ok := err.(StoreError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if se.Reason != ReasonQuotaExceeded {
		t.Fatalf("unexpected reason %s", se.Reason)
	}
}

func TestStorePutSegmentTooLarge(t *testing.T) {
	s := NewStore(Config{
		MaxSegmentSize: 10,
		Now:            func() time.Time { return time.Now() },
	})
	err := s.Put("space", "channel", "seg", 11)
	if err == nil {
		t.Fatalf("expected segment too large")
	}
	se, ok := err.(StoreError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if se.Reason != ReasonSegmentTooLarge {
		t.Fatalf("unexpected reason %s", se.Reason)
	}
}

func TestStorePruneRetention(t *testing.T) {
	base := time.Now()
	cur := base
	s := NewStore(Config{
		Retention: time.Minute,
		Now:       func() time.Time { return cur },
	})

	_ = s.Put("space", "channel", "seg-new", 10)
	cur = base.Add(2 * time.Minute)
	_ = s.Put("space", "channel", "seg-current", 5)
	pruned := s.Prune()
	if len(pruned) != 1 {
		t.Fatalf("expected one pruned segment, got %d", len(pruned))
	}
	if pruned[0].Reason != ReasonRetentionPolicy {
		t.Fatalf("unexpected prune reason %s", pruned[0].Reason)
	}
	if s.SpaceUsage("space") != 5 {
		t.Fatalf("expected space usage 5, got %d", s.SpaceUsage("space"))
	}
}
