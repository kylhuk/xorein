package backfill

import (
	"errors"
	"testing"
)

func TestBackfillManagerDeniesAfterAttempts(t *testing.T) {
	manager := NewManager(1)
	req := BackfillRequest{SpaceID: "space", ChannelID: "chan", Range: TimeRange{Start: 0, End: 10}}

	fetch := func() ([]Segment, error) {
		return []Segment{{ID: "s-1"}}, nil
	}
	apply := func(Segment) error { return nil }

	if err := manager.Backfill(req, fetch, apply); err != nil {
		t.Fatalf("expected first backfill to pass, got %v", err)
	}

	if err := manager.Backfill(req, fetch, apply); !errors.Is(err, ErrBackfillDenied) {
		t.Fatalf("expected ErrBackfillDenied, got %v", err)
	}
}

func TestBackfillManagerInvalidRange(t *testing.T) {
	manager := NewManager(1)
	req := BackfillRequest{SpaceID: "space", ChannelID: "chan", Range: TimeRange{Start: 10, End: 5}}

	err := manager.Backfill(req, func() ([]Segment, error) {
		return []Segment{{ID: "s-1"}}, nil
	}, func(Segment) error { return nil })
	if !errors.Is(err, ErrBackfillInvalidRange) {
		t.Fatalf("expected ErrBackfillInvalidRange, got %v", err)
	}

	if report := manager.Progress(req); report.Reason != ReasonInvalidRange {
		t.Fatalf("expected invalid range reason, got %s", report.Reason)
	}
}

func TestBackfillReportAppliesProgress(t *testing.T) {
	manager := NewManager(2)
	req := BackfillRequest{SpaceID: "space", ChannelID: "chan", Range: TimeRange{Start: 0, End: 5}}
	segments := []Segment{{ID: "s-1"}, {ID: "s-2"}}

	applyErr := errors.New("apply failure")
	err := manager.Backfill(req, func() ([]Segment, error) {
		return segments, nil
	}, func(segment Segment) error {
		if segment.ID == "s-2" {
			return applyErr
		}
		return nil
	})
	if !errors.Is(err, ErrBackfillApplyFailed) {
		t.Fatalf("expected ErrBackfillApplyFailed, got %v", err)
	}

	report := manager.Progress(req)
	if report.Applied != 1 || report.Total != len(segments) || report.Reason != ReasonApplyFailed {
		t.Fatalf("unexpected progress state: %+v", report)
	}
}

func TestBackfillProgressReportsCompletion(t *testing.T) {
	manager := NewManager(2)
	req := BackfillRequest{SpaceID: "space", ChannelID: "chan", Range: TimeRange{Start: 0, End: 5}}
	segments := []Segment{{ID: "s-1"}, {ID: "s-2"}}

	err := manager.Backfill(req, func() ([]Segment, error) {
		return segments, nil
	}, func(Segment) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected backfill error: %v", err)
	}

	report := manager.Progress(req)
	if !report.Completed || report.Applied != len(segments) || report.Total != len(segments) {
		t.Fatalf("expected completed progress, got %+v", report)
	}
}
